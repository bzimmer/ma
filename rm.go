package ma

import (
	"fmt"

	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

// imageID returns id with its serial: as given when it already has one,
// else looked up among the album's images, read once into images.
func imageID(c *cli.Context, id string, images *map[string]*smugmug.Image) (string, error) {
	if imageRE.MatchString(id) {
		return id, nil
	}
	if *images == nil {
		var err error
		_, *images, err = existing(c, func(img *smugmug.Image) string {
			return img.ImageKey
		})
		if err != nil {
			return "", err
		}
	}
	image, ok := (*images)[id]
	if !ok {
		return "", &InvalidVersionError{ImageKey: id}
	}
	return fmt.Sprintf("%s-%d", id, image.Serial), nil
}

// remove deletes one image, or with dryrun only counts it.
func remove(c *cli.Context, albumKey, id string, dryrun bool) (bool, error) {
	if dryrun {
		runtime(c).Metrics.IncrCounter([]string{"rm", c.Command.Name, "dryrun"}, 1)
		return false, nil
	}
	runtime(c).Metrics.IncrCounter([]string{"rm", c.Command.Name, "attempt"}, 1)
	res, err := runtime(c).Smugmug().Image.Delete(c.Context, albumKey, id)
	if err != nil {
		runtime(c).Metrics.IncrCounter([]string{"rm", c.Command.Name, "failure"}, 1)
		return false, err
	}
	runtime(c).Metrics.IncrCounter([]string{"rm", c.Command.Name, "success"}, 1)
	return res, nil
}

func rm(c *cli.Context) error {
	dryrun := c.Bool("dryrun")
	albumKey := c.String("album")
	var images map[string]*smugmug.Image
	for i := range c.NArg() {
		id, err := imageID(c, c.Args().Get(i), &images)
		if err != nil {
			return err
		}
		res, err := remove(c, albumKey, id, dryrun)
		if err != nil {
			return err
		}
		if err = runtime(c).Encoder.Encode(map[string]any{
			"AlbumKey": albumKey,
			"ImageKey": id,
			"Status":   res,
		}); err != nil {
			return err
		}
		log.Debug().Str("albumKey", albumKey).Str("imageKey", id).Msg("delete")
	}
	return nil
}

func CommandRemove() *cli.Command {
	return &cli.Command{
		Name:        "rm",
		HelpName:    "rm",
		Usage:       "Delete an entity",
		Description: "Delete an entity",
		Subcommands: []*cli.Command{
			{
				Name:        "image",
				HelpName:    "image",
				Usage:       "Delete an image from an album",
				Description: "Delete an image from an album",
				ArgsUsage:   "<image key> [<image key>, ...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "album",
						Required: true,
						Usage:    "The album from which the image is to be deleted",
					},
					&cli.BoolFlag{
						Name:     "dryrun",
						Usage:    "Prepare to upload but don't actually do it",
						Aliases:  []string{"n"},
						Value:    false,
						Required: false,
					},
				},
				Action: rm,
			},
		},
	}
}
