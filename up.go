package ma

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/bzimmer/smugmug"
	"github.com/bzimmer/smugmug/uploadable/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
)

func visit(c *cli.Context) filesystem.PreFunc {
	return func(fs afero.Fs, filename string) (bool, error) {
		runtime(c).Metrics.IncrCounter([]string{"uploadable.fs", "visit"}, 1)
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
			runtime(c).Metrics.IncrCounter([]string{"uploadable.fs", "skip", "unsupported"}, 1)
			log.Info().Str("reason", "unsupported").Str("path", filename).Msg("skipping")
		}
		return ok, err
	}
}

func open(c *cli.Context) filesystem.UseFunc {
	return func(up *smugmug.Uploadable) (*smugmug.Uploadable, error) {
		runtime(c).Metrics.IncrCounter([]string{"uploadable.fs", "open"}, 1)
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
			runtime(c).Metrics.IncrCounter([]string{"uploadable.fs", "skip", "md5"}, 1)
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
			runtime(c).Metrics.IncrCounter([]string{"uploadable.fs", "replace"}, 1)
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

func uploadable(
	c *cli.Context, album *smugmug.Album, images map[string]*smugmug.Image) (<-chan *smugmug.Upload, <-chan error) {
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

func do(c *cli.Context, uploadc <-chan *smugmug.Upload, errc <-chan error) error {
	enc := runtime(c).Encoder
	for {
		select {
		case <-c.Done():
			return c.Err()
		case err := <-errc:
			return err
		case up, ok := <-uploadc:
			if !ok {
				return nil
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
				return err
			}
		}
	}
}

func deleteImages(c *cli.Context, album *smugmug.Album, images map[string]*smugmug.Image) error {
	mg := runtime(c).Client
	enc := runtime(c).Encoder
	met := runtime(c).Metrics
	dryrun := c.Bool("dryrun")
	log.Info().Int("count", len(images)).Msg("existing images to remove")
	for filename, image := range images {
		id := fmt.Sprintf("%s-%d", image.ImageKey, image.Serial)
		log.Info().
			Bool("dryrun", dryrun).
			Str("filename", filename).
			Str("albumKey", album.AlbumKey).
			Str("imageKey", id).
			Msg("delete")
		switch dryrun {
		case true:
			met.IncrCounter([]string{c.Command.Name, "mirror", "dryrun"}, 1)
		case false:
			met.IncrCounter([]string{c.Command.Name, "mirror", "delete"}, 1)
			met.IncrCounter([]string{c.Command.Name, "delete", "attempt"}, 1)
			res, err := mg.Image.Delete(c.Context, album.AlbumKey, id)
			if err != nil {
				return err
			}
			met.IncrCounter([]string{c.Command.Name, "delete", "success"}, 1)
			if err = enc.Encode(map[string]any{
				"AlbumKey": album.AlbumKey,
				"ImageKey": id,
				"Status":   res,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func mirror(c *cli.Context) error {
	var m sync.RWMutex
	album, images, err := existing(c, func(img *smugmug.Image) string {
		return img.FileName
	})
	if err != nil {
		return err
	}
	u, err := filesystem.NewFsUploadable(album.AlbumKey)
	if err != nil {
		return err
	}
	u.Pre(filesystem.Extensions(c.StringSlice("ext")...))
	u.Pre(func(fs afero.Fs, filename string) (bool, error) {
		m.Lock()
		delete(images, filepath.Base(filename))
		m.Unlock()
		return false, nil
	})
	uploadc, errc := filesystem.
		NewFsUploadables(runtime(c).Fs, c.Args().Slice(), u).
		Uploadables(c.Context)
	for {
		select {
		case <-c.Done():
			return c.Err()
		case err = <-errc:
			return err
		case _, ok := <-uploadc:
			if !ok {
				return deleteImages(c, album, images)
			}
		}
	}
}

func up(c *cli.Context) error {
	album, images, err := existing(c, func(img *smugmug.Image) string {
		return img.FileName
	})
	if err != nil {
		return err
	}
	uc, ec := uploadable(c, album, images)
	if err = do(c, uc, ec); err != nil {
		return err
	}
	if c.Bool("mirror") {
		if err = mirror(c); err != nil {
			return err
		}
	}
	log.Info().Msg("done")
	return nil
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
				Usage:    "mirror the local filesystem with a SmugMug gallery",
				Value:    false,
				Required: false,
			},
		},
		Action: up,
	}
}
