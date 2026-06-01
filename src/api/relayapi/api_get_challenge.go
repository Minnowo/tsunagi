package relayapi

import (
	"context"
	"crypto/rand"
	"tsunagi/src/rpc"
)

func (this *RelayApi) GetChallenge(ctx context.Context, req *rpc.AuthRequest) (*rpc.AuthChallenge, error) {

	var nonce [64]byte

	rand.Read(nonce[:])

	ch := &rpc.AuthChallenge{
		Nonce: nonce[:],
	}

	return ch, nil
}
