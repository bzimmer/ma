package ma_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/ma"
)

func TestTitle(t *testing.T) {
	a := assert.New(t)

	before := func(c *cli.Context) error {
		c.App.Writer = new(bytes.Buffer)
		runtime(c).Encoder = json.NewEncoder(c.App.Writer)
		return nil
	}
	after := func(u string) cli.AfterFunc {
		return func(c *cli.Context) error {
			data := decode(a, c.App.Writer.(io.Reader))
			a.Equal(u, data["Title"])
			return nil
		}
	}
	for _, tt := range []harness{
		{
			name:   "default",
			args:   []string{"-j", "title", "foobar"},
			before: before,
			after:  after("Foobar"),
		},
		{
			name:   "title",
			args:   []string{"-j", "title", "--caser", "title", "foobar"},
			before: before,
			after:  after("Foobar"),
		},
		{
			name:   "upper",
			args:   []string{"-j", "title", "--caser", "upper", "foobar"},
			before: before,
			after:  after("FOOBAR"),
		},
		{
			name:   "lower",
			args:   []string{"-j", "title", "--caser", "lower", "foOBar"},
			before: before,
			after:  after("foobar"),
		},
		{
			name:   "unknown caser",
			args:   []string{"-j", "title", "--caser", "orange", "foOBar"},
			before: before,
			err:    "unknown caser: orange",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, nil, ma.CommandTitle)
		})
	}
}
