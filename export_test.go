package ma_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/httpwares"
	"github.com/bzimmer/ma"
	"github.com/bzimmer/smugmug"
)

func TestExport(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	handler := func(mux *http.ServeMux) {
		mux.HandleFunc("/node/VsQ7zr!parents", func(w http.ResponseWriter, r *http.Request) {
			a.NoError(copyFile(w, "testdata/node_VsQ7zr_parents.json"))
		})
		mux.HandleFunc("/node/VsQ7zr", func(w http.ResponseWriter, r *http.Request) {
			a.NoError(copyFile(w, "testdata/node_VsQ7zr.json"))
		})
		mux.HandleFunc("/album/TDZWbg!images", func(w http.ResponseWriter, r *http.Request) {
			a.NoError(copyFile(w, "testdata/album_TDZWbg_images.json"))
		})
	}

	tests := []struct {
		name     string
		args     []string
		err      string
		counters map[string]int
		handler  func(*http.ServeMux)
		before   func(app *cli.App)
		after    func(app *cli.App)
	}{
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
			handler: handler,
			before: func(app *cli.App) {
				runtime(app).Grab = &http.Client{Transport: &httpwares.TestDataTransport{
					Status:      http.StatusOK,
					Filename:    "Nikon_D70.jpg",
					ContentType: "image/jpg",
				}}
			},
			after: func(app *cli.App) {
				stat, err := runtime(app).Fs.Stat("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.NoError(err)
				a.NotNil(stat)
			},
		},
		{
			name: "export album image not found",
			args: []string{"ma", "export", "VsQ7zr", "/foo/bar"},
			counters: map[string]int{
				"ma.export.download.failed.not_found": 1,
			},
			handler: handler,
			before: func(app *cli.App) {
				runtime(app).Grab = &http.Client{Transport: &httpwares.TestDataTransport{
					Status: http.StatusNotFound,
				}}
			},
			after: func(app *cli.App) {
				stat, err := runtime(app).Fs.Stat("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.Nil(stat)
				a.Error(err)
				a.True(os.IsNotExist(err))
			},
		},
		{
			name: "export album image server error",
			args: []string{"ma", "export", "VsQ7zr", "/foo/bar"},
			counters: map[string]int{
				"ma.export.download.failed.internal_server_error": 1,
			},
			err:     "download failed",
			handler: handler,
			before: func(app *cli.App) {
				runtime(app).Grab = &http.Client{Transport: &httpwares.TestDataTransport{
					Status: http.StatusInternalServerError,
				}}
			},
			after: func(app *cli.App) {
				stat, err := runtime(app).Fs.Stat("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.Nil(stat)
				a.Error(err)
				a.True(os.IsNotExist(err))
			},
		},
		{
			name: "skip existing image",
			args: []string{"ma", "export", "VsQ7zr", "/foo/bar"},
			counters: map[string]int{
				"ma.export.download.skipping.exists": 1,
			},
			handler: handler,
			before: func(app *cli.App) {
				runtime(app).Grab = &http.Client{Transport: &httpwares.TestDataTransport{
					Status: http.StatusOK,
				}}
				fp, err := runtime(app).Fs.Create("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Nikon_D70.jpg"))
				a.NoError(fp.Close())
			},
			after: func(app *cli.App) {
				stat, err := runtime(app).Fs.Stat("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.NoError(err)
				a.NotNil(stat)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			mux := http.NewServeMux()
			if tt.handler != nil {
				tt.handler(mux)
			}
			svr := httptest.NewServer(mux)
			defer svr.Close()

			app := NewTestApp(t, tt.name, ma.CommandExport(), smugmug.WithBaseURL(svr.URL))

			if tt.before != nil {
				tt.before(app)
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
				tt.after(app)
			}
		})
	}
}
