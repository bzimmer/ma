package ma

import (
	"fmt"

	"github.com/bzimmer/smugmug"
	"github.com/urfave/cli/v2"
)

func image(c *cli.Context) error {
	mg := runtime(c).Smugmug()
	for i := range c.NArg() {
		id := c.Args().Get(i)
		if !imageRE.MatchString(id) {
			id = fmt.Sprintf("%s-0", id)
		}
		image, err := mg.Image.Image(c.Context, id, smugmug.WithExpansions("ImageAlbum", "ImageMetadata"))
		if err != nil {
			return err
		}
		f := imageIterFunc(c, image.Album, "ls")
		if _, err = f(image); err != nil {
			return err
		}
	}
	return nil
}

func album(c *cli.Context) error {
	mg := runtime(c).Smugmug()
	f := albumIterFunc(c, "ls", smugmug.WithExpansions("ImageAlbum", "ImageMetadata"))
	for _, id := range c.Args().Slice() {
		album, err := mg.Album.Album(c.Context, id)
		if err != nil {
			return err
		}
		if _, err = f(album); err != nil {
			return err
		}
	}
	return nil
}

func node(c *cli.Context) error {
	mg := runtime(c).Smugmug()
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
		Usage:       "List nodes, albums, and/or images",
		Description: "List the details of albums, nodes, and/or images",
		Subcommands: []*cli.Command{
			{
				Name:        "album",
				HelpName:    "album",
				Usage:       "List albums",
				Description: "List the contents of an album(s)",
				ArgsUsage:   "<album key> [<album key>, ...]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "image",
						Aliases: []string{"i", "R"},
						Usage:   "Include images in the query",
					},
				},
				Action: album,
			},
			{
				Name:        "node",
				HelpName:    "node",
				Usage:       "List nodes",
				Description: "List the contents of a node(s)",
				ArgsUsage:   "<node id> [<node id>, ...]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "album",
						Aliases: []string{"a"},
						Usage:   "Include albums in the query",
					},
					&cli.BoolFlag{
						Name:    "node",
						Aliases: []string{"n", "f"},
						Usage:   "Include nodes in the query",
					},
					&cli.BoolFlag{
						Name:    "image",
						Aliases: []string{"i"},
						Usage:   "Include images in the query",
					},
					&cli.BoolFlag{
						Name:    "recurse",
						Aliases: []string{"R"},
						Usage:   "Walk the node tree",
					},
					&cli.IntFlag{
						Name:  "depth",
						Value: -1,
						Usage: "Walk the node tree to the specified depth",
					},
				},
				Before: albumOrNode,
				Action: node,
			},
			{
				Name:        "image",
				HelpName:    "image",
				Usage:       "List images",
				Description: "List the details of an image(s)",
				ArgsUsage:   "<image key> [<image key>, ...]",
				Action:      image,
			},
		},
	}
}
