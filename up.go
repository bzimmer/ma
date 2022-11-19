package ma

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/bzimmer/smugmug"
	"github.com/bzimmer/smugmug/uploadable/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
)

func input(c *cli.Context) ([]string, error) {
	if !c.Bool("0") {
		return c.Args().Slice(), nil
	}
	log.Info().Msg("reading paths from stdin")
	data, err := io.ReadAll(c.App.Reader)
	if err != nil {
		return nil, err
	}
	var out []string
	for {
		x := bytes.IndexByte(data, 0)
		if x == -1 {
			break
		}
		out = append(out, string(data[0:x]))
		data = data[x+1:]
	}
	return out, nil
}

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
	return func(up *smugmug.Uploadable) error {
		runtime(c).Metrics.IncrCounter([]string{"uploadable.fs", "open"}, 1)
		return nil
	}
}

func skip(c *cli.Context, images map[string]*smugmug.Image) filesystem.UseFunc {
	f := filesystem.Skip(false, images)
	return func(up *smugmug.Uploadable) error {
		if err := f(up); err != nil {
			if errors.Is(err, filesystem.ErrSkip) {
				runtime(c).Metrics.IncrCounter([]string{"uploadable.fs", "skip", "md5"}, 1)
				log.Info().Str("reason", "md5").Str("path", up.Name).Msg("skipping")
				return filesystem.ErrSkip
			}
			return err
		}
		return nil
	}
}

func replace(c *cli.Context, images map[string]*smugmug.Image) filesystem.UseFunc {
	f := filesystem.Replace(true, images)
	return func(up *smugmug.Uploadable) error {
		if err := f(up); err != nil {
			if errors.Is(err, filesystem.ErrSkip) {
				return nil
			}
			return err
		}
		if up.Replaces != "" {
			runtime(c).Metrics.IncrCounter([]string{"uploadable.fs", "replace"}, 1)
		}
		return nil
	}
}

func attempt(c *cli.Context) filesystem.UseFunc {
	return func(up *smugmug.Uploadable) error {
		info := log.Info().
			Str("name", up.Name).
			Str("album", up.AlbumKey).
			Str("replaces", up.Replaces)
		if c.Bool("dryrun") {
			info.Str("status", "dryrun").Msg("upload")
			runtime(c).Metrics.IncrCounter([]string{"upload", "dryrun"}, 1)
			return filesystem.ErrSkip
		}
		info.Str("status", "attempt").Msg("upload")
		runtime(c).Metrics.IncrCounter([]string{"upload", "attempt"}, 1)
		return nil
	}
}

type upload struct{}

func (x *upload) upload(c *cli.Context) error {
	in, err := input(c)
	if err != nil {
		return err
	}
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
	u.Pre(visit(c), extensions(c))
	u.Use(open(c), skip(c, images), replace(c, images), attempt(c))
	uc, ec := runtime(c).Client.Upload.Uploads(
		c.Context, filesystem.NewFsUploadables(runtime(c).Fs, in, u))
	return x.up(c, uc, ec)
}

func (x *upload) up(c *cli.Context, uploadc <-chan *smugmug.Upload, errc <-chan error) error {
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

type mirror struct{}

func (x *mirror) mirror(c *cli.Context) error {
	var m sync.RWMutex
	in, err := input(c)
	if err != nil {
		return err
	}
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
		NewFsUploadables(runtime(c).Fs, in, u).
		Uploadables(c.Context)
	for {
		select {
		case <-c.Done():
			return c.Err()
		case err = <-errc:
			return err
		case _, ok := <-uploadc:
			if !ok {
				return x.delete(c, album, images)
			}
		}
	}
}

func (x *mirror) delete(c *cli.Context, album *smugmug.Album, images map[string]*smugmug.Image) error {
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

func up(c *cli.Context) error {
	var u upload
	if err := u.upload(c); err != nil {
		return err
	}
	if c.Bool("mirror") {
		var m mirror
		if err := m.mirror(c); err != nil {
			return err
		}
	}
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
			&cli.BoolFlag{
				Name:     "0",
				Usage:    "read null byte terminated strings from stdin",
				Value:    false,
				Required: false,
			},
		},
		Action: up,
	}
}
