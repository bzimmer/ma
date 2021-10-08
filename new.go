package ma

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func knew(c *cli.Context) error {
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

	node, err := client(c).Node.Create(c.Context, c.String("parent"), nodelet)
	if err != nil {
		return err
	}
	metric(c).IncrCounter([]string{"album", "new"}, 1)
	msg := log.Info()
	msg = msg.Str("name", node.Name)
	msg = msg.Str("nodeID", node.NodeID)
	msg = msg.Str("nodeURI", node.URI)
	msg = msg.Str("urlName", node.URLName)
	msg = msg.Str("webURI", node.WebURI)
	if nodelet.Type == "Album" {
		node, err = client(c).Node.Node(c.Context, node.NodeID, smugmug.WithExpansions("Album"))
		if err != nil {
			return err
		}
		msg = msg.Str("albumKey", node.Album.AlbumKey)
	}
	msg.Msg("new")
	return encoder(c).Encode(node)
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
	return &cli.Command{
		Name:      "new",
		HelpName:  "new",
		Usage:     "create a new node",
		ArgsUsage: "<node name> [<node url>]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "parent",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "privacy",
				Value: "",
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:   "album",
				Before: before,
				Action: knew,
			},
			{
				Name:   "folder",
				Before: before,
				Action: knew,
			},
		},
	}
}
