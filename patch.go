package ma

import (
	"errors"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

const (
	patchName     = "Name"
	patchURLName  = "UrlName"
	patchKeywords = "KeywordArray"
)

type patches map[string]interface{}

type patchFunc func(c *cli.Context) (bool, string, interface{}, error)

type patcher interface {
	finalize(c *cli.Context, patches patches) error
	patch(c *cli.Context, keyName string, patches patches) error
}

func patchFuncs() []patchFunc {
	return []patchFunc{
		keywords("keyword"),
		str("name"),
		str("title"),
		str("caption"),
		float("altitude"),
		float("latitude"),
		float("longitude"),
		url("urlname"),
	}
}

func keywords(key string) patchFunc {
	return func(c *cli.Context) (bool, string, interface{}, error) {
		if !c.IsSet(key) {
			return false, key, nil, nil
		}
		var kws []string
		for _, kw := range c.StringSlice(key) {
			switch kw {
			case "":
				return true, patchKeywords, []string{}, nil
			default:
				kws = append(kws, kw)
			}
		}
		return true, patchKeywords, kws, nil
	}
}

func str(key string) patchFunc {
	title := strings.Title(key)
	return func(c *cli.Context) (bool, string, interface{}, error) {
		if !c.IsSet(key) {
			return false, key, nil, nil
		}
		return true, title, c.String(key), nil
	}
}

func url(key string) patchFunc {
	return func(c *cli.Context) (bool, string, interface{}, error) {
		if !c.IsSet(key) {
			return false, key, nil, nil
		}
		url := c.String(key)
		if err := validateURLName(url); err != nil {
			log.Error().Err(err).Str("urlname", url).Msg("invalid")
			return false, key, nil, err
		}
		return true, patchURLName, url, nil
	}
}

func float(key string) patchFunc {
	title := strings.Title(key)
	return func(c *cli.Context) (bool, string, interface{}, error) {
		if !c.IsSet(key) {
			return false, key, nil, nil
		}
		return true, title, c.Float64(key), nil
	}
}

type imagePatcher struct{}

func (p *imagePatcher) finalize(c *cli.Context, patches patches) error {
	return nil
}

func (p *imagePatcher) patch(c *cli.Context, imageKey string, patches patches) error {
	img, err := runtime(c).Client.Image.Patch(c.Context, imageKey, patches)
	if err != nil {
		return err
	}
	f := imageIterFunc(c, nil, "patch")
	if _, err := f(img); err != nil {
		return err
	}
	return nil
}

type albumPatcher struct{}

func (p *albumPatcher) finalize(c *cli.Context, patches patches) error {
	if !c.Bool("auto-urlname") {
		return nil
	}
	if name, ok := patches[patchName]; ok {
		if v, ok := name.(string); ok {
			patches[patchURLName] = urlname(v)
		}
	}
	return nil
}

func (p *albumPatcher) patch(c *cli.Context, albumKey string, patches patches) error {
	album, err := runtime(c).Client.Album.Patch(c.Context, albumKey, patches)
	if err != nil {
		return err
	}
	f := albumIterFunc(c, "patch")
	if _, err := f(album); err != nil {
		return err
	}
	return nil
}

func patch(keyName string, p patcher) cli.ActionFunc {
	return func(c *cli.Context) error {
		ps := make(patches)
		for _, f := range patchFuncs() {
			ok, key, value, err := f(c)
			if err != nil {
				return err
			}
			if ok {
				ps[key] = value
			}
		}
		if err := p.finalize(c, ps); err != nil {
			return err
		}
		for _, x := range c.Args().Slice() {
			switch {
			case len(ps) == 0:
				log.Warn().Str(keyName, x).Msg("no patches to apply")
			case !c.Bool("force"):
				runtime(c).Metrics.IncrCounter([]string{"patch", c.Command.Name, "dryrun"}, 1)
				log.Info().Str(keyName, x).Interface("patches", ps).Msg("dryrun")
			default:
				log.Info().Str(keyName, x).Interface("patches", ps).Msg("applying")
				if err := p.patch(c, x, ps); err != nil {
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
		Name:      "album",
		HelpName:  "album",
		Usage:     "patch an album (or albums)",
		ArgsUsage: "<album key> [<album key>, ...]",
		Flags: []cli.Flag{
			forceFlag(),
			&cli.BoolFlag{
				Name:  "auto-urlname",
				Usage: "if enabled, and an album name provided as a flag, the urlname will be auto-generated from the name",
			},
			&cli.StringSliceFlag{
				Name: "keyword",
			},
			&cli.StringFlag{
				Name: "name",
			},
			&cli.StringFlag{
				Name: "urlname",
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
		Action: patch("albumKey", &albumPatcher{}),
	}
}

func imagePatch() *cli.Command {
	return &cli.Command{
		Name:      "image",
		HelpName:  "image",
		Usage:     "patch an image (or images)",
		ArgsUsage: "<image key> [<image key>, ...]",
		Flags: []cli.Flag{
			forceFlag(),
			&cli.StringSliceFlag{
				Name: "keyword",
			},
			&cli.StringFlag{
				Name: "caption",
			},
			&cli.StringFlag{
				Name: "title",
			},
			&cli.Float64Flag{
				Name: "latitude",
			},
			&cli.Float64Flag{
				Name: "longitude",
			},
			&cli.Float64Flag{
				Name: "altitude",
			},
		},
		Action: patch("imageKey", &imagePatcher{}),
	}
}

func CommandPatch() *cli.Command {
	return &cli.Command{
		Name:     "patch",
		HelpName: "patch",
		Usage:    "patch the metadata for albums and images",
		Subcommands: []*cli.Command{
			albumPatch(),
			imagePatch(),
		},
	}
}
