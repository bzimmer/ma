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

	node, err := runtime(c).Client.Node.Create(c.Context, c.String("parent"), nodelet)
	if err != nil {
		return err
	}
	runtime(c).Metrics.IncrCounter([]string{"album", c.Command.Name}, 1)
	msg := log.Info()
	msg = msg.Str("name", node.Name)
	msg = msg.Str("nodeID", node.NodeID)
	msg = msg.Str("nodeURI", node.URI)
	msg = msg.Str("urlName", node.URLName)
	msg = msg.Str("webURI", node.WebURI)
	if nodelet.Type == smugmug.TypeAlbum {
		node, err = runtime(c).Client.Node.Node(c.Context, node.NodeID, smugmug.WithExpansions("Album"))
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
		Usage:       "create a new node",
		Description: "create a new album or folder",
		ArgsUsage:   "<node name> [<node url>]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "parent",
				Required: true,
				Usage:    "the parent node at which the new node will be rooted",
			},
			&cli.StringFlag{
				Name:  "privacy",
				Value: "",
				Usage: "the privacy settings for the new album",
			},
		},
		Subcommands: []*cli.Command{
			{
				Name:        "album",
				Usage:       "create a new album",
				Description: "create a new album for images",
				Before:      before,
				Action:      knew,
			},
			{
				Name:        "folder",
				Usage:       "create a new folder",
				Description: "create a new folder for albums",
				Before:      before,
				Action:      knew,
			},
		},
	}
}
