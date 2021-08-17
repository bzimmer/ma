package ma

import (
	"github.com/bzimmer/smugmug"
	"github.com/urfave/cli/v2"
)

func imageIterFunc(enc Encoder, albumKey string, op string) smugmug.ImageIterFunc {
	return func(image *smugmug.Image) (bool, error) {
		err := enc.Encode(op, map[string]interface{}{
			"albumKey": albumKey,
			"imageKey": image.ImageKey,
			"imageURI": image.URI,
			"filename": image.FileName,
			"type":     "Image",
			"caption":  image.Caption,
			"version":  image.Serial,
			"keywords": image.KeywordArray,
		})
		return err == nil, err
	}
}

func albumIterFunc(c *cli.Context, op string) smugmug.AlbumIterFunc {
	imageq := c.Bool("image")
	return func(album *smugmug.Album) (bool, error) {
		enc, err := encoder(c)
		if err != nil {
			return false, err
		}
		if err := enc.Encode(op, map[string]interface{}{
			"name":       album.Name,
			"type":       "Album",
			"nodeID":     album.NodeID,
			"albumKey":   album.AlbumKey,
			"imageCount": album.ImageCount,
		}); err != nil {
			return false, err
		}
		if imageq {
			mg, err := client(c)
			if err != nil {
				return false, err
			}
			f := imageIterFunc(enc, album.AlbumKey, op)
			if err := mg.Image.ImagesIter(c.Context, album.AlbumKey, f); err != nil {
				return false, err
			}
		}

		return true, nil
	}
}

func nodeIterFunc(c *cli.Context, recurse bool, op string) smugmug.NodeIterFunc {
	nodeq := c.Bool("node")
	albumq := c.Bool("album")
	imageq := c.Bool("image")
	return func(node *smugmug.Node) (bool, error) {
		var albumKey string

		enc, err := encoder(c)
		if err != nil {
			return false, err
		}

		msg := map[string]interface{}{
			"name":   node.Name,
			"type":   node.Type,
			"nodeID": node.NodeID,
		}
		if node.Parent != nil {
			msg["parentID"] = node.Parent.NodeID
		}

		switch node.Type {
		case "Album":
			if !albumq {
				return recurse, nil
			}
			msg["albumKey"] = node.Album.AlbumKey
			msg["imageCount"] = node.Album.ImageCount
			if imageq {
				albumKey = node.Album.AlbumKey
			}
		case "Folder":
			if !nodeq {
				return recurse, nil
			}
		}

		if err := enc.Encode(op, msg); err != nil {
			return false, err
		}

		if albumKey != "" {
			mg, err := client(c)
			if err != nil {
				return false, err
			}
			f := imageIterFunc(enc, albumKey, op)
			if err := mg.Image.ImagesIter(c.Context, albumKey, f); err != nil {
				return false, err
			}
		}

		return recurse, nil
	}
}
