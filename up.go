package ma

import (
	"github.com/bzimmer/smugmug"
	"github.com/bzimmer/smugmug/uploadable/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func up(c *cli.Context) error {
	mg, err := client(c)
	if err != nil {
		return err
	}

	metric, err := metric(c)
	if err != nil {
		return err
	}

	albumID := c.String("album")
	images := make(map[string]*smugmug.Image)

	log.Info().Msg("querying existing gallery images")
	if err := mg.Image.ImagesIter(c.Context, albumID, func(img *smugmug.Image) (bool, error) {
		images[img.FileName] = img
		return true, nil
	}); err != nil {
		return err
	}
	log.Info().Int("count", len(images)).Msg("existing gallery images")

	u, err := filesystem.NewFsUploadable(
		filesystem.WithMetrics(metric),
		filesystem.WithExtensions(".jpg"),
		filesystem.WithImages(albumID, images),
	)
	if err != nil {
		return err
	}
	fsys := filesystem.RelativeFS("/")
	fsup := filesystem.NewFsUploadables(fsys, c.Args().Slice(), u)
	uploadc, errc := mg.Upload.Uploads(c.Context, fsup)
	for {
		select {
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
				Value:    "",
				Required: true,
			},
		},
		Action: up,
		After:  stats,
	}
}
