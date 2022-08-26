package ma_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/ma"
	"github.com/bzimmer/smugmug"
)

type errorEncoder struct{}

func (e *errorEncoder) Encode(v any) error {
	return errors.New("error encoder")
}

func TestUpload(t *testing.T) { //nolint
	a := assert.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/album/qety", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		http.ServeFile(w, r, "testdata/album_qety_404.json")
	})
	mux.HandleFunc("/album/qety!images", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		http.ServeFile(w, r, "testdata/album_qety_404.json")
	})
	mux.HandleFunc("/album/TDZWbg!images", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_TDZWbg_images.json")
	})
	mux.HandleFunc("/album/TDZWbg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_TDZWbg.json")
	})
	mux.HandleFunc("/Fujifilm_FinePix6900ZOOM.jpg", func(w http.ResponseWriter, r *http.Request) {
		a.Equal(http.MethodPut, r.Method)
		http.ServeFile(w, r, "testdata/album_vVjSft_upload.json")
	})

	for _, tt := range []harness{
		{
			name: "upload with no arguments",
			args: []string{"upload"},
			err:  "Required flag \"album\" not set",
		},
		{
			name: "upload invalid album",
			args: []string{"upload", "--album", "qety", "/tmp/foo/bar"},
			err:  "Not Found",
		},
		{
			name: "upload directory does not exist",
			args: []string{"upload", "--album", "TDZWbg", "/tmp/foo/bar"},
			err:  "file does not exist",
		},
		{
			name: "upload no valid files",
			args: []string{"upload", "--album", "TDZWbg", "/foo/bar"},
			before: func(c *cli.Context) error {
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar", 0755))
				return nil
			},
		},
		{
			name: "upload file already exists",
			args: []string{"upload", "--album", "TDZWbg", "/foo/bar"},
			counters: map[string]int{
				"uploadable.fs.visit":    1,
				"uploadable.fs.open":     1,
				"uploadable.fs.skip.md5": 1,
			},
			before: func(c *cli.Context) error {
				fp, err := runtime(c).Fs.Create("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Nikon_D70.jpg"))
				a.NoError(fp.Close())
				return nil
			},
		},
		{
			name: "upload skip unsupported file",
			args: []string{"upload", "--album", "TDZWbg", "/foo/bar"},
			counters: map[string]int{
				"uploadable.fs.visit":            1,
				"uploadable.fs.skip.unsupported": 1,
			},
			before: func(c *cli.Context) error {
				fp, err := runtime(c).Fs.Create("/foo/bar/Nikon_D70.xmp")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(fp.Close())
				return nil
			},
		},
		{
			name: "upload replace existing image (dryrun)",
			args: []string{"upload", "--dryrun", "--album", "TDZWbg", "/foo/bar"},
			counters: map[string]int{
				"uploadable.fs.visit":   1,
				"uploadable.fs.open":    1,
				"uploadable.fs.replace": 1,
				"upload.dryrun":         1,
			},
			before: func(c *cli.Context) error {
				// create a file of the same name as a previously uploaded file but copy the
				//  contents of a different file to force the md5s to be different
				fp, err := runtime(c).Fs.Create("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Fujifilm_FinePix6900ZOOM.jpg"))
				a.NoError(fp.Close())
				return nil
			},
		},
		{
			name: "upload new image",
			args: []string{"upload", "--album", "TDZWbg", "/foo/bar"},
			counters: map[string]int{
				"uploadable.fs.visit": 1,
				"upload.attempt":      1,
				"uploadable.fs.open":  1,
				"upload.success":      1,
			},
			before: func(c *cli.Context) error {
				fp, err := runtime(c).Fs.Create("/foo/bar/hdxDH/VsQ7zr/Fujifilm_FinePix6900ZOOM.jpg")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Fujifilm_FinePix6900ZOOM.jpg"))
				a.NoError(fp.Close())
				return nil
			},
		},
		{
			name: "upload new image encoding error",
			args: []string{"-j", "upload", "--album", "TDZWbg", "/foo/bar"},
			err:  "error encoder",
			before: func(c *cli.Context) error {
				fp, err := runtime(c).Fs.Create("/foo/bar/hdxDH/VsQ7zr/Fujifilm_FinePix6900ZOOM.jpg")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Fujifilm_FinePix6900ZOOM.jpg"))
				a.NoError(fp.Close())
				runtime(c).Encoder = &errorEncoder{}
				return nil
			},
		},
		{
			name: "upload new image (dryrun)",
			args: []string{"upload", "--dryrun", "--album", "TDZWbg", "/foo/bar"},
			counters: map[string]int{
				"uploadable.fs.visit": 1,
				"upload.dryrun":       1,
				"uploadable.fs.open":  1,
			},
			before: func(c *cli.Context) error {
				fp, err := runtime(c).Fs.Create("/foo/bar/hdxDH/VsQ7zr/Fujifilm_FinePix6900ZOOM.jpg")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Fujifilm_FinePix6900ZOOM.jpg"))
				a.NoError(fp.Close())
				return nil
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, mux, ma.CommandUpload)
		})
	}
}

func TestMirrorDryRun(t *testing.T) {
	a := assert.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/album/TDZWbg!images", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_TDZWbg_images.json")
	})
	mux.HandleFunc("/album/TDZWbg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_TDZWbg.json")
	})
	mux.HandleFunc("/Fujifilm_FinePix6900ZOOM.jpg", func(w http.ResponseWriter, r *http.Request) {
		a.Equal(http.MethodPut, r.Method)
		http.ServeFile(w, r, "testdata/album_vVjSft_upload.json")
	})

	for _, tt := range []harness{
		{
			name: "upload new image dryrun",
			args: []string{"upload", "--album", "TDZWbg", "--mirror", "--dryrun", "/foo/bar"},
			counters: map[string]int{
				"up.mirror.dryrun":    1,
				"uploadable.fs.open":  1,
				"uploadable.fs.visit": 1,
			},
			before: func(c *cli.Context) error {
				fp, err := runtime(c).Fs.Create("/foo/bar/hdxDH/VsQ7zr/Fujifilm_FinePix6900ZOOM.jpg")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Fujifilm_FinePix6900ZOOM.jpg"))
				a.NoError(fp.Close())
				return nil
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, mux, ma.CommandUpload)
		})
	}
}

func TestMirror(t *testing.T) {
	a := assert.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/album/TDZWbg!images", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_TDZWbg_images.json")
	})
	mux.HandleFunc("/album/TDZWbg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_TDZWbg.json")
	})
	mux.HandleFunc("/album/TDZWbg/image/TL4PJfh-0", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete:
			enc := json.NewEncoder(w)
			a.NoError(enc.Encode(struct {
				Response struct {
					Image *smugmug.Image `json:"Image"`
				} `json:"Response"`
				Code    int    `json:"Code"`
				Message string `json:"Message"`
			}{
				Code:    200,
				Message: "OK",
			}))
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}
	})
	mux.HandleFunc("/Fujifilm_FinePix6900ZOOM.jpg", func(w http.ResponseWriter, r *http.Request) {
		a.Equal(http.MethodPut, r.Method)
		http.ServeFile(w, r, "testdata/album_vVjSft_upload.json")
	})

	for _, tt := range []harness{
		{
			name: "mirror filesystem",
			args: []string{"upload", "--album", "TDZWbg", "--mirror", "/foo/bar"},
			counters: map[string]int{
				"upload.success": 1,
				// "up.mirror.delete":    1,
				// "up.delete.attempt":   1,
				"uploadable.fs.open":  1,
				"uploadable.fs.visit": 1,
			},
			before: func(c *cli.Context) error {
				fp, err := runtime(c).Fs.Create("/foo/bar/hdxDH/VsQ7zr/Fujifilm_FinePix6900ZOOM.jpg")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Fujifilm_FinePix6900ZOOM.jpg"))
				a.NoError(fp.Close())
				return nil
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, mux, ma.CommandUpload)
		})
	}
}
