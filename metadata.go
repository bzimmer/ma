package ma

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/armon/go-metrics"
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type Encoder interface {
	Encode(op string, msg interface{}) error
}

type EncoderJSON struct {
	Encoder *json.Encoder
}

func (e *EncoderJSON) Encode(op string, msg interface{}) error {
	return e.Encoder.Encode(msg)
}

type EncoderLog struct{}

func (e *EncoderLog) Encode(op string, msg interface{}) error {
	g, ok := msg.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{}, found %z", msg)
	}
	m := log.Info()
	for key, val := range g {
		switch x := val.(type) {
		case string:
			m = m.Str(key, x)
		case int:
			m = m.Int(key, x)
		case []string:
			m = m.Strs(key, x)
		case float64:
			m = m.Float64(key, x)
		default:
			log.Warn().Str("key", key).Msg("unhandled")
			m = m.Interface(key, val)
		}
	}
	m.Msg(op)
	return nil
}

func encoder(c *cli.Context) (Encoder, error) {
	t, ok := c.App.Metadata["encoder"]
	if !ok {
		return nil, errors.New("missing encoder")
	}
	switch x := t.(type) {
	case Encoder:
		return x, nil
	default:
		return nil, errors.New("missing encoder")
	}
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
