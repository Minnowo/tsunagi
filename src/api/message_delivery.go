package api

import (
	"context"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
)

func (this *TsunagiBase) DeliverMessage(ctx context.Context, req *rpc.RelayEvent) error {

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

func (this *TsunagiBase) DeliverNoiseHandshake(ctx context.Context, id data.Identifier, noiseMsg []byte) error {

	// check if id user is online, if so, pass message directly to them

	// else save message in the database

	return nil
}

func (this *TsunagiBase) DeliverMessagePayload(ctx context.Context, id data.Identifier, cipherText []byte) error {

	// check if id user is online, if so, pass message directly to them
	ok := this.ClientConns.PutRelayMsg(id, &rpc.RelayEvent{
		Body: &rpc.RelayEvent_MessagePayload{
			MessagePayload: &rpc.MessagePayload{
				CipherText: cipherText,
			},
		},
	})

	if ok {
		log.Info().Msg("message delivered to client")
		return nil
	}

	// else save message in the database
	log.Info().Msg("message put in the DB (not really)")

	return nil
}
