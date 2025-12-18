package ma

import (
	"fmt"

	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func rm(c *cli.Context) error {
	dryrun := c.Bool("dryrun")
	albumKey := c.String("album")
	var (
		res    bool
		err    error
		images map[string]*smugmug.Image
	)
	for i := range c.NArg() {
		id := c.Args().Get(i)
		if !imageRE.MatchString(id) {
			if images == nil {
				_, images, err = existing(c, func(img *smugmug.Image) string {
					return img.ImageKey
				})
				if err != nil {
					return err
				}
			}
			image, ok := images[id]
			if !ok {
				return &InvalidVersionError{ImageKey: id}
			}
			id = fmt.Sprintf("%s-%d", id, image.Serial)
		}
		if dryrun {
			res = false
			runtime(c).Metrics.IncrCounter([]string{"rm", c.Command.Name, "dryrun"}, 1)
		} else {
			runtime(c).Metrics.IncrCounter([]string{"rm", c.Command.Name, "attempt"}, 1)
			res, err = runtime(c).Smugmug().Image.Delete(c.Context, albumKey, id)
			if err != nil {
				runtime(c).Metrics.IncrCounter([]string{"rm", c.Command.Name, "failure"}, 1)
				return err
			}
			runtime(c).Metrics.IncrCounter([]string{"rm", c.Command.Name, "success"}, 1)
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
