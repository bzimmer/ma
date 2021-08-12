package ma

import (
	"github.com/bzimmer/smugmug"
	"github.com/urfave/cli/v2"
)

func list(c *cli.Context) error {
	mg, err := client(c)
	if err != nil {
		return err
	}

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
		Usage:   "list albums and/or folders",
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
		Action: list,
	}
}
