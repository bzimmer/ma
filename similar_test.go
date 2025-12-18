package ma_test

import (
	"fmt"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/ma"
)

func TestSimilar(t *testing.T) { //nolint:gocognit // test harness
	a := assert.New(t)
	tests := []harness{
		{
			name: "similar files",
			args: []string{"similar", "/foo/bar"},
			counters: map[string]int{
				"similar.analyze.true":  1,
				"similar.analyze.false": 2,
				"similar.path":          4,
				"similar.icon.skipped":  1,
			},
			before: func(c *cli.Context) error {
				afs := runtime(c).Fs
				a.NoError(afs.MkdirAll("/foo/bar", 0755))
				for name, file := range map[string]string{
					"/foo/bar/A.jpg":          "testdata/Nikon_D70.jpg",
					"/foo/bar/B.jpg":          "testdata/Fujifilm_FinePix6900ZOOM.jpg",
					"/foo/bar/C.jpg":          "testdata/Fujifilm_FinePix6900ZOOM.jpg",
					"/foo/bar/user_cmac.json": "testdata/user_cmac.json",
				} {
					fp, err := afs.Create(name)
					if err != nil {
						a.NoError(err)
					}
					if err = copyFile(fp, file); err != nil {
						a.NoError(err)
					}
					a.NoError(fp.Close())
				}
				return nil
			},
		},
		{
			name: "permission denied",
			args: []string{"similar", "/foo/bar"},
			err:  "permission denied",
			counters: map[string]int{
				"similar.path":       1,
				"similar.icon.error": 1,
			},
			before: func(c *cli.Context) error {
				afs := runtime(c).Fs
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar", 0755))
				fp, err := afs.Create("/foo/bar/user_foo.json")
				if err != nil {
					a.NoError(err)
				}
				if err = copyFile(fp, "testdata/user_cmac.json"); err != nil {
					a.NoError(err)
				}
				a.NoError(fp.Close())
				runtime(c).Fs = &ErrFs{
					Fs:   runtime(c).Fs,
					err:  fs.ErrPermission,
					name: "/foo/bar/user_foo.json"}
				return nil
			},
		},
		{
			name: "no image files",
			args: []string{"similar", "-c", "4", "/foo/bar"},
			counters: map[string]int{
				"similar.path":         10,
				"similar.icon.skipped": 10,
			},
			before: func(c *cli.Context) error {
				afs := runtime(c).Fs
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar", 0755))
				for i := range 10 {
					fp, err := afs.Create(fmt.Sprintf("/foo/bar/user_foo_%02d.json", i))
					if err != nil {
						a.NoError(err)
					}
					if err = copyFile(fp, "testdata/user_cmac.json"); err != nil {
						a.NoError(err)
					}
					a.NoError(fp.Close())
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, nil, ma.CommandSimilar)
		})
	}
}
