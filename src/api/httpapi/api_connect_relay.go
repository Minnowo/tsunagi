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
	Type         string `json:"type"` // "noise_handshake" | "message_payload"
	MessageID    uint64 `json:"message_id"`
	DeliverToPub []byte `json:"deliver_to_pub_key"`
	HandshakeMsg []byte `json:"handshake_msg"` // noise_handshake
	CipherText   []byte `json:"cipher_text"`   // message_payload
}

// wsRelayAck is the JSON shape sent back to the relay after delivery.
type wsRelayAck struct {
	Type      string `json:"type"`
	MessageID uint64 `json:"message_id"`
}

func (ev *wsRelayInEvent) toProto() *rpc.RelayEvent {
	switch ev.Type {
	case "noise_handshake":
		return &rpc.RelayEvent{
			Body: &rpc.RelayEvent_NoiseHandshake{
				NoiseHandshake: &rpc.NoiseHandshake{
					MessageID:       ev.MessageID,
					DeliverToPubKey: ev.DeliverToPub,
					HandshakeMsg:    ev.HandshakeMsg,
				},
			},
		}
	case "message_payload":
		return &rpc.RelayEvent{
			Body: &rpc.RelayEvent_MessagePayload{
				MessagePayload: &rpc.MessagePayload{
					MessageID:       ev.MessageID,
					DeliverToPubKey: ev.DeliverToPub,
					CipherText:      ev.CipherText,
				},
			},
		}
	}
	return &rpc.RelayEvent{}
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
	log.Debug().Str("deviceID", id.String()).Msg("relay connected")

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

		msgID, err := h.DeliverMessage(ctx, ev.toProto())
		if err != nil {
			log.Error().Err(err).Msg("connect_relay: deliver error")
		}

		ack, _ := json.Marshal(wsRelayAck{Type: "relay_ack", MessageID: msgID})
		if err := conn.Write(ctx, websocket.MessageText, ack); err != nil {
			log.Debug().Err(err).Msg("connect_relay: ack write error")
			return
		}
	}
}
