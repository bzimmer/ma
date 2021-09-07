package ma

import (
	"github.com/bzimmer/smugmug"
	"github.com/bzimmer/smugmug/uploadable/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
)

var exts = []string{".jpg"}

func up(c *cli.Context) error {
	mg := client(c)
	albumKey := c.String("album")
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
		filesystem.WithExtensions(c.StringSlice("exts")...),
		filesystem.WithImages(albumKey, images),
	)
	if err != nil {
		return err
	}

	fs := afero.NewOsFs()
	fsup := filesystem.NewFsUploadables(fs, c.Args().Slice(), u)
	uploadc, errc := mg.Upload.Uploads(c.Context, fsup)
	for {
		select {
		case <-c.Context.Done():
			return c.Context.Err()
		case err := <-errc:
			return err
		case _, ok := <-uploadc:
			if !ok {
				log.Info().Msg("complete")
				return nil
			}
		}
	}
}

func CommandUp() *cli.Command {
	return &cli.Command{
		Name:  "up",
		Usage: "upload images to SmugMug",
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
