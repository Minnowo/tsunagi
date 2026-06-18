package relayapi

import (
	"context"
	"tsunagi/src/rpc"

	"github.com/minnowo/tsunagi/mod/tcrypto"
)

func (this *RelayApi) GetChallenge(ctx context.Context, req *rpc.AuthRequest) (*rpc.AuthChallenge, error) {

	state, err := tcrypto.NewResponderAuthHandshakeState()

	if err != nil {
		return nil, err
	}

	_, _, _, err = state.ReadMessage(nil, req.HandshakeInitMsg)

	if err != nil {
		return nil, err
	}

	msg, enc, _, err := state.WriteMessage(nil, nil)

	if err != nil {
		return nil, err
	}

	proof, err := enc.Encrypt(nil, nil, []byte{1, 2, 3})

	if err != nil {
		return nil, err
	}

	ch := &rpc.AuthChallenge{
		HandshakeDoneMsg: msg,
		AuthChallenge: proof,
	}

	return ch, nil
}
