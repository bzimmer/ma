package ma

import (
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
		runtime(c).Metrics.IncrCounter([]string{"delete", c.Command.Name, "attempt"}, 1)
		res, err := runtime(c).Client.Image.Delete(c.Context, albumKey, id)
		if err != nil {
			runtime(c).Metrics.IncrCounter([]string{"delete", c.Command.Name, "failure"}, 1)
			return err
		}
		runtime(c).Metrics.IncrCounter([]string{"delete", c.Command.Name, "success"}, 1)
		if err := runtime(c).Encoder.Encode(map[string]any{
			"AlbumKey": albumKey,
			"ImageKey": id,
			"Status":   res,
		}); err != nil {
			return err
		}
	}
	return nil
}

func CommandRemove() *cli.Command {
	return &cli.Command{
		Name:        "rm",
		HelpName:    "rm",
		Usage:       "delete an entity",
		Description: "delete an entity",
		Subcommands: []*cli.Command{
			{
				Name:        "image",
				Usage:       "delete an image from an album",
				Description: "delete an image from an album",
				ArgsUsage:   "IMAGE_KEY [, IMAGE-KEY, ...]",
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
