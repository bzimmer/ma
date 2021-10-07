package ma

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/armon/go-metrics"
	"github.com/otiai10/copy"
	"github.com/rs/zerolog/log"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

const (
	defaultBufferSize = 1024 * 1024
	defaultDateFormat = "2006/2006-01/02"
)

var defaultImages = []string{".raf", ".nef", ".dng", ".jpg", ".jpeg"}

func split(fullname string) (dirname, basename string) {
	dirname, filename := filepath.Split(fullname)
	n := strings.LastIndexFunc(filename, func(s rune) bool {
		return s == '.'
	})
	switch n {
	case -1:
		basename = filename
	default:
		basename = filename[0:n]
	}
	dirname = filepath.Clean(dirname)
	return
}

type dateTimeExif struct {
	fs   afero.Fs
	src  string
	ext  string
	info fs.FileInfo
}

func (b *dateTimeExif) bufferSize() (int64, error) {
	switch b.ext {
	case ".orf", ".dng", ".nef":
		return b.info.Size(), nil
	default:
		return defaultBufferSize, nil
	}
}

func (b *dateTimeExif) dateTime() (time.Time, error) {
	fp, err := b.fs.Open(filepath.Join(b.src, b.info.Name()))
	if err != nil {
		return time.Time{}, err
	}
	defer fp.Close()
	size, err := b.bufferSize()
	if err != nil {
		return time.Time{}, err
	}
	data := make([]byte, size)
	_, err = fp.Read(data)
	if err != nil {
		return time.Time{}, err
	}
	x, err := exif.Decode(bytes.NewBuffer(data))
	if err != nil {
		return time.Time{}, err
	}
	tm, err := x.DateTime()
	if err != nil {
		return time.Time{}, err
	}
	return tm, err
}

type fileSet struct {
	files []fs.FileInfo
}

func (f *fileSet) add(info fs.FileInfo) {
	f.files = append(f.files, info)
}

func (f *fileSet) dateTime(fs afero.Fs, dirname string) (time.Time, error) {
	// for every file in the fileset attempt to create a time.Time
	times := make(map[string]time.Time)
	for i := range f.files {
		info := f.files[i]
		ext := strings.ToLower(filepath.Ext(info.Name()))
		switch ext {
		case ".jpg", ".jpeg", ".raf", ".dng", ".nef":
			dt := &dateTimeExif{fs: fs, src: dirname, ext: ext, info: info}
			t, err := dt.dateTime()
			if err != nil {
				return time.Time{}, err
			}
			times[ext] = t
		case ".mp4", ".mov", ".avi":
			// @todo(movies)
		case ".orf":
			// @todo(orf)
		case "", ".xmp":
			// not trustworthy for valid dates
		}
	}

	// in priority order, find the first non-zero time.Time
	for _, ext := range defaultImages {
		t, ok := times[ext]
		if ok {
			return t, nil
		}
	}

	// found no time
	return time.Time{}, nil
}

type copyFunc func(src, dest string) error

type entangle struct {
	source  string
	fileSet *fileSet
}

type entangler struct {
	fs          afero.Fs
	metric      *metrics.Metrics
	copyFunc    copyFunc
	concurrency int
	dryrun      bool
	dateFormat  string
}

func (c *entangler) cp(ctx context.Context, sources []string, destination string) error {
	q := make(chan *entangle)
	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		defer close(q)
		sets := make(map[string]map[string]*fileSet)
		for i := range sources {
			if err := afero.Walk(c.fs, sources[i], c.filesets(sets)); err != nil {
				return err
			}
		}
		for dirname, filesets := range sets {
			for _, fileset := range filesets {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case q <- &entangle{source: dirname, fileSet: fileset}:
					c.metric.IncrCounter([]string{"cp", "filesets"}, 1)
				}
			}
		}
		return nil
	})
	for i := 0; i < c.concurrency; i++ {
		grp.Go(c.copyFileSet(q, destination))
	}
	return grp.Wait()
}

func (c *entangler) copyFileSet(q <-chan *entangle, destination string) func() error {
	return func() error {
		for ent := range q {
			c.metric.IncrCounter([]string{"cp", "fileset", "attempt"}, 1)
			dt, err := ent.fileSet.dateTime(c.fs, ent.source)
			if err != nil {
				return err
			}
			if dt.IsZero() {
				c.metric.IncrCounter([]string{"cp", "fileset", "skip", "unsupported"}, 1)
				for i := range ent.fileSet.files {
					filename := ent.fileSet.files[i].Name()
					ext := filepath.Ext(filename)
					ext = strings.TrimPrefix(ext, ".")
					if ext == "" {
						ext = "<none>"
					}
					log.Warn().Str("filename", filename).Str("reason", "unsupported."+ext).Msg("skip")
					c.metric.IncrCounter([]string{"cp", "skip", "unsupported", ext}, 1)
				}
			}
			df := dt.Format(c.dateFormat)
			for i := range ent.fileSet.files {
				src := filepath.Join(ent.source, ent.fileSet.files[i].Name())
				dst := filepath.Join(destination, df, ent.fileSet.files[i].Name())
				if err := c.copyFile(src, dst); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func (c *entangler) copyFile(source, destination string) error {
	defer c.metric.MeasureSince([]string{"cp", "elapsed", "file"}, time.Now())
	c.metric.IncrCounter([]string{"cp", "file", "attempt"}, 1)
	stat, err := c.fs.Stat(destination)
	if err != nil {
		if !errors.Is(err, afero.ErrFileNotFound) {
			return err
		}
	}
	if stat != nil {
		c.metric.IncrCounter([]string{"cp", "skip", "exists"}, 1)
		log.Info().Str("src", source).Str("dst", destination).Str("reason", "exists").Msg("skip")
		return nil
	}
	log.Info().Str("src", source).Str("dst", destination).Msg("cp")
	if c.dryrun {
		c.metric.IncrCounter([]string{"cp", "file", "dryrun"}, 1)
		return nil
	}
	if err := c.copyFunc(source, destination); err != nil {
		c.metric.IncrCounter([]string{"cp", "file", "failed"}, 1)
		return err
	}
	c.metric.IncrCounter([]string{"cp", "file", "success"}, 1)
	return nil
}

// filesets creates fileSets from a directory traversal
func (c *entangler) filesets(sets map[string]map[string]*fileSet) filepath.WalkFunc {
	return func(fullname string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			c.metric.IncrCounter([]string{"cp", "visited", "directories"}, 1)
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			c.metric.IncrCounter([]string{"cp", "skip", "hidden"}, 1)
			return nil
		}
		c.metric.IncrCounter([]string{"cp", "visited", "files"}, 1)

		dirname, basename := split(fullname)
		dirs, ok := sets[dirname]
		if !ok {
			dirs = make(map[string]*fileSet)
			sets[dirname] = dirs
		}
		fs, ok := dirs[basename]
		if !ok {
			fs = new(fileSet)
			dirs[basename] = fs
		}
		fs.add(info)

		return nil
	}
}

func cp(c *cli.Context) error {
	if c.NArg() < 2 {
		return fmt.Errorf("expected 2+ arguments, not {%d}", c.NArg())
	}

	opts := copy.Options{
		Sync:          true,
		PreserveTimes: true,
		Skip: func(string) (bool, error) {
			return false, nil // Never skip
		},
		OnDirExists: func(src, dest string) copy.DirExistsAction {
			return copy.Merge
		},
	}

	defer metric(c).MeasureSince([]string{"cp", "elapsed"}, time.Now())
	en := &entangler{
		fs:          runtime(c).Fs,
		metric:      metric(c),
		concurrency: c.Int("concurrency"),
		copyFunc: func(src, dst string) error {
			return copy.Copy(src, dst, opts)
		},
		dryrun:     c.Bool("dryrun"),
		dateFormat: c.String("format"),
	}
	args := c.Args().Slice()
	destination, err := filepath.Abs(args[len(args)-1])
	if err != nil {
		return err
	}
	return en.cp(c.Context, args[0:len(args)-1], destination)
}

func CommandCopy() *cli.Command {
	return &cli.Command{
		Name:  "cp",
		Usage: "copy files to a pre-determined directory structure",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "dryrun",
				Aliases:  []string{"n"},
				Value:    false,
				Required: false,
			},
			&cli.StringFlag{
				Name:     "format",
				Value:    defaultDateFormat,
				Required: false,
			},
			&cli.IntFlag{
				Name:    "concurrency",
				Aliases: []string{"c"},
				Usage:   "the number of concurrent copies",
				Value:   2,
			},
		},
		Action: cp,
	}
}
