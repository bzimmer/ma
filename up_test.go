package ma_test

import (
	"net/http"
	"testing"

	"github.com/bzimmer/ma"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestUpload(t *testing.T) { //nolint
	t.Parallel()
	a := assert.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/album/qety", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		a.NoError(copyFile(w, "testdata/album_qety_404.json"))
	})
	mux.HandleFunc("/album/qety!images", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		a.NoError(copyFile(w, "testdata/album_qety_404.json"))
	})
	mux.HandleFunc("/album/TDZWbg!images", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/album_TDZWbg_images.json"))
	})
	mux.HandleFunc("/album/TDZWbg", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/album_TDZWbg.json"))
	})
	mux.HandleFunc("/Fujifilm_FinePix6900ZOOM.jpg", func(w http.ResponseWriter, r *http.Request) {
		a.Equal(http.MethodPut, r.Method)
		a.NoError(copyFile(w, "testdata/album_vVjSft_upload.json"))
	})

	for _, tt := range []harness{
		{
			name: "upload with no arguments",
			args: []string{"ma", "upload"},
			err:  "Required flag \"album\" not set",
		},
		{
			name: "upload invalid album",
			args: []string{"ma", "upload", "--album", "qety", "/tmp/foo/bar"},
			err:  "Not Found",
		},
		{
			name: "upload directory does not exist",
			args: []string{"ma", "upload", "--album", "TDZWbg", "/tmp/foo/bar"},
			err:  "file does not exist",
		},
		{
			name: "upload no valid files",
			args: []string{"ma", "upload", "--album", "TDZWbg", "/foo/bar"},
			before: func(app *cli.App) {
				a.NoError(runtime(app).Fs.MkdirAll("/foo/bar", 0777))
			},
		},
		{
			name: "upload file already exists",
			args: []string{"ma", "upload", "--album", "TDZWbg", "/foo/bar"},
			counters: map[string]int{
				"ma.fsUploadable.visit":    1,
				"ma.fsUploadable.open":     1,
				"ma.fsUploadable.skip.md5": 1,
			},
			before: func(app *cli.App) {
				fp, err := runtime(app).Fs.Create("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Nikon_D70.jpg"))
				a.NoError(fp.Close())
			},
		},
		{
			name: "upload skip unsupported file",
			args: []string{"ma", "upload", "--album", "TDZWbg", "/foo/bar"},
			counters: map[string]int{
				"ma.fsUploadable.visit":            1,
				"ma.fsUploadable.skip.unsupported": 1,
			},
			before: func(app *cli.App) {
				fp, err := runtime(app).Fs.Create("/foo/bar/Nikon_D70.xmp")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(fp.Close())
			},
		},
		{
			name: "upload replace existing image (dryrun)",
			args: []string{"ma", "upload", "--dryrun", "--album", "TDZWbg", "/foo/bar"},
			counters: map[string]int{
				"ma.fsUploadable.visit":   1,
				"ma.fsUploadable.open":    1,
				"ma.fsUploadable.replace": 1,
				"ma.upload.dryrun":        1,
			},
			before: func(app *cli.App) {
				// create a file of the same name as a previously uploaded file but copy the
				//  contents of a different file to force the md5s to be different
				fp, err := runtime(app).Fs.Create("/foo/bar/hdxDH/VsQ7zr/Nikon_D70.jpg")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Fujifilm_FinePix6900ZOOM.jpg"))
				a.NoError(fp.Close())
			},
		},
		{
			name: "upload new image",
			args: []string{"ma", "upload", "--album", "TDZWbg", "/foo/bar"},
			counters: map[string]int{
				"ma.fsUploadable.visit": 1,
				"ma.upload.attempt":     1,
				"ma.fsUploadable.open":  1,
				"ma.upload.success":     1,
			},
			before: func(app *cli.App) {
				fp, err := runtime(app).Fs.Create("/foo/bar/hdxDH/VsQ7zr/Fujifilm_FinePix6900ZOOM.jpg")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Fujifilm_FinePix6900ZOOM.jpg"))
				a.NoError(fp.Close())
			},
		},
		{
			name: "upload new image (dryrun)",
			args: []string{"ma", "upload", "--dryrun", "--album", "TDZWbg", "/foo/bar"},
			counters: map[string]int{
				"ma.fsUploadable.visit": 1,
				"ma.upload.dryrun":      1,
				"ma.fsUploadable.open":  1,
			},
			before: func(app *cli.App) {
				fp, err := runtime(app).Fs.Create("/foo/bar/hdxDH/VsQ7zr/Fujifilm_FinePix6900ZOOM.jpg")
				a.NotNil(fp)
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Fujifilm_FinePix6900ZOOM.jpg"))
				a.NoError(fp.Close())
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			harnessFunc(t, tt, mux, ma.CommandUpload)
		})
	}
}
