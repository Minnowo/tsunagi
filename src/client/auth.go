package client

import (
	"context"
	"encoding/hex"
	"tsunagi/src/rpc"

	"google.golang.org/grpc/metadata"
)

func GetAuthContext(conn *TsunagiConn, ctx context.Context) (context.Context, error) {

	challenge, err := conn.Auth.GetChallenge(ctx, &rpc.AuthRequest{})

	if err != nil {
		return nil, err
	}

	authToken, err := conn.Auth.ProveChallenge(ctx, &rpc.AuthProof{
		Signature: challenge.Nonce,
	})

	if err != nil {
		return nil, err
	}

	authCtx := metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("authorization", "Bearer "+hex.EncodeToString(authToken.GetToken())),
	)

	return authCtx, nil
}
