package ma

import (
	"github.com/bzimmer/smugmug"
	"github.com/urfave/cli/v2"
)

func find(c *cli.Context) error {
	mg, err := client(c)
	if err != nil {
		return err
	}

	scope := c.String("scope")
	if scope == "" {
		user, err := mg.User.AuthUser(c.Context)
		if err != nil {
			return err
		}
		scope = user.URI
	}

	options := []smugmug.APIOption{smugmug.WithSearch(scope, c.Args().First())}
	if c.Bool("node") {
		options = append(options, smugmug.WithExpansions("ParentNode"))
		if err := mg.Node.SearchIter(c.Context, nodeIterFunc(c, "find"), options...); err != nil {
			return err
		}
	}
	if c.Bool("album") {
		options = append(options, smugmug.WithExpansions("Node"))
		if err := mg.Album.SearchIter(c.Context, albumIterFunc(c, "find"), options...); err != nil {
			return err
		}
	}
	return nil
}

func CommandFind() *cli.Command {
	return &cli.Command{
		Name:    "find",
		Aliases: []string{"f"},
		Usage:   "search for album or folders by name",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "album",
				Aliases: []string{"a"},
			},
			&cli.BoolFlag{
				Name:    "node",
				Aliases: []string{"n", "f"},
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
		Action: find,
	}
}
