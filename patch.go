package ma

import (
	"errors"

	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type patchKey int

const (
	patchKeyAlbum patchKey = iota
	patchKeyImage
)

func (p patchKey) String() string {
	var key string
	switch p {
	case patchKeyAlbum:
		key = "albumKey"
	case patchKeyImage:
		key = "imageKey"
	}
	return key
}

type patcher struct {
	c       *cli.Context
	err     error
	patches map[string]interface{}
}

func with(c *cli.Context) *patcher {
	return &patcher{c: c, patches: make(map[string]interface{})}
}

func (p *patcher) patch(k patchKey, key string) error {
	if p.err != nil {
		return p.err
	}
	switch k {
	case patchKeyAlbum:
		album, err := runtime(p.c).Client.Album.Patch(p.c.Context, key, p.patches)
		if err != nil {
			return err
		}
		f := albumIterFunc(p.c, "patch")
		_, err = f(album)
		return err
	case patchKeyImage:
		image, err := runtime(p.c).Client.Image.Patch(p.c.Context, key, p.patches)
		if err != nil {
			return err
		}
		f := imageIterFunc(p.c, nil, "patch")
		_, err = f(image)
		return err
	}
	return nil
}

func (p *patcher) str(key string) *patcher {
	if p.err != nil || !p.c.IsSet(key) {
		return p
	}
	p.patches[titlecase(p.c, key)] = p.c.String(key)
	return p
}

func (p *patcher) float(key string) *patcher {
	if p.err != nil || !p.c.IsSet(key) {
		return p
	}
	p.patches[titlecase(p.c, key)] = p.c.Float64(key)
	return p
}

func (p *patcher) urlname() *patcher {
	if p.err != nil {
		return p
	}
	var url string
	switch {
	case p.c.Bool("auto-urlname"):
		url = smugmug.URLName(p.c.String("name"), runtime(p.c).Language)
	default:
		if !p.c.IsSet("urlname") {
			return p
		}
		url = p.c.String("urlname")
	}
	if err := validate(url); err != nil {
		log.Error().Err(err).Str("urlname", url).Msg("invalid")
		p.err = err
		return p
	}
	p.patches["UrlName"] = url
	return p
}

func (p *patcher) keywords(key string) *patcher {
	if p.err != nil || !p.c.IsSet(key) {
		return p
	}
	var kws []string
	for _, kw := range p.c.StringSlice(key) {
		switch kw {
		case "":
			kws = []string{}
		default:
			kws = append(kws, kw)
		}
	}
	p.patches["KeywordArray"] = kws
	return p
}

func patch(key patchKey) cli.ActionFunc {
	return func(c *cli.Context) error {
		p := with(c).keywords("keyword").urlname()
		for _, flag := range []string{"title", "name", "caption"} {
			p = p.str(flag)
		}
		for _, flag := range []string{"latitude", "longitude", "altitude"} {
			p = p.float(flag)
		}
		if p.err != nil {
			return p.err
		}
		if len(p.patches) == 0 {
			log.Warn().Msg("no patches to apply")
			return nil
		}
		for _, x := range c.Args().Slice() {
			msg := log.Info().Str(key.String(), x).Interface("patches", p.patches)
			switch {
			case !c.Bool("force"):
				msg.Msg("dryrun")
				runtime(c).Metrics.IncrCounter([]string{"patch", c.Command.Name, "dryrun"}, 1)
			default:
				msg.Msg("apply")
				runtime(c).Metrics.IncrCounter([]string{"patch", c.Command.Name, "apply"}, 1)
				if err := p.patch(key, x); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func forceFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:    "force",
		Aliases: []string{"f"},
		Usage:   "force must be specified to apply the patch",
		Value:   false,
	}
}

func albumPatch() *cli.Command {
	return &cli.Command{
		Name:        "album",
		HelpName:    "album",
		Usage:       "patch an album ",
		Description: "patch the metadata of a single album",
		ArgsUsage:   "<album key> [<album key>, ...]",
		Flags: []cli.Flag{
			forceFlag(),
			&cli.BoolFlag{
				Name:  "auto-urlname",
				Usage: "if enabled, and an album name provided as a flag, the urlname will be auto-generated from the name",
			},
			&cli.StringSliceFlag{
				Name:  "keyword",
				Usage: "a set of keywords describing the album",
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "the name of the album",
			},
			&cli.StringFlag{
				Name:  "urlname",
				Usage: "the urlname of the album (see `--auto-urlname` to set this automatically based on the album name)",
			},
		},
		Before: func(c *cli.Context) error {
			switch {
			case c.IsSet("auto-urlname") && c.IsSet("urlname"):
				return errors.New("only one of `auto-urlname` or `urlname` may be specified")
			case c.IsSet("auto-urlname") && !c.IsSet("name"):
				return errors.New("cannot specify `auto-urlname` without `name`")
			}
			switch c.NArg() {
			case 0:
				return errors.New("expected one albumKey argument")
			case 1:
				return nil
			default:
				return errors.New("expected only one albumKey argument")
			}
		},
		Action: patch(patchKeyAlbum),
	}
}

func imagePatch() *cli.Command {
	return &cli.Command{
		Name:        "image",
		HelpName:    "image",
		Usage:       "patch an image (or images)",
		Description: "patch the metadata of an image (not the image itself though)",
		ArgsUsage:   "<image key> [<image key>, ...]",
		Flags: []cli.Flag{
			forceFlag(),
			&cli.StringSliceFlag{
				Name:  "keyword",
				Usage: "specifies keywords describing the image",
			},
			&cli.StringFlag{
				Name:  "caption",
				Usage: "the caption of the image",
			},
			&cli.StringFlag{
				Name:  "title",
				Usage: "the title of the image",
			},
			&cli.Float64Flag{
				Name:  "latitude",
				Usage: "the latitude of the image location",
			},
			&cli.Float64Flag{
				Name:  "longitude",
				Usage: "the longitude of the image location",
			},
			&cli.Float64Flag{
				Name:  "altitude",
				Usage: "the altitude of the image location",
			},
		},
		Action: patch(patchKeyImage),
	}
}

func CommandPatch() *cli.Command {
	return &cli.Command{
		Name:        "patch",
		HelpName:    "patch",
		Usage:       "patch the metadata of albums and images",
		Description: "patch enables updating the metadata of both albums and images",
		Subcommands: []*cli.Command{
			albumPatch(),
			imagePatch(),
		},
	}
}
