package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"tsunagi/src/rpc"
)

func (this *TsunagiBase) ForwardMessage(ctx context.Context, req *rpc.ClientEvent) error {

	relayAddr, err := url.ParseRequestURI(req.RelayAddr)

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

		switch v := req.Body.(type) {

		case *rpc.ClientEvent_MessagePayload:

			return this.RelayClient.Send(address, &rpc.RelayEvent{
				DeviceID: req.DeviceID,
				Body: &rpc.RelayEvent_MessagePayload{
					MessagePayload: v.MessagePayload,
				},
			})

		case *rpc.ClientEvent_NoiseHandshake:

			return this.RelayClient.Send(address, &rpc.RelayEvent{
				DeviceID: req.DeviceID,
				Body: &rpc.RelayEvent_NoiseHandshake{
					NoiseHandshake: v.NoiseHandshake,
				},
			})

		}
	}

	return nil
}
