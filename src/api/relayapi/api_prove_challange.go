package relayapi

import (
	"context"
	"tsunagi/src/rpc"
)

func (this *RelayApi) ProveChallenge(ctx context.Context, proof *rpc.AuthProof) (*rpc.AuthToken, error) {

	token := &rpc.AuthToken{
		Token: proof.Signature,
	}

	return token, nil
}
