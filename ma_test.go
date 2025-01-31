package ma_test

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/language"

	"github.com/bzimmer/smugmug"

	"github.com/bzimmer/ma"
)

const RuntimeKey = "github.com/bzimmer/ma#testRuntimeKey"

type Runtime struct {
	*ma.Runtime
	URL string
}

func mg(url string) func() *smugmug.Client {
	return func() *smugmug.Client {
		tracing := false // zerolog.GlobalLevel() == zerolog.DebugLevel
		client, err := smugmug.NewClient(
			smugmug.WithBaseURL(url),
			smugmug.WithUploadURL(url),
			smugmug.WithHTTPTracing(tracing))
		if err != nil {
			panic(err)
		}
		return client
	}
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
	return c.App.Metadata[RuntimeKey].(*Runtime) //nolint:errcheck // cannot happen
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
	name := strings.ReplaceAll(tt.name, " ", "-")
	return &cli.App{
		Name:     name,
		HelpName: name,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "json",
				Aliases:  []string{"j"},
				Value:    false,
				Required: false,
			},
		},
		Before: func(c *cli.Context) error {
			cfg := metrics.DefaultConfig(c.App.Name)
			cfg.EnableRuntimeMetrics = false
			cfg.TimerGranularity = time.Second
			sink := metrics.NewInmemSink(time.Hour*24, time.Hour*24)
			metric, err := metrics.New(cfg, sink)
			if err != nil {
				t.Error(err)
			}

			writer := io.Discard
			if c.Bool("json") {
				writer = c.App.Writer
			}

			rt := &ma.Runtime{
				Smugmug:  mg(url),
				Metrics:  metric,
				Sink:     sink,
				Encoder:  json.NewEncoder(writer),
				Grab:     new(http.Client),
				Fs:       afero.NewMemMapFs(),
				Language: language.English,
				Start:    time.Now(),
			}
			c.App.Metadata = map[string]any{
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
			if err := walkfs(c); err != nil {
				return err
			}
			if err := ma.Metrics(c); err != nil {
				return err
			}
			return counters(c.App.Name, t, tt.counters)(c)
		},
		Commands: []*cli.Command{cmd},
	}
}

func walkfs(c *cli.Context) error {
	return afero.Walk(runtime(c).Fs, "/", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrPermission) {
				fmt.Fprintf(c.App.ErrWriter, "%s (%s)\n", path, info.Mode().Perm().String())
				return filepath.SkipDir
			}
			return err
		}
		fmt.Fprintf(c.App.ErrWriter, "%s (%s)\n", path, info.Mode().Perm().String())
		return nil
	})
}

func counters(name string, t *testing.T, expected map[string]int) cli.AfterFunc {
	a := assert.New(t)
	return func(c *cli.Context) error {
		data := runtime(c).Sink.Data()
		for key, value := range expected {
			var found bool
			key = fmt.Sprintf("%s.%s", name, key)
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
	context   func(context.Context) context.Context
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

	ctx := context.Background()
	if tt.context != nil {
		ctx = tt.context(ctx)
	}
	err := app.RunContext(ctx, append([]string{app.Name}, tt.args...))
	if tt.err == "" {
		a.NoError(err)
		return
	}
	a.Error(err)
	if err != nil { // avoids a panic if err is nil
		a.Contains(err.Error(), tt.err)
	}
}

type ErrFs struct {
	afero.Fs
	name string
	err  error
}

func (p *ErrFs) Open(name string) (afero.File, error) {
	switch name {
	case p.name:
		return nil, p.err
	default:
		return p.Fs.Open(name)
	}
}
