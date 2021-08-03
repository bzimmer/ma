package ma

import (
	"encoding/json"
	"errors"

	"github.com/armon/go-metrics"
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func encoder(c *cli.Context) *json.Encoder {
	return json.NewEncoder(c.App.Writer)
}

func client(c *cli.Context) (*smugmug.Client, error) {
	t, ok := c.App.Metadata["client"]
	if !ok {
		return nil, errors.New("missing client")
	}
	switch x := t.(type) {
	case *smugmug.Client:
		return x, nil
	default:
		return nil, errors.New("missing client")
	}
}

func sink(c *cli.Context) (*metrics.InmemSink, error) {
	t, ok := c.App.Metadata["sink"]
	if !ok {
		return nil, errors.New("missing sink")
	}
	switch x := t.(type) {
	case *metrics.InmemSink:
		return x, nil
	default:
		return nil, errors.New("missing sink")
	}
}

func metric(c *cli.Context) (*metrics.Metrics, error) {
	t, ok := c.App.Metadata["metrics"]
	if !ok {
		return nil, errors.New("missing metrics")
	}
	switch x := t.(type) {
	case *metrics.Metrics:
		return x, nil
	default:
		return nil, errors.New("missing metrics")
	}
}

func stats(c *cli.Context) error {
	snk, err := sink(c)
	if err != nil {
		return err
	}
	data := snk.Data()
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
