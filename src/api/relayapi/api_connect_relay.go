package relayapi

import (
	"context"
	"io"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (this *RelayApi) ConnectRelay(stream grpc.ClientStreamingServer[rpc.RelayEvent, rpc.Empty]) error {

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

	log.Debug().Hex("deviceID", pubkey).Msg("relay connected")

	for {
		event, err := stream.Recv()

		if err != nil {

			log.Debug().Err(err).Msg("api_connect: read error")

			if err == io.EOF {
				return nil
			}

			return err
		}

		err = this.DeliverMessage(stream.Context(), event)
		if err != nil {
			log.Error().Err(err).Msg("error delivering message")
		}
	}
}

func (this *RelayApi) DeliverMessage(ctx context.Context, req *rpc.RelayEvent) error {

	var deviceID data.Identifier

	if err := deviceID.FromBytes(req.PubKey); err != nil {
		return err
	}

	switch v := req.Body.(type) {

	case *rpc.RelayEvent_MessagePayload:
		return this.DeliverMessagePayload(ctx, deviceID, v.MessagePayload.CipherText)

	case *rpc.RelayEvent_NoiseHandshake:
		return this.DeliverNoiseHandshake(ctx, deviceID, v.NoiseHandshake.State)
	}

	return nil
}
