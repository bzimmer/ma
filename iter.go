package ma

import (
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func nodeIterFunc(c *cli.Context, op string) smugmug.NodeIterFunc {
	enc := encoder(c)
	nodeb := c.Bool("node")
	album := c.Bool("album")
	return func(node *smugmug.Node) (bool, error) {
		msg := map[string]interface{}{
			"name":   node.Name,
			"type":   node.Type,
			"nodeID": node.NodeID,
		}
		info := log.Info().
			Str("name", node.Name).
			Str("type", node.Type).
			Str("nodeID", node.NodeID)

		if node.Parent != nil {
			msg["parentID"] = node.Parent.NodeID
			info = info.Str("parentID", node.Parent.NodeID)
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
		switch c.Bool("json") {
		case true:
			if err := enc.Encode(msg); err != nil {
				return false, err
			}
		default:
			info.Msg(op)
		}
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
		switch c.Bool("json") {
		case true:
			if err := enc.Encode(msg); err != nil {
				return false, err
			}
		default:
			log.Info().
				Str("name", album.Name).
				Str("albumKey", album.AlbumKey).
				Str("nodeID", album.NodeID).
				Int("imageCount", album.ImageCount).
				Msg(op)
		}
		return true, nil
	}
}
