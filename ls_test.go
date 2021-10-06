package ma_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
	"github.com/bzimmer/smugmug"
)

func TestList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		counter string
		count   int
		errmsg  string
	}{
		{
			name:    "album",
			args:    []string{"ma", "ls", "album", "RM4BL2"},
			counter: "ma.ls.album",
			count:   1,
		},
		{
			name:    "node",
			args:    []string{"ma", "ls", "node"},
			counter: "ma.ls.node",
			count:   1,
		},
		{
			name:    "image with version",
			args:    []string{"ma", "ls", "image", "B2fHSt7-0"},
			counter: "ma.ls.image",
			count:   1,
		},
		{
			name:    "image with auto-versioning",
			args:    []string{"ma", "ls", "image", "--zero-version", "B2fHSt7"},
			counter: "ma.ls.image",
			count:   1,
		},
		{
			name:    "image with no version and no auto-versioning",
			args:    []string{"ma", "ls", "image", "B2fHSt7"},
			counter: "",
			count:   0,
			errmsg:  "no version specified",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var filename string
				switch {
				case strings.HasPrefix(r.URL.Path, "/!authuser"):
					filename = "testdata/user_cmac.json"
				case strings.HasPrefix(r.URL.Path, "/node/zx4Fx"):
					filename = "testdata/node_zx4Fx.json"
				case strings.HasPrefix(r.URL.Path, "/image/B2fHSt7"):
					filename = "testdata/image_B2fHSt7-0.json"
				case strings.HasPrefix(r.URL.Path, "/album/RM4BL2"):
					filename = "testdata/album_RM4BL2.json"
				default:
					a.FailNow("unexpected path", r.URL.Path)
				}
				fp, err := os.Open(filename)
				a.NoError(err)
				defer fp.Close()
				_, err = io.Copy(w, fp)
				a.NoError(err)
			}))
			defer svr.Close()

			app := NewTestApp(t, ma.CommandList(), smugmug.WithBaseURL(svr.URL))

			_, err := findCounter(app, tt.counter)
			a.Error(err)

			switch {
			case tt.errmsg != "":
				err := app.RunContext(context.TODO(), tt.args)
				a.True(strings.Contains(err.Error(), tt.errmsg))
			default:
				a.NoError(app.RunContext(context.TODO(), tt.args))
			}

			if tt.counter == "" {
				return
			}
			counter, err := findCounter(app, tt.counter)
			a.NoError(err)
			a.Equal(tt.count, counter.Count)
		})
	}
}

func TestListIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "ls",
			args: []string{"-j", "ls", "node"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			ma := Command(tt.args...)
			out, err := ma.Output()
			a.NoError(err)
			res := make(map[string]interface{})
			dec := json.NewDecoder(bytes.NewBuffer(out))
			a.NoError(dec.Decode(&res))
			a.Equal("Folder", res["Type"])
		})
	}
}
