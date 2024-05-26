package ma_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/bzimmer/smugmug"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/ma"
)

type errorEncoder struct{}

func (e *errorEncoder) Encode(_ any) error {
	return errors.New("error encoder")
}

func TestUpload(t *testing.T) {
	a := assert.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/album/qety", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/album/qety!images", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/album/TDZWbg!images", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_TDZWbg_images.json")
	})
	mux.HandleFunc("/album/TDZWbg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_TDZWbg.json")
	})
	mux.HandleFunc("/photo.jpg", func(w http.ResponseWriter, r *http.Request) {
		a.Equal(http.MethodPut, r.Method)
		http.ServeFile(w, r, "testdata/album_vVjSft_upload.json")
	})

	for _, tt := range []harness{
		{
			name: "upload with no arguments",
			args: []string{"upload"},
			err:  `Required flag "album" not set`,
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
			name: "upload images from null strings",
			args: []string{"upload", "--album", "TDZWbg", "--0"},
			counters: map[string]int{
				"upload.attempt":      2,
				"upload.success":      2,
				"uploadable.fs.open":  2,
				"uploadable.fs.visit": 2,
			},
			before: func(c *cli.Context) error {
				var in bytes.Buffer
				input, err := os.Open("testdata/null_input.txt")
				a.NoError(err)
				_, err = io.Copy(&in, input)
				a.NoError(err)
				c.App.Reader = &in
				fp, err := runtime(c).Fs.Create("/foo/bar/_DSC6073.JPG")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Fujifilm_FinePix6900ZOOM.jpg"))
				a.NoError(fp.Close())
				fp, err = runtime(c).Fs.Create("/foo/bar/IMG_0827.JPG")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Nikon_D70.jpg"))
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

func TestMirror(t *testing.T) {
	a := assert.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/album/dDfCWW!images", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_dDfCWW_images.json")
	})
	mux.HandleFunc("/album/dDfCWW", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_dDfCWW.json")
	})
	mux.HandleFunc("/album/dDfCWW/image/qPzttW4-0", func(w http.ResponseWriter, r *http.Request) {
		a.Equal(http.MethodDelete, r.Method)
		a.NoError(json.NewEncoder(w).Encode(struct {
			Response struct {
				Image *smugmug.Image `json:"Image"`
			} `json:"Response"`
			Code    int    `json:"Code"`
			Message string `json:"Message"`
		}{
			Code:    200,
			Message: "OK",
		}))
	})

	for _, tt := range []harness{
		{
			name: "mirror filesystem dryrun",
			args: []string{"upload", "--album", "dDfCWW", "--mirror", "--dryrun", "/foo/bar"},
			counters: map[string]int{
				"up.mirror.dryrun":       1,
				"uploadable.fs.open":     2,
				"uploadable.fs.skip.md5": 2,
				"uploadable.fs.visit":    2,
			},
			before: func(c *cli.Context) error {
				for _, name := range []string{"Fujifilm_FinePix6900ZOOM.jpg", "Nikon_D70.jpg"} {
					fp, err := runtime(c).Fs.Create("/foo/bar/baz/" + name)
					a.NotNil(fp)
					a.NoError(err)
					a.NoError(copyFile(fp, "testdata/"+name))
					a.NoError(fp.Close())
				}
				return nil
			},
		},
		{
			name: "mirror filesystem",
			args: []string{"upload", "--album", "dDfCWW", "--mirror", "/foo/bar"},
			counters: map[string]int{
				"up.mirror.delete":       1,
				"uploadable.fs.open":     2,
				"uploadable.fs.skip.md5": 2,
				"uploadable.fs.visit":    2,
				"up.delete.attempt":      1,
				"up.delete.success":      1,
			},
			before: func(c *cli.Context) error {
				for _, name := range []string{"Fujifilm_FinePix6900ZOOM.jpg", "Nikon_D70.jpg"} {
					fp, err := runtime(c).Fs.Create("/foo/bar/baz/" + name)
					a.NotNil(fp)
					a.NoError(err)
					a.NoError(copyFile(fp, "testdata/"+name))
					a.NoError(fp.Close())
				}
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
