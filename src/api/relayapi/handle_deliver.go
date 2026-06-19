package relayapi

import (
	"context"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
)

func (this *RelayApi) DeliverMessage(ctx context.Context, req *rpc.RelayEvent) error {

	var deviceID data.Identifier

	if err := deviceID.FromBytes(req.DeviceID); err != nil {
		return err
	}

	switch v := req.Body.(type) {

	case *rpc.RelayEvent_MessagePayload:

		err := this.inbox.PutMsg(deviceID, v.MessagePayload.CipherText)

		if err != nil {
			return err
		}

		log.Info().
			Hex("device", req.DeviceID[:]).
			Bytes("mgs", v.MessagePayload.CipherText).
			Msg("message was delivered to inbox")

	case *rpc.RelayEvent_NoiseHandshake:

		log.Info().
			Hex("device", req.DeviceID[:]).
			Bytes("mgs", v.NoiseHandshake.State).
			Msg("got noise handshake")

	}

	return nil
}
