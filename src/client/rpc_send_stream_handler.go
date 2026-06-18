package client

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	ErrReqTimeout = fmt.Errorf("request deadline hit")
)

// _SendStream represents an RPC client stream.
type _SendStream interface {
	IsConnected() bool
	connect(ctx context.Context) error
	disconnect()
	rpcSend(req any) error
}

// processSends handles sending events from the given recieve channel into the RPC stream.
// It will reconnect as needed.
func processSends(
	exit chan struct{},
	recieve chan SendEventRequest,
	stream _SendStream,
) {

	var req *SendEventRequest
	for {

		if req != nil && req.IsDeadline() {
			req.PutErr(ErrReqTimeout)
			req = nil
		}

		select {
		case <-exit:
			stream.disconnect()
			return
		default:
			break
		}

		if !stream.IsConnected() {

			var ctx context.Context

			if req == nil {
				ctx = context.Background()
			} else {
				ctx = req.ctx
			}

			if err := stream.connect(ctx); err != nil {

				log.Warn().Err(err).Msg("failed to connect, retrying in 1s")
				time.Sleep(time.Second)

				continue
			}
		}

		if req != nil {

			if err := stream.rpcSend(req.event); err != nil {

				log.Warn().Err(err).Msg("stream send failed, reconnecting")

				stream.disconnect()
				time.Sleep(time.Second)

				continue
			} else {
				req.PutErr(nil)
				req = nil
			}
		}

		select {
		case nextEvent, ok := <-recieve:

			if !ok {
				log.Debug().Msg("send channel closed")
				return
			}

			req = &nextEvent
		}
	}
}
