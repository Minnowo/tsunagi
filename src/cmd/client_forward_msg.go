package cmd

import (
	"context"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func CmdClientForwardMsg(ctx context.Context, c *cli.Command) error {

	address1 := c.Value("addr1").(string)
	address2 := c.Value("addr2").(string)
	device := c.Value("device").(string)
	msg := c.Value("msg").(string)

	var deviceID data.Identifier

	if err := deviceID.FromString(device); err != nil {
		return err
	}

	conn, err := grpc.NewClient(
		address1,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return err
	}

	defer conn.Close()

	client := rpc.NewTsunagiClient(conn)

	_, err = client.ForwardMessage(ctx, &rpc.ForwardRequest{
		RelayAddr:  address2,
		DeviceID:   deviceID[:],
		CipherText: []byte(msg),
	})

	if err != nil {
		return err
	}

	log.Info().
		Str("sent2", address1).
		Str("forwarded2", address2).
		Msg("forwarded message")

	return nil
}
