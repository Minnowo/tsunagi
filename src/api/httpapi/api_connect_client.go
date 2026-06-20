package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"tsunagi/src/api"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/coder/websocket"
	"github.com/rs/zerolog/log"
)

// wsClientEvent is the JSON shape for messages the client sends over the WebSocket.
type wsClientEvent struct {
	PubKey     []byte `json:"pub_key"`
	RelayAddr  string `json:"relay_addr"`
	Type       string `json:"type"`        // "noise_handshake" | "message_payload"
	State      []byte `json:"state"`       // noise_handshake
	CipherText []byte `json:"cipher_text"` // message_payload
}

// wsRelayEvent is the JSON shape for messages the server sends to the client.
type wsRelayEvent struct {
	PubKey     []byte `json:"pub_key"`
	Type       string `json:"type"`
	State      []byte `json:"state"`
	CipherText []byte `json:"cipher_text"`
}

func relayEventToWS(ev *rpc.RelayEvent) wsRelayEvent {

	out := wsRelayEvent{PubKey: ev.PubKey}

	switch v := ev.Body.(type) {
	case *rpc.RelayEvent_NoiseHandshake:
		out.Type = "noise_handshake"
		out.State = v.NoiseHandshake.State
	case *rpc.RelayEvent_MessagePayload:
		out.Type = "message_payload"
		out.CipherText = v.MessagePayload.CipherText
	}

	return out
}

func (ev *wsClientEvent) toProto() *rpc.ClientEvent {

	ce := &rpc.ClientEvent{
		PubKey:    ev.PubKey,
		RelayAddr: ev.RelayAddr,
	}

	switch ev.Type {
	case "noise_handshake":
		ce.Body = &rpc.ClientEvent_NoiseHandshake{
			NoiseHandshake: &rpc.NoiseHandshake{State: ev.State},
		}
	case "message_payload":
		ce.Body = &rpc.ClientEvent_MessagePayload{
			MessagePayload: &rpc.MessagePayload{CipherText: ev.CipherText},
		}
	}

	return ce
}

func (h *HttpRelayApi) apiConnectClient(w http.ResponseWriter, r *http.Request) {

	pubkey, err := h.getAuthIdentity(r)
	if err != nil {
		api.Unauthorized(w)
		return
	}

	var id data.Identifier
	if err := id.FromBytes(pubkey); err != nil {
		api.Unauthorized(w)
		return
	}
	log.Debug().Str("deviceID", id.String()).Msg("client device connected")

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Debug().Err(err).Msg("connect_client: accept error")
		return
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	ch := api.ClientConn{
		SendCh: make(chan *rpc.RelayEvent, 16),
		Ctx:    ctx,
	}

	if !h.ClientConns.AddConn(id, &ch) {
		conn.Close(websocket.StatusPolicyViolation, "already connected")
		return
	}
	defer h.ClientConns.RemoveConn(id)

	// goroutine: relay events -> JSON -> WS
	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-ch.SendCh:
				if !ok {
					return
				}
				msg, err := json.Marshal(relayEventToWS(ev))
				if err != nil {
					log.Debug().Err(err).Msg("connect_client: marshal error")
					return
				}
				if err := conn.Write(ctx, websocket.MessageText, msg); err != nil {
					log.Debug().Err(err).Msg("connect_client: write error")
					return
				}
			}
		}
	}()

	// main loop: WS -> ClientEvent -> forward
	for {
		_, raw, err := conn.Read(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("connect_client: read error")
			return
		}

		var ev wsClientEvent

		if err := json.Unmarshal(raw, &ev); err != nil {
			log.Debug().Err(err).Msg("connect_client: unmarshal error")
			continue
		}

		log.Info().Interface("msg", ev).Msg("got msg")

		if err := h.ForwardMessage(ctx, ev.toProto()); err != nil {
			log.Error().Err(err).Msg("connect_client: forward error")
		}
	}
}
