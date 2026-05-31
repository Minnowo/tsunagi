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

func CmdClientDeliverMsg(ctx context.Context, c *cli.Command) error {

	address := c.Value("addr").(string)
	device := c.Value("device").(string)
	msg := c.Value("msg").(string)
	println(address)

	var deviceID data.Identifier

	if err := deviceID.FromString(device); err!=nil{
		return err
	}

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return err
	}
    
	defer conn.Close()

    client := rpc.NewTsunagiClient(conn)

	_, err = client.DeliverMessage(ctx, &rpc.DeliverRequest{
		DeviceID:   deviceID[:],
		CipherText: []byte(msg),
	})

	if err != nil {
		return err
	}

	log.Info().
		Str("addr", address).
		Msg("delivered message via gRPC")

	return nil
}
