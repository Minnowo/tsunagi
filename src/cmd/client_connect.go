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

func CmdClientConnect(ctx context.Context, c *cli.Command) error {

	address := c.Value("addr").(string)

	var deviceID data.Identifier
	deviceID.GenNew()
	log.Info().Str("id", deviceID.String()).Msg("tmp device")

	client := client.NewClientRelayClient(0)

	log.Info().Msg("connected - type messages, 'exit' to quit")

	go func() {
		read, exit, err := client.GetReadHandle(address)

		if err != nil {
			log.Panic().Err(err).Msg("could not get read handle on stream")
		}

		for {
			select {
			case event := <-read:
				log.Info().Interface("event", event).Msg("read event")
			case <-exit:
				log.Info().Msg("client shutdown")
				return
			}
		}
	}()

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

		case "forward":

			var other data.Identifier
			err = other.FromString(args[1])
			if err != nil {
				break
			}
			err = client.Send(address, &rpc.ClientEvent{
				DeviceID:  other[:],
				RelayAddr: args[2],
				Body: &rpc.ClientEvent_MessagePayload{
					MessagePayload: &rpc.MessagePayload{
						CipherText: []byte(args[3]),
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
