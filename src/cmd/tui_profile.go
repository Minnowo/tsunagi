package cmd

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/flynn/noise"
	"github.com/minnowo/tsunagi/mod/tcrypto"
)

const profilesDir = "profiles"

// ── Profile data ──────────────────────────────────────────────────────────────

type Profile struct {
	Name    string   `json:"name"`
	PubKey  string   `json:"pub_key"` // base64
	PriKey  string   `json:"pri_key"` // base64
	Friends []Friend `json:"friends"`
}

func (p Profile) Title() string       { return p.Name }
func (p Profile) Description() string { return p.PubKey[:min(24, len(p.PubKey))] + "…" }
func (p Profile) FilterValue() string { return p.Name }

func (p *Profile) Identity() (noise.DHKey, error) {
	pub, err := base64.StdEncoding.DecodeString(p.PubKey)
	if err != nil {
		return noise.DHKey{}, err
	}
	pri, err := base64.StdEncoding.DecodeString(p.PriKey)
	if err != nil {
		return noise.DHKey{}, err
	}
	return noise.DHKey{Public: pub, Private: pri}, nil
}

// ── Persistence ───────────────────────────────────────────────────────────────

func profilePath(name string) string {
	safe := strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == 0 {
			return '_'
		}
		return r
	}, name)
	return filepath.Join(profilesDir, safe+".json")
}

func SaveProfile(p *Profile) error {
	if err := os.MkdirAll(profilesDir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(profilePath(p.Name), data, 0o600)
}

func LoadAllProfiles() ([]*Profile, error) {
	entries, err := os.ReadDir(profilesDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var profiles []*Profile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(profilesDir, e.Name()))
		if err != nil {
			continue
		}
		var p Profile
		if err := json.Unmarshal(data, &p); err != nil {
			continue
		}
		profiles = append(profiles, &p)
	}
	return profiles, nil
}

func NewProfile(name string) (*Profile, error) {
	key, err := tcrypto.GenerateNoiseKeypair()
	if err != nil {
		return nil, err
	}
	p := &Profile{
		Name:   name,
		PubKey: base64.StdEncoding.EncodeToString(key.Public),
		PriKey: base64.StdEncoding.EncodeToString(key.Private),
	}
	return p, SaveProfile(p)
}

// ── Profile list items ────────────────────────────────────────────────────────

func profilesToListItems(profiles []*Profile) []list.Item {
	items := make([]list.Item, len(profiles))
	for i, p := range profiles {
		items[i] = p
	}
	return items
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
