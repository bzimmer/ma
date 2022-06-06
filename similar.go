package ma

import (
	"errors"
	goimage "image"
	"io/fs"

	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"github.com/vitali-fedulov/images3"
)

// https://github.com/vitali-fedulov/images3

func icon(afs afero.Fs, path string) (*images3.IconT, error) {
	file, err := afs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := goimage.Decode(file)
	if err != nil {
		return nil, err
	}
	ic := images3.Icon(img, path)
	return &ic, err
}

func similar(c *cli.Context) error {
	afs := runtime(c).Fs
	var icons []images3.IconT
	for i := 0; i < c.NArg(); i++ {
		err := afero.Walk(afs, c.Args().Get(i), func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			log.Info().Str("path", path).Msg("reading")
			ic, err := icon(afs, path)
			if err != nil {
				if !errors.Is(err, goimage.ErrFormat) {
					return err
				}
				log.Warn().Str("path", path).Msg("unknown format")
				return nil
			}
			icons = append(icons, *ic)
			return nil
		})
		if err != nil {
			return err
		}
	}

	n := len(icons)
	enc := runtime(c).Encoder
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			b := images3.Similar(icons[i], icons[j])
			if b {
				log.Info().
					Int("i", i).Int("j", j).
					Str("A", icons[i].Path).Str("B", icons[j].Path).
					Bool("similar", b).
					Msg(c.Command.Name)
			}
			if err := enc.Encode([]string{icons[i].Path, icons[j].Path}); err != nil {
				return err
			}
		}
	}
	return nil
}

func CommandSimilar() *cli.Command {
	return &cli.Command{
		Name:        "similar",
		HelpName:    "similar",
		Usage:       "identify similar images",
		Description: "identify similar images",
		Action:      similar,
	}
}
