package ma_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/httpwares"
	"github.com/bzimmer/ma"
	"github.com/bzimmer/smugmug"
)

func TestExport(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	copy := func(w io.Writer, filename string) {
		fp, err := os.Open(filename)
		a.NoError(err)
		defer fp.Close()
		_, err = io.Copy(w, fp)
		a.NoError(err)
	}

	setup := func(mux *http.ServeMux) {
		mux.HandleFunc("/node/VsQ7zr!parents", func(w http.ResponseWriter, r *http.Request) {
			copy(w, "testdata/node_VsQ7zr_parents.json")
		})
		mux.HandleFunc("/node/VsQ7zr", func(w http.ResponseWriter, r *http.Request) {
			copy(w, "testdata/node_VsQ7zr.json")
		})
		mux.HandleFunc("/album/TDZWbg!images", func(w http.ResponseWriter, r *http.Request) {
			copy(w, "testdata/album_TDZWbg_images.json")
		})
	}

	tests := []struct {
		name      string
		args      []string
		err       string
		counters  map[string]int
		handlers  func(*http.ServeMux)
		transport http.RoundTripper
		before    func(runtime *ma.Runtime) error
		after     func(runtime *ma.Runtime) error
	}{
		{
			name: "export with no arguments",
			args: []string{"ma", "export"},
			err:  "expected two arguments, not {0}",
		},
		{
			name: "export album",
			args: []string{"ma", "export", "VsQ7zr", "/foo/bar"},
			transport: &httpwares.TestDataTransport{
				Status:      http.StatusOK,
				Filename:    "Nikon_D70.jpg",
				ContentType: "image/jpg",
			},
			counters: map[string]int{
				"ma.export.download.ok": 1,
			},
			handlers: setup,
			after: func(runtime *ma.Runtime) error {
				stat, err := runtime.Fs.Stat("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.NoError(err)
				a.NotNil(stat)
				return err
			},
		},
		{
			name: "export album image not found",
			args: []string{"ma", "export", "VsQ7zr", "/foo/bar"},
			transport: &httpwares.TestDataTransport{
				Status: http.StatusNotFound,
			},
			counters: map[string]int{
				"ma.export.download.failed.not_found": 1,
			},
			handlers: setup,
			after: func(runtime *ma.Runtime) error {
				stat, err := runtime.Fs.Stat("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.Nil(stat)
				a.Error(err)
				a.True(os.IsNotExist(err))
				return nil
			},
		},
		{
			name: "export album image server error",
			args: []string{"ma", "export", "VsQ7zr", "/foo/bar"},
			transport: &httpwares.TestDataTransport{
				Status: http.StatusInternalServerError,
			},
			counters: map[string]int{
				"ma.export.download.failed.internal_server_error": 1,
			},
			err:      "download failed",
			handlers: setup,
			after: func(runtime *ma.Runtime) error {
				stat, err := runtime.Fs.Stat("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.Nil(stat)
				a.Error(err)
				a.True(os.IsNotExist(err))
				return nil
			},
		},
		{
			name: "skip existing image",
			args: []string{"ma", "export", "VsQ7zr", "/foo/bar"},
			transport: &httpwares.TestDataTransport{
				Status: http.StatusOK,
			},
			counters: map[string]int{
				"ma.export.download.skipping.exists": 1,
			},
			handlers: setup,
			before: func(runtime *ma.Runtime) error {
				fp, err := runtime.Fs.Create("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.NotNil(fp)
				a.NoError(err)
				copy(fp, "testdata/Nikon_D70.jpg")
				return nil
			},
			after: func(runtime *ma.Runtime) error {
				stat, err := runtime.Fs.Stat("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.NoError(err)
				a.NotNil(stat)
				return nil
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			mux := http.NewServeMux()
			if tt.handlers != nil {
				tt.handlers(mux)
			}
			svr := httptest.NewServer(mux)
			defer svr.Close()

			app := NewTestApp(t, tt.name, ma.CommandExport(), smugmug.WithBaseURL(svr.URL))
			runtime(app).Grab = &http.Client{Transport: tt.transport}

			if tt.before != nil {
				a.NoError(tt.before(runtime(app)))
			}

			err := app.RunContext(context.TODO(), tt.args)
			switch tt.err == "" {
			case true:
				a.NoError(err)
			case false:
				a.Error(err)
				a.Contains(err.Error(), tt.err)
			}

			for key, value := range tt.counters {
				counter, err := findCounter(app, key)
				a.NoError(err)
				a.Equalf(value, counter.Count, key)
			}

			if tt.after != nil {
				a.NoError(tt.after(runtime(app)))
			}
		})
	}
}
