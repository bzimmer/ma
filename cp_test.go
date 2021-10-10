package ma_test

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
)

func createTestFile(fs afero.Fs) (afero.File, error) {
	if err := fs.MkdirAll("/foo/bar", 0777); err != nil {
		return nil, err
	}
	fp, err := fs.Create("/foo/bar/Nikon_D70.jpg")
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	if err := copyFile(fp, "testdata/Nikon_D70.jpg"); err != nil {
		return nil, err
	}
	return fp, nil
}

func TestCopy(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	tests := []struct {
		name     string
		args     []string
		err      string
		counters map[string]int
		before   func(runtime *ma.Runtime)
		after    func(runtime *ma.Runtime)
	}{
		{
			name: "one argument",
			args: []string{"ma", "cp", "/foo/bar"},
			err:  "expected 2+ arguments",
		},
		{
			name: "empty directory",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories": 1,
			},
			before: func(runtime *ma.Runtime) {
				a.NoError(runtime.Fs.MkdirAll("/foo/bar", 0777))
			},
		},
		{
			name: "hidden files",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories": 1,
				"ma.cp.skip.hidden":         2,
			},
			before: func(runtime *ma.Runtime) {
				a.NoError(runtime.Fs.MkdirAll("/foo/bar", 0777))
				fp, err := runtime.Fs.Create("/foo/bar/.something")
				a.NoError(err)
				a.NoError(fp.Close())
				fp, err = runtime.Fs.Create("/foo/bar/.else")
				a.NoError(err)
				a.NoError(fp.Close())
			},
		},
		{
			name: "filename with no extension",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories":     1,
				"ma.cp.skip.unsupported.<none>": 1,
			},
			before: func(runtime *ma.Runtime) {
				a.NoError(runtime.Fs.MkdirAll("/foo/bar", 0777))
				fp, err := runtime.Fs.Create("/foo/bar/something")
				a.NoError(err)
				a.NoError(fp.Close())
			},
		},
		{
			name: "unsupported files",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories":  2,
				"ma.cp.skip.unsupported.UKN": 1,
				"ma.cp.skip.unsupported.txt": 1,
			},
			before: func(runtime *ma.Runtime) {
				a.NoError(runtime.Fs.MkdirAll("/foo/bar/boo", 0777))
				fp, err := runtime.Fs.Create("/foo/bar/DSC18920.UKN")
				a.NoError(err)
				a.NoError(fp.Close())
				fp, err = runtime.Fs.Create("/foo/bar/schedule.txt")
				a.NoError(err)
				a.NoError(fp.Close())
			},
		},
		{
			name: "single image dng",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories": 1,
			},
			before: func(runtime *ma.Runtime) {
				image, err := createTestFile(runtime.Fs)
				a.NoError(err)
				a.NoError(image.Close())
				// a bit of hack to test reading the entire contents of a .dng file
				// the exif parser doesn't care about file extensions, it sees only bytes
				name := image.Name()
				name = strings.Replace(name, ".jpg", ".dng", 1)
				a.NoError(runtime.Fs.Rename(image.Name(), name))
			},
		},
		{
			name: "single image",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories": 1,
			},
			before: func(runtime *ma.Runtime) {
				image, err := createTestFile(runtime.Fs)
				a.NoError(err)
				a.NoError(image.Close())

				tm := time.Date(2008, time.March, 15, 11, 22, 0, 0, time.Local)
				a.NoError(runtime.Fs.Chtimes(image.Name(), tm, tm))

				dst, err := runtime.Fs.Stat(image.Name())
				a.NoError(err)
				a.NotNil(dst)
				log.Info().Str("src", image.Name()).Time("dst", dst.ModTime()).Msg("set test modification times")
			},
			after: func(runtime *ma.Runtime) {
				dst, err := runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				a.NoError(err)
				a.NotNil(dst)
				if false {
					// @todo memfs chtimes doesn't seem to work properly -- need to investigate
					t := time.Date(2008, time.March, 15, 11, 22, 0, 0, time.Local)
					log.Info().Time("src", t).Time("dst", dst.ModTime()).Msg("asserting modification times")
					a.Equalf(t, dst.ModTime(), "expected identical modification times")
				}
			},
		},
		{
			name: "image exists",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories": 1,
				"ma.cp.skip.exists":         1,
			},
			before: func(runtime *ma.Runtime) {
				a.NoError(runtime.Fs.MkdirAll("/foo/bar", 0777))
				for _, filename := range []string{"/foo/bar/Nikon_D70.xmp", "/foo/baz/2008/2008-03/15/Nikon_D70.jpg"} {
					fp, err := runtime.Fs.Create(filename)
					a.NoError(err)
					a.NoError(fp.Close())
					image, err := runtime.Fs.Create("/foo/bar/Nikon_D70.jpg")
					a.NoError(err)
					fp, err = os.Open("testdata/Nikon_D70.jpg")
					a.NoError(err)
					_, err = io.Copy(image, fp)
					a.NoError(err)
					a.NoError(image.Sync())
					a.NoError(image.Close())
					a.NoError(fp.Close())
				}
			},
		},
		{
			name: "image + xmp",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories": 1,
			},
			before: func(runtime *ma.Runtime) {
				fp, err := createTestFile(runtime.Fs)
				a.NoError(err)
				a.NoError(fp.Close())
				fp, err = runtime.Fs.Create("/foo/bar/Nikon_D70.xmp")
				a.NoError(err)
				a.NoError(fp.Close())
			},
			after: func(runtime *ma.Runtime) {
				stat, err := runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				a.NoError(err)
				a.NotNil(stat)
				stat, err = runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.xmp")
				a.NoError(err)
				a.NotNil(stat)
			},
		},
		{
			name: "image + xmp dry-run",
			args: []string{"ma", "cp", "-n", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories": 1,
				"ma.cp.file.dryrun":         2,
			},
			before: func(runtime *ma.Runtime) {
				fp, err := createTestFile(runtime.Fs)
				a.NoError(err)
				a.NoError(fp.Close())
				fp, err = runtime.Fs.Create("/foo/bar/Nikon_D70.xmp")
				a.NoError(err)
				a.NoError(fp.Close())
			},
			after: func(runtime *ma.Runtime) {
				_, err := runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				a.Error(err)
				a.True(os.IsNotExist(err))
				_, err = runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.xmp")
				a.Error(err)
				a.True(os.IsNotExist(err))
			},
		},
		{
			name: "image + xmp in different directories",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.files":            2,
				"ma.cp.visited.directories":      2,
				"ma.cp.fileset.skip.unsupported": 1,
			},
			before: func(runtime *ma.Runtime) {
				fp, err := createTestFile(runtime.Fs)
				a.NoError(err)
				a.NoError(fp.Close())
				a.NoError(runtime.Fs.MkdirAll("/foo/bar/boo", 0777))
				fp, err = runtime.Fs.Create("/foo/bar/boo/Nikon_D70.xmp")
				a.NoError(err)
				a.NoError(fp.Close())
			},
			after: func(runtime *ma.Runtime) {
				stat, err := runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				a.NoError(err)
				a.NotNil(stat)
				_, err = runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.xmp")
				a.Error(err)
				a.True(os.IsNotExist(err))
			},
		},
		{
			name: "image with garbage exif",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			err:  "EOF",
			counters: map[string]int{
				"ma.cp.filesets":            1,
				"ma.cp.visited.directories": 1,
				"ma.cp.visited.files":       1,
				"ma.cp.fileset.failed.exif": 1,
			},
			before: func(runtime *ma.Runtime) {
				fp, err := runtime.Fs.Create("/foo/bar/Nikon_D70.jpg")
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/user_cmac.json"))
				a.NoError(fp.Close())
			},
		},
		{
			name: "fail on copy",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			err:  "operation not permitted",
			counters: map[string]int{
				"ma.cp.visited.directories": 1,
			},
			before: func(runtime *ma.Runtime) {
				fp, err := createTestFile(runtime.Fs)
				a.NoError(err)
				a.NoError(fp.Close())
				runtime.Fs = afero.NewReadOnlyFs(runtime.Fs)
			},
			after: func(runtime *ma.Runtime) {
				stat, err := runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				a.True(os.IsNotExist(err))
				a.Nil(stat)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			app := NewTestApp(t, tt.name, ma.CommandCopy())

			if tt.before != nil {
				tt.before(runtime(app))
			}

			err := app.RunContext(context.TODO(), tt.args)
			switch tt.err == "" {
			case true:
				a.NoError(err)
			case false:
				a.Contains(err.Error(), tt.err)
			}

			for key, value := range tt.counters {
				counter, err := findCounter(app, key)
				a.NoError(err)
				a.Equalf(value, counter.Count, key)
			}

			if tt.after != nil {
				tt.after(runtime(app))
			}
		})
	}
}
