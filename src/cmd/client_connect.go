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

func splitShell(s string) []string {
	var args []string
	var cur strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, r := range s {
		switch {
		case (r == '"' || r == '\'') && !inQuotes:
			inQuotes = true
			quoteChar = r
		case r == quoteChar && inQuotes:
			inQuotes = false
		case r == ' ' && !inQuotes:
			if cur.Len() > 0 {
				args = append(args, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}

	if cur.Len() > 0 {
		args = append(args, cur.String())
	}
	return args
}

func CmdClientConnect(ctx context.Context, c *cli.Command) error {

	address := c.Value("addr").(string)
	device := c.Value("device").(string)

	var deviceID data.Identifier

	if err := deviceID.FromString(device); err != nil {
		return err
	}

	client := client.New()

	log.Info().Msg("connected - type messages, 'exit' to quit")

	go func() {
		read, exit := client.Read(address)
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
			err = client.ForwardMsg(address, &rpc.ForwardRequest{
				DeviceID:   deviceID[:],
				RelayAddr:  args[1],
				CipherText: []byte(args[2]),
			})

		case "deliver":
			err = client.DeliverMsg(address,
				&rpc.DeliverRequest{
					DeviceID:   deviceID[:],
					CipherText: []byte(args[1]),
				})
		}

		if err != nil {
			log.Error().Err(err).Msg("send failed")
			return err
		}
	}
}
