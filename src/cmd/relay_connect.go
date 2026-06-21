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

func CmdRelayConnect(ctx context.Context, c *cli.Command) error {

	address := c.Value("addr").(string)

	identity, err := tcrypto.GenerateNoiseKeypair()
	if err != nil {
		return err
	}

	pubB64 := base64.StdEncoding.EncodeToString(identity.Public)
	fmt.Printf("your public key: %s\n", pubB64)
	fmt.Printf("connecting to   %s\n", address)
	fmt.Println("commands: deliver <dest-pub-b64> <message...>  |  whoami  |  exit  |  help")

	cl := client.NewRelayRelayClient(identity, 64, 64)

	var senderID data.Identifier
	if err := senderID.FromBytes(identity.Public); err != nil {
		return err
	}

	go func() {
		ackChan := cl.ReadAck()
		for {
			select {
			case ack, ok := <-ackChan:
				if !ok {
					return
				}
				fmt.Printf("[ack] messageID=%d\n", ack.MessageID)
			case <-ctx.Done():
				return
			}
		}
	}()

	msgID := uint64(0)
	scanner := bufio.NewScanner(os.Stdin)

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

		parts := strings.SplitN(line, " ", 3)
		cmd := parts[0]

		switch cmd {

		case "exit", "quit":
			return nil

		case "whoami":
			fmt.Printf("pub: %s\n", pubB64)

		case "help":
			fmt.Println("  deliver <dest-pub-b64> <message>   deliver a message payload to a client")
			fmt.Println("  whoami                             show your public key")
			fmt.Println("  exit                               disconnect and quit")

		case "deliver":
			if len(parts) < 3 {
				fmt.Println("usage: deliver <dest-pub-b64> <message>")
				continue
			}
			var dest data.Identifier
			if err := dest.FromString(parts[1]); err != nil {
				fmt.Printf("bad dest key: %v\n", err)
				continue
			}
			msgID++
			msg := parts[2]
			err = cl.Send(address, senderID, &rpc.RelayEvent{
				Body: &rpc.RelayEvent_MessagePayload{
					MessagePayload: &rpc.MessagePayload{
						MessageID:       msgID,
						DeliverToPubKey: dest[:],
						CipherText:      []byte(msg),
					},
				},
			})
			if err != nil {
				fmt.Printf("send error: %v\n", err)
				log.Error().Err(err).Msg("deliver failed")
			} else {
				fmt.Printf("sent (msgID=%d)\n", msgID)
			}

		default:
			fmt.Printf("unknown command %q - type 'help'\n", cmd)
		}
	}
}
