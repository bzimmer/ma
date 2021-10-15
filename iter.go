package ma

import (
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func imageIterFunc(c *cli.Context, album *smugmug.Album, op string) smugmug.ImageIterFunc {
	enc := encoder(c)
	var albumKey string
	if album != nil {
		albumKey = album.AlbumKey
	}
	return func(image *smugmug.Image) (bool, error) {
		metric(c).IncrCounter([]string{op, "image"}, 1)
		if album != nil && image.Album == nil {
			image.Album = album
		}
		log.Info().
			Str("type", "Image").
			Str("albumKey", albumKey).
			Str("imageKey", image.ImageKey).
			Str("imageURI", image.URI).
			Str("filename", image.FileName).
			Str("caption", image.Caption).
			Int("version", image.Serial).
			Strs("keywords", image.KeywordArray).
			Msg(op)
		err := enc.Encode(image)
		return err == nil, err
	}
}

func albumIterFunc(c *cli.Context, op string) smugmug.AlbumIterFunc {
	mg := client(c)
	enc := encoder(c)
	imageq := c.Bool("image")
	return func(album *smugmug.Album) (bool, error) {
		metric(c).IncrCounter([]string{op, "album"}, 1)
		log.Info().
			Str("type", smugmug.TypeAlbum).
			Str("name", album.Name).
			Str("nodeID", album.NodeID).
			Str("albumKey", album.AlbumKey).
			Int("imageCount", album.ImageCount).
			Msg(op)
		err := enc.Encode(album)
		if err != nil {
			return false, err
		}
		if imageq {
			f := imageIterFunc(c, album, op)
			if err := mg.Image.ImagesIter(c.Context, album.AlbumKey, f); err != nil {
				return false, err
			}
		}
		return true, nil
	}
}

func nodeIterFunc(c *cli.Context, recurse bool, op string) smugmug.NodeIterFunc {
	enc := encoder(c)
	nodeq := c.Bool("node")
	albumq := c.Bool("album")
	imageq := c.Bool("image")
	return func(node *smugmug.Node) (bool, error) {
		metric(c).IncrCounter([]string{op, "node"}, 1)
		msg := log.Info()
		msg = msg.Str("type", node.Type)
		msg = msg.Str("name", node.Name)
		msg = msg.Str("nodeID", node.NodeID)

		if node.Parent != nil {
			msg = msg.Str("parentID", node.Parent.NodeID)
		}

		switch node.Type {
		case smugmug.TypeAlbum:
			if !albumq {
				return recurse, nil
			}
			msg = msg.Str("albumKey", node.Album.AlbumKey)
			msg = msg.Int("imageCount", node.Album.ImageCount)
		case smugmug.TypeFolder:
			if !nodeq {
				return recurse, nil
			}
		}

		msg.Msg(op)
		if err := enc.Encode(node); err != nil {
			return false, err
		}

		if imageq && node.Album != nil {
			albumKey := node.Album.AlbumKey
			f := imageIterFunc(c, node.Album, op)
			if err := client(c).Image.ImagesIter(c.Context, albumKey, f); err != nil {
				return false, err
			}
		}

		return recurse, nil
	}
}
