package ma

import (
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func rm(c *cli.Context) error {
	zv := c.Bool("zero-version")
	albumKey := c.String("album")
	for i := 0; i < c.NArg(); i++ {
		id, err := zero(c.Args().Get(i), zv)
		if err != nil {
			return err
		}
		runtime(c).Metrics.IncrCounter([]string{"rm", c.Command.Name, "attempt"}, 1)
		res, err := runtime(c).Client.Image.Delete(c.Context, albumKey, id)
		if err != nil {
			runtime(c).Metrics.IncrCounter([]string{"rm", c.Command.Name, "failure"}, 1)
			return err
		}
		runtime(c).Metrics.IncrCounter([]string{"rm", c.Command.Name, "success"}, 1)
		if err = runtime(c).Encoder.Encode(map[string]any{
			"AlbumKey": albumKey,
			"ImageKey": id,
			"Status":   res,
		}); err != nil {
			return err
		}
		log.Info().Str("albumKey", albumKey).Str("imageKey", id).Msg("delete")
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
				Usage:       "delete an image from an album",
				Description: "delete an image from an album",
				ArgsUsage:   "<image key> [<image key>, ...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "album",
						Required: true,
						Usage:    "the album from which the image is to be deleted",
					},
					zeroFlag(),
				},
				Action: rm,
			},
		},
	}
}
