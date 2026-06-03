package cmd

import (
	"context"
	"fmt"
	"net/http"
	"tsunagi/src/api/clientapi"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

func CmdClientServerMain(ctx context.Context, c *cli.Command) error {

	host := c.Value("host").(string)
	port := c.Value("port").(uint32)

	addr := fmt.Sprintf("%s:%d", host, port)

	r := chi.NewRouter()

	server := clientapi.ClientApi{}
	server.Register(r)

	log.Info().Str("addr", addr).Msg("client api listening")

	return http.ListenAndServe(addr, r)
}
