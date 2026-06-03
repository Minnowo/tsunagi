package clientapi

import (
	"net/http"

	"github.com/coder/websocket"
)

func (this *ClientApi) HandleWebSocket(w http.ResponseWriter, r *http.Request) {

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})

	if err != nil {
		logger.Debug().Err(err).Msg("ws_connect: accept error")
		return
	}

	defer conn.CloseNow()

	ctx := r.Context()

	for {
		msgType, data, err := conn.Read(ctx)

		if err != nil {
			logger.Debug().Err(err).Msg("ws_connect: read error")
			return
		}

		if err := conn.Write(ctx, msgType, data); err != nil {
			logger.Debug().Err(err).Msg("ws_connect: write error")
			return
		}
	}
}
