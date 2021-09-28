package ma

import (
	"fmt"
	"regexp"

	"github.com/bzimmer/smugmug"
	"github.com/urfave/cli/v2"
)

var imageRE = regexp.MustCompile("[a-zA-Z0-9]+-[0-9]+")

func image(c *cli.Context) error {
	mg := client(c)
	zv := c.Bool("zero-version")
	for _, id := range c.Args().Slice() {
		// preempt a common mistake
		ok := imageRE.MatchString(id)
		if !ok {
			if !zv {
				return fmt.Errorf("no version specified for image key {%s}", id)
			}
			id = fmt.Sprintf("%s-0", id)
		}
		image, err := mg.Image.Image(c.Context, id, smugmug.WithExpansions("ImageAlbum"))
		if err != nil {
			return err
		}
		f := imageIterFunc(c, image.Album, "ls")
		if _, err := f(image); err != nil {
			return err
		}
	}
	return nil
}

func album(c *cli.Context) error {
	mg := client(c)
	f := albumIterFunc(c, "ls")
	for _, id := range c.Args().Slice() {
		album, err := mg.Album.Album(c.Context, id)
		if err != nil {
			return err
		}
		if _, err := f(album); err != nil {
			return err
		}
	}
	return nil
}

func node(c *cli.Context) error {
	mg := client(c)
	nodeIDs := c.Args().Slice()
	if len(nodeIDs) == 0 {
		user, err := mg.User.AuthUser(c.Context, smugmug.WithExpansions("Node"))
		if err != nil {
			return err
		}
		nodeIDs = []string{user.Node.NodeID}
	}
	depth := c.Int("depth")
	f := nodeIterFunc(c, c.Bool("recurse"), "ls")
	for i := range nodeIDs {
		if err := mg.Node.WalkN(c.Context, nodeIDs[i], f, depth, smugmug.WithExpansions("Album", "ParentNode")); err != nil {
			return err
		}
	}
	return nil
}

func CommandList() *cli.Command {
	return &cli.Command{
		Name:    "ls",
		Aliases: []string{"list"},
		Usage:   "list nodes, albums, and/or images",
		Subcommands: []*cli.Command{
			{
				Name:      "album",
				Usage:     "list albums",
				ArgsUsage: "<album key> [<album key>, ...]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "image",
						Aliases: []string{"i", "R"},
					},
				},
				Action: album,
			},
			{
				Name:      "node",
				Usage:     "list nodes",
				ArgsUsage: "<node id> [<node id>, ...]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "album",
						Aliases: []string{"a"},
					},
					&cli.BoolFlag{
						Name:    "node",
						Aliases: []string{"n", "f"},
					},
					&cli.BoolFlag{
						Name:    "image",
						Aliases: []string{"i"},
					},
					&cli.BoolFlag{
						Name:    "recurse",
						Aliases: []string{"R"},
					},
					&cli.IntFlag{
						Name:  "depth",
						Value: -1,
					},
				},
				Before: albumOrNode,
				Action: node,
			},
			{
				Name:      "image",
				Usage:     "list images",
				ArgsUsage: "<image key> [<image key>, ...]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:     "zero-version",
						Aliases:  []string{"z"},
						Value:    false,
						Required: false,
					},
				},
				Action: image,
			},
		},
	}
}
