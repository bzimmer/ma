package ma_test

import (
	"context"
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/bzimmer/ma"
)

func createTestFile(t *testing.T, afs afero.Fs) afero.File {
	if err := afs.MkdirAll("/foo/bar", 0755); err != nil {
		t.Error(err)
	}
	fp, err := afs.Create("/foo/bar/Nikon_D70.jpg")
	if err != nil {
		t.Error(err)
	}
	defer fp.Close()
	if err := copyFile(fp, "testdata/Nikon_D70.jpg"); err != nil {
		t.Error(err)
	}
	return fp
}

func TestCopy(t *testing.T) { //nolint
	a := assert.New(t)
	tests := []harness{
		{
			name: "one argument",
			args: []string{"cp", "/foo/bar"},
			err:  "expected 2+ arguments",
		},
		{
			name: "empty directory",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"cp.visited.directories": 1,
			},
			before: func(c *cli.Context) error {
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar", 0755))
				return nil
			},
		},
		{
			name: "hidden files",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"cp.visited.directories": 1,
				"cp.skip.hidden":         2,
			},
			before: func(c *cli.Context) error {
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar", 0755))
				fp, err := runtime(c).Fs.Create("/foo/bar/.something")
				a.NoError(err)
				a.NoError(fp.Close())
				fp, err = runtime(c).Fs.Create("/foo/bar/.else")
				a.NoError(err)
				a.NoError(fp.Close())
				return nil
			},
		},
		{
			name: "filename with no extension",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"cp.visited.directories":     1,
				"cp.skip.unsupported.<none>": 1,
			},
			before: func(c *cli.Context) error {
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar", 0755))
				fp, err := runtime(c).Fs.Create("/foo/bar/something")
				a.NoError(err)
				a.NoError(fp.Close())
				return nil
			},
		},
		{
			name: "unsupported files",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"cp.visited.directories":  2,
				"cp.skip.unsupported.UKN": 1,
				"cp.skip.unsupported.txt": 1,
			},
			before: func(c *cli.Context) error {
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar/boo", 0755))
				fp, err := runtime(c).Fs.Create("/foo/bar/DSC18920.UKN")
				a.NoError(err)
				a.NoError(fp.Close())
				fp, err = runtime(c).Fs.Create("/foo/bar/schedule.txt")
				a.NoError(err)
				a.NoError(fp.Close())
				return nil
			},
		},
		{
			name: "single image dng",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"cp.visited.directories": 1,
			},
			before: func(c *cli.Context) error {
				image := createTestFile(t, runtime(c).Fs)
				a.NoError(image.Close())
				// a bit of hack to test reading the entire contents of a .dng file
				// the exif parser doesn't care about file extensions, it sees only bytes
				name := image.Name()
				name = strings.Replace(name, ".jpg", ".dng", 1)
				a.NoError(runtime(c).Fs.Rename(image.Name(), name))
				return nil
			},
		},
		{
			name: "single image",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"cp.visited.directories": 1,
			},
			before: func(c *cli.Context) error {
				image := createTestFile(t, runtime(c).Fs)
				a.NoError(image.Close())

				tm := time.Date(2008, time.March, 15, 11, 22, 0, 0, time.Local)
				a.NoError(runtime(c).Fs.Chtimes(image.Name(), tm, tm))

				dst, err := runtime(c).Fs.Stat(image.Name())
				a.NoError(err)
				a.NotNil(dst)
				log.Info().Str("src", image.Name()).Time("dst", dst.ModTime()).Msg("set test modification times")
				return nil
			},
			after: func(c *cli.Context) error {
				dst, err := runtime(c).Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				a.NoError(err)
				a.NotNil(dst)
				if false {
					// @todo memfs chtimes doesn't seem to work properly -- need to investigate
					t := time.Date(2008, time.March, 15, 11, 22, 0, 0, time.Local)
					log.Info().Time("src", t).Time("dst", dst.ModTime()).Msg("asserting modification times")
					a.Equalf(t, dst.ModTime(), "expected identical modification times")
				}
				return nil
			},
		},
		{
			name: "image exists",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"cp.visited.directories": 1,
				"cp.skip.exists":         1,
			},
			before: func(c *cli.Context) error {
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar", 0755))
				for _, filename := range []string{"/foo/bar/Nikon_D70.xmp", "/foo/baz/2008/2008-03/15/Nikon_D70.jpg"} {
					fp, err := runtime(c).Fs.Create(filename)
					a.NoError(err)
					a.NoError(fp.Close())
					image, err := runtime(c).Fs.Create("/foo/bar/Nikon_D70.jpg")
					a.NoError(err)
					fp, err = os.Open("testdata/Nikon_D70.jpg")
					a.NoError(err)
					_, err = io.Copy(image, fp)
					a.NoError(err)
					a.NoError(image.Sync())
					a.NoError(image.Close())
					a.NoError(fp.Close())
				}
				return nil
			},
		},
		{
			name: "image + xmp",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"cp.visited.directories": 1,
			},
			before: func(c *cli.Context) error {
				fp := createTestFile(t, runtime(c).Fs)
				a.NoError(fp.Close())
				fp, err := runtime(c).Fs.Create("/foo/bar/Nikon_D70.xmp")
				a.NoError(err)
				a.NoError(fp.Close())
				return nil
			},
			after: func(c *cli.Context) error {
				stat, err := runtime(c).Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				a.NoError(err)
				a.NotNil(stat)
				stat, err = runtime(c).Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.xmp")
				a.NoError(err)
				a.NotNil(stat)
				return nil
			},
		},
		{
			name: "two valid files",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"cp.visited.directories": 1,
				"cp.visited.files":       2,
			},
			before: func(c *cli.Context) error {
				fp := createTestFile(t, runtime(c).Fs)
				a.NoError(fp.Close())
				fp, err := runtime(c).Fs.Create("/foo/bar/Nikon_D70_0.jpg")
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/Nikon_D70.jpg"))
				a.NoError(fp.Close())
				return nil
			},
			after: func(c *cli.Context) error {
				stat, err := runtime(c).Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				a.NoError(err)
				a.NotNil(stat)
				stat, err = runtime(c).Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70_0.jpg")
				a.NoError(err)
				a.NotNil(stat)
				return nil
			},
		},
		{
			name: "image + xmp dry-run",
			args: []string{"cp", "-n", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"cp.visited.directories": 1,
				"cp.file.dryrun":         2,
			},
			before: func(c *cli.Context) error {
				fp := createTestFile(t, runtime(c).Fs)
				a.NoError(fp.Close())
				fp, err := runtime(c).Fs.Create("/foo/bar/Nikon_D70.xmp")
				a.NoError(err)
				a.NoError(fp.Close())
				return nil
			},
			after: func(c *cli.Context) error {
				_, err := runtime(c).Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				a.Error(err)
				a.True(os.IsNotExist(err))
				_, err = runtime(c).Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.xmp")
				a.Error(err)
				a.True(os.IsNotExist(err))
				return nil
			},
		},
		{
			name: "image + xmp in different directories",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"cp.visited.files":            2,
				"cp.visited.directories":      2,
				"cp.fileset.skip.unsupported": 1,
			},
			before: func(c *cli.Context) error {
				fp := createTestFile(t, runtime(c).Fs)
				a.NoError(fp.Close())
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar/boo", 0755))
				fp, err := runtime(c).Fs.Create("/foo/bar/boo/Nikon_D70.xmp")
				a.NoError(err)
				a.NoError(fp.Close())
				return nil
			},
			after: func(c *cli.Context) error {
				stat, err := runtime(c).Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				a.NoError(err)
				a.NotNil(stat)
				_, err = runtime(c).Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.xmp")
				a.Error(err)
				a.True(os.IsNotExist(err))
				return nil
			},
		},
		{
			name: "image with garbage exif",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"cp.filesets":            1,
				"cp.visited.directories": 1,
				"cp.visited.files":       1,
				"cp.fileset.failed.exif": 1,
			},
			before: func(c *cli.Context) error {
				fp, err := runtime(c).Fs.Create("/foo/bar/Nikon_D70.jpg")
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/user_cmac.json"))
				a.NoError(fp.Close())
				return nil
			},
		},
		{
			name: "fail on copy",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			err:  "operation not permitted",
			counters: map[string]int{
				"cp.visited.directories": 1,
			},
			before: func(c *cli.Context) error {
				fp := createTestFile(t, runtime(c).Fs)
				a.NoError(fp.Close())
				runtime(c).Fs = afero.NewReadOnlyFs(runtime(c).Fs)
				return nil
			},
			after: func(c *cli.Context) error {
				stat, err := runtime(c).Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				a.True(os.IsNotExist(err))
				a.Nil(stat)
				return nil
			},
		},
		{
			name: "directory without read/execute permissions",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"cp.skip.denied":         1,
				"cp.visited.directories": 5,
			},
			before: func(c *cli.Context) error {
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar/boo0", 0755))
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar/boo1", 0755))
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar/boo2", 0600))
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar/boo3", 0755))
				runtime(c).Fs = &ErrFs{Fs: runtime(c).Fs, err: fs.ErrPermission, name: "/foo/bar/boo2"}
				return nil
			},
		},
		{
			name: "directory walk error",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			err:  "invalid argument",
			before: func(c *cli.Context) error {
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar/boo0", 0755))
				a.NoError(runtime(c).Fs.MkdirAll("/foo/bar/boo3", 0755))
				runtime(c).Fs = &ErrFs{Fs: runtime(c).Fs, err: fs.ErrInvalid, name: "/foo/bar/boo3"}
				return nil
			},
		},
		{
			name: "canceled context",
			args: []string{"cp", "/foo/bar", "/foo/baz"},
			err:  context.Canceled.Error(),
			before: func(c *cli.Context) error {
				fp := createTestFile(t, runtime(c).Fs)
				a.NoError(fp.Close())
				return nil
			},
			context: func(ctx context.Context) context.Context {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				<-ctx.Done()
				return ctx
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, nil, ma.CommandCopy)
		})
	}
}
