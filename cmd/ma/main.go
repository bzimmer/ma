package main

//go:generate go run main.go manual -o ../../docs/commands.md ../../docs/commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/armon/go-metrics"
	"github.com/bzimmer/httpwares"
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/language"

	"github.com/bzimmer/manual"

	"github.com/bzimmer/ma"
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
			Usage:    "disable colored output",
			Value:    false,
		},
		&cli.BoolFlag{
			Name:     "debug",
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
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = false
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:        c.App.ErrWriter,
			NoColor:    c.Bool("monochrome"),
			TimeFormat: time.RFC3339,
		},
	)
	return nil
}

type mg struct {
	ctx *cli.Context
}

func (c *mg) Client() *smugmug.Client {
	httpclient, err := smugmug.NewHTTPClient(
		c.ctx.String("smugmug-client-key"),
		c.ctx.String("smugmug-client-secret"),
		c.ctx.String("smugmug-access-token"),
		c.ctx.String("smugmug-token-secret"))
	if err != nil {
		panic(err)
	}
	client, err := smugmug.NewClient(
		smugmug.WithConcurrency(c.ctx.Int("concurrency")),
		smugmug.WithHTTPClient(httpclient),
		smugmug.WithPretty(c.ctx.Bool("debug")),
		smugmug.WithHTTPTracing(c.ctx.Bool("debug")))
	if err != nil {
		panic(err)
	}
	return client
}

func main() {
	app := &cli.App{
		Name:        "ma",
		HelpName:    "ma",
		Usage:       "CLI for managing local and Smugmug-hosted photos",
		Description: "CLI for managing local and Smugmug-hosted photos",
		Flags:       flags(),
		ExitErrHandler: func(c *cli.Context, err error) {
			if err == nil {
				return
			}
			log.Error().Stack().Err(err).Msg(c.App.Name)
		},
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
			if c.Bool("debug") {
				grab.Transport = &httpwares.VerboseTransport{}
			}

			writer := io.Discard
			if c.Bool("json") {
				writer = c.App.Writer
			}

			c.App.Metadata = map[string]any{
				ma.RuntimeKey: &ma.Runtime{
					Smugmug:  &mg{c},
					Sink:     sink,
					Grab:     grab,
					Metrics:  metric,
					Encoder:  json.NewEncoder(writer),
					Fs:       afero.NewOsFs(),
					Exif:     ma.NewGoExif(),
					Language: language.English,
					Start:    time.Now(),
				},
			}

			return nil
		},
		After: ma.Metrics,
		Commands: []*cli.Command{
			ma.CommandCopy(),
			ma.CommandExif(),
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
	var err error
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				log.Error().Err(v).Msg(app.Name)
			case string:
				log.Error().Err(errors.New(v)).Msg(app.Name)
			default:
				log.Error().Err(fmt.Errorf("%v", v)).Msg(app.Name)
			}
			os.Exit(1)
		}
		if err != nil {
			log.Error().Err(err).Msg(app.Name)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	err = app.RunContext(context.Background(), os.Args)
}
