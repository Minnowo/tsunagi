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
	server.Init(nil)
	server.Register(s)

	log.Info().Str("host", lis.Addr().String()).Msg("listening")

	if err := s.Serve(lis); err != nil {
		return err
	}

	return nil
}
