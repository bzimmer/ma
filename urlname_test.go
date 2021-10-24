package ma_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/bzimmer/ma"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestURLName(t *testing.T) {
	a := assert.New(t)

	before := func(c *cli.Context) error {
		c.App.Writer = new(bytes.Buffer)
		runtime(c).Encoder = ma.NewJSONEncoder(json.NewEncoder(c.App.Writer))
		return nil
	}
	after := func(u string) cli.AfterFunc {
		return func(c *cli.Context) error {
			data := decode(a, c.App.Writer.(io.Reader))
			a.Equal(u, data["UrlName"])
			return nil
		}
	}
	for _, tt := range []harness{
		{
			name:   "valid",
			args:   []string{"ma", "-j", "urlname", "foobar"},
			before: before,
			after:  after("Foobar"),
		},
		{
			name:   "remove `'s` and `-`",
			args:   []string{"ma", "-j", "urlname", "Foo's - The Best"},
			before: before,
			after:  after("Foos-The-Best"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			run(t, tt, nil, ma.CommandURLName)
		})
	}
}
