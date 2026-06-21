package cmd

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"tsunagi/src/client"
	"tsunagi/src/data"
	"tsunagi/src/rpc"

	"github.com/minnowo/tsunagi/mod/tcrypto"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

func CmdClientConnect(ctx context.Context, c *cli.Command) error {

	address := c.Value("addr").(string)

	identity, err := tcrypto.GenerateNoiseKeypair()
	if err != nil {
		return err
	}

	pubB64 := base64.StdEncoding.EncodeToString(identity.Public)
	fmt.Printf("your public key: %s\n", pubB64)
	fmt.Printf("connecting to   %s\n", address)
	fmt.Println("commands: send <dest-pub-b64> <relay-addr> <message...>  |  whoami  |  exit  |  help")

	cl := client.NewClientRelayClient(identity, 0)

	read, exit, err := cl.GetReadHandle(address)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-read:
				if !ok {
					return
				}
				switch v := event.Body.(type) {
				case *rpc.RelayEvent_MessagePayload:
					fmt.Printf("[msg from %s]: %s\n",
						base64.StdEncoding.EncodeToString(v.MessagePayload.DeliverToPubKey),
						string(v.MessagePayload.CipherText))
				case *rpc.RelayEvent_NoiseHandshake:
					fmt.Printf("[noise from %s]: %d bytes\n",
						base64.StdEncoding.EncodeToString(v.NoiseHandshake.DeliverToPubKey),
						len(v.NoiseHandshake.HandshakeMsg))
				case *rpc.RelayEvent_RelayAck:
					fmt.Printf("[ack] messageID=%d\n", v.RelayAck.MessageID)
				default:
					log.Info().Interface("event", event).Msg("recv")
				}
			case <-exit:
				fmt.Println("[disconnected]")
				return
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	msgID := uint64(0)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		fmt.Print("> ")
		if !scanner.Scan() {
			return scanner.Err()
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 4)
		cmd := parts[0]

		switch cmd {

		case "exit", "quit":
			return nil

		case "whoami":
			fmt.Printf("pub: %s\n", pubB64)

		case "help":
			fmt.Println("  send <dest-pub-b64> <relay-addr> <message>   forward a message payload")
			fmt.Println("  whoami                                        show your public key")
			fmt.Println("  exit                                          disconnect and quit")

		case "send":
			if len(parts) < 4 {
				fmt.Println("usage: send <dest-pub-b64> <relay-addr> <message>")
				continue
			}
			var dest data.Identifier
			if err := dest.FromString(parts[1]); err != nil {
				fmt.Printf("bad dest key: %v\n", err)
				continue
			}
			relayAddr := parts[2]
			msg := parts[3]
			msgID++
			err = cl.Send(address, &rpc.ClientEvent{
				RelayAddr: relayAddr,
				Body: &rpc.ClientEvent_MessagePayload{
					MessagePayload: &rpc.MessagePayload{
						MessageID:       msgID,
						DeliverToPubKey: dest[:],
						CipherText:      []byte(msg),
					},
				},
			})
			if err != nil {
				fmt.Printf("send error: %v\n", err)
			} else {
				fmt.Println("sent")
			}

		default:
			fmt.Printf("unknown command %q - type 'help'\n", cmd)
		}
	}
}
