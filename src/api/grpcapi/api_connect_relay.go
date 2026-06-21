package grpcapi

import (
	"io"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (this *RelayApi) ConnectRelay(stream grpc.BidiStreamingServer[rpc.RelayEvent, rpc.RelayAck]) error {

	md, ok := metadata.FromIncomingContext(stream.Context())

	if !ok {
		log.Debug().Msg("relay failed to connect - no auth context")
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	pubkey, err := this.GetAuthIdentity(md)

	if err != nil {
		log.Debug().Err(err).Msg("relay failed to connect - bad token")
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	var id data.Identifier

	if err := id.FromBytes(pubkey); err != nil {
		log.Debug().Err(err).Msg("relay failed to connect - pubkey could not fit inside identifier")
		return status.Error(codes.Unauthenticated, "bad token")
	}

	log.Debug().Str("deviceID", id.String()).Msg("relay connected")

	ctx := stream.Context()

	for {
		event, err := stream.Recv()

		if err != nil {

			log.Debug().Err(err).Msg("api_connect: read error")

			if err == io.EOF {
				return nil
			}

			return err
		}

		msgID, err := this.DeliverMessage(ctx, event)

		if err != nil {
			log.Error().Err(err).Msg("error delivering message")
		}

		stream.SendMsg(&rpc.RelayAck{
			MessageID: msgID,
		})
	}
}
