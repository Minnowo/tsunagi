package cmd

import (
	"bufio"
	"context"
	"os"
	"strings"
	"tsunagi/src/client"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

func CmdRelayConnect(ctx context.Context, c *cli.Command) error {

	address := c.Value("addr").(string)

	client := client.NewRelayRelayClient(0)

	log.Info().Msg("connected - type messages, 'exit' to quit")

	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		text, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		text = strings.TrimSpace(text)

		if text == "" {
			continue
		}

		args := splitShell(text)

		switch args[0] {
		case "exit":
			log.Info().Msg("closing")
			return nil

		case "deliver":

			var other data.Identifier
			err = other.FromString(args[1])
			if err != nil {
				break
			}
			err = client.Send(address, &rpc.RelayEvent{
				PubKey: other[:],
				Body: &rpc.RelayEvent_MessagePayload{
					MessagePayload: &rpc.MessagePayload{
						CipherText: []byte(args[2]),
					},
				},
			})

		}

		if err != nil {
			log.Error().Err(err).Msg("send failed")
			return err
		}
	}
}
