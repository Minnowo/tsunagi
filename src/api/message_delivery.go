package api

import (
	"context"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
)

func (this *TsunagiBase) DeliverMessage(ctx context.Context, req *rpc.RelayEvent) (uint64, error) {

	var deviceID data.Identifier

	switch v := req.Body.(type) {
	case *rpc.RelayEvent_MessagePayload:

		if err := deviceID.FromBytes(v.MessagePayload.DeliverToPubKey); err != nil {
			return v.MessagePayload.MessageID, err
		}

		// check if id user is online, if so, pass message directly to them
		ok := this.ClientConns.PutRelayMsg(deviceID, req)

		if ok {
			log.Info().Str("device", deviceID.String()).Msg("message delivered to client")
			return v.MessagePayload.MessageID, nil
		}

		// else save message in the database
		log.Info().Str("device", deviceID.String()).Msg("message put in the DB (not really)")

		return v.MessagePayload.MessageID, nil

	case *rpc.RelayEvent_NoiseHandshake:

		if err := deviceID.FromBytes(v.NoiseHandshake.DeliverToPubKey); err != nil {
			return v.NoiseHandshake.MessageID, err
		}

		// check if id user is online, if so, pass message directly to them
		ok := this.ClientConns.PutRelayMsg(deviceID, req)

		if ok {
			log.Info().Str("device", deviceID.String()).Msg("message delivered to client")
			return v.NoiseHandshake.MessageID, nil
		}

		// else save message in the database
		log.Info().Str("device", deviceID.String()).Msg("message put in the DB (not really)")

		return v.NoiseHandshake.MessageID, nil

		// case *rpc.RelayEvent_RelayAck:

		// 	return this.DeliverNoiseHandshake(ctx, deviceID, v.RelayAck.HandshakeMsg)
	}

	return 0, nil
}
