package ma

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/armon/go-metrics"
	"github.com/otiai10/copy"
	"github.com/rs/zerolog/log"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/urfave/cli/v2"
	xmpexif "trimmer.io/go-xmp/models/exif"
	"trimmer.io/go-xmp/xmp"
)

// @todo(bzimmer) add support for orf, videos

const (
	defaultBufferSize = 1024 * 1024
	defaultDateFormat = "2006/2006-01/02"
)

type dateTime interface {
	dateTime() (time.Time, error)
}

type dateTimeXmp struct {
	src string
}

func (b *dateTimeXmp) dateTime() (time.Time, error) {
	fp, err := os.Open(b.src)
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
	src   string
	ext   string
	entry fs.DirEntry
}

func (b *dateTimeExif) bufferSize() (int64, error) {
	switch b.ext {
	case ".orf", ".dng", ".nef":
		info, err := b.entry.Info()
		if err != nil {
			return 0, err
		}
		return info.Size(), nil
	default:
		return defaultBufferSize, nil
	}
}

func (b *dateTimeExif) dateTime() (time.Time, error) {
	fp, err := os.Open(b.src)
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

type copier struct {
	dryrun     bool
	dateFormat string
	options    copy.Options
	metric     *metrics.Metrics
}

func (c *copier) dateTime(src string, d fs.DirEntry) dateTime {
	if strings.HasPrefix(d.Name(), ".") {
		c.metric.IncrCounter([]string{"cp", "skip", "unsupported", "hidden"}, 1)
		return nil
	}
	ext := strings.ToLower(filepath.Ext(d.Name()))
	switch ext {
	case ".jpg", ".raf", ".dng", ".nef", ".jpeg":
		return &dateTimeExif{src: src, ext: ext, entry: d}
	case ".xmp":
		return &dateTimeXmp{src: src}
	case ".mp4", ".mov", ".avi":
		fallthrough
	case ".orf":
		fallthrough
	default:
		ext = ext[1:]
		log.Info().Str("src", src).Str("reason", "unsupported").Str("ext", ext).Msg("skip")
		c.metric.IncrCounter([]string{"cp", "skip", "unsupported", ext}, 1)
		return nil
	}
}

func (c *copier) cp(root, dest string) error {
	return filepath.WalkDir(root, func(src string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			c.metric.IncrCounter([]string{"cp", "visited", "directories"}, 1)
			return nil
		}
		c.metric.IncrCounter([]string{"cp", "visited", "files"}, 1)

		dt := c.dateTime(src, d)
		if dt == nil {
			return nil // not an error but not supported
		}

		tm, err := dt.dateTime()
		if err != nil {
			log.Error().Err(err).Str("src", src).Msg("dateTime")
			return err
		}

		dst := filepath.Join(dest, tm.Format(c.dateFormat), d.Name())
		stat, err := os.Stat(dst)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		if stat != nil {
			c.metric.IncrCounter([]string{"cp", "skip", "exists"}, 1)
			log.Info().Str("src", src).Str("dst", dst).Str("reason", "exists").Msg("skip")
			return nil
		}

		log.Info().Str("src", src).Str("dst", dst).Msg("cp")
		if c.dryrun {
			c.metric.IncrCounter([]string{"cp", "dryrun"}, 1)
			return nil
		}
		c.metric.IncrCounter([]string{"cp", "attempt"}, 1)
		if err := copy.Copy(src, dst, c.options); err != nil {
			c.metric.IncrCounter([]string{"cp", "failed"}, 1)
			return err
		}
		c.metric.IncrCounter([]string{"cp", "success"}, 1)
		return nil
	})
}

func cp(c *cli.Context) error {
	met, err := metric(c)
	if err != nil {
		return err
	}
	n := c.NArg()
	cpr := &copier{
		metric:     met,
		dateFormat: defaultDateFormat,
		dryrun:     c.Bool("dryrun"),
		options: copy.Options{
			PreserveTimes: true,
			Skip: func(string) (bool, error) {
				return false, nil // Don't skip
			},
			OnDirExists: func(src, dest string) copy.DirExistsAction {
				return copy.Merge
			},
		}}
	dest := c.Args().Get(n - 1)
	for i := 0; i < n-1; i++ {
		if err := cpr.cp(c.Args().Get(i), dest); err != nil {
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
