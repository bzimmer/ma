package ma

import (
	"fmt"
	"regexp"

	"github.com/bzimmer/smugmug"
	"github.com/urfave/cli/v2"
)

var imageRE = regexp.MustCompile(`[a-zA-Z0-9]+-\d+`)

func image(c *cli.Context) error {
	mg := runtime(c).Client
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
	mg := runtime(c).Client
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
	mg := runtime(c).Client
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
		Name:        "ls",
		HelpName:    "ls",
		Aliases:     []string{"list"},
		Usage:       "list nodes, albums, and/or images",
		Description: "list the deails of albums, nodes, and/or images",
		Subcommands: []*cli.Command{
			{
				Name:        "album",
				Usage:       "list albums",
				Description: "list the contents of an album(s)",
				ArgsUsage:   "<album key> [<album key>, ...]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "image",
						Aliases: []string{"i", "R"},
						Usage:   "include images in the query",
					},
				},
				Action: album,
			},
			{
				Name:        "node",
				Usage:       "list nodes",
				Description: "list the contents of a node(s)",
				ArgsUsage:   "<node id> [<node id>, ...]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "album",
						Aliases: []string{"a"},
						Usage:   "include albums in the query",
					},
					&cli.BoolFlag{
						Name:    "node",
						Aliases: []string{"n", "f"},
						Usage:   "include nodes in the query",
					},
					&cli.BoolFlag{
						Name:    "image",
						Aliases: []string{"i"},
						Usage:   "include images in the query",
					},
					&cli.BoolFlag{
						Name:    "recurse",
						Aliases: []string{"R"},
						Usage:   "walk the node tree",
					},
					&cli.IntFlag{
						Name:  "depth",
						Value: -1,
						Usage: "walk the node tree to the specified depth",
					},
				},
				Before: albumOrNode,
				Action: node,
			},
			{
				Name:        "image",
				Usage:       "list images",
				Description: "list the details of an image(s)",
				ArgsUsage:   "<image key> [<image key>, ...]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:     "zero-version",
						Aliases:  []string{"z"},
						Usage:    "if no version is specified, append `-0`",
						Value:    false,
						Required: false,
					},
				},
				Action: image,
			},
		},
	}
}
