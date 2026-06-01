package cmd

import (
	"context"
	"fmt"
	"net"
	"tsunagi/src/api/relayapi"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
)

func CmdServerMain(ctx context.Context, c *cli.Command) error {

	host := c.Value("host").(string)
	port := c.Value("port").(uint32)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))

	if err != nil {
		return err
	}

	s := grpc.NewServer()

	server := relayapi.RelayApi{}
	server.Init()
	server.Register(s)

	log.Info().Str("host", lis.Addr().String()).Msg("listening")

	if err := s.Serve(lis); err != nil {
		return err
	}

	// // Each node is a symmetric peer in the network.
	// // Every node runs both an HTTP server (to receive requests)
	// // and an HTTP client (to send requests to other peers).
	// // NOTE: v1 is only http client/server, this may change to support more protocols in the future.

	// // Get known peers.
	// // A peer consists of:
	// //  - an identity (static long-term public key)
	// //  - ---DEPRECATED----------------[a list of network addresses (Yggdrasil IPv6)]-----------------
	// // This includes the responding node.
	// // If no peers are known, the node returns itself as a peer
	// // if it is accepting incoming connections.
	// r.Get("/peers", tmp)

	// // Perform a Noise handshake over HTTP.
	// // This establishes a shared symmetric session key using
	// // Diffie-Hellman between static and ephemeral keys.
	// // The result is a session ID that maps to the derived session state
	// // on both client and server.
	// //
	// // The session ID is NOT cryptographic; it is only a reference
	// // to the session state stored locally.
	// //
	// // Both sides independently derive the same session keys;
	// // no secret key is transmitted over the network.
	// r.Post("/handshake", tmp)

	// // Attempt receive encrypted messages for a session.
	// //
	// // Each message is encrypted using the session's symmetric key.
	// // A sequence number (nonce/counter) is included for replay protection
	// // and ordering.
	// //
	// // The server uses (session_id + sequence number) to locate the
	// // correct session state and decrypt the payload.
	// //
	// // The payload must contain a message_id for deduplication and responds with ACK if everything goes well. The caller can consider the message delivered on this ACK, but otherwise should keep trying.
	// //
	// // Example decrypted message:
	// // {
	// //   "message_id": "",
	// //   "seq": "",
	// //   "sender": "sender-identity-key",
	// //   "message": "hello bob",
	// //   "time": 1231232
	// // }
	// // Responds with ACK:
	// // {
	// //   "ok": true,
	// //   "message_id": ""
	// // }
	// r.Post("/message", tmp)

	// // client gets messages for itself
	// r.Post("/inbox", tmp)

	// r.Get("/", tmp)

	// http.ListenAndServe(":3000", r)

	return nil
}
