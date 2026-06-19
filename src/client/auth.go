package client

import (
	"context"
	"encoding/base64"
	"tsunagi/src/rpc"

	"github.com/minnowo/tsunagi/mod/tcrypto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"
)

func GetAuthContext(conn *TsunagiConn, ctx context.Context) (context.Context, error) {

	keypair, err := tcrypto.GenerateNoiseKeypair()

	if err != nil {
		return nil, err
	}

	state, err := tcrypto.NewSenderHandshakeIN(keypair)

	if err != nil {
		return nil, err
	}

	msg, err := tcrypto.SenderHandshakeINStep1(state)

	if err != nil {
		return nil, err
	}

	challenge, err := conn.Auth.GetChallenge(ctx, &rpc.AuthRequest{
		PubKey:           keypair.Public,
		HandshakeInitMsg: msg,
	})

	if err != nil {
		return nil, err
	}

	log.Debug().Hex("cipher", challenge.AuthChallenge).Msg("got auth challenge from relay")

	cipher, err := tcrypto.SenderHandshakeINStep2(challenge.HandshakeDoneMsg, state)

	if err != nil {
		return nil, err
	}

	proof, err := cipher.Dec.Decrypt(nil, nil, challenge.AuthChallenge)

	if err != nil {
		return nil, err
	}

	log.Debug().Msg("challenge decrypted")

	authToken, err := conn.Auth.ProveChallenge(ctx, &rpc.AuthProof{
		Signature: proof,
	})

	if err != nil {
		return nil, err
	}

	log.Debug().Hex("token", authToken.Token).Msg("got auth token")

	authCtx := metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("authorization", base64.StdEncoding.EncodeToString(authToken.Token)),
	)

	return authCtx, nil
}
