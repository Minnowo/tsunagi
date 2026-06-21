package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"tsunagi/src/client"
	"tsunagi/src/rpc"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/minnowo/log4zero"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

// ── styles ────────────────────────────────────────────────────────────────────

var (
	stylePanelBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240"))

	styleFormLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Width(12)

	styleHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 1)

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Padding(0, 1)

	styleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("76")).
			Padding(0, 1)

	styleMuted = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	styleMsgSelf  = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	styleMsgOther = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	styleMsgTime  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleMsgErr   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// ── Friend list item ──────────────────────────────────────────────────────────

type Friend struct {
	Name      string `json:"name"`
	PubKeyB64 string `json:"pub_key"`
	RelayAddr string `json:"relay_addr"`
}

func (f Friend) Title() string       { return f.Name }
func (f Friend) Description() string { return trimAddr(f.RelayAddr) }
func (f Friend) FilterValue() string { return f.Name + " " + f.RelayAddr }

func trimAddr(s string) string {
	for _, pfx := range []string{"grpc://", "https://", "http://"} {
		s = strings.TrimPrefix(s, pfx)
	}
	return s
}

// ── add-friend form ───────────────────────────────────────────────────────────

const (
	formName = iota
	formPubKey
	formRelay
	formFieldCount
)

var formLabels = [formFieldCount]string{"Name", "Public key", "Relay addr"}

type addForm struct {
	inputs [formFieldCount]textinput.Model
	focus  int
	err    string
	ok     string
}

func newAddForm() addForm {
	placeholders := [formFieldCount]string{
		"Alice",
		"base64-encoded public key",
		"grpc://localhost:7471",
	}
	f := addForm{}
	for i := range f.inputs {
		t := textinput.New()
		t.Placeholder = placeholders[i]
		t.CharLimit = 512
		t.Width = 52
		f.inputs[i] = t
	}
	_ = f.inputs[formName].Focus()
	return f
}

func (f *addForm) focusField(i int) {
	for j := range f.inputs {
		if j == i {
			_ = f.inputs[j].Focus()
		} else {
			f.inputs[j].Blur()
		}
	}
	f.focus = i
}

func (f addForm) submit() (Friend, string) {
	name   := strings.TrimSpace(f.inputs[formName].Value())
	pubkey := strings.TrimSpace(f.inputs[formPubKey].Value())
	relay  := strings.TrimSpace(f.inputs[formRelay].Value())
	switch {
	case name == "":
		return Friend{}, "name is required"
	case pubkey == "":
		return Friend{}, "public key is required"
	case relay == "":
		return Friend{}, "relay address is required"
	}
	if !strings.HasPrefix(relay, "grpc://") {
		relay = "grpc://" + relay
	}
	return Friend{Name: name, PubKeyB64: pubkey, RelayAddr: relay}, ""
}

// ── chat ──────────────────────────────────────────────────────────────────────

type chatLine struct {
	ts   time.Time
	from string
	text string
}

type chatState struct {
	friend      Friend
	lines       []chatLine
	vp          viewport.Model
	input       textinput.Model
	relayClient *client.ClientRelayClient
}

func newChatState(f Friend, rc *client.ClientRelayClient) chatState {
	inp := textinput.New()
	inp.Placeholder = "type a message…"
	inp.CharLimit = 2048
	_ = inp.Focus()

	vp := viewport.New(40, 10)

	return chatState{
		friend:      f,
		input:       inp,
		vp:          vp,
		relayClient: rc,
	}
}

func (cs *chatState) appendLine(from, text string) {
	cs.lines = append(cs.lines, chatLine{ts: time.Now(), from: from, text: text})
	cs.vp.SetContent(cs.renderLines())
	cs.vp.GotoBottom()
}

func (cs *chatState) renderLines() string {
	var b strings.Builder
	const tsW = 6 // "HH:MM "
	vpW := cs.vp.Width
	for _, l := range cs.lines {
		ts := styleMsgTime.Render(l.ts.Format("15:04"))

		var label string
		var textStyle lipgloss.Style
		switch l.from {
		case "me":
			label = "you"
			textStyle = styleMsgSelf
		case "sys":
			label = "sys"
			textStyle = styleMsgTime
		case "err":
			label = "err"
			textStyle = styleMsgErr
		default:
			label = l.from
			textStyle = styleMsgOther
		}

		prefixW := tsW + len(label) + 2
		textW := vpW - prefixW
		if textW < 10 {
			textW = 10
		}

		indent := strings.Repeat(" ", prefixW)
		parts := wrapToWidth(l.text, textW)
		for i, part := range parts {
			if i == 0 {
				fmt.Fprintf(&b, "%s %s: %s\n", ts, textStyle.Render(label), textStyle.Render(part))
			} else {
				fmt.Fprintf(&b, "%s%s\n", indent, textStyle.Render(part))
			}
		}
	}
	return b.String()
}

// wrapToWidth splits s into lines of at most n runes, breaking at spaces where possible.
func wrapToWidth(s string, n int) []string {
	if n <= 0 {
		return []string{s}
	}
	var lines []string
	for len(s) > n {
		cut := n
		if i := strings.LastIndex(s[:n], " "); i > 0 {
			cut = i + 1
		}
		lines = append(lines, strings.TrimRight(s[:cut], " "))
		s = s[cut:]
	}
	if len(s) > 0 {
		lines = append(lines, s)
	}
	return lines
}

// ── tea messages ──────────────────────────────────────────────────────────────

type relayEventMsg struct{ ev *rpc.RelayEvent }
type relayExitMsg struct{}

func waitForEvent(read <-chan *rpc.RelayEvent, exit <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		select {
		case ev, ok := <-read:
			if !ok {
				return relayExitMsg{}
			}
			return relayEventMsg{ev}
		case <-exit:
			return relayExitMsg{}
		}
	}
}

// ── views ─────────────────────────────────────────────────────────────────────

type tuiView int

const (
	viewProfiles tuiView = iota
	viewNewProfile
	viewFriends
	viewAddFriend
	viewChat
)

// ── model ─────────────────────────────────────────────────────────────────────

type tuiModel struct {
	serverAddr string
	width      int
	height     int

	// profile picker
	profileList    list.Model
	allProfiles    []*Profile
	newProfileInput textinput.Model
	newProfileErr  string

	// active session (set when a profile is selected)
	profile     *Profile
	pubKeyB64   string
	relayClient *client.ClientRelayClient

	view    tuiView
	friends list.Model
	form    addForm
	chat    *chatState

	readCh <-chan *rpc.RelayEvent
	exitCh <-chan struct{}
}

const listPanelW = 32

func newProfileListModel(profiles []*Profile, h int) list.Model {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(2)
	delegate.SetSpacing(0)
	l := list.New(profilesToListItems(profiles), delegate, 40, h)
	l.Title = "Profiles"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	return l
}

func newFriendListModel(friends []Friend, h int) list.Model {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(2)
	delegate.SetSpacing(0)
	items := make([]list.Item, len(friends))
	for i, f := range friends {
		items[i] = f
	}
	l := list.New(items, delegate, listPanelW-4, h)
	l.Title = "Friends"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	return l
}

func newTuiModel(serverAddr string, profiles []*Profile) tuiModel {
	inp := textinput.New()
	inp.Placeholder = "profile name"
	inp.CharLimit = 64
	inp.Width = 40

	return tuiModel{
		serverAddr:      serverAddr,
		view:            viewProfiles,
		allProfiles:     profiles,
		profileList:     newProfileListModel(profiles, 20),
		newProfileInput: inp,
		friends:         newFriendListModel(nil, 20),
	}
}

func (m tuiModel) Init() tea.Cmd { return nil }

// ── update ────────────────────────────────────────────────────────────────────

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listH := m.height - 6
		if listH < 4 {
			listH = 4
		}
		m.profileList.SetSize(32, listH)
		m.friends.SetSize(listPanelW-4, listH)
		if m.chat != nil {
			m.resizeChatPanels()
		}
		return m, nil

	case relayEventMsg:
		if m.chat != nil {
			m.handleRelayEvent(msg.ev)
		}
		return m, waitForEvent(m.readCh, m.exitCh)

	case relayExitMsg:
		if m.chat != nil {
			m.chat.appendLine("err", "disconnected from relay")
		}
		return m, nil
	}

	switch m.view {
	case viewProfiles:
		return m.updateProfiles(msg)
	case viewNewProfile:
		return m.updateNewProfile(msg)
	case viewFriends:
		return m.updateFriends(msg)
	case viewAddFriend:
		return m.updateAddForm(msg)
	case viewChat:
		return m.updateChat(msg)
	}
	return m, nil
}

func (m tuiModel) updateProfiles(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if ok {
		switch key.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "n":
			m.view = viewNewProfile
			m.newProfileInput.SetValue("")
			m.newProfileErr = ""
			_ = m.newProfileInput.Focus()
			return m, nil
		case "enter":
			sel := m.profileList.SelectedItem()
			if sel == nil {
				return m, nil
			}
			return m.activateProfile(sel.(*Profile))
		}
	}
	var cmd tea.Cmd
	m.profileList, cmd = m.profileList.Update(msg)
	return m, cmd
}

func (m tuiModel) updateNewProfile(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.newProfileInput, cmd = m.newProfileInput.Update(msg)
		return m, cmd
	}

	switch key.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.view = viewProfiles
	case "enter":
		name := strings.TrimSpace(m.newProfileInput.Value())
		if name == "" {
			m.newProfileErr = "name is required"
			return m, nil
		}
		p, err := NewProfile(name)
		if err != nil {
			m.newProfileErr = err.Error()
			return m, nil
		}
		m.allProfiles = append(m.allProfiles, p)
		m.profileList.SetItems(profilesToListItems(m.allProfiles))
		return m.activateProfile(p)
	default:
		var cmd tea.Cmd
		m.newProfileInput, cmd = m.newProfileInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m tuiModel) updateFriends(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if ok {
		switch key.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			// back to profile picker
			m.view = viewProfiles
			m.profile = nil
			m.relayClient = nil
			m.chat = nil
			return m, nil
		case "a":
			m.view = viewAddFriend
			m.form = newAddForm()
			return m, nil
		case "enter":
			sel := m.friends.SelectedItem()
			if sel == nil {
				return m, nil
			}
			return m.openChat(sel.(Friend))
		}
	}
	var cmd tea.Cmd
	m.friends, cmd = m.friends.Update(msg)
	return m, cmd
}

func (m tuiModel) updateAddForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.form.inputs[m.form.focus], cmd = m.form.inputs[m.form.focus].Update(msg)
		return m, cmd
	}

	switch key.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.view = viewFriends
	case "tab", "down":
		m.form.focusField((m.form.focus + 1) % formFieldCount)
	case "shift+tab", "up":
		m.form.focusField((m.form.focus + formFieldCount - 1) % formFieldCount)
	case "enter":
		if m.form.focus < formFieldCount-1 {
			m.form.focusField(m.form.focus + 1)
		} else {
			friend, errMsg := m.form.submit()
			if errMsg != "" {
				m.form.err = errMsg
				m.form.ok = ""
			} else {
				_ = m.friends.InsertItem(len(m.friends.Items()), friend)
				m.form.ok = fmt.Sprintf("added %q", friend.Name)
				m.form.err = ""
				// persist
				m.saveCurrentProfile()
				m.form = newAddForm()
				m.form.ok = fmt.Sprintf("added %q", friend.Name)
			}
		}
	default:
		var cmd tea.Cmd
		m.form.inputs[m.form.focus], cmd = m.form.inputs[m.form.focus].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m tuiModel) updateChat(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmds []tea.Cmd
		var c tea.Cmd
		m.chat.vp, c = m.chat.vp.Update(msg)
		cmds = append(cmds, c)
		m.chat.input, c = m.chat.input.Update(msg)
		cmds = append(cmds, c)
		return m, tea.Batch(cmds...)
	}

	switch key.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.view = viewFriends
		m.chat = nil
	case "enter":
		text := strings.TrimSpace(m.chat.input.Value())
		if text == "" {
			return m, nil
		}
		m.chat.input.SetValue("")
		m.sendMessage(text)
	default:
		var cmd tea.Cmd
		m.chat.input, cmd = m.chat.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (m tuiModel) activateProfile(p *Profile) (tuiModel, tea.Cmd) {
	identity, err := p.Identity()
	if err != nil {
		m.newProfileErr = "bad keys in profile: " + err.Error()
		return m, nil
	}

	m.profile = p
	m.pubKeyB64 = p.PubKey
	m.relayClient = client.NewClientRelayClient(identity, 0)
	m.relayClient.AutoMsgID=true
	m.friends = newFriendListModel(p.Friends, m.height-6)
	m.view = viewFriends
	m.chat = nil
	return m, nil
}

func (m *tuiModel) saveCurrentProfile() {
	if m.profile == nil {
		return
	}
	items := m.friends.Items()
	friends := make([]Friend, 0, len(items))
	for _, it := range items {
		friends = append(friends, it.(Friend))
	}
	m.profile.Friends = friends
	_ = SaveProfile(m.profile)
}

func (m *tuiModel) openChat(f Friend) (tuiModel, tea.Cmd) {
	cs := newChatState(f, m.relayClient)
	m.chat = &cs
	m.resizeChatPanels()
	m.view = viewChat

	read, exit, err := m.relayClient.GetReadHandle(m.serverAddr)
	if err != nil {
		m.chat.appendLine("err", err.Error())
		return *m, nil
	}
	m.readCh = read
	m.exitCh = exit

	m.chat.appendLine("sys", fmt.Sprintf("your pub: %s…", m.pubKeyB64[:min(16, len(m.pubKeyB64))]))
	return *m, waitForEvent(read, exit)
}

func (m *tuiModel) resizeChatPanels() {
	if m.chat == nil {
		return
	}
	chatW := m.width - listPanelW - 6
	if chatW < 20 {
		chatW = 20
	}
	vpH := m.height - 7
	if vpH < 4 {
		vpH = 4
	}
	m.chat.vp.Width = chatW - 4
	m.chat.vp.Height = vpH
	m.chat.input.Width = chatW - 6
}

func (m *tuiModel) sendMessage(text string) {
	f := m.chat.friend
	destBytes, err := base64.StdEncoding.DecodeString(f.PubKeyB64)
	if err != nil {
		m.chat.appendLine("err", "bad dest pub key: "+err.Error())
		return
	}
	err = m.relayClient.Send(m.serverAddr, &rpc.ClientEvent{
		RelayAddr: f.RelayAddr,
		Body: &rpc.ClientEvent_MessagePayload{
			MessagePayload: &rpc.MessagePayload{
				DeliverToPubKey: destBytes,
				CipherText:      []byte(text),
			},
		},
	})
	if err != nil {
		m.chat.appendLine("err", err.Error())
	} else {
		m.chat.appendLine("me", text)
	}
}

func (m *tuiModel) handleRelayEvent(ev *rpc.RelayEvent) {
	switch v := ev.Body.(type) {
	case *rpc.RelayEvent_MessagePayload:
		m.chat.appendLine(m.chat.friend.Name, string(v.MessagePayload.CipherText))
	case *rpc.RelayEvent_NoiseHandshake:
		m.chat.appendLine("sys", fmt.Sprintf("noise handshake (%d bytes)", len(v.NoiseHandshake.HandshakeMsg)))
	case *rpc.RelayEvent_RelayAck:
		m.chat.appendLine("sys", fmt.Sprintf("ack msgID=%d", v.RelayAck.MessageID))
	}
}

// ── view ──────────────────────────────────────────────────────────────────────

func (m tuiModel) View() string {
	switch m.view {
	case viewProfiles:
		return m.viewProfiles()
	case viewNewProfile:
		return m.viewNewProfile()
	case viewFriends:
		return m.viewFriends()
	case viewAddFriend:
		return m.viewAddForm()
	case viewChat:
		return m.viewChat()
	}
	return ""
}

func (m tuiModel) viewProfiles() string {
	panelH := m.height - 4
	if panelH < 4 {
		panelH = 4
	}
	listW := 36
	m.profileList.SetSize(listW-4, panelH-2)

	rightW := m.width - listW - 6
	if rightW < 20 {
		rightW = 20
	}

	var detail string
	if sel := m.profileList.SelectedItem(); sel != nil {
		p := sel.(*Profile)
		detail = strings.Join([]string{
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Render(p.Name),
			"",
			styleFormLabel.Render("pub key:"),
			styleMuted.Render(strings.Join(wrapToWidth(p.PubKey, rightW-4), "\n")),
			"",
			styleMuted.Render(fmt.Sprintf("%d friend(s)", len(p.Friends))),
			"",
			styleMuted.Render("press enter to log in"),
		}, "\n")
	} else {
		detail = styleMuted.Render("no profiles yet\n\npress n to create one")
	}

	leftPanel  := stylePanelBorder.Width(listW).Height(panelH).Render(m.profileList.View())
	rightPanel := stylePanelBorder.Width(rightW).Height(panelH).Render(detail)
	top        := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)
	help       := styleHelp.Render("↑/↓ navigate · enter select · n new profile · q quit")
	return lipgloss.JoinVertical(lipgloss.Left, top, help)
}

func (m tuiModel) viewNewProfile() string {
	var b strings.Builder
	fmt.Fprintln(&b, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250")).Render("New profile"))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, styleFormLabel.Render("Name:"))
	fmt.Fprintln(&b, "  "+m.newProfileInput.View())
	if m.newProfileErr != "" {
		fmt.Fprintln(&b)
		fmt.Fprint(&b, styleError.Render("✗ "+m.newProfileErr))
	}
	panelW := 50
	if m.width > 0 && m.width-4 < panelW {
		panelW = m.width - 4
	}
	panel := stylePanelBorder.Width(panelW).Render(b.String())
	help  := styleHelp.Render("enter create · esc back")
	return lipgloss.JoinVertical(lipgloss.Left, panel, help)
}

func (m tuiModel) viewFriends() string {
	panelH := m.height - 4
	if panelH < 4 {
		panelH = 4
	}
	m.friends.SetSize(listPanelW-4, panelH-2)
	leftPanel := stylePanelBorder.Width(listPanelW).Height(panelH).Render(m.friends.View())

	rightW := m.width - listPanelW - 6
	if rightW < 20 {
		rightW = 20
	}
	var detail string
	if sel := m.friends.SelectedItem(); sel != nil {
		f := sel.(Friend)
		detail = strings.Join([]string{
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Render(f.Name),
			"",
			styleFormLabel.Render("relay:") + "  " + f.RelayAddr,
			"",
			styleFormLabel.Render("pub key:"),
			styleMuted.Render(strings.Join(wrapToWidth(f.PubKeyB64, rightW-6), "\n")),
			"",
			styleMuted.Render("press enter to open chat"),
		}, "\n")
	} else {
		detail = styleMuted.Render("select a friend")
	}
	profileLine := styleHelp.Render(fmt.Sprintf("profile: %s", m.profile.Name))
	rightPanel  := stylePanelBorder.Width(rightW).Height(panelH).Render(profileLine + "\n\n" + detail)

	top  := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)
	help := styleHelp.Render("↑/↓ navigate · enter open chat · a add friend · q switch profile")
	return lipgloss.JoinVertical(lipgloss.Left, top, help)
}

func (m tuiModel) viewAddForm() string {
	var b strings.Builder
	fmt.Fprintln(&b, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250")).Render("Add friend"))
	fmt.Fprintln(&b)
	for i := 0; i < formFieldCount; i++ {
		fmt.Fprintln(&b, styleFormLabel.Render(formLabels[i]+":"))
		fmt.Fprintln(&b, "  "+m.form.inputs[i].View())
		fmt.Fprintln(&b)
	}
	switch {
	case m.form.err != "":
		fmt.Fprint(&b, styleError.Render("✗ "+m.form.err))
	case m.form.ok != "":
		fmt.Fprint(&b, styleSuccess.Render("✓ "+m.form.ok))
	}
	panelW := m.width - 4
	if panelW > 72 {
		panelW = 72
	}
	panel := stylePanelBorder.Width(panelW).Render(b.String())
	help  := styleHelp.Render("tab/↑↓ move · enter next/submit · esc back")
	return lipgloss.JoinVertical(lipgloss.Left, panel, help)
}

func (m tuiModel) viewChat() string {
	if m.chat == nil {
		return ""
	}
	panelH := m.height - 4
	if panelH < 4 {
		panelH = 4
	}
	m.friends.SetSize(listPanelW-4, panelH-2)
	leftPanel := stylePanelBorder.Width(listPanelW).Height(panelH).Render(m.friends.View())

	rightW := m.width - listPanelW - 6
	if rightW < 20 {
		rightW = 20
	}
	chatTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250")).
		Render("Chat · " + m.chat.friend.Name)
	vpView   := m.chat.vp.View()
	inputBox := stylePanelBorder.Width(rightW - 2).Render(m.chat.input.View())

	rightContent := lipgloss.JoinVertical(lipgloss.Left, chatTitle, vpView, inputBox)
	rightPanel   := stylePanelBorder.Width(rightW).Height(panelH).Render(rightContent)

	top  := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)
	help := styleHelp.Render("enter send · esc friends · q switch profile · ctrl+c quit")
	return lipgloss.JoinVertical(lipgloss.Left, top, help)
}

// ── command ───────────────────────────────────────────────────────────────────

func CmdTuiClient(_ context.Context, c *cli.Command) error {
	for _, logger := range log4zero.LoggerRegistry {
		*logger = logger.Level(zerolog.Disabled)
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.Disabled)

	profiles, err := LoadAllProfiles()
	if err != nil {
		return fmt.Errorf("load profiles: %w", err)
	}

	m := newTuiModel(c.Value("addr").(string), profiles)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
