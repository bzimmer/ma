package ma

import (
	"net/http"

	"github.com/armon/go-metrics"
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
)

// RuntimeKey in app metadata
const RuntimeKey = "_runtime"

// Runtime for access to runtime components
type Runtime struct {
	// Encoder encodes a struct
	Encoder Encoder
	// Client for SmugMug
	Client *smugmug.Client
	// Sink for metrics
	Sink *metrics.InmemSink
	// Fs for file access
	Fs afero.Fs
	// Metrics for capturing metrics
	Metrics *metrics.Metrics
	// Grab for bulk querying images
	Grab *http.Client
}

// Encoder encodes a struct to a specific format
type Encoder interface {
	// Encode writes the encoding of v
	Encode(v interface{}) error
}

func runtime(c *cli.Context) *Runtime {
	return c.App.Metadata[RuntimeKey].(*Runtime)
}

func encoder(c *cli.Context) Encoder {
	return runtime(c).Encoder
}

func client(c *cli.Context) *smugmug.Client {
	return runtime(c).Client
}

func sink(c *cli.Context) *metrics.InmemSink {
	return runtime(c).Sink
}

func metric(c *cli.Context) *metrics.Metrics {
	return runtime(c).Metrics
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

// Stats logs and encodes (if requested) the stats
func Stats(c *cli.Context) error {
	data := sink(c).Data()
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
	return encoder(c).Encode(data)
}
