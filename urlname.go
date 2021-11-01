package ma

import (
	"errors"
	"strings"
	"unicode"

	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

var ErrInvalidURLName = errors.New("node url name must start with a number or capital letter")

func validate(urlName string) error {
	if urlName == "" {
		return ErrInvalidURLName
	}
	v := rune(urlName[0])
	switch {
	case unicode.IsNumber(v), unicode.IsUpper(v):
		return nil
	default:
		return ErrInvalidURLName
	}
}

// urlname returns a valid SmugMug UrlName for `name`
// This function replaces "unpleasant" values such as `'s` and `-` to make for a cleaner UrlName
func urlname(name string) string {
	for _, x := range [][]string{
		{"'s", "s"}, {"-", " "},
	} {
		name = strings.ReplaceAll(name, x[0], x[1])
	}
	return smugmug.URLName(name)
}

func CommandURLName() *cli.Command {
	return &cli.Command{
		Name:        "urlname",
		HelpName:    "urlname",
		Usage:       "process the argument as a url name",
		Description: "create a clean url for the argument by removing \"unpleasant\" values such as `'s` and `-`",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "validate",
				Aliases: []string{"a"},
				Usage:   "validate the url name",
			},
		},
		Action: func(c *cli.Context) error {
			enc := runtime(c).Encoder
			for i := 0; i < c.NArg(); i++ {
				arg, url, valid := c.Args().Get(i), "", true
				switch c.Bool("validate") {
				case true:
					if err := validate(arg); err != nil {
						valid = false
					}
				case false:
					url = urlname(arg)
				}
				runtime(c).Metrics.IncrCounter([]string{"urlname", c.Command.Name}, 1)
				log.Info().Str("name", arg).Str("url", url).Bool("valid", valid).Msg(c.Command.Name)
				if err := enc.Encode(map[string]interface{}{
					"Name": arg, "UrlName": url, "Valid": valid,
				}); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
