package internal

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/bzimmer/httpwares"
	"github.com/bzimmer/ma"
	"github.com/bzimmer/manual"
	"github.com/bzimmer/smugmug"
	"github.com/hashicorp/go-metrics"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/language"
)

func flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "smugmug-client-key",
			Usage:   "smugmug client key",
			EnvVars: []string{"SMUGMUG_CLIENT_KEY"},
		},
		&cli.StringFlag{
			Name:    "smugmug-client-secret",
			Usage:   "smugmug client secret",
			EnvVars: []string{"SMUGMUG_CLIENT_SECRET"},
		},
		&cli.StringFlag{
			Name:    "smugmug-access-token",
			Usage:   "smugmug access token",
			EnvVars: []string{"SMUGMUG_ACCESS_TOKEN"},
		},
		&cli.StringFlag{
			Name:    "smugmug-token-secret",
			Usage:   "smugmug token secret",
			EnvVars: []string{"SMUGMUG_TOKEN_SECRET"},
		},
		&cli.BoolFlag{
			Name:     "json",
			Aliases:  []string{"j"},
			Usage:    "emit all results as JSON and print to stdout",
			Value:    false,
			Required: false,
		},
		&cli.BoolFlag{
			Name:     "monochrome",
			Required: false,
			Usage:    "disable colored loggingoutput",
			Value:    false,
		},
		&cli.BoolFlag{
			Name:     "debug",
			Required: false,
			Usage:    "enable verbose debugging",
			Value:    false,
		},
		&cli.BoolFlag{
			Name:     "trace",
			Required: false,
			Usage:    "enable debugging of http requests",
			Value:    false,
		},
	}
}

func initLogging(c *cli.Context) error {
	level := zerolog.InfoLevel
	if c.Bool("debug") {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)
	zerolog.DurationFieldUnit = time.Millisecond //nolint:reassign // configuration
	zerolog.DurationFieldInteger = false         //nolint:reassign // configuration
	log.Logger = log.Output(                     //nolint:reassign // configuration
		zerolog.ConsoleWriter{
			Out:        c.App.ErrWriter,
			NoColor:    c.Bool("monochrome"),
			TimeFormat: time.RFC3339,
		},
	)
	return nil
}

func mg(c *cli.Context) func() *smugmug.Client {
	return func() *smugmug.Client {
		httpclient, err := smugmug.NewHTTPClient(
			c.String("smugmug-client-key"),
			c.String("smugmug-client-secret"),
			c.String("smugmug-access-token"),
			c.String("smugmug-token-secret"))
		if err != nil {
			panic(err)
		}
		client, err := smugmug.NewClient(
			smugmug.WithConcurrency(c.Int("concurrency")),
			smugmug.WithHTTPClient(httpclient),
			smugmug.WithPretty(c.Bool("trace")),
			smugmug.WithHTTPTracing(c.Bool("trace")))
		if err != nil {
			panic(err)
		}
		return client
	}
}

func App() *cli.App {
	return &cli.App{
		Name:        "ma",
		HelpName:    "ma",
		Usage:       "CLI for managing local and Smugmug-hosted photos",
		Description: "CLI for managing local and Smugmug-hosted photos",
		Flags:       flags(),
		Before: func(c *cli.Context) error {
			if err := initLogging(c); err != nil {
				return err
			}

			cfg := metrics.DefaultConfig(c.App.Name)
			cfg.EnableRuntimeMetrics = false
			cfg.TimerGranularity = time.Second
			sink := metrics.NewInmemSink(time.Hour*24, time.Hour*24)
			metric, err := metrics.New(cfg, sink)
			if err != nil {
				return err
			}

			grab := &http.Client{}
			if c.Bool("trace") {
				grab.Transport = &httpwares.VerboseTransport{}
			}

			writer := io.Discard
			if c.Bool("json") {
				writer = c.App.Writer
			}

			c.App.Metadata = map[string]any{
				ma.RuntimeKey: &ma.Runtime{
					Smugmug:  mg(c),
					Sink:     sink,
					Grab:     grab,
					Metrics:  metric,
					Encoder:  json.NewEncoder(writer),
					Fs:       afero.NewOsFs(),
					Language: language.English,
					Start:    time.Now(),
				},
			}

			return nil
		},
		After: ma.Metrics,
		Commands: []*cli.Command{
			ma.CommandExport(),
			ma.CommandFind(),
			ma.CommandList(),
			ma.CommandNew(),
			ma.CommandPatch(),
			ma.CommandRemove(),
			ma.CommandSimilar(),
			ma.CommandTitle(),
			ma.CommandUpload(),
			ma.CommandURLName(),
			ma.CommandUser(),
			ma.CommandVersion(),
			manual.EnvVars(),
			manual.Manual(),
		},
	}
}
