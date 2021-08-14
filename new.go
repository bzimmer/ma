package ma

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/bzimmer/smugmug"
	"github.com/urfave/cli/v2"
)

func new(c *cli.Context) error {
	var url string
	name := c.Args().First()
	switch c.NArg() {
	case 1:
		url = smugmug.URLName(name)
	case 2:
		url = c.Args().Get(1)
		if !unicode.IsUpper(rune(url[0])) {
			return fmt.Errorf("node url name must start with a capital letter")
		}
	}

	nodelet := &smugmug.Nodelet{
		Name:    name,
		Type:    strings.Title(c.Command.Name),
		URLName: url,
		Privacy: c.String("privacy"),
	}

	mg, err := client(c)
	if err != nil {
		return err
	}
	node, err := mg.Node.Create(c.Context, c.String("parent"), nodelet)
	if err != nil {
		return err
	}
	enc, err := encoder(c)
	if err != nil {
		return err
	}
	return enc.Encode("new", map[string]interface{}{
		"name":    node.Name,
		"nodeID":  node.NodeID,
		"nodeURI": node.URI,
		"urlName": node.URLName,
	})
}

func CommandNew() *cli.Command {
	before := func(c *cli.Context) error {
		n := c.NArg()
		if n == 0 || n > 2 {
			return fmt.Errorf("expected one or two arguments, not {%d}", n)
		}
		privacy := c.String("privacy")
		switch privacy {
		case "":
		case "Unlisted", "Private", "Public":
		default:
			return fmt.Errorf("privacy one of [Unlisted, Private, Public], not {%s}", privacy)
		}
		return nil
	}
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     "parent",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "privacy",
			Value: "",
		},
	}
	return &cli.Command{
		Name:      "new",
		Usage:     "create a new node",
		ArgsUsage: "<node name> [<node url>]",
		Subcommands: []*cli.Command{
			{
				Name:   "album",
				Flags:  flags,
				Before: before,
				Action: new,
			},
			{
				Name:   "folder",
				Flags:  flags,
				Before: before,
				Action: new,
			},
		},
	}
}
