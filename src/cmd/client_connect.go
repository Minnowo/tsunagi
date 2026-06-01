package cmd

import (
	"context"
	"io"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func CmdClientConnect(ctx context.Context, c *cli.Command) error {

	address := c.Value("addr").(string)
	device := c.Value("device").(string)

	var deviceID data.Identifier

	if err := deviceID.FromString(device); err != nil {
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

	req := rpc.ConnectRequest{
		DeviceID: deviceID[:],
	}

	stream, err := client.Connect(context.Background(), &req)

	if err != nil {
		return err
	}

	for {
		evt, err := stream.Recv()

		if err == io.EOF {
			log.Info().Msg("server gave EOF")
			break
		}

		if err != nil {
			return err
		}

		log.Info().Bytes("msg", evt.GetCipherText()).Msg("got message")
	}

	return nil
}
