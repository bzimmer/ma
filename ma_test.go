package ma_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/armon/go-metrics"
	"github.com/bzimmer/ma"
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	os.Exit(m.Run())
}

func runtime(app *cli.App) *ma.Runtime {
	return app.Metadata[ma.RuntimeKey].(*ma.Runtime)
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

type encoderBlackhole struct{}

func (e *encoderBlackhole) Encode(_ interface{}) error {
	return nil
}

func NewTestApp(t *testing.T, name string, cmd *cli.Command, opts ...smugmug.Option) *cli.App {
	cfg := metrics.DefaultConfig("ma")
	cfg.EnableRuntimeMetrics = false
	cfg.TimerGranularity = time.Second
	sink := metrics.NewInmemSink(time.Hour*24, time.Hour*24)
	metric, err := metrics.New(cfg, sink)
	if err != nil {
		t.Error(err)
	}

	client, err := smugmug.NewClient(opts...)
	if err != nil {
		t.Error(err)
	}

	return &cli.App{
		Name:     name,
		HelpName: name,
		After: func(c *cli.Context) error {
			t.Log(name)
			switch v := runtime(c.App).Fs.(type) {
			case *afero.MemMapFs:
				v.List()
			default:
			}
			return ma.Stats(c)
		},
		Commands: []*cli.Command{cmd},
		Metadata: map[string]interface{}{
			ma.RuntimeKey: &ma.Runtime{
				Client:  client,
				Metrics: metric,
				Sink:    sink,
				Grab:    new(http.Client),
				Encoder: new(encoderBlackhole),
				Fs:      afero.NewMemMapFs(),
			},
		},
	}
}

func findCounter(app *cli.App, name string) (metrics.SampledValue, error) {
	sink := runtime(app).Sink
	for i := range sink.Data() {
		im := sink.Data()[i]
		if sample, ok := im.Counters[name]; ok {
			return sample, nil
		}
	}
	return metrics.SampledValue{}, fmt.Errorf("cannot find sample value for {%s}", name)
}

// Root finds the root of the source tree by recursively ascending until 'go.mod' is located
func Root() string {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	paths := []string{string(os.PathSeparator)}
	paths = append(paths, strings.Split(path, string(os.PathSeparator))...)
	for len(paths) > 0 {
		x := filepath.Join(paths...)
		root := filepath.Join(x, "go.mod")
		if _, err := os.Stat(root); os.IsNotExist(err) {
			paths = paths[:len(paths)-1]
		} else {
			return x
		}
	}
	panic("unable to find go.mod")
}

func Command(args ...string) *exec.Cmd {
	return exec.Command(filepath.Join(Root(), "dist", "ma"), args...) //nolint:gosec
}

type harness struct {
	name, err     string
	args          []string
	counters      map[string]int
	before, after func(app *cli.App)
}

func harnessFunc(t *testing.T, tt harness, mux *http.ServeMux, cmd func() *cli.Command) {
	a := assert.New(t)

	svr := httptest.NewServer(mux)
	defer svr.Close()

	app := NewTestApp(t, tt.name, cmd(), smugmug.WithBaseURL(svr.URL), smugmug.WithHTTPTracing(false))

	if tt.before != nil {
		tt.before(app)
	}

	err := app.RunContext(context.TODO(), tt.args)
	switch tt.err == "" {
	case true:
		a.NoError(err)
	case false:
		a.Error(err)
		a.Contains(err.Error(), tt.err)
	}

	for key, value := range tt.counters {
		counter, err := findCounter(app, key)
		a.NoError(err)
		a.Equalf(value, counter.Count, key)
	}

	if tt.after != nil {
		tt.after(app)
	}
}
