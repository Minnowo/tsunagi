package relayapi

import (
	"context"
	"io"
	"tsunagi/src/api"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (this *RelayApi) ConnectClient(stream grpc.BidiStreamingServer[rpc.ClientEvent, rpc.RelayEvent]) error {

	log.Debug().Msg("client trying to connect")

	md, ok := metadata.FromIncomingContext(stream.Context())

	if !ok {
		log.Debug().Msg("client failed to connect - no auth context")
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	pubkey, err := this.GetAuthIdentity(md)

	if err != nil {
		log.Debug().Err(err).Msg("client failed to connect - bad token")
		return status.Error(codes.Unauthenticated, "bad token")
	}

	var id data.Identifier

	if err := id.FromBytes(pubkey); err != nil {
		log.Debug().Err(err).Msg("client failed to connect - pubkey could not fit inside identifier")
		return status.Error(codes.Unauthenticated, "bad token")
	}

	log.Debug().Str("deviceID", id.String()).Msg("client device connected")

	ctx, cancel := context.WithCancel(stream.Context())

	conn := api.ClientConn{
		SendCh: make(chan *rpc.RelayEvent),
		Ctx:    ctx,
	}

	if !this.ClientConns.AddConn(id, &conn) {
		return api.ErrClientConnExists
	}
	defer this.ClientConns.RemoveConn(id)

	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return

			case msg, ok := <-conn.SendCh:
				if !ok {
					return
				}

				log.Debug().Msg("sending message to client")

				err := stream.Send(msg)

				if err != nil {
					log.Debug().Err(err).Msg("send failed, closing stream")
					cancel()
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

		err = this.ForwardMessage(stream.Context(), event)

		if err != nil {
			log.Error().Err(err).Msg("error forwarding message")
		}
	}
}
