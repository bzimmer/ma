package ma_test

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/ma"
)

func TestExif(t *testing.T) {
	a := assert.New(t)
	tests := []harness{
		{
			name: "no arguments",
			args: []string{"exif"},
		},
		{
			name: "supported exif file",
			args: []string{"exif", "/foo/bar/Nikon_D70.jpg"},
			before: func(c *cli.Context) error {
				fp := createTestFile(t, runtime(c).Fs)
				a.NotNil(fp)
				defer fp.Close()
				return nil
			},
		},
		{
			name: "error does not exist",
			args: []string{"exif", "/foo/bar/Nikon_D70.jpg"},
			err:  os.ErrNotExist.Error(),
		},
		{
			name: "error opening file",
			args: []string{"exif", "/foo/bar/"},
			err:  os.ErrPermission.Error(),
			before: func(c *cli.Context) error {
				fp := createTestFile(t, runtime(c).Fs)
				fp.Close()
				runtime(c).Fs = &ErrFs{Fs: runtime(c).Fs, err: fs.ErrPermission, name: fp.Name()}
				return nil
			},
		},
		{
			name: "unsupported exif file",
			args: []string{"exif", "/foo/bar/Olympus.orf"},
			before: func(c *cli.Context) error {
				afs := runtime(c).Fs
				a.NoError(afs.MkdirAll("/foo/bar", 0755))
				fp, err := afs.Create("/foo/bar/Olympus.orf")
				a.NoError(err)
				a.NotNil(fp)
				defer fp.Close()
				return nil
			},
		},
		{
			name: "file with no exif data",
			args: []string{"exif", "/foo/bar/user_cmac.json"},
			before: func(c *cli.Context) error {
				afs := runtime(c).Fs
				a.NoError(afs.MkdirAll("/foo/bar", 0755))
				fp, err := afs.Create("/foo/bar/user_cmac.json")
				a.NoError(err)
				a.NotNil(fp)
				a.NoError(copyFile(fp, "testdata/user_cmac.json"))
				defer fp.Close()
				return nil
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, nil, ma.CommandExif)
		})
	}
}
