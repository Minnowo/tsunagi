package cmd

import (
	"context"
	"fmt"
	"net/http"
	"tsunagi/src/api/clientapi"
	"tsunagi/src/database/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

func CmdClientServerMain(ctx context.Context, c *cli.Command) error {

	host := c.Value("host").(string)
	port := c.Value("port").(uint32)

	db, err := sqlite.OpenSqliteDatabase(ctx, "/tmp/tsunagi_tmp.sqlite3")

	if err != nil {
		return err
	}

	err = db.Migrate(ctx)

	if err != nil {
		return err
	}

	r := chi.NewRouter()

	server := clientapi.ClientApi{
		DB: db,
	}
	server.Register(r)

	addr := fmt.Sprintf("%s:%d", host, port)
	log.Info().Str("addr", addr).Msg("client api listening")

	return http.ListenAndServe(addr, r)
}
