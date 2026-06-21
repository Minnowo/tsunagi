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
	RelayAddr    string `json:"relay_addr"`
	Type         string `json:"type"` // "noise_handshake" | "message_payload"
	MessageID    uint64 `json:"message_id"`
	DeliverToPub []byte `json:"deliver_to_pub_key"`
	HandshakeMsg []byte `json:"handshake_msg"` // noise_handshake
	CipherText   []byte `json:"cipher_text"`   // message_payload
}

// wsRelayEvent is the JSON shape for messages the server sends to the client.
type wsRelayEvent struct {
	Type         string `json:"type"`
	MessageID    uint64 `json:"message_id"`
	DeliverToPub []byte `json:"deliver_to_pub_key,omitempty"`
	HandshakeMsg []byte `json:"handshake_msg,omitempty"`
	CipherText   []byte `json:"cipher_text,omitempty"`
}

func relayEventToWS(ev *rpc.RelayEvent) wsRelayEvent {
	switch v := ev.Body.(type) {
	case *rpc.RelayEvent_NoiseHandshake:
		return wsRelayEvent{
			Type:         "noise_handshake",
			MessageID:    v.NoiseHandshake.MessageID,
			DeliverToPub: v.NoiseHandshake.DeliverToPubKey,
			HandshakeMsg: v.NoiseHandshake.HandshakeMsg,
		}
	case *rpc.RelayEvent_MessagePayload:
		return wsRelayEvent{
			Type:         "message_payload",
			MessageID:    v.MessagePayload.MessageID,
			DeliverToPub: v.MessagePayload.DeliverToPubKey,
			CipherText:   v.MessagePayload.CipherText,
		}
	case *rpc.RelayEvent_RelayAck:
		return wsRelayEvent{
			Type:      "relay_ack",
			MessageID: v.RelayAck.MessageID,
		}
	}
	return wsRelayEvent{}
}

func (ev *wsClientEvent) toProto() *rpc.ClientEvent {
	ce := &rpc.ClientEvent{RelayAddr: ev.RelayAddr}
	switch ev.Type {
	case "noise_handshake":
		ce.Body = &rpc.ClientEvent_NoiseHandshake{
			NoiseHandshake: &rpc.NoiseHandshake{
				MessageID:       ev.MessageID,
				DeliverToPubKey: ev.DeliverToPub,
				HandshakeMsg:    ev.HandshakeMsg,
			},
		}
	case "message_payload":
		ce.Body = &rpc.ClientEvent_MessagePayload{
			MessagePayload: &rpc.MessagePayload{
				MessageID:       ev.MessageID,
				DeliverToPubKey: ev.DeliverToPub,
				CipherText:      ev.CipherText,
			},
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
		ClientID: id,
		SendCh:   make(chan *rpc.RelayEvent, 16),
		Ctx:      ctx,
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

		if err := h.ForwardMessage(&ch, ev.toProto()); err != nil {
			log.Error().Err(err).Msg("connect_client: forward error")
		}
	}
}
