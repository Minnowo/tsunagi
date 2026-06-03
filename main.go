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
						Name:     "port",
						Aliases:  []string{"p"},
						Usage:    "Bind to this port",
						Value:    7471,
						Required: false,
					},
				},
			},
			{
				Name:        "client-server",
				Usage:       "Run the client WebSocket server",
				Description: "",
				Action:      cmd.CmdClientServerMain,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "host",
						Aliases:  []string{"b"},
						Usage:    "Bind to this host",
						Value:    "0.0.0.0",
						Required: false,
					},
					&cli.Uint32Flag{
						Name:     "port",
						Aliases:  []string{"p"},
						Usage:    "Bind to this port",
						Value:    8080,
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
								Name:     "addr",
								Usage:    "The remote address (e.g. tcp://localhost:7471/)",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "device",
								Usage:    "The target device ID",
								Value:    "0000000000000000000000000000000000000000000",
								Required: false,
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
								Name:     "addr",
								Usage:    "The remote address (e.g. tcp://localhost:7471/)",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "device",
								Usage:    "The target device ID",
								Value:    "0000000000000000000000000000000000000000000",
								Required: false,
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
