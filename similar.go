package ma

import (
	"context"
	"errors"
	goimage "image"
	"io/fs"
	"strconv"

	"github.com/armon/go-metrics"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"github.com/vitali-fedulov/images3"
	"golang.org/x/sync/errgroup"
)

type analyzer struct {
	afs afero.Fs
	met *metrics.Metrics
	enc Encoder
}

func (y *analyzer) icon(path string) (images3.IconT, error) {
	file, err := y.afs.Open(path)
	if err != nil {
		return images3.IconT{}, err
	}
	defer file.Close()
	img, _, err := goimage.Decode(file)
	if err != nil {
		return images3.IconT{}, err
	}
	return images3.Icon(img, path), nil
}

func (y *analyzer) icons(ctx context.Context, paths <-chan string, icons chan<- images3.IconT) error {
	for path := range paths {
		log.Info().Str("path", path).Msg("reading")
		icon, err := y.icon(path)
		if err != nil {
			if !errors.Is(err, goimage.ErrFormat) {
				log.Error().Err(err).Str("path", path).Msg("icons")
				y.met.IncrCounter([]string{"similar", "icon", "error"}, 1)
				return err
			}
			log.Warn().Err(err).Str("path", path).Msg("reading")
			y.met.IncrCounter([]string{"similar", "icon", "skipped"}, 1)
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case icons <- icon:
			y.met.IncrCounter([]string{"similar", "icon", "success"}, 1)
		}
	}
	return nil
}

func (y *analyzer) paths(ctx context.Context, paths chan<- string, root ...string) error {
	f := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		select {
		case <-ctx.Done():
			log.Error().Err(ctx.Err()).Msg("paths")
			return ctx.Err()
		case paths <- path:
			y.met.IncrCounter([]string{"similar", "path"}, 1)
		}
		return nil
	}
	for i := 0; i < len(root); i++ {
		if err := afero.Walk(y.afs, root[i], f); err != nil {
			return err
		}
	}
	return nil
}

func (y *analyzer) analyze(iconss []images3.IconT) error {
	n := len(iconss)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			b := images3.Similar(iconss[i], iconss[j])
			if b {
				log.Info().
					Str("A", iconss[i].Path).
					Str("B", iconss[j].Path).
					Bool("similar", b).
					Msg("similar")
			}
			if err := y.enc.Encode(map[string]any{
				"A":       iconss[i].Path,
				"B":       iconss[j].Path,
				"Similar": b}); err != nil {
				return err
			}
			y.met.IncrCounter([]string{"similar", "analyze", strconv.FormatBool(b)}, 1)
		}
	}
	return nil
}

func similar(c *cli.Context) error {
	y := &analyzer{
		afs: runtime(c).Fs,
		met: runtime(c).Metrics,
		enc: runtime(c).Encoder,
	}

	pathsc := make(chan string)
	iconsc := make(chan images3.IconT)
	igrp, ictx := errgroup.WithContext(c.Context)
	for i := 0; i < c.Int("concurrency"); i++ {
		igrp.Go(func() error {
			defer func() {
				log.Debug().Msg("done creating icons")
			}()
			return y.icons(ictx, pathsc, iconsc)
		})
	}
	grp, ctx := errgroup.WithContext(ictx)
	grp.Go(func() error {
		defer close(iconsc)
		defer func() {
			log.Debug().Msg("closing icon channel")
		}()
		return igrp.Wait()
	})
	grp.Go(func() error {
		defer close(pathsc)
		defer func() {
			log.Debug().Msg("closing path channel")
		}()
		return y.paths(ctx, pathsc, c.Args().Slice()...)
	})

	icons := make([]images3.IconT, 0)
	grp.Go(func() error {
		for icon := range iconsc {
			log.Debug().Str("path", icon.Path).Msg("gathering")
			icons = append(icons, icon)
		}
		log.Info().Int("num", len(icons)).Msg("images")
		return nil
	})
	if err := grp.Wait(); err != nil {
		return err
	}
	return y.analyze(icons)
}

func CommandSimilar() *cli.Command {
	return &cli.Command{
		Name:        "similar",
		HelpName:    "similar",
		Usage:       "identify similar images",
		ArgsUsage:   "FILE-OR-DIRECTORY, [FILE-OR-DIRECTORY, ...]",
		Description: "identifies similar images",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "concurrency",
				Aliases: []string{"c"},
				Usage:   "the number of concurrent image reads",
				Value:   4,
			},
		},
		Action: similar,
	}
}
