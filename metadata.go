package ma

import (
	"errors"

	"github.com/armon/go-metrics"
	"github.com/bzimmer/smugmug"
	"github.com/urfave/cli/v2"
)

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
