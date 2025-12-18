package ma

import (
	"errors"
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

func CommandURLName() *cli.Command {
	return &cli.Command{
		Name:        "urlname",
		HelpName:    "urlname",
		Usage:       "Create a clean urlname for each argument",
		Description: "Create a clean urlname for each argument",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "validate",
				Aliases: []string{"a"},
				Usage:   "Validate the url name",
			},
		},
		Action: func(c *cli.Context) error {
			val := c.Bool("validate")
			enc := runtime(c).Encoder
			for i := range c.NArg() {
				arg, url, valid := c.Args().Get(i), "", true
				if val {
					if err := validate(arg); err != nil {
						valid = false
					}
				} else {
					url = smugmug.URLName(arg, runtime(c).Language)
				}
				runtime(c).Metrics.IncrCounter([]string{"urlname", c.Command.Name}, 1)
				log.Debug().Str("name", arg).Str("url", url).Bool("valid", valid).Msg(c.Command.Name)
				if err := enc.Encode(map[string]any{
					"Name": arg, "UrlName": url, "Valid": valid,
				}); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
