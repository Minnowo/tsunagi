package httpapi

import (
	"encoding/json"
	"net/http"
	"tsunagi/src/api"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/coder/websocket"
	"github.com/rs/zerolog/log"
)

// wsRelayInEvent is the JSON shape for RelayEvents a relay pushes to us.
type wsRelayInEvent struct {
	PubKey     []byte `json:"pub_key"`
	Type       string `json:"type"`        // "noise_handshake" | "message_payload"
	State      []byte `json:"state"`       // noise_handshake
	CipherText []byte `json:"cipher_text"` // message_payload
}

func (ev *wsRelayInEvent) toProto() *rpc.RelayEvent {

	re := &rpc.RelayEvent{PubKey: ev.PubKey}

	switch ev.Type {
	case "noise_handshake":
		re.Body = &rpc.RelayEvent_NoiseHandshake{
			NoiseHandshake: &rpc.NoiseHandshake{State: ev.State},
		}
	case "message_payload":
		re.Body = &rpc.RelayEvent_MessagePayload{
			MessagePayload: &rpc.MessagePayload{CipherText: ev.CipherText},
		}
	}

	return re
}

func (h *HttpRelayApi) apiConnectRelay(w http.ResponseWriter, r *http.Request) {

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
	log.Debug().Hex("deviceID", pubkey).Msg("relay connected")

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})

	if err != nil {
		log.Debug().Err(err).Msg("connect_relay: accept error")
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()

	for {
		_, raw, err := conn.Read(ctx)
		if err != nil {
			log.Debug().Err(err).Msg("connect_relay: read error")
			return
		}

		var ev wsRelayInEvent

		if err := json.Unmarshal(raw, &ev); err != nil {
			log.Debug().Err(err).Msg("connect_relay: unmarshal error")
			continue
		}

		if err := h.DeliverMessage(ctx, ev.toProto()); err != nil {
			log.Error().Err(err).Msg("connect_relay: deliver error")
		}
	}
}
