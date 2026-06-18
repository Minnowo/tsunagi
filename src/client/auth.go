package client

import (
	"context"
	"encoding/hex"
	"tsunagi/src/rpc"

	"github.com/minnowo/tsunagi/mod/tcrypto"
	"google.golang.org/grpc/metadata"
)

func GetAuthContext(conn *TsunagiConn, ctx context.Context) (context.Context, error) {

	keypair, err := tcrypto.GenerateNoiseKeypair()

	if err != nil {
		return nil, err
	}

	state, err := tcrypto.NewSenderAuthHandshakeState(keypair)

	if err != nil {
		return nil, err
	}

	msg, _ ,_, err := state.WriteMessage(nil, nil)

	if err != nil {
		return nil, err
	}

	challenge, err := conn.Auth.GetChallenge(ctx, &rpc.AuthRequest{
		DeviceID: keypair.Public,
		PubKey: keypair.Public,
		HandshakeInitMsg: msg,
	})

	if err != nil {
		return nil, err
	}

	_, dec, _, err := state.ReadMessage(nil, challenge.HandshakeDoneMsg)

	if err != nil {
		return nil, err
	}

	proof, err := dec.Decrypt(nil, nil, challenge.AuthChallenge)

	if err != nil {
		return nil, err
	}

	authToken, err := conn.Auth.ProveChallenge(ctx, &rpc.AuthProof{
		Signature: proof,
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
