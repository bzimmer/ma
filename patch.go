package ma

import (
	"errors"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type patchFunc func(c *cli.Context) (bool, string, interface{}, error)

type patcherFunc func(*cli.Context, string, map[string]interface{}) error

func patchFuncs() []patchFunc {
	return []patchFunc{
		keywords,
		str("name"),
		str("title"),
		str("caption"),
		float("altitude"),
		float("latitude"),
		float("longitude"),
		urlname("urlname"),
	}
}

func keywords(c *cli.Context) (bool, string, interface{}, error) {
	if !c.IsSet("keyword") {
		return false, "keywords", nil, nil
	}
	var kws []string
	keyword := "KeywordArray"
	for _, kw := range c.StringSlice("keyword") {
		switch kw {
		case "":
			return true, keyword, []string{}, nil
		default:
			kws = append(kws, kw)
		}
	}
	return true, keyword, kws, nil
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

func urlname(key string) patchFunc {
	title := "UrlName"
	return func(c *cli.Context) (bool, string, interface{}, error) {
		if !c.IsSet(key) {
			return false, key, nil, nil
		}
		url := c.String(key)
		if err := validateURLName(url); err != nil {
			log.Error().Err(err).Str("urlname", url).Msg("invalid")
			return false, key, nil, err
		}
		return true, title, url, nil
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

func imagePatcher(c *cli.Context, imageKey string, patches map[string]interface{}) error {
	img, err := client(c).Image.Patch(c.Context, imageKey, patches)
	if err != nil {
		return err
	}
	f := imageIterFunc(c, nil, "patch")
	if _, err := f(img); err != nil {
		return err
	}
	return nil
}

func albumPatcher(c *cli.Context, albumKey string, patches map[string]interface{}) error {
	album, err := client(c).Album.Patch(c.Context, albumKey, patches)
	if err != nil {
		return err
	}
	f := albumIterFunc(c, "patch")
	if _, err := f(album); err != nil {
		return err
	}
	return nil
}

func patch(keyName string, patcher patcherFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		patches := make(map[string]interface{})
		for _, f := range patchFuncs() {
			ok, key, value, err := f(c)
			if err != nil {
				return err
			}
			if ok {
				patches[key] = value
			}
		}
		for _, x := range c.Args().Slice() {
			switch {
			case len(patches) == 0:
				log.Warn().Str(keyName, x).Msg("no patches to apply")
			case !c.Bool("force"):
				metric(c).IncrCounter([]string{"patch", c.Command.Name, "dryrun"}, 1)
				log.Info().Str(keyName, x).Interface("patches", patches).Msg("dryrun")
			default:
				log.Info().Str(keyName, x).Interface("patches", patches).Msg("applying")
				if err := patcher(c, x, patches); err != nil {
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
			switch c.NArg() {
			case 0:
				return errors.New("expected one albumKey argument")
			case 1:
				return nil
			default:
				return errors.New("expected only one albumKey argument")
			}
		},
		Action: patch("albumKey", albumPatcher),
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
		Action: patch("imageKey", imagePatcher),
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
