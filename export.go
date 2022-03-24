package ma

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/armon/go-metrics"
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

type Grab interface {
	Do(req *http.Request) (*http.Response, error)
}

type request struct {
	URI         string
	URL         string
	Filename    string
	HTTPRequest *http.Request
}

type response struct {
	Request      *request
	HTTPResponse *http.Response
}

type exporter struct {
	mg          *smugmug.Client
	fs          afero.Fs
	grab        Grab
	metrics     *metrics.Metrics
	force       bool
	concurrency int
}

func (x *exporter) parents(ctx context.Context, nodeID string) ([]string, error) {
	var nodeIDs []string
	if err := x.mg.Node.ParentsIter(ctx, nodeID, func(node *smugmug.Node) (bool, error) {
		nodeIDs = append(nodeIDs, node.NodeID)
		return true, nil
	}); err != nil {
		return nil, err
	}
	last := len(nodeIDs) - 1
	for i := 0; i < len(nodeIDs)/2; i++ {
		nodeIDs[i], nodeIDs[last-i] = nodeIDs[last-i], nodeIDs[i]
	}
	return nodeIDs, nil
}

func (x *exporter) request(image *smugmug.Image, destination string) (*request, error) {
	original := image.ImageSizeDetails.ImageSizeOriginal
	if !x.force {
		stat, err := x.fs.Stat(destination)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		}
		if stat != nil && stat.Size() == original.Size {
			return nil, nil
		}
	}
	req, err := http.NewRequest(http.MethodGet, original.URL, http.NoBody)
	if err != nil {
		return nil, err
	}
	return &request{
		URI:         image.URI,
		URL:         original.URL,
		Filename:    destination,
		HTTPRequest: req,
	}, nil
}

func (x *exporter) do(ctx context.Context, req *request) (*response, error) {
	defer func(t time.Time) {
		x.metrics.AddSample([]string{"export", "download"}, float32(time.Since(t).Seconds()))
	}(time.Now())
	res, err := x.grab.Do(req.HTTPRequest.WithContext(ctx)) // nolint:bodyclose
	if err != nil {
		return nil, err
	}
	resp := &response{
		Request:      req,
		HTTPResponse: res,
	}
	return resp, nil
}

func (x *exporter) write(res *response) error {
	if res.HTTPResponse != nil && res.HTTPResponse.Body != nil {
		defer res.HTTPResponse.Body.Close()
	}
	code := res.HTTPResponse.StatusCode
	text := strings.ToLower(http.StatusText(code))
	switch code {
	case http.StatusOK:
		x.metrics.IncrCounter([]string{"export", "download", text}, 1)
		log.Info().Str("uri", res.Request.URI).Str("filename", res.Request.Filename).Msg("downloaded")
	case http.StatusNotFound:
		x.metrics.IncrCounter([]string{"export", "download", "failed", text}, 1)
		log.Warn().
			Int("code", code).
			Str("uri", res.Request.URI).
			Str("url", res.Request.URL).
			Msg("download")
		return nil
	default:
		x.metrics.IncrCounter([]string{"export", "download", "failed", text}, 1)
		log.Error().
			Int("code", code).
			Str("uri", res.Request.URI).
			Str("url", res.Request.URL).
			Msg("download")
		return errors.New("download failed")
	}
	fp, err := x.fs.Create(res.Request.Filename)
	if err != nil {
		log.Error().Err(err).Str("filename", res.Request.Filename).Msg("failed to create file")
		return err
	}
	_, err = io.Copy(fp, res.HTTPResponse.Body)
	if err != nil {
		log.Error().Err(err).Str("filename", res.Request.Filename).Msg("failed to write file contents")
		return err
	}
	if err := fp.Close(); err != nil {
		log.Error().Err(err).Str("filename", res.Request.Filename).Msg("failed to close file")
		return err
	}
	return nil
}

func (x *exporter) download(ctx context.Context, reqs []*request) error {
	requestsc := make(chan *request)

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		defer close(requestsc)
		for i := 0; i < len(reqs); i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case requestsc <- reqs[i]:
			}
		}
		return nil
	})
	for i := 0; i < x.concurrency; i++ {
		grp.Go(func() error {
			for req := range requestsc {
				res, err := x.do(ctx, req)
				if err != nil {
					return err
				}
				if err := x.write(res); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return grp.Wait()
}

func (x *exporter) export(ctx context.Context, destination string) smugmug.AlbumIterFunc {
	return func(album *smugmug.Album) (bool, error) {
		ps, err := x.parents(ctx, album.NodeID)
		if err != nil {
			return false, err
		}
		out := filepath.Join(destination, filepath.Join(ps...))
		log.Info().
			Str("albumName", album.Name).
			Str("albumKey", album.AlbumKey).
			Str("nodeID", album.NodeID).
			Str("destination", out).
			Msg("album")

		var reqs []*request
		err = x.mg.Image.ImagesIter(ctx, album.AlbumKey, func(image *smugmug.Image) (bool, error) {
			var req *request
			dest := filepath.Join(out, image.FileName)
			req, err = x.request(image, dest)
			if err != nil {
				return false, err
			}
			if req == nil {
				x.metrics.IncrCounter([]string{"export", "download", "skipping", "exists"}, 1)
				log.Info().Str("imageKey", image.ImageKey).Str("filename", dest).Msg("skipping")
				return true, nil
			}
			x.metrics.IncrCounter([]string{"export", "download", "enqueued"}, 1)
			log.Info().
				Str("imageKey", image.ImageKey).
				Str("uri", req.URI).
				Str("url", req.URL).
				Str("destination", dest).
				Msg("enqueued")
			reqs = append(reqs, req)
			return true, nil
		}, smugmug.WithExpansions("ImageSizeDetails"))
		if err != nil {
			return false, err
		}

		if len(reqs) == 0 {
			return false, nil
		}

		err = x.fs.MkdirAll(out, 0755)
		if err != nil {
			return false, err
		}

		err = x.download(ctx, reqs)
		return err == nil, err
	}
}

func export(c *cli.Context) error {
	if c.NArg() < 2 {
		return fmt.Errorf("expected two arguments, not {%d}", c.NArg())
	}
	x := &exporter{
		mg:          runtime(c).Client,
		fs:          runtime(c).Fs,
		grab:        runtime(c).Grab,
		force:       c.Bool("force"),
		metrics:     runtime(c).Metrics,
		concurrency: c.Int("concurrency")}
	f := x.export(c.Context, c.Args().Get(1))
	return x.mg.Node.Walk(c.Context, c.Args().Get(0), func(node *smugmug.Node) (bool, error) {
		if node.Type == smugmug.TypeAlbum {
			return f(node.Album)
		}
		return true, nil
	}, smugmug.WithExpansions(smugmug.TypeAlbum))
}

func CommandExport() *cli.Command {
	return &cli.Command{
		Name:        "export",
		HelpName:    "export",
		Usage:       "export images from albums",
		Description: "export images from albums to local disk",
		ArgsUsage:   "<node id> <directory>",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "concurrency",
				Aliases: []string{"c"},
				Usage:   "the number of concurrent downloads",
				Value:   3,
			},
			&cli.BoolFlag{
				Name:  "force",
				Usage: "overwrite existing files",
				Value: false,
			},
		},
		Action: export,
	}
}
