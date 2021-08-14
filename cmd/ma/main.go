package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/armon/go-metrics"
	"github.com/bzimmer/ma"
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:     "ma",
		HelpName: "ma",
		Usage:    "CLI for managing photos locally and at SmugMug",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "smugmug-client-key",
				Required: true,
				Usage:    "smugmug client key",
				EnvVars:  []string{"SMUGMUG_CLIENT_KEY"},
			},
			&cli.StringFlag{
				Name:     "smugmug-client-secret",
				Required: true,
				Usage:    "smugmug client secret",
				EnvVars:  []string{"SMUGMUG_CLIENT_SECRET"},
			},
			&cli.StringFlag{
				Name:     "smugmug-access-token",
				Required: true,
				Usage:    "smugmug access token",
				EnvVars:  []string{"SMUGMUG_ACCESS_TOKEN"},
			},
			&cli.StringFlag{
				Name:     "smugmug-token-secret",
				Required: true,
				Usage:    "smugmug token secret",
				EnvVars:  []string{"SMUGMUG_TOKEN_SECRET"},
			},
			&cli.BoolFlag{
				Name:     "json",
				Aliases:  []string{"j"},
				Value:    false,
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "debug",
				Required: false,
				Usage:    "enable debugging",
				Value:    false,
			},
		},
		ExitErrHandler: func(c *cli.Context, err error) {
			if err == nil {
				return
			}
			log.Error().Err(err).Msg(c.App.Name)
		},
		Before: func(c *cli.Context) error {
			level := zerolog.InfoLevel
			if c.Bool("debug") {
				level = zerolog.DebugLevel
			}
			zerolog.SetGlobalLevel(level)
			zerolog.DurationFieldUnit = time.Millisecond
			zerolog.DurationFieldInteger = false
			log.Logger = log.Output(
				zerolog.ConsoleWriter{
					Out:        c.App.ErrWriter,
					NoColor:    false,
					TimeFormat: time.RFC3339,
				},
			)

			cfg := metrics.DefaultConfig("ma")
			cfg.EnableRuntimeMetrics = false
			cfg.TimerGranularity = time.Second
			sink := metrics.NewInmemSink(time.Hour*24, time.Hour*24)
			metric, err := metrics.New(cfg, sink)
			if err != nil {
				return err
			}

			httpclient, err := smugmug.NewHTTPClient(
				c.String("smugmug-client-key"),
				c.String("smugmug-client-secret"),
				c.String("smugmug-access-token"),
				c.String("smugmug-token-secret"))
			if err != nil {
				return err
			}

			client, err := smugmug.NewClient(
				smugmug.WithMetrics(metric),
				smugmug.WithHTTPClient(httpclient),
				smugmug.WithPretty(c.Bool("debug")),
				smugmug.WithHTTPTracing(c.Bool("debug")))
			if err != nil {
				return err
			}

			var enc ma.Encoder
			switch {
			case c.Bool("json"):
				enc = &ma.EncoderJSON{Encoder: json.NewEncoder(c.App.Writer)}
			default:
				enc = &ma.EncoderLog{}
			}

			c.App.Metadata = map[string]interface{}{
				"client":  client,
				"metrics": metric,
				"sink":    sink,
				"encoder": enc,
			}

			return nil
		},
		Commands: []*cli.Command{
			ma.CommandUser(),
			ma.CommandFind(),
			ma.CommandList(),
			ma.CommandUp(),
			ma.CommandCopy(),
			ma.CommandExport(),
		},
	}
	if err := app.RunContext(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
