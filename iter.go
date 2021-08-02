package ma

import (
	"encoding/json"

	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func encoder(c *cli.Context) *json.Encoder {
	return json.NewEncoder(c.App.Writer)
}

func nodeIterFunc(c *cli.Context, op string) smugmug.NodeIterFunc {
	enc := encoder(c)
	nodeb := c.Bool("node")
	album := c.Bool("album")
	return func(node *smugmug.Node) (bool, error) {
		msg := map[string]interface{}{
			"nodeID": node.NodeID,
			"type":   node.Type,
			"name":   node.Name,
		}
		info := log.Info().
			Str("nodeID", node.NodeID).
			Str("type", node.Type).
			Str("name", node.Name)

		if node.Parent != nil {
			info = info.Str("parentID", node.Parent.NodeID)
			msg["parentID"] = node.Parent.NodeID
		}

		switch node.Type {
		case "Album":
			if !album {
				return true, nil
			}
			msg["albumKey"] = node.Album.AlbumKey
			msg["imageCount"] = node.Album.ImageCount
			info = info.Str("albumKey", node.Album.AlbumKey).Int("imageCount", node.Album.ImageCount)
		case "Folder":
			if !nodeb {
				return true, nil
			}
		}

		if enc != nil {
			if err := enc.Encode(msg); err != nil {
				return false, err
			}
		}

		info.Msg(op)

		return true, nil
	}
}

func albumIterFunc(c *cli.Context, op string) smugmug.AlbumIterFunc {
	enc := encoder(c)
	return func(album *smugmug.Album) (bool, error) {
		msg := map[string]interface{}{
			"name":       album.Name,
			"nodeID":     album.NodeID,
			"albumKey":   album.AlbumKey,
			"imageCount": album.ImageCount,
		}
		if enc != nil {
			if err := enc.Encode(msg); err != nil {
				return false, err
			}
		}
		log.Info().
			Str("name", album.Name).
			Str("albumKey", album.AlbumKey).
			Str("nodeID", album.NodeID).
			Int("imageCount", album.ImageCount).
			Msg(op)
		return true, nil
	}
}
