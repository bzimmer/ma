package ma

import (
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

func imageIterFunc(c *cli.Context, album *smugmug.Album, op string) smugmug.ImageIterFunc {
	enc := runtime(c).Encoder
	var albumKey string
	if album != nil {
		albumKey = album.AlbumKey
	}
	return func(image *smugmug.Image) (bool, error) {
		runtime(c).Metrics.IncrCounter([]string{op, "image"}, 1)
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
	mg := runtime(c).Smugmug()
	enc := runtime(c).Encoder
	imageq := c.Bool("image")
	return func(album *smugmug.Album) (bool, error) {
		runtime(c).Metrics.IncrCounter([]string{op, "album"}, 1)
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
			if err = mg.Image.ImagesIter(c.Context, album.AlbumKey, f); err != nil {
				return false, err
			}
		}
		return true, nil
	}
}

func nodeIterFunc(c *cli.Context, recurse bool, op string) smugmug.NodeIterFunc {
	enc := runtime(c).Encoder
	nodeq := c.Bool("node")
	albumq := c.Bool("album")
	imageq := c.Bool("image")
	return func(node *smugmug.Node) (bool, error) {
		runtime(c).Metrics.IncrCounter([]string{op, "node"}, 1)
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
			msg = msg.Str("scope", node.URI)
		}

		msg.Msg(op)
		if err := enc.Encode(node); err != nil {
			return false, err
		}

		client := runtime(c).Smugmug()
		if imageq && node.Album != nil {
			albumKey := node.Album.AlbumKey
			f := imageIterFunc(c, node.Album, op)
			if err := client.Image.ImagesIter(c.Context, albumKey, f); err != nil {
				return false, err
			}
		}

		return recurse, nil
	}
}

func existing[T comparable](c *cli.Context, f func(*smugmug.Image) T) (*smugmug.Album, map[T]*smugmug.Image, error) {
	mg := runtime(c).Smugmug()
	albumKey := c.String("album")
	albumc := make(chan *smugmug.Album, 1)
	imagesc := make(chan map[T]*smugmug.Image, 1)
	grp, ctx := errgroup.WithContext(c.Context)
	grp.Go(func() error {
		defer close(imagesc)
		images := make(map[T]*smugmug.Image)
		log.Info().Str("albumKey", albumKey).Msg("querying existing gallery images")
		if err := mg.Image.ImagesIter(ctx, albumKey, func(img *smugmug.Image) (bool, error) {
			images[f(img)] = img
			return true, nil
		}); err != nil {
			return err
		}
		imagesc <- images
		return nil
	})
	grp.Go(func() error {
		defer close(albumc)
		album, err := mg.Album.Album(ctx, albumKey)
		if err != nil {
			return err
		}
		albumc <- album
		return nil
	})
	if err := grp.Wait(); err != nil {
		log.Error().Err(err).Msg("failed to query album or album images")
		return nil, nil, err
	}
	album, images := <-albumc, <-imagesc
	log.Info().
		Int("count", len(images)).
		Str("name", album.Name).
		Str("albumKey", album.AlbumKey).
		Msg("existing gallery images")
	return album, images, nil
}
