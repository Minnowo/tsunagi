package relayapi

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
)

func (this *RelayApi) ForwardMessage(ctx context.Context, req *rpc.ClientEvent) error {

	relayAddr, err := url.ParseRequestURI(req.RelayAddr)

	log.Info().Interface("payload", req).Msg("got forward")
	if err != nil {
		return err
	}

	switch strings.ToLower(relayAddr.Scheme) {

	default:
		return fmt.Errorf("bad protocol: %s", relayAddr.Scheme)

	case "grpc", "tcp", "http", "https", "dns":

		// // gRPC ignores http/https scheme semantics; we just use host:port
		address := relayAddr.Host

		if address == "" {
			return fmt.Errorf("missing host in relay address")
		}

		var err error

		switch v := req.Body.(type){

		case *rpc.ClientEvent_MessagePayload:

		err = this.relayClient.Send(address, &rpc.RelayEvent{
			DeviceID:   req.DeviceID,
			Body: &rpc.RelayEvent_MessagePayload{
				MessagePayload: v.MessagePayload ,
			},
		})

		case *rpc.ClientEvent_NoiseHandshake:

		err = this.relayClient.Send(address, &rpc.RelayEvent{
			DeviceID:   req.DeviceID,
			Body: &rpc.RelayEvent_NoiseHandshake{
				NoiseHandshake: v.NoiseHandshake,
			},
		})

		}

		if err != nil {
			return err
		}
	}

	return nil
}
