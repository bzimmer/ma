package ma_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/ma"
)

func TestMetrics(t *testing.T) {
	a := assert.New(t)
	tests := []harness{
		{
			name: "no runtime",
			args: []string{"matrix"},
			before: func(c *cli.Context) error {
				c.App.After = nil
				return nil
			},
			after: func(c *cli.Context) error {
				c.App.Metadata = map[string]any{}
				a.Nil(c.App.Metadata[RuntimeKey])
				return ma.Metrics(c)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, nil, func() *cli.Command {
				return &cli.Command{
					Name: "matrix",
					Action: func(_ *cli.Context) error {
						return nil
					},
				}
			})
		})
	}
}
