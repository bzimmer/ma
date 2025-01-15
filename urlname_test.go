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

func TestURLName(t *testing.T) {
	a := assert.New(t)

	before := func(c *cli.Context) error {
		c.App.Writer = new(bytes.Buffer)
		runtime(c).Encoder = json.NewEncoder(c.App.Writer)
		return nil
	}
	after := func(u string, valid bool) cli.AfterFunc {
		return func(c *cli.Context) error {
			data := decode(a, c.App.Writer.(io.Reader)) //nolint:errcheck // cannot happen
			a.Equal(u, data["UrlName"])
			a.Equal(valid, data["Valid"])
			return nil
		}
	}
	for _, tt := range []harness{
		{
			name:   "valid",
			args:   []string{"-j", "urlname", "foobar"},
			before: before,
			after:  after("Foobar", true),
		},
		{
			name:   "remove `'s` and `-`",
			args:   []string{"-j", "urlname", "Foo's - The Best"},
			before: before,
			after:  after("Foos-The-Best", true),
		},
		{
			name:   "empty name",
			args:   []string{"-j", "urlname", "-a", "\"\""},
			before: before,
			after:  after("", false),
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, nil, ma.CommandURLName)
		})
	}
}
