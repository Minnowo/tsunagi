package main

import (
	"context"
	"os"
	"time"
	"tsunagi/src/cmd"

	"github.com/minnowo/log4zero"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

func main() {

	log4zero.Init("./log-config.json")
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	cmd := &cli.Command{
		Name:  "tsunagi",
		Usage: "pulsar router",
		Commands: []*cli.Command{
			{
				Name:        "session",
				Usage:       "Session related commands",
				Description: "",
				Commands: []*cli.Command{
					{
						Name:        "init",
						Usage:       "Create a new identity",
						Description: "Create a new identity",
						Action:      cmd.CmdSessionInit,
					},
				},
			},
			{
				Name:        "run",
				Usage:       "Run the server",
				Description: "",
				Action:      cmd.CmdServerMain,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "host",
						Aliases:  []string{"b"},
						Usage:    "Bind to this host",
						Value:    "0.0.0.0",
						Required: false,
					},
					&cli.Uint32Flag{
						Name:     "grpc-port",
						Aliases:  []string{"p"},
						Usage:    "Bind grpc to this port",
						Value:    7471,
						Required: false,
					},
					&cli.Uint32Flag{
						Name:     "http-port",
						Aliases:  []string{"P"},
						Usage:    "Bind http to this port",
						Value:    7470,
						Required: false,
					},
				},
			},
			{
				Name:        "client",
				Usage:       "Commandline client",
				Description: "",
				Commands: []*cli.Command{
					{
						Name:   "connect",
						Usage:  "Connect to recieve updates",
						Action: cmd.CmdClientConnect,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "addr",
								Usage: "The remote address (e.g. tcp://localhost:7471/)",
								Value: "localhost:7471",
							},
						},
					},
				},
			},
			{
				Name:        "relay",
				Usage:       "Commandline client",
				Description: "",
				Commands: []*cli.Command{
					{
						Name:   "connect",
						Usage:  "Connect to recieve updates",
						Action: cmd.CmdRelayConnect,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "addr",
								Usage: "The remote address (e.g. tcp://localhost:7471/)",
								Value: "localhost:7471",
							},
						},
					},
				},
			},
			{
				Name:        "tui",
				Usage:       "Terminal UI client",
				Description: "",
				Commands: []*cli.Command{
					{
						Name:   "client",
						Usage:  "Launch the TUI client",
						Action: cmd.CmdTuiClient,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "addr",
								Usage: "The remote address (e.g. localhost:7470)",
								Value: "localhost:7471",
							},
						},
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Error().Err(err).Msg("")
	}
}
