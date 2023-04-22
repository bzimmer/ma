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

func (p patchKey) Title() string {
	var key string
	switch p {
	case patchKeyAlbum:
		key = "AlbumKey"
	case patchKeyImage:
		key = "ImageKey"
	}
	return key
}

type patcher struct {
	c       *cli.Context
	err     error
	patches map[string]any
}

func with(c *cli.Context) *patcher {
	return &patcher{c: c, patches: make(map[string]any)}
}

func (p *patcher) patch(k patchKey, key string) error {
	if p.err != nil {
		return p.err
	}
	client := runtime(p.c).Smugmug()
	switch k {
	case patchKeyAlbum:
		album, err := client.Album.Patch(p.c.Context, key, p.patches)
		if err != nil {
			return err
		}
		f := albumIterFunc(p.c, "patch")
		_, err = f(album)
		return err
	case patchKeyImage:
		image, err := client.Image.Patch(p.c.Context, key, p.patches)
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
	case p.c.Bool("auto"):
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
		for _, flag := range []string{"title", "name", "caption", "description"} {
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
		enc := runtime(c).Encoder
		for i := 0; i < c.NArg(); i++ {
			id := c.Args().Get(i)
			msg := log.Info().Str(key.String(), id).Interface("patches", p.patches)
			if err := enc.Encode(map[string]any{
				key.Title(): id,
				"Patches":   p.patches,
			}); err != nil {
				return err
			}
			switch {
			case c.Bool("dryrun"):
				msg.Msg("dryrun")
				runtime(c).Metrics.IncrCounter([]string{"patch", c.Command.Name, "dryrun"}, 1)
			default:
				msg.Msg("apply")
				runtime(c).Metrics.IncrCounter([]string{"patch", c.Command.Name, "apply"}, 1)
				if err := p.patch(key, id); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func dryrunFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:    "dryrun",
		Aliases: []string{"n"},
		Usage:   "dryrun the patches",
		Value:   false,
	}
}

func albumPatch() *cli.Command {
	return &cli.Command{
		Name:        "album",
		HelpName:    "album",
		Usage:       "Patch an album ",
		Description: "Patch the metadata of a single album",
		ArgsUsage:   "<album key> [<album key>, ...]",
		Flags: []cli.Flag{
			dryrunFlag(),
			&cli.StringFlag{
				Name:  "urlname",
				Usage: "the urlname of the album",
			},
			&cli.BoolFlag{
				Name:    "auto",
				Aliases: []string{"A", "auto-urlname"},
				Usage:   "auto-generate the urlname (album name is required)",
			},
			&cli.StringSliceFlag{
				Name:  "keyword",
				Usage: "a set of keywords describing the album",
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "the album name",
			},
			&cli.StringFlag{
				Name:  "description",
				Usage: "the album description",
			},
		},
		Before: func(c *cli.Context) error {
			switch {
			case c.IsSet("auto") && c.IsSet("urlname"):
				return errors.New("only one of `auto` or `urlname` may be specified")
			case c.IsSet("auto") && !c.IsSet("name"):
				return errors.New("cannot specify `auto` without `name`")
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
			dryrunFlag(),
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
