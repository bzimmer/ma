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

	q := mg.Node.ChildrenIter
	if c.IsSet("recurse") {
		q = mg.Node.Walk
	}

	nodeIDs := c.Args().Slice()
	if len(nodeIDs) == 0 {
		user, err := mg.User.AuthUser(c.Context, smugmug.WithExpansions("Node"))
		if err != nil {
			return err
		}
		nodeIDs = []string{user.Node.NodeID}
	}

	f := nodeIterFunc(c, "ls")
	for i := range nodeIDs {
		if err := q(c.Context, nodeIDs[i], f, smugmug.WithExpansions("Album", "ParentNode")); err != nil {
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
				Name:    "recurse",
				Aliases: []string{"R"},
			},
			&cli.BoolFlag{
				Name:     "json",
				Aliases:  []string{"j"},
				Value:    false,
				Required: false,
			},
		},
		Before: albumOrNode,
		Action: list,
	}
}
