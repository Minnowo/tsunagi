package client

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

func processReads(r *ClientRelayStream) {

	for {

		select {
		case <-r.exit:
			return
		default:
			break
		}

		if !r.IsConnected() {

			if err := r.connect(context.Background()); err != nil {

				log.Warn().Err(err).Msg("failed to connect, retrying in 1s")

				time.Sleep(time.Second)

				continue
			}
		}

		event, err := r.stream.Recv()

		if err != nil {

			log.Warn().Err(err).Msg("stream recv failed")

			r.disconnect()

			continue
		}

		r.read <- event
	}
}
