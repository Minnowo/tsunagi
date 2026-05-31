package relayapi

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)


func(this*RelayApi) ForwardMessage(ctx context.Context, req *rpc.ForwardRequest) (*rpc.Empty, error) {

	relayAddr, err := url.ParseRequestURI(req.RelayAddr)

	log.Info().Interface("payload", req).Msg("got forward")
	if err != nil {
		return nil, err
	}

  switch strings.ToLower(relayAddr.Scheme) {

    default:
        return nil, fmt.Errorf("bad protocol: %s", relayAddr.Scheme)

    case "grpc", "tcp", "http", "https", "dns":

        // gRPC ignores http/https scheme semantics; we just use host:port
        address := relayAddr.Host
        if address == "" {
            return nil, fmt.Errorf("missing host in relay address")
        }

		log.Info().Str("addr", address).Msg("forwarding to")

        conn, err := grpc.NewClient(
            address,
            grpc.WithTransportCredentials(insecure.NewCredentials()),
        )

        if err != nil {
            return nil, err
        }
        defer conn.Close()

        client := rpc.NewTsunagiClient(conn)

        _, err = client.DeliverMessage(ctx, &rpc.DeliverRequest{
            DeviceID:   req.DeviceID,
            CipherText: req.CipherText,
        })

        if err != nil {
            return nil, err
        }

        log.Info().
            Str("addr", address).
            Bytes("text", req.CipherText).
            Msg("delivered message via gRPC")
    }

	return &rpc.Empty{}, nil
}
