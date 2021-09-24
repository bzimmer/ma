package ma

import (
	"github.com/bzimmer/smugmug"
	"github.com/bzimmer/smugmug/uploadable/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

func visit(c *cli.Context) filesystem.PreFunc {
	return func(fs afero.Fs, filename string) (bool, error) {
		metric(c).IncrCounter([]string{"fsUploadable", "visit"}, 1)
		return true, nil
	}
}

func extensions(c *cli.Context) filesystem.PreFunc {
	f := filesystem.Extensions(c.StringSlice("ext")...)
	return func(fs afero.Fs, filename string) (bool, error) {
		ok, err := f(fs, filename)
		if err != nil {
			return false, err
		}
		if !ok {
			metric(c).IncrCounter([]string{"fsUploadable", "skip", "unsupported"}, 1)
			log.Info().Str("reason", "unsupported").Str("path", filename).Msg("skipping")
		}
		return ok, err
	}
}

func open(c *cli.Context) filesystem.UseFunc {
	return func(up *smugmug.Uploadable) (*smugmug.Uploadable, error) {
		metric(c).IncrCounter([]string{"fsUploadable", "open"}, 1)
		return up, nil
	}
}

func skip(c *cli.Context, images map[string]*smugmug.Image) filesystem.UseFunc {
	f := filesystem.Skip(false, images)
	return func(up *smugmug.Uploadable) (*smugmug.Uploadable, error) {
		sup, err := f(up)
		if err != nil {
			return nil, err
		}
		if sup == nil {
			metric(c).IncrCounter([]string{"fsUploadable", "skip", "md5"}, 1)
			log.Info().Str("reason", "md5").Str("path", up.Name).Msg("skipping")
			return nil, err
		}
		return sup, nil
	}
}

func replace(c *cli.Context, images map[string]*smugmug.Image) filesystem.UseFunc {
	f := filesystem.Replace(true, images)
	return func(up *smugmug.Uploadable) (*smugmug.Uploadable, error) {
		up, err := f(up)
		if err != nil {
			return nil, err
		}
		if up == nil {
			return nil, nil
		}
		if up.Replaces != "" {
			metric(c).IncrCounter([]string{"fsUploadable", "replace"}, 1)
		}
		return up, err
	}
}

func upload(c *cli.Context) filesystem.UseFunc {
	return func(up *smugmug.Uploadable) (*smugmug.Uploadable, error) {
		info := log.Info().
			Str("name", up.Name).
			Str("album", up.AlbumKey).
			Str("replaces", up.Replaces)
		if c.Bool("dryrun") {
			info.Str("status", "dryrun").Msg("upload")
			metric(c).IncrCounter([]string{"upload", "dryrun"}, 1)
			return nil, nil
		}
		info.Str("status", "attempt").Msg("upload")
		metric(c).IncrCounter([]string{"upload", "attempt"}, 1)
		return up, nil
	}
}

func up(c *cli.Context) error {
	mg := client(c)
	albumKey := c.String("album")
	images := make(map[string]*smugmug.Image)
	log.Info().Msg("querying existing gallery images")
	if err := client(c).Image.ImagesIter(c.Context, albumKey, func(img *smugmug.Image) (bool, error) {
		images[img.FileName] = img
		return true, nil
	}); err != nil {
		return err
	}
	log.Info().Int("count", len(images)).Msg("existing gallery images")

	u, err := filesystem.NewFsUploadable(albumKey)
	if err != nil {
		return err
	}
	u.Pre(visit(c), extensions(c))
	u.Use(open(c), skip(c, images), replace(c, images), upload(c))

	grp, ctx := errgroup.WithContext(c.Context)

	albumbc := make(chan *smugmug.Album, 1)
	grp.Go(func() error {
		defer close(albumbc)
		album, err := mg.Album.Album(ctx, albumKey)
		if err != nil {
			return err
		}
		albumbc <- album
		return nil
	})
	grp.Go(func() error {
		ups := filesystem.NewFsUploadables(afs(c), c.Args().Slice(), u)
		uploadc, errc := mg.Upload.Uploads(ctx, ups)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case err := <-errc:
				return err
			case up, ok := <-uploadc:
				if !ok {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case album := <-albumbc:
						log.Info().Str("albumKey", album.AlbumKey).Str("webURI", album.WebURI).Msg("complete")
					}
					return nil
				}
				metric(c).IncrCounter([]string{"upload", "success"}, 1)
				metric(c).AddSample([]string{"upload", "upload"}, float32(up.Elapsed.Seconds()))
				log.Info().
					Str("name", up.Uploadable.Name).
					Str("album", up.Uploadable.AlbumKey).
					Dur("elapsed", up.Elapsed).
					Str("uri", up.ImageURI).
					Str("status", "success").
					Msg("upload")
			}
		}
	})
	return grp.Wait()
}

func CommandUp() *cli.Command {
	return &cli.Command{
		Name:    "up",
		Aliases: []string{"upload"},
		Usage:   "upload images to SmugMug",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "album",
				Aliases:  []string{"a"},
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "ext",
				Aliases:  []string{"x"},
				Required: false,
				Value:    cli.NewStringSlice(".jpg"),
			},
			&cli.BoolFlag{
				Name:     "dryrun",
				Aliases:  []string{"n"},
				Value:    false,
				Required: false,
			},
		},
		Action: up,
		After:  stats,
	}
}
