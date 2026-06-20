package grpcapi

import (
	"context"
	"time"
	"tsunagi/src/rpc"

	"github.com/minnowo/tsunagi/mod/tcrypto"
	"github.com/rs/zerolog/log"
)

func (this *RelayApi) GetChallenge(ctx context.Context, req *rpc.AuthRequest) (*rpc.AuthChallenge, error) {

	state, err := tcrypto.NewResponderHandshakeIN()

	if err != nil {
		return nil, err
	}

	handshakeDoneMsg, cipher, err := tcrypto.ResponderHandshakeINStep1(req.HandshakeInitMsg, state)

	if err != nil {
		return nil, err
	}

	token, err := tcrypto.BuildAuthToken(req.PubKey, time.Hour, this.MacKey[:])

	if err != nil {
		return nil, err
	}

	proof, err := cipher.Enc.Encrypt(nil, nil, token)

	if err != nil {
		return nil, err
	}

	ch := &rpc.AuthChallenge{
		HandshakeDoneMsg: handshakeDoneMsg,
		AuthChallenge:    proof,
	}

	log.Info().Msg("send AuthChallenge")

	return ch, nil
}
