package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/flynn/noise"
	"github.com/urfave/cli/v3"
)


func CmdSessionInit(ctx context.Context, c *cli.Command) error {

	keypair, err := noise.DH25519.GenerateKeypair(rand.Reader)

	if err != nil {return err }

	fmt.Println("private:", hex.EncodeToString(keypair.Private))
	fmt.Println("public: ", hex.EncodeToString(keypair.Public))

	return nil
}
