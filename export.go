package ma

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bzimmer/smugmug"
	"github.com/cavaliercoder/grab"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type exporter struct {
	mg         *smugmug.Client
	concurrent int
}

func (x *exporter) parents(ctx context.Context, nodeID string) (string, error) {
	var nodeIDs []string
	if err := x.mg.Node.ParentsIter(ctx, nodeID, func(node *smugmug.Node) (bool, error) {
		nodeIDs = append(nodeIDs, node.NodeID)
		return true, nil
	}); err != nil {
		return "", err
	}
	path := nodeIDs[len(nodeIDs)-1]
	for i := len(nodeIDs) - 2; i >= 0; i-- {
		path = filepath.Join(path, nodeIDs[i])
	}
	return path, nil
}

func (x *exporter) request(ctx context.Context, image *smugmug.Image, destination string) (*grab.Request, error) {
	original := image.ImageSizeDetails.ImageSizeOriginal
	url := original.URL
	stat, err := os.Stat(destination)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	if stat != nil && stat.Size() == int64(original.Size) {
		return nil, nil
	}
	req, err := grab.NewRequest(destination, url)
	if err != nil {
		return nil, err
	}
	req.Label = image.URI
	req.NoCreateDirectories = false
	return req.WithContext(ctx), nil
}

func (x *exporter) download(ctx context.Context, reqs []*grab.Request) error {
	client := grab.NewClient()
	for res := range client.DoBatch(x.concurrent, reqs...) {
		err := res.Err()
		if err == nil {
			log.Info().Str("uri", res.Request.Label).Str("filename", res.Filename).Msg("downloaded")
			continue
		}
		code := res.HTTPResponse.StatusCode
		switch code {
		case http.StatusOK:
			log.Error().
				Err(err).
				Int("code", code).
				Str("uri", res.Request.Label).
				Str("url", res.Request.URL().String()).
				Msg("download")
		case http.StatusNotFound:
			log.Warn().
				Int("code", code).
				Str("uri", res.Request.Label).
				Str("url", res.Request.URL().String()).
				Msg("download")
		default:
			log.Error().
				Int("code", code).
				Str("uri", res.Request.Label).
				Str("url", res.Request.URL().String()).
				Msg("download")
			return err
		}
	}
	return nil
}

func (x *exporter) export(ctx context.Context, destination string) smugmug.AlbumIterFunc {
	return func(album *smugmug.Album) (bool, error) {
		ps, err := x.parents(ctx, album.NodeID)
		if err != nil {
			return false, err
		}
		log.Info().
			Str("albumName", album.Name).
			Str("albumKey", album.AlbumKey).
			Str("nodeID", album.NodeID).
			Msg("album")

		var reqs []*grab.Request
		err = x.mg.Image.ImagesIter(ctx, album.AlbumKey, func(image *smugmug.Image) (bool, error) {
			dest := filepath.Join(destination, ps, image.FileName)
			req, err := x.request(ctx, image, dest)
			if err != nil {
				return false, err
			}
			if req == nil {
				log.Info().Str("imageKey", image.ImageKey).Msg("skipping")
				return true, nil
			}
			log.Info().
				Str("imageKey", image.ImageKey).
				Str("url", req.URL().String()).
				Msg("enqueued")
			reqs = append(reqs, req)
			return true, nil
		}, smugmug.WithExpansions("ImageSizeDetails"))
		if err != nil {
			return false, err
		}

		err = x.download(ctx, reqs)
		return err == nil, err
	}
}

func export(c *cli.Context) error {
	mg, err := client(c)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	x := &exporter{mg: mg, concurrent: c.Int("concurrent")}
	f := x.export(ctx, c.Args().Get(1))
	return mg.Node.Walk(ctx, c.Args().Get(0), func(node *smugmug.Node) (bool, error) {
		if node.Type == "Album" {
			return f(node.Album)
		}
		return true, nil
	}, smugmug.WithExpansions("Album"))
}

func CommandExport() *cli.Command {
	return &cli.Command{
		Name:  "export",
		Usage: "export images from albums",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "size",
				Usage: "the size of the image to export",
				Value: "Original",
			},
			&cli.IntFlag{
				Name:  "concurrent",
				Usage: "the number of concurrent downloads",
				Value: 2,
			},
		},
		Before: func(c *cli.Context) error {
			if c.NArg() < 2 {
				return fmt.Errorf("expected two arguments, not {%d}", c.NArg())
			}
			return nil
		},
		Action: export,
	}
}
