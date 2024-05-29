package ma

import (
	"fmt"
	"regexp"
	"time"

	"github.com/bzimmer/smugmug"
	"github.com/hashicorp/go-metrics"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// RuntimeKey in app metadata
const RuntimeKey = "github.com/bzimmer/ma#RuntimeKey"

// Runtime for access to runtime components
type Runtime struct {
	// Encoder encodes a struct
	Encoder Encoder
	// Smugmug returns a smugmug client
	Smugmug SmugmugFunc
	// Sink for metrics
	Sink *metrics.InmemSink
	// Metrics for capturing metrics
	Metrics *metrics.Metrics
	// Fs for file access
	Fs afero.Fs
	// Grab for bulk querying images
	Grab Grab
	// Language for title case
	Language language.Tag
	// Start time of the execution
	Start time.Time
}

// SmugmugFunc returns a smugmug client
// panics if credentials are not provided
type SmugmugFunc func() *smugmug.Client

// Encoder encodes a struct to a specific format
type Encoder interface {
	// Encode writes the encoding of v
	Encode(v any) error
}

func runtime(c *cli.Context) *Runtime {
	return c.App.Metadata[RuntimeKey].(*Runtime)
}

func albumOrNode(c *cli.Context) error {
	node := c.Bool("node")
	album := c.Bool("album")
	if !(album || node) {
		if err := c.Set("node", "true"); err != nil {
			return err
		}
		if err := c.Set("album", "true"); err != nil {
			return err
		}
	}
	return nil
}

// Metrics (if enabled) emits the metrics as json
func Metrics(c *cli.Context) error {
	if len(c.App.Metadata) == 0 {
		return nil
	}
	runtime(c).Metrics.AddSample(
		[]string{"elapsed"}, float32(time.Since(runtime(c).Start).Milliseconds()))
	data := runtime(c).Sink.Data()
	for i := range data {
		for key, val := range data[i].Counters {
			log.Info().
				Int("count", val.Count).
				Str("metric", key).
				Msg("counters")
		}
		for key, val := range data[i].Samples {
			as := val.AggregateSample
			log.Info().
				Int("count", val.Count).
				Str("metric", key).
				Float64("min", as.Min).
				Float64("max", as.Max).
				Float64("mean", as.Mean()).
				Float64("stddev", as.Stddev()).
				Msg("samples")
		}
	}
	return nil
}

func titlecase(c *cli.Context, s string) string {
	title := cases.Title(runtime(c).Language)
	return title.String(s)
}

var imageRE = regexp.MustCompile(`[a-zA-Z0-9]+-\d+`)

type InvalidVersionError struct {
	ImageKey string
}

func (x *InvalidVersionError) Error() string {
	return fmt.Sprintf("no version specified for image key {%s}", x.ImageKey)
}
