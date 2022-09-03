package ma

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/armon/go-metrics"
	"github.com/bzimmer/smugmug"
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
	// Client for SmugMug
	Client *smugmug.Client
	// Sink for metrics
	Sink *metrics.InmemSink
	// Metrics for capturing metrics
	Metrics *metrics.Metrics
	// Fs for file access
	Fs afero.Fs
	// Grab for bulk querying images
	Grab Grab
	// Exif for accessing EXIF metadata
	Exif Exif
	// Language for title case
	Language language.Tag
	// Start time of the execution
	Start time.Time
}

// Encoder encodes a struct to a specific format
type Encoder interface {
	// Encode writes the encoding of v
	Encode(v any) error
}

type encoderBlackhole struct{}

func (e *encoderBlackhole) Encode(_ any) error {
	return nil
}

func NewBlackholeEncoder() Encoder {
	return &encoderBlackhole{}
}

type encoderJSON struct {
	encoder *json.Encoder
}

func (e *encoderJSON) Encode(v any) error {
	return e.encoder.Encode(v)
}

func NewJSONEncoder(enc *json.Encoder) Encoder {
	return &encoderJSON{encoder: enc}
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
	if c.Bool("metrics") {
		return runtime(c).Encoder.Encode(data)
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
