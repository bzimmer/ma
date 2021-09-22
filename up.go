package ma

import (
	"github.com/bzimmer/smugmug"
	"github.com/bzimmer/smugmug/uploadable/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

var exts = []string{".jpg"}

func up(c *cli.Context) error {
	mg := client(c)
	albumKey := c.String("album")
	albumbc := make(chan *smugmug.Album, 1)
	images := make(map[string]*smugmug.Image)

	log.Info().Msg("querying existing gallery images")
	if err := mg.Image.ImagesIter(c.Context, albumKey, func(img *smugmug.Image) (bool, error) {
		images[img.FileName] = img
		return true, nil
	}); err != nil {
		return err
	}
	log.Info().Int("count", len(images)).Msg("existing gallery images")

	u, err := filesystem.NewFsUploadable(
		filesystem.WithMetrics(metric(c)),
		filesystem.WithExtensions(c.StringSlice("ext")...),
		filesystem.WithImages(albumKey, images),
	)
	if err != nil {
		return err
	}

	grp, ctx := errgroup.WithContext(c.Context)
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
		fs := afero.NewOsFs()
		fsup := filesystem.NewFsUploadables(fs, c.Args().Slice(), u)
		uploadc, errc := mg.Upload.Uploads(ctx, fsup)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case err := <-errc:
				return err
			case _, ok := <-uploadc:
				if !ok {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case album := <-albumbc:
						log.Info().Str("albumKey", album.AlbumKey).Str("webURI", album.WebURI).Msg("complete")
					}
					return nil
				}
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
				Value:    cli.NewStringSlice(exts...),
			},
		},
		Action: up,
		After:  stats,
	}
}
