package ma

import (
	"strings"

	"github.com/bzimmer/smugmug"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

func find(c *cli.Context) error {
	mg := runtime(c).Smugmug()
	scope := c.String("scope")
	if scope == "" {
		user, err := mg.User.AuthUser(c.Context)
		if err != nil {
			return err
		}
		scope = user.URI
	}

	q := strings.Join(c.Args().Slice(), " ")
	options := []smugmug.APIOption{smugmug.WithSearch(scope, q)}

	grp, ctx := errgroup.WithContext(c.Context)
	if c.Bool("node") {
		grp.Go(func() error {
			return mg.Node.SearchIter(ctx, nodeIterFunc(c, true, "find"), options...)
		})
	}
	if c.Bool("album") {
		grp.Go(func() error {
			return mg.Album.SearchIter(c.Context, albumIterFunc(c, "find"), options...)
		})
	}
	return grp.Wait()
}

func CommandFind() *cli.Command {
	return &cli.Command{
		Name:     "find",
		HelpName: "find",
		Aliases:  []string{"search"},
		Usage:    "Search for albums or folders by name",
		Description: `Find albums or folders by name

(if neither --album nor --node is specified, both will be searched)
`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "scope",
				Usage: "Root the search at the scope, if not specified the authenticated user's scope will be used",
			},
			&cli.BoolFlag{
				Name:    "album",
				Usage:   "Search for albums",
				Aliases: []string{"a"},
			},
			&cli.BoolFlag{
				Name:    "node",
				Usage:   "Search for nodes",
				Aliases: []string{"n", "f"},
			},
		},
		Before: albumOrNode,
		Action: find,
	}
}
