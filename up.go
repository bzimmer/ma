package ma

import (
	"fmt"

	"github.com/bzimmer/smugmug"
	"github.com/bzimmer/smugmug/uploadable/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

func visit(c *cli.Context) filesystem.PreFunc {
	return func(fs afero.Fs, filename string) (bool, error) {
		runtime(c).Metrics.IncrCounter([]string{"fsUploadable", "visit"}, 1)
		return true, nil
	}
}

func extensions(c *cli.Context) filesystem.PreFunc {
	f := filesystem.Extensions(c.StringSlice("ext")...)
	return func(fs afero.Fs, filename string) (bool, error) {
		ok, err := f(fs, filename)
		if err != nil {
			return false, err
		}
		if !ok {
			runtime(c).Metrics.IncrCounter([]string{"fsUploadable", "skip", "unsupported"}, 1)
			log.Info().Str("reason", "unsupported").Str("path", filename).Msg("skipping")
		}
		return ok, err
	}
}

func open(c *cli.Context) filesystem.UseFunc {
	return func(up *smugmug.Uploadable) (*smugmug.Uploadable, error) {
		runtime(c).Metrics.IncrCounter([]string{"fsUploadable", "open"}, 1)
		return up, nil
	}
}

func skip(c *cli.Context, images map[string]*smugmug.Image) filesystem.UseFunc {
	f := filesystem.Skip(false, images)
	return func(up *smugmug.Uploadable) (*smugmug.Uploadable, error) {
		sup, err := f(up)
		if err != nil {
			return nil, err
		}
		if sup == nil {
			runtime(c).Metrics.IncrCounter([]string{"fsUploadable", "skip", "md5"}, 1)
			log.Info().Str("reason", "md5").Str("path", up.Name).Msg("skipping")
			return nil, err
		}
		return sup, nil
	}
}

func replace(c *cli.Context, images map[string]*smugmug.Image) filesystem.UseFunc {
	f := filesystem.Replace(true, images)
	return func(up *smugmug.Uploadable) (*smugmug.Uploadable, error) {
		up, err := f(up)
		if err != nil {
			return nil, err
		}
		if up == nil {
			return nil, nil
		}
		if up.Replaces != "" {
			runtime(c).Metrics.IncrCounter([]string{"fsUploadable", "replace"}, 1)
		}
		return up, err
	}
}

func upload(c *cli.Context) filesystem.UseFunc {
	return func(up *smugmug.Uploadable) (*smugmug.Uploadable, error) {
		info := log.Info().
			Str("name", up.Name).
			Str("album", up.AlbumKey).
			Str("replaces", up.Replaces)
		if c.Bool("dryrun") {
			info.Str("status", "dryrun").Msg("upload")
			runtime(c).Metrics.IncrCounter([]string{"upload", "dryrun"}, 1)
			return nil, nil
		}
		info.Str("status", "attempt").Msg("upload")
		runtime(c).Metrics.IncrCounter([]string{"upload", "attempt"}, 1)
		return up, nil
	}
}

func existing(c *cli.Context) (*smugmug.Album, map[string]*smugmug.Image, error) {
	mg := runtime(c).Client
	albumKey := c.String("album")
	albumc := make(chan *smugmug.Album, 1)
	imagesc := make(chan map[string]*smugmug.Image, 1)
	grp, ctx := errgroup.WithContext(c.Context)
	grp.Go(func() error {
		images := make(map[string]*smugmug.Image)
		log.Info().Str("albumKey", albumKey).Msg("querying existing gallery images")
		if err := mg.Image.ImagesIter(ctx, albumKey, func(img *smugmug.Image) (bool, error) {
			images[img.FileName] = img
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

func uploadable(c *cli.Context, album *smugmug.Album, images map[string]*smugmug.Image) (<-chan *smugmug.Upload, <-chan error) {
	mg := runtime(c).Client
	u, err := filesystem.NewFsUploadable(album.AlbumKey)
	if err != nil {
		return nil, nil
	}
	u.Pre(visit(c), extensions(c))
	u.Use(open(c), skip(c, images), replace(c, images), upload(c))
	return mg.Upload.Uploads(
		c.Context, filesystem.NewFsUploadables(runtime(c).Fs, c.Args().Slice(), u))
}

func do(c *cli.Context, uploadc <-chan *smugmug.Upload, errc <-chan error) ([]string, error) {
	var filenames []string
	enc := runtime(c).Encoder
Done:
	for {
		select {
		case <-c.Done():
			return nil, c.Err()
		case err := <-errc:
			return nil, err
		case up, ok := <-uploadc:
			if !ok {
				break Done
			}
			runtime(c).Metrics.IncrCounter([]string{"upload", "success"}, 1)
			runtime(c).Metrics.AddSample([]string{"upload", "upload"}, float32(up.Elapsed.Seconds()))
			log.Info().
				Str("name", up.Uploadable.Name).
				Str("album", up.Uploadable.AlbumKey).
				Dur("elapsed", up.Elapsed).
				Str("uri", up.ImageURI).
				Str("status", "success").
				Msg("upload")
			if err := enc.Encode(up); err != nil {
				return nil, err
			}
			filenames = append(filenames, up.Uploadable.Name)
		}
	}
	log.Info().Msg("complete")
	return filenames, nil
}

func mirror(c *cli.Context, images map[string]*smugmug.Image, filenames []string) error {
	if !c.Bool("mirror") {
		return nil
	}
	mg := runtime(c).Client
	album := c.String("album")
	dryrun := c.Bool("dryrun")
	f := func(filename string, image *smugmug.Image) error {
		imageKey := fmt.Sprintf("%s-%d", image.ImageKey, image.Serial)
		log.Info().
			Str("filename", filename).
			Str("imageKey", imageKey).
			Bool("dryrun", dryrun).
			Msg("mirror")
		switch dryrun {
		case true:
			runtime(c).Metrics.IncrCounter([]string{"mirror", "dryrun"}, 1)
		case false:
			runtime(c).Metrics.IncrCounter([]string{"mirror", "delete"}, 1)
			if _, err := mg.Image.Delete(c.Context, album, imageKey); err != nil {
				return err
			}
		}
		return nil
	}
	for _, filename := range filenames {
		log.Info().Msg(filename)
		delete(images, filename)
	}
	for filename, image := range images {
		if err := f(filename, image); err != nil {
			return err
		}
	}
	return nil
}

func up(c *cli.Context) error {
	album, images, err := existing(c)
	if err != nil {
		return err
	}

	uc, ec := uploadable(c, album, images)
	filenames, err := do(c, uc, ec)
	if err != nil {
		return err
	}

	return mirror(c, images, filenames)
}

func CommandUpload() *cli.Command {
	return &cli.Command{
		Name:        "up",
		Aliases:     []string{"upload"},
		Usage:       "Upload images to SmugMug",
		Description: "Upload image files to the specified album, selectively including specific file extensions",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "album",
				Usage:    "the album to which image files will be uploaded",
				Aliases:  []string{"a"},
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "ext",
				Usage:    "the set of file extensions suitable for uploading",
				Aliases:  []string{"x"},
				Required: false,
				Value:    cli.NewStringSlice(".jpg", ".jpeg"),
			},
			&cli.BoolFlag{
				Name:     "dryrun",
				Usage:    "prepare to upload but don't actually do it",
				Aliases:  []string{"n"},
				Value:    false,
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "mirror",
				Usage:    "mirror the smugmug gallery to the local filesystem",
				Value:    false,
				Required: false,
			},
		},
		Action: up,
	}
}
