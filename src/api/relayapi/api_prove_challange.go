package relayapi

import (
	"context"
	"tsunagi/src/rpc"

	"github.com/minnowo/tsunagi/mod/tcrypto"
)

func (this *RelayApi) ProveChallenge(ctx context.Context, proof *rpc.AuthProof) (*rpc.AuthToken, error) {

	_, err := tcrypto.ParseAuthToken(proof.Signature, this.macKey)

	if err != nil {
		return nil, err
	}

	token := &rpc.AuthToken{
		Token: proof.Signature,
	}

	return token, nil
}
