package ma

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/cases"
)

func CommandTitle() *cli.Command {
	return &cli.Command{
		Name:        "title",
		HelpName:    "title",
		Usage:       "Create a title following the specified convention",
		Description: "Create a title following the specified convention",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "caser",
				Usage:    "The case algorithm to use, one of 'upper', 'lower', or 'title'",
				Value:    "title",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			enc := runtime(c).Encoder
			tag := runtime(c).Language
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
			for i := 0; i < c.NArg(); i++ {
				title := c.Args().Get(i)
				runtime(c).Metrics.IncrCounter([]string{c.Command.Name, c.String("caser")}, 1)
				log.Info().Str("title", title).Str("caser", c.String("caser")).Str("lang", tag.String()).Msg(c.Command.Name)
				if err := enc.Encode(map[string]string{"Title": caser.String(title)}); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
