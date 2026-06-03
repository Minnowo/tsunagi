package relayapi

import (
	"context"
	"io"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (this *RelayApi) ConnectRelay(stream grpc.ClientStreamingServer[rpc.Event, rpc.Empty]) error {

	md, ok := metadata.FromIncomingContext(stream.Context())

	if !ok || !this.ValidAuth(md) {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

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
