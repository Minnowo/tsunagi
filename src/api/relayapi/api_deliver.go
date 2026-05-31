package relayapi

import (
	"context"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
)

func(this*RelayApi) DeliverMessage(ctx context.Context, req *rpc.DeliverRequest) (*rpc.Empty, error) {

	var deviceID data.Identifier

	if err := deviceID.FromBytes(req.DeviceID) ; err != nil {
		return nil, err
	}

	err := this.inbox.PutMsg(deviceID, req.CipherText)

	if err != nil  {
		return nil, err
	}

	log.Info().
		Hex("device", req.DeviceID[:]).
		Msg("message was delivered to inbox")

	return &rpc.Empty{}, nil
}
