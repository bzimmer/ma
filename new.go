package ma

import (
	"fmt"

	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func knew(c *cli.Context) error {
	var url string
	name := c.Args().First()
	switch c.NArg() {
	case 1:
		url = smugmug.URLName(name, runtime(c).Language)
	case 2:
		url = c.Args().Get(1)
		if err := validate(url); err != nil {
			return err
		}
	}

	nodelet := &smugmug.Nodelet{
		Name:    name,
		Type:    titlecase(c, c.Command.Name),
		URLName: url,
		Privacy: c.String("privacy"),
	}

	client := runtime(c).Smugmug()
	node, err := client.Node.Create(c.Context, c.String("parent"), nodelet)
	if err != nil {
		return err
	}
	runtime(c).Metrics.IncrCounter([]string{"album", c.Command.Name}, 1)
	msg := log.Debug().
		Str("name", node.Name).
		Str("nodeID", node.NodeID).
		Str("nodeURI", node.URI).
		Str("urlName", node.URLName).
		Str("webURI", node.WebURI)
	if nodelet.Type == smugmug.TypeAlbum {
		node, err = client.Node.Node(c.Context, node.NodeID, smugmug.WithExpansions("Album"))
		if err != nil {
			return err
		}
		msg = msg.Str("albumKey", node.Album.AlbumKey)
	}
	msg.Msg(c.Command.Name)
	return runtime(c).Encoder.Encode(node)
}

func CommandNew() *cli.Command {
	before := func(c *cli.Context) error {
		n := c.NArg()
		if n == 0 || n > 2 {
			return fmt.Errorf("expected one or two arguments, not {%d}", n)
		}
		switch p := c.String("privacy"); p {
		case "":
		case "Unlisted", "Private", "Public":
		default:
			return fmt.Errorf("privacy one of [Unlisted, Private, Public], not {%s}", p)
		}
		return nil
	}
	return &cli.Command{
		Name:        "new",
		HelpName:    "new",
		Aliases:     []string{"create"},
		Usage:       "Create a new node",
		Description: "Create a new album or folder",
		ArgsUsage:   "<node name> [<node url>]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "parent",
				Required: true,
				Usage:    "The parent node at which the new node will be rooted",
			},
			&cli.StringFlag{
				Name:  "privacy",
				Value: "",
				Usage: "The privacy settings for the new album",
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:        "album",
				HelpName:    "album",
				Usage:       "Create a new album",
				Description: "Create a new album for images",
				Before:      before,
				Action:      knew,
			},
			{
				Name:        "folder",
				HelpName:    "folder",
				Usage:       "Create a new folder",
				Description: "Create a new folder for albums",
				Before:      before,
				Action:      knew,
			},
		},
	}
}
