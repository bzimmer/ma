package ma

import (
	"errors"
	"unicode"

	"github.com/armon/go-metrics"
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
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
}

// Encoder encodes a struct to a specific format
type Encoder interface {
	// Encode writes the encoding of v
	Encode(v interface{}) error
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

// Stats logs and encodes (if enabled) the stats
func Stats(c *cli.Context) error {
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
	return runtime(c).Encoder.Encode(data)
}

var ErrInvalidURLName = errors.New("node url name must start with a number or capital letter")

func validateURLName(urlName string) error {
	v := rune(urlName[0])
	switch {
	case unicode.IsNumber(v), unicode.IsUpper(v):
		return nil
	default:
		return ErrInvalidURLName
	}
}
