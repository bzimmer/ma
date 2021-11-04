package ma_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/ma"
)

type grab struct {
	url    string
	status int
}

func (g *grab) Do(req *http.Request) (*http.Response, error) {
	if g.status > 0 {
		res := &http.Response{
			StatusCode:    g.status,
			ContentLength: 0,
			Body:          io.NopCloser(bytes.NewBuffer(nil)),
			Header:        make(map[string][]string),
			Request:       req,
		}
		return res, nil
	}
	url := fmt.Sprintf("%s%s", g.url, req.URL.Path)
	proxy, err := http.NewRequestWithContext(req.Context(), req.Method, url, req.Body)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(proxy)
}

func TestExport(t *testing.T) { //nolint
	a := assert.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/node/VsQ7zr!parents", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/node_VsQ7zr_parents.json")
	})
	mux.HandleFunc("/node/VsQ7zr", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/node_VsQ7zr.json")
	})
	mux.HandleFunc("/album/TDZWbg!images", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_TDZWbg_images.json")
	})
	mux.HandleFunc("/photos/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/Nikon_D70.jpg")
	})

	tests := []harness{
		{
			name: "export with no arguments",
			args: []string{"ma", "export"},
			err:  "expected two arguments, not {0}",
		},
		{
			name: "export album",
			args: []string{"ma", "export", "VsQ7zr", "/foo/bar"},
			counters: map[string]int{
				"ma.export.download.ok": 1,
			},
			before: func(c *cli.Context) error {
				runtime(c).Grab = &grab{url: runtime(c).URL}
				return nil
			},
			after: func(c *cli.Context) error {
				stat, err := runtime(c).Fs.Stat("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.NoError(err)
				a.NotNil(stat)
				return nil
			},
		},
		{
			name: "export album image not found",
			args: []string{"ma", "export", "VsQ7zr", "/foo/bar"},
			counters: map[string]int{
				"ma.export.download.failed.not_found": 1,
			},
			before: func(c *cli.Context) error {
				runtime(c).Grab = &grab{
					url:    runtime(c).URL,
					status: http.StatusNotFound,
				}
				return nil
			},
			after: func(c *cli.Context) error {
				stat, err := runtime(c).Fs.Stat("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.Nil(stat)
				a.Error(err)
				a.True(os.IsNotExist(err))
				return nil
			},
		},
		{
			name: "export album image server error",
			args: []string{"ma", "export", "VsQ7zr", "/foo/bar"},
			counters: map[string]int{
				"ma.export.download.failed.internal_server_error": 1,
			},
			err: "download failed",
			before: func(c *cli.Context) error {
				runtime(c).Grab = &grab{
					url:    runtime(c).URL,
					status: http.StatusInternalServerError,
				}
				return nil
			},
			after: func(c *cli.Context) error {
				stat, err := runtime(c).Fs.Stat("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.Nil(stat)
				a.Error(err)
				a.True(os.IsNotExist(err))
				return nil
			},
		},
		{
			name: "skip existing image",
			args: []string{"ma", "export", "VsQ7zr", "/foo/bar"},
			counters: map[string]int{
				"ma.export.download.skipping.exists": 1,
			},
			before: func(c *cli.Context) error {
				runtime(c).Grab = &grab{url: runtime(c).URL}
				fp, err := runtime(c).Fs.Create("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Nikon_D70.jpg"))
				a.NoError(fp.Close())
				return nil
			},
			after: func(c *cli.Context) error {
				stat, err := runtime(c).Fs.Stat("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.NoError(err)
				a.NotNil(stat)
				return nil
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, mux, ma.CommandExport)
		})
	}
}
