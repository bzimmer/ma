package ma

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func CommandTitle() *cli.Command {
	return &cli.Command{
		Name:        "title",
		HelpName:    "title",
		Usage:       "Create a title following the specified convention",
		Description: "Create a title following the specified convention",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "language",
				Aliases:  []string{"lang", "l"},
				Usage:    "The language being cased as a BCP 47 language tag (eg 'en', 'de')",
				Value:    "en",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "caser",
				Usage:    "The case algorithm to use, one of 'upper', 'lower', or 'title'",
				Value:    "title",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			enc := runtime(c).Encoder
			tag, err := language.Parse(c.String("language"))
			if err != nil {
				return err
			}
			var caser cases.Caser
			switch c.String("caser") {
			case "lower":
				caser = cases.Lower(tag)
			case "upper":
				caser = cases.Upper(tag)
			case "title":
				caser = cases.Title(tag, cases.NoLower)
			default:
				return fmt.Errorf("unknown caser: %s", c.String("caser"))
			}
			runtime(c).Metrics.IncrCounter([]string{c.Command.Name, c.String("language")}, 1)
			log.Info().Str("caser", c.String("caser")).Str("lang", tag.String()).Msg(c.Command.Name)
			return enc.Encode(map[string]string{
				"Title": caser.String(strings.Join(c.Args().Slice(), " ")),
			})
		},
	}
}
