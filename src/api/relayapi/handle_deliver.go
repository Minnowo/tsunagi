package relayapi

import (
	"context"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
)

func (this *RelayApi) DeliverMessage(ctx context.Context, req *rpc.DeliverRequest) error {

	var deviceID data.Identifier

	if err := deviceID.FromBytes(req.DeviceID); err != nil {
		return err
	}

	err := this.inbox.PutMsg(deviceID, req.CipherText)

	if err != nil {
		return err
	}

	log.Info().
		Hex("device", req.DeviceID[:]).
		Bytes("mgs", req.CipherText[:]).
		Msg("message was delivered to inbox")

	return nil
}
