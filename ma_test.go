package ma_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/armon/go-metrics"
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/ma"
)

const RuntimeKey = "github.com/bzimmer/ma#testRuntimeKey"

type Runtime struct {
	*ma.Runtime
	URL string
}

func TestMain(m *testing.M) {
	// hijack the `go test` verbose flag to manage logging
	verbose := flag.CommandLine.Lookup("test.v")
	if verbose.Value.String() != "" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	os.Exit(m.Run())
}

func runtime(c *cli.Context) *Runtime {
	return c.App.Metadata[RuntimeKey].(*Runtime)
}

func copyFile(w io.Writer, filename string) error {
	fp, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fp.Close()
	_, err = io.Copy(w, fp)
	return err
}

func NewTestApp(t *testing.T, tt *harness, cmd *cli.Command, url string) *cli.App {
	return &cli.App{
		Name:     tt.name,
		HelpName: tt.name,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "json",
				Aliases:  []string{"j"},
				Value:    false,
				Required: false,
			},
		},
		Before: func(c *cli.Context) error {
			cfg := metrics.DefaultConfig("ma")
			cfg.EnableRuntimeMetrics = false
			cfg.TimerGranularity = time.Second
			sink := metrics.NewInmemSink(time.Hour*24, time.Hour*24)
			metric, err := metrics.New(cfg, sink)
			if err != nil {
				t.Error(err)
			}

			client, err := smugmug.NewClient(
				smugmug.WithBaseURL(url),
				smugmug.WithHTTPTracing(zerolog.GlobalLevel() == zerolog.DebugLevel))
			if err != nil {
				t.Error(err)
			}

			var enc ma.Encoder
			switch {
			case c.Bool("json"):
				enc = ma.NewJSONEncoder(json.NewEncoder(c.App.Writer))
			default:
				enc = ma.NewBlackholeEncoder()
			}

			rt := &ma.Runtime{
				Client:  client,
				Metrics: metric,
				Sink:    sink,
				Encoder: enc,
				Grab:    new(http.Client),
				Fs:      afero.NewMemMapFs(),
			}
			c.App.Metadata = map[string]interface{}{
				ma.RuntimeKey: rt,
				RuntimeKey: &Runtime{
					Runtime: rt,
					URL:     url,
				},
			}
			return nil
		},
		After: func(c *cli.Context) error {
			t.Logf("***** %s *****\n", tt.name)
			switch v := runtime(c).Fs.(type) {
			case *afero.MemMapFs:
				v.List()
			default:
			}
			if err := ma.Stats(c); err != nil {
				return err
			}
			return counters(t, tt.counters)(c)
		},
		Commands: []*cli.Command{cmd},
	}
}

func counters(t *testing.T, expected map[string]int) cli.AfterFunc {
	a := assert.New(t)
	return func(c *cli.Context) error {
		data := runtime(c).Sink.Data()
		for key, value := range expected {
			var found bool
			for i := range data {
				if counter, ok := data[i].Counters[key]; ok {
					found = true
					a.Equalf(value, counter.Count, key)
					break
				}
			}
			if !found {
				return fmt.Errorf("cannot find sample value for {%s}", key)
			}
		}
		return nil
	}
}

type harness struct {
	name, err string
	args      []string
	counters  map[string]int
	before    cli.BeforeFunc
	after     cli.AfterFunc
}

func run(t *testing.T, tt *harness, handler http.Handler, cmd func() *cli.Command) {
	a := assert.New(t)

	svr := httptest.NewServer(handler)
	defer svr.Close()

	app := NewTestApp(t, tt, cmd(), svr.URL)

	if tt.before != nil {
		f := app.Before
		app.Before = func(c *cli.Context) error {
			for _, f := range []cli.BeforeFunc{f, tt.before} {
				if f != nil {
					if err := f(c); err != nil {
						return err
					}
				}
			}
			return nil
		}
	}
	if tt.after != nil {
		f := app.After
		app.After = func(c *cli.Context) error {
			for _, f := range []cli.AfterFunc{f, tt.after} {
				if f != nil {
					if err := f(c); err != nil {
						return err
					}
				}
			}
			return nil
		}
	}

	err := app.RunContext(context.Background(), tt.args)
	switch tt.err == "" {
	case true:
		a.NoError(err)
	case false:
		a.Error(err)
		a.Contains(err.Error(), tt.err)
	}
}
