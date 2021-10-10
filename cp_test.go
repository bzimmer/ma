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
		before   func(runtime *ma.Runtime) error
		after    func(a *assert.Assertions, runtime *ma.Runtime) error
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
			before: func(runtime *ma.Runtime) error {
				return runtime.Fs.MkdirAll("/foo/bar", 0777)
			},
		},
		{
			name: "hidden files",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories": 1,
				"ma.cp.skip.hidden":         2,
			},
			before: func(runtime *ma.Runtime) error {
				if err := runtime.Fs.MkdirAll("/foo/bar", 0777); err != nil {
					return err
				}
				if _, err := runtime.Fs.Create("/foo/bar/.something"); err != nil {
					return err
				}
				if _, err := runtime.Fs.Create("/foo/bar/.else"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "filename with no extension",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories":     1,
				"ma.cp.skip.unsupported.<none>": 1,
			},
			before: func(runtime *ma.Runtime) error {
				if err := runtime.Fs.MkdirAll("/foo/bar", 0777); err != nil {
					return err
				}
				if _, err := runtime.Fs.Create("/foo/bar/something"); err != nil {
					return err
				}
				return nil
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
			before: func(runtime *ma.Runtime) error {
				if err := runtime.Fs.MkdirAll("/foo/bar/boo", 0777); err != nil {
					return err
				}
				if _, err := runtime.Fs.Create("/foo/bar/DSC18920.UKN"); err != nil {
					return err
				}
				if _, err := runtime.Fs.Create("/foo/bar/schedule.txt"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			name: "single image dng",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories": 1,
			},
			before: func(runtime *ma.Runtime) error {
				image, err := createTestFile(runtime.Fs)
				if err != nil {
					return err
				}
				// a bit of hack to test reading the entire contents of a .dng file
				// the exif parser doesn't care about file extensions, it sees only bytes
				name := image.Name()
				name = strings.Replace(name, ".jpg", ".dng", 1)
				return runtime.Fs.Rename(image.Name(), name)
			},
		},
		{
			name: "single image",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories": 1,
			},
			before: func(runtime *ma.Runtime) error {
				image, err := createTestFile(runtime.Fs)
				if err != nil {
					return err
				}

				t := time.Date(2008, time.March, 15, 11, 22, 0, 0, time.Local)
				if err := runtime.Fs.Chtimes(image.Name(), t, t); err != nil {
					return err
				}

				dst, err := runtime.Fs.Stat(image.Name())
				if err != nil {
					return err
				}

				log.Info().Str("src", image.Name()).Time("dst", dst.ModTime()).Msg("set test modification times")
				return nil
			},
			after: func(a *assert.Assertions, runtime *ma.Runtime) error {
				dst, err := runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				if err != nil {
					return err
				}
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
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories": 1,
				"ma.cp.skip.exists":         1,
			},
			before: func(runtime *ma.Runtime) error {
				if err := runtime.Fs.MkdirAll("/foo/bar", 0777); err != nil {
					return err
				}
				for _, filename := range []string{"/foo/bar/Nikon_D70.xmp", "/foo/baz/2008/2008-03/15/Nikon_D70.jpg"} {
					err := func() error {
						if _, err := runtime.Fs.Create(filename); err != nil {
							return err
						}
						image, err := runtime.Fs.Create("/foo/bar/Nikon_D70.jpg")
						if err != nil {
							return err
						}
						defer image.Close()
						fp, err := os.Open("testdata/Nikon_D70.jpg")
						if err != nil {
							return err
						}
						defer fp.Close()
						if _, err := io.Copy(image, fp); err != nil {
							return err
						}
						return image.Sync()

					}()
					if err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			name: "image + xmp",
			args: []string{"ma", "cp", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories": 1,
			},
			before: func(runtime *ma.Runtime) error {
				_, err := createTestFile(runtime.Fs)
				if err != nil {
					return err
				}
				_, err = runtime.Fs.Create("/foo/bar/Nikon_D70.xmp")
				return err

			},
			after: func(a *assert.Assertions, runtime *ma.Runtime) error {
				_, err := runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				if err != nil {
					return err
				}
				_, err = runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.xmp")
				return err
			},
		},
		{
			name: "image + xmp dry-run",
			args: []string{"ma", "cp", "-n", "/foo/bar", "/foo/baz"},
			counters: map[string]int{
				"ma.cp.visited.directories": 1,
				"ma.cp.file.dryrun":         2,
			},
			before: func(runtime *ma.Runtime) error {
				_, err := createTestFile(runtime.Fs)
				if err != nil {
					return err
				}
				_, err = runtime.Fs.Create("/foo/bar/Nikon_D70.xmp")
				return err

			},
			after: func(a *assert.Assertions, runtime *ma.Runtime) error {
				_, err := runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				a.Error(err)
				a.True(os.IsNotExist(err))
				_, err = runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.xmp")
				a.Error(err)
				a.True(os.IsNotExist(err))
				return nil
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
			before: func(runtime *ma.Runtime) error {
				_, err := createTestFile(runtime.Fs)
				if err != nil {
					return err
				}
				if err := runtime.Fs.MkdirAll("/foo/bar/boo", 0777); err != nil {
					return err
				}
				_, err = runtime.Fs.Create("/foo/bar/boo/Nikon_D70.xmp")
				return err
			},
			after: func(a *assert.Assertions, runtime *ma.Runtime) error {
				_, err := runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.jpg")
				if err != nil {
					return err
				}
				_, err = runtime.Fs.Stat("/foo/baz/2008/2008-03/15/Nikon_D70.xmp")
				a.Error(err)
				a.True(os.IsNotExist(err))
				return nil
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
			before: func(runtime *ma.Runtime) error {
				fp, err := runtime.Fs.Create("/foo/bar/Nikon_D70.jpg")
				a.NoError(err)
				a.NoError(copyFile(fp, "testdata/user_cmac.json"))
				return nil
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			app := NewTestApp(t, tt.name, ma.CommandCopy())

			if tt.before != nil {
				a.NoError(tt.before(runtime(app)))
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
				a.NoError(tt.after(a, runtime(app)))
			}
		})
	}
}
