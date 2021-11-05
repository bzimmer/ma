package ma

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/armon/go-metrics"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

const defaultDateFormat = "2006/2006-01/02"

func split(fullname string) identifier {
	dirname, filename := filepath.Split(fullname)
	n := strings.LastIndexFunc(filename, func(s rune) bool {
		return s == '.'
	})
	var basename string
	switch n {
	case -1:
		basename = filename
	default:
		basename = filename[0:n]
	}
	return identifier{dirname: filepath.Clean(dirname), basename: basename}
}

type identifier struct {
	dirname  string
	basename string
}

type fileset struct {
	identifier identifier
	files      []fs.FileInfo
}

// dateTime attempts to create a time.Time for for every file in the fileset
func (f *fileset) dateTime(afs afero.Fs, ex Exif) (time.Time, error) {
	var infos []fs.FileInfo
	for i := range f.files {
		info := f.files[i]
		ext := strings.ToLower(filepath.Ext(info.Name()))
		switch ext {
		case "", ".xmp":
			// not trustworthy for valid dates
		default:
			infos = append(infos, info)
		}
	}
	if len(infos) == 0 {
		return time.Time{}, nil
	}
	var times []time.Time
	mds := ex.Extract(afs, f.identifier.dirname, infos...)
	for i := range mds {
		if mds[i].Err != nil {
			return time.Time{}, mds[i].Err
		}
		times = append(times, mds[i].DateTime)
	}
	sort.SliceStable(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})
	// @todo(bzimmer) ensure dates are consistent (within a ~second or so)
	return times[0], nil
}

type entangler struct {
	fs          afero.Fs
	exif        Exif
	metrics     *metrics.Metrics
	concurrency int
	dryrun      bool
	dateFormat  string
}

func (c *entangler) cp(ctx context.Context, sources []string, destination string) error {
	q := make(chan *fileset)
	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		defer close(q)
		sets := make(map[identifier][]fs.FileInfo)
		for i := range sources {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := afero.Walk(c.fs, sources[i], c.filesets(sets)); err != nil {
					return err
				}
			}
		}
		for id, files := range sets {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case q <- &fileset{identifier: id, files: files}:
				c.metrics.IncrCounter([]string{"cp", "filesets"}, 1)
			}
		}
		return nil
	})
	for i := 0; i < c.concurrency; i++ {
		grp.Go(c.copyFileset(q, destination))
	}
	return grp.Wait()
}

func (c *entangler) copyFileset(q <-chan *fileset, destination string) func() error {
	return func() error {
		for x := range q {
			c.metrics.IncrCounter([]string{"cp", "fileset", "attempt"}, 1)
			dt, err := x.dateTime(c.fs, c.exif)
			if err != nil {
				c.metrics.IncrCounter([]string{"cp", "fileset", "failed", "exif"}, 1)
				return err
			}
			if dt.IsZero() {
				c.metrics.IncrCounter([]string{"cp", "fileset", "skip", "unsupported"}, 1)
				for i := range x.files {
					filename := filepath.Join(x.identifier.dirname, x.files[i].Name())
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
			for i := range x.files {
				src := filepath.Join(x.identifier.dirname, x.files[i].Name())
				dst := filepath.Join(destination, df, x.files[i].Name())
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

// filesets creates filesets from a directory traversal
func (c *entangler) filesets(sets map[identifier][]fs.FileInfo) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrPermission) {
				c.metrics.IncrCounter([]string{"cp", "skip", "denied"}, 1)
				log.Warn().Str("path", path).Err(err).Msg("skip")
				return filepath.SkipDir
			}
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

		id := split(path)
		sets[id] = append(sets[id], info)

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
		exif:        runtime(c).Exif,
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
		Description: "copy files from a source(s) to a destination using the image date to layout the directory structure",
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
