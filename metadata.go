package ma

import (
	"github.com/armon/go-metrics"
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
)

// Encoder encodes a struct to a specific format
type Encoder interface {
	// Encode writes the encoding of v
	Encode(v interface{}) error
}

func encoder(c *cli.Context) Encoder {
	return c.App.Metadata["encoder"].(Encoder)
}

func client(c *cli.Context) *smugmug.Client {
	return c.App.Metadata["client"].(*smugmug.Client)
}

func sink(c *cli.Context) *metrics.InmemSink {
	return c.App.Metadata["sink"].(*metrics.InmemSink)
}

func metric(c *cli.Context) *metrics.Metrics {
	return c.App.Metadata["metrics"].(*metrics.Metrics)
}

func afs(c *cli.Context) afero.Fs {
	return c.App.Metadata["fs"].(afero.Fs)
}

func stats(c *cli.Context) error {
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
