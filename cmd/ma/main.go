package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/armon/go-metrics"
	"github.com/bzimmer/httpwares"
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/ma"
)

type encoderBlackhole struct{}

func (e *encoderBlackhole) Encode(_ interface{}) error {
	return nil
}

type encoderJSON struct {
	encoder *json.Encoder
}

func (e *encoderJSON) Encode(v interface{}) error {
	return e.encoder.Encode(v)
}

func flags() []cli.Flag {
	return []cli.Flag{
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
		&cli.IntFlag{
			Name:     "concurrency",
			Value:    2,
			Required: false,
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
	}
}

func initLogging(c *cli.Context) {
	level := zerolog.InfoLevel
	if c.Bool("debug") {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = false
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:        c.App.ErrWriter,
			NoColor:    false,
			TimeFormat: time.RFC3339,
		},
	)
}

func main() {
	app := &cli.App{
		Name:     "ma",
		HelpName: "ma",
		Usage:    "CLI for managing photos locally and at SmugMug",
		Flags:    flags(),
		ExitErrHandler: func(c *cli.Context, err error) {
			if err == nil {
				return
			}
			log.Error().Stack().Err(err).Msg(c.App.Name)
		},
		Before: func(c *cli.Context) error {
			initLogging(c)

			cfg := metrics.DefaultConfig("ma")
			cfg.EnableRuntimeMetrics = false
			cfg.TimerGranularity = time.Second
			sink := metrics.NewInmemSink(time.Hour*24, time.Hour*24)
			metric, err := metrics.New(cfg, sink)
			if err != nil {
				return err
			}

			grab := &http.Client{}
			if c.Bool("debug") {
				grab.Transport = &httpwares.VerboseTransport{}
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
				smugmug.WithConcurrency(c.Int("concurrency")),
				smugmug.WithHTTPClient(httpclient),
				smugmug.WithPretty(c.Bool("debug")),
				smugmug.WithHTTPTracing(c.Bool("debug")))
			if err != nil {
				return err
			}

			var enc ma.Encoder
			switch {
			case c.Bool("json"):
				enc = &encoderJSON{encoder: json.NewEncoder(c.App.Writer)}
			default:
				enc = &encoderBlackhole{}
			}

			c.App.Metadata = map[string]interface{}{
				ma.RuntimeKey: &ma.Runtime{
					Client:  client,
					Sink:    sink,
					Grab:    grab,
					Metrics: metric,
					Encoder: enc,
					Fs:      afero.NewOsFs(),
				},
			}

			return nil
		},
		After: ma.Stats,
		Commands: []*cli.Command{
			ma.CommandUser(),
			ma.CommandFind(),
			ma.CommandList(),
			ma.CommandNew(),
			ma.CommandPatch(),
			ma.CommandUpload(),
			ma.CommandCopy(),
			ma.CommandExport(),
			ma.CommandVersion(),
		},
	}
	if err := app.RunContext(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
