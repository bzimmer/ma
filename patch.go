package ma

import (
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

type patchFunc func(c *cli.Context) (bool, string, interface{})

var patchFuncs = []patchFunc{
	keywords,
	str("caption"),
	str("title"),
	float("altitude"),
	float("latitude"),
	float("longitude"),
}

func keywords(c *cli.Context) (bool, string, interface{}) {
	if !c.IsSet("keyword") {
		return false, "keywords", nil
	}
	var kws []string
	for _, kw := range c.StringSlice("keyword") {
		switch kw {
		case "":
			return true, "KeywordArray", []string{}
		default:
			kws = append(kws, kw)
		}
	}
	return true, "KeywordArray", kws
}

func str(key string) patchFunc {
	title := strings.Title(key)
	return func(c *cli.Context) (bool, string, interface{}) {
		if !c.IsSet(key) {
			return false, key, nil
		}
		return true, title, c.String(key)
	}
}

func float(key string) patchFunc {
	title := strings.Title(key)
	return func(c *cli.Context) (bool, string, interface{}) {
		if !c.IsSet(key) {
			return false, key, nil
		}
		return true, title, c.Float64(key)
	}
}

func patch(c *cli.Context) error {
	patches := make(map[string]interface{})
	for _, f := range patchFuncs {
		ok, key, value := f(c)
		if ok {
			patches[key] = value
		}
	}

	for _, imageKey := range c.Args().Slice() {
		switch {
		case len(patches) == 0:
			log.Warn().Str("imageKey", imageKey).Msg("no patches to apply")
		case !c.Bool("force"):
			metric(c).IncrCounter([]string{"patch", "patched", "dryrun"}, 1)
			log.Info().Str("imageKey", imageKey).Interface("patches", patches).Msg("dryrun")
		default:
			metric(c).IncrCounter([]string{"patch", "patched"}, 1)
			log.Info().Str("imageKey", imageKey).Interface("patches", patches).Msg("applying")
			img, err := client(c).Image.Patch(c.Context, imageKey, patches)
			if err != nil {
				return err
			}
			f := imageIterFunc(c, nil, "patch")
			if _, err := f(img); err != nil {
				return err
			}
		}
	}
	return nil
}

func CommandPatch() *cli.Command {
	return &cli.Command{
		Name:      "patch",
		HelpName:  "patch",
		Usage:     "patch an image (or images)",
		ArgsUsage: "<image key> [<image key>, ...]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "force",
				Usage: "force must be specified to apply the patch",
				Value: false,
			},
			&cli.StringSliceFlag{
				Name:     "keyword",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "caption",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "title",
				Required: false,
			},
			&cli.Float64Flag{
				Name:     "latitude",
				Required: false,
			},
			&cli.Float64Flag{
				Name:     "longitude",
				Required: false,
			},
			&cli.Float64Flag{
				Name:     "altitude",
				Required: false,
			},
		},
		Action: patch,
	}
}
