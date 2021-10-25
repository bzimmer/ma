package ma

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/armon/go-metrics"
	"github.com/rs/zerolog/log"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

const (
	defaultBufferSize = 1024 * 1024
	defaultDateFormat = "2006/2006-01/02"
	dng               = ".dng"
	jpeg              = ".jpeg"
	jpg               = ".jpg"
	nef               = ".nef"
	orf               = ".orf"
	raf               = ".raf"
	xmp               = ".xmp"
)

func defaultImages() []string {
	return []string{raf, nef, dng, jpg, jpeg}
}

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

func (b *dateTimeExif) bufferSize() int64 {
	switch b.ext {
	case orf, dng, nef:
		return b.info.Size()
	default:
		return defaultBufferSize
	}
}

func (b *dateTimeExif) dateTime() (time.Time, error) {
	fp, err := b.fs.Open(filepath.Join(b.src, b.info.Name()))
	if err != nil {
		return time.Time{}, err
	}
	defer fp.Close()
	data := make([]byte, b.bufferSize())
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

func (f *fileSet) dateTime(filesystem afero.Fs, dirname string) (time.Time, error) {
	// for every file in the fileset attempt to create a time.Time
	times := make(map[string]time.Time)
	for i := range f.files {
		info := f.files[i]
		ext := strings.ToLower(filepath.Ext(info.Name()))
		switch ext {
		case jpg, jpeg, raf, dng, nef:
			dt := &dateTimeExif{fs: filesystem, src: dirname, ext: ext, info: info}
			t, err := dt.dateTime()
			if err != nil {
				return time.Time{}, err
			}
			times[ext] = t
		case ".mp4", ".mov", ".avi":
			// @todo(movies)
		case orf:
			// @todo(orf)
		case "", xmp:
			// not trustworthy for valid dates
		}
	}

	// in priority order, find the first non-zero time.Time
	for _, ext := range defaultImages() {
		t, ok := times[ext]
		if ok {
			return t, nil
		}
	}

	// found no time
	return time.Time{}, nil
}

type entangle struct {
	source  string
	fileSet *fileSet
}

type entangler struct {
	fs          afero.Fs
	metrics     *metrics.Metrics
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
			if err := afero.Walk(c.fs, sources[i], c.fileSets(sets)); err != nil {
				return err
			}
		}
		for dirname, filesets := range sets {
			for _, fileset := range filesets {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case q <- &entangle{source: dirname, fileSet: fileset}:
					c.metrics.IncrCounter([]string{"cp", "filesets"}, 1)
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
			c.metrics.IncrCounter([]string{"cp", "fileset", "attempt"}, 1)
			dt, err := ent.fileSet.dateTime(c.fs, ent.source)
			if err != nil {
				c.metrics.IncrCounter([]string{"cp", "fileset", "failed", "exif"}, 1)
				return err
			}
			if dt.IsZero() {
				c.metrics.IncrCounter([]string{"cp", "fileset", "skip", "unsupported"}, 1)
				for i := range ent.fileSet.files {
					filename := ent.fileSet.files[i].Name()
					ext := filepath.Ext(filename)
					ext = strings.TrimPrefix(ext, ".")
					if ext == "" {
						ext = "<none>"
					}
					log.Warn().Str("filename", filename).Str("reason", "unsupported."+ext).Msg("skip")
					c.metrics.IncrCounter([]string{"cp", "skip", "unsupported", ext}, 1)
				}
				continue
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

func (c *entangler) copy(src, dst string) error {
	dirname, _ := filepath.Split(dst)
	if err := c.fs.MkdirAll(dirname, 0755); err != nil {
		return err
	}
	out, err := c.fs.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	in, err := c.fs.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	err = out.Sync()
	if err != nil {
		return err
	}
	info, err := in.Stat()
	if err != nil {
		return err
	}
	mtime := info.ModTime()
	return c.fs.Chtimes(dst, mtime, mtime)
}

func (c *entangler) copyFile(source, destination string) error {
	defer c.metrics.MeasureSince([]string{"cp", "elapsed", "file"}, time.Now())
	c.metrics.IncrCounter([]string{"cp", "file", "attempt"}, 1)
	stat, err := c.fs.Stat(destination)
	if err != nil {
		if !errors.Is(err, afero.ErrFileNotFound) {
			return err
		}
	}
	if stat != nil {
		c.metrics.IncrCounter([]string{"cp", "skip", "exists"}, 1)
		log.Info().Str("src", source).Str("dst", destination).Str("reason", "exists").Msg("skip")
		return nil
	}
	log.Info().Str("src", source).Str("dst", destination).Msg("cp")
	if c.dryrun {
		c.metrics.IncrCounter([]string{"cp", "file", "dryrun"}, 1)
		return nil
	}
	if err := c.copy(source, destination); err != nil {
		c.metrics.IncrCounter([]string{"cp", "file", "failed"}, 1)
		return err
	}
	c.metrics.IncrCounter([]string{"cp", "file", "success"}, 1)
	return nil
}

// fileSets creates fileSets from a directory traversal
func (c *entangler) fileSets(sets map[string]map[string]*fileSet) filepath.WalkFunc {
	return func(fullname string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			c.metrics.IncrCounter([]string{"cp", "visited", "directories"}, 1)
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			c.metrics.IncrCounter([]string{"cp", "skip", "hidden"}, 1)
			return nil
		}
		c.metrics.IncrCounter([]string{"cp", "visited", "files"}, 1)

		dirname, basename := split(fullname)
		dirs, ok := sets[dirname]
		if !ok {
			dirs = make(map[string]*fileSet)
			sets[dirname] = dirs
		}
		filesystem, ok := dirs[basename]
		if !ok {
			filesystem = new(fileSet)
			dirs[basename] = filesystem
		}
		filesystem.add(info)

		return nil
	}
}

func cp(c *cli.Context) error {
	if c.NArg() < 2 {
		return fmt.Errorf("expected 2+ arguments, not {%d}", c.NArg())
	}
	defer runtime(c).Metrics.MeasureSince([]string{"cp", "elapsed"}, time.Now())
	en := &entangler{
		fs:          runtime(c).Fs,
		metrics:     runtime(c).Metrics,
		concurrency: c.Int("concurrency"),
		dryrun:      c.Bool("dryrun"),
		dateFormat:  c.String("format"),
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
		Name:        "cp",
		HelpName:    "cp",
		Usage:       "copy files to a the directory structure of `--format`",
		Description: "copy files from a source(s) to a destination using the Exif format to create the directory structure",
		ArgsUsage:   "<file-or-directory> [, <file-or-directory>] <file-or-directory>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "dryrun",
				Aliases:  []string{"n"},
				Usage:    "prepare to copy but don't actually do it",
				Value:    false,
				Required: false,
			},
			&cli.StringFlag{
				Name:     "format",
				Usage:    "the date format used for the destination directory",
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
