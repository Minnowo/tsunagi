package relayapi

import (
	"context"
	"crypto/rand"
	"io"
	"time"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (this *RelayApi) Connect(stream grpc.BidiStreamingServer[rpc.Event, rpc.Event]) error {

	md, ok := metadata.FromIncomingContext(stream.Context())

	if !ok || !this.ValidAuth(md) {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	go func() {

		// this is temporary to test the client.Client reads
		var deviceId [32]byte
		ticker := time.NewTicker(time.Second * 5)
		for {
			select {
			case <-ticker.C:
				log.Info().Msg("tick")
				rand.Read(deviceId[:])

				err := stream.Send(
					&rpc.Event{
						Body: &rpc.Event_DeliverRequest{
							DeliverRequest: &rpc.DeliverRequest{
								DeviceID: deviceId[:],
							},
						},
					})
				if err != nil {
					log.Info().Msg("client disconnected")
					return
				}
			}
		}

	}()

	for {
		event, err := stream.Recv()

		if err != nil {

			log.Debug().Err(err).Msg("api_connect: read error")

			if err == io.EOF {
				return nil
			}

			return err
		}

		switch v := event.Body.(type) {

		case *rpc.Event_DeliverRequest:

			dctx := context.Background()
			this.DeliverMessage(dctx, v.DeliverRequest)

		case *rpc.Event_ForwardRequest:

			fctx := context.Background()
			this.ForwardMessage(fctx, v.ForwardRequest)
		}
	}
}
