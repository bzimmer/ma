package ma

import (
	"context"

	"github.com/bzimmer/smugmug"
	"github.com/urfave/cli/v2"
)

func list(c *cli.Context) error {
	mg, err := client(c)
	if err != nil {
		return err
	}

	var q func(context.Context, string, smugmug.NodeIterFunc, ...smugmug.APIOption) error
	if c.IsSet("recurse") {
		q = mg.Node.Walk
	} else {
		q = mg.Node.ChildrenIter
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
		Aliases: []string{"l", "list"},
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
		},
		Before: func(c *cli.Context) error {
			node := c.Bool("node")
			album := c.Bool("album")
			if !(album || node) {
				if err := c.Set("node", "true"); err != nil {
					return err
				}
				if err := c.Set("album", "true"); err != nil {
					return err
				}
			}
			return nil
		},
		Action: list,
	}
}
