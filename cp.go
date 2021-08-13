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
	"github.com/otiai10/copy"
	"github.com/rs/zerolog/log"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	xmpexif "trimmer.io/go-xmp/models/exif"
	"trimmer.io/go-xmp/xmp"
)

const (
	defaultBufferSize = 1024 * 1024
	defaultDateFormat = "2006/2006-01/02"
)

type dateTime interface {
	dateTime() (time.Time, error)
}

type dateTimeXmp struct {
	fs  afero.Fs
	src string
}

func (b *dateTimeXmp) dateTime() (time.Time, error) {
	fp, err := b.fs.Open(b.src)
	if err != nil {
		return time.Time{}, nil
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return time.Time{}, nil
	}
	var document xmp.Document
	if err := xmp.Unmarshal(data, &document); err != nil {
		return time.Time{}, nil
	}
	x := xmpexif.FindModel(&document)
	if x == nil {
		return time.Time{}, nil
	}
	return x.DateTimeOriginal.Value(), nil
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
	fp, err := b.fs.Open(b.src)
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

type req struct {
	src string
	dst string
}

type copyFunc func(src, dest string) error

type copier struct {
	fs         afero.Fs
	dryrun     bool
	dateFormat string
	concurrent int
	copyFunc   copyFunc
	metric     *metrics.Metrics
}

func (c *copier) dateTime(src string, info fs.FileInfo) dateTime {
	if strings.HasPrefix(info.Name(), ".") {
		c.metric.IncrCounter([]string{"cp", "skip", "unsupported", "hidden"}, 1)
		return nil
	}
	ext := strings.ToLower(filepath.Ext(info.Name()))
	switch ext {
	case ".jpg", ".raf", ".dng", ".nef", ".jpeg":
		return &dateTimeExif{fs: c.fs, src: src, ext: ext, info: info}
	case ".xmp":
		return &dateTimeXmp{fs: c.fs, src: src}
	case "":
		ext = ".<none>"
	case ".mp4", ".mov", ".avi":
		// @todo(movies)
	case ".orf":
		// @todo(orf)
	default:
	}
	ext = ext[1:]
	log.Info().Str("src", src).Str("reason", "unsupported").Str("ext", ext).Msg("skip")
	c.metric.IncrCounter([]string{"cp", "skip", "unsupported", ext}, 1)
	return nil
}

func (c *copier) copy(q req) error {
	log.Info().Str("src", q.src).Str("dst", q.dst).Msg("cp")
	if c.dryrun {
		c.metric.IncrCounter([]string{"cp", "dryrun"}, 1)
		return nil
	}
	if err := c.copyFunc(q.src, q.dst); err != nil {
		c.metric.IncrCounter([]string{"cp", "failed"}, 1)
		return err
	}
	c.metric.IncrCounter([]string{"cp", "success"}, 1)
	return nil
}

func (c *copier) walker(ctx context.Context, q chan<- req, dest string) filepath.WalkFunc {
	return func(src string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			c.metric.IncrCounter([]string{"cp", "visited", "directories"}, 1)
			return nil
		}
		c.metric.IncrCounter([]string{"cp", "visited", "files"}, 1)

		dt := c.dateTime(src, info)
		if dt == nil {
			return nil // not an error but not supported
		}

		tm, err := dt.dateTime()
		if err != nil {
			log.Error().Err(err).Str("src", src).Msg("dateTime")
			return err
		}

		dst := filepath.Join(dest, tm.Format(c.dateFormat), info.Name())
		stat, err := c.fs.Stat(dst)
		if err != nil {
			if !errors.Is(err, afero.ErrFileNotFound) {
				return err
			}
		}
		if stat != nil {
			c.metric.IncrCounter([]string{"cp", "skip", "exists"}, 1)
			log.Info().Str("src", src).Str("dst", dst).Str("reason", "exists").Msg("skip")
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case q <- req{src: src, dst: dst}:
			c.metric.IncrCounter([]string{"cp", "attempt"}, 1)
		}
		return nil
	}
}

func (c *copier) cp(ctx context.Context, root, dest string) error {
	q := make(chan req)
	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		defer close(q)
		return afero.Walk(c.fs, root, c.walker(ctx, q, dest))
	})
	log.Info().Str("root", root).Int("concurrent", c.concurrent).Msg("cp")
	for i := 0; i < c.concurrent; i++ {
		grp.Go(func() error {
			for x := range q {
				if err := c.copy(x); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return grp.Wait()
}

func cp(c *cli.Context) error {
	met, err := metric(c)
	if err != nil {
		return err
	}
	opts := copy.Options{
		Sync:          true,
		PreserveTimes: true,
		Skip: func(string) (bool, error) {
			return false, nil // Don't skip
		},
		OnDirExists: func(src, dest string) copy.DirExistsAction {
			return copy.Merge
		},
	}
	cpr := &copier{
		fs:     afero.NewOsFs(),
		metric: met,
		copyFunc: func(src, dst string) error {
			return copy.Copy(src, dst, opts)
		},
		dryrun:     c.Bool("dryrun"),
		dateFormat: c.String("format"),
		concurrent: c.Int("concurrent"),
	}
	n := c.NArg()
	dest := c.Args().Get(n - 1)
	for i := 0; i < n-1; i++ {
		if err := cpr.cp(c.Context, c.Args().Get(i), dest); err != nil {
			return err
		}
	}
	return nil
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
				Name:    "concurrent",
				Aliases: []string{"c"},
				Usage:   "the number of concurrent copies",
				Value:   2,
			},
		},
		Before: func(c *cli.Context) error {
			if c.NArg() < 2 {
				return fmt.Errorf("expected 2+ arguments, not {%d}", c.NArg())
			}
			return nil
		},
		Action: cp,
		After:  stats,
	}
}
