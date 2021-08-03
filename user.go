package ma

import (
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func user(c *cli.Context) error {
	mg, err := client(c)
	if err != nil {
		return err
	}

	user, err := mg.User.AuthUser(c.Context, smugmug.WithExpansions("Node"))
	if err != nil {
		return err
	}

	switch c.Bool("json") {
	case true:
		enc := encoder(c)
		msg := map[string]interface{}{
			"nickname": user.NickName,
			"uri":      user.URI,
		}
		if err := enc.Encode(msg); err != nil {
			return err
		}
	default:
		log.Info().Str("nickname", user.NickName).Str("uri", user.URI).Msg("user")
	}

	return nil
}

func CommandUser() *cli.Command {
	return &cli.Command{
		Name:   "user",
		Usage:  "query the authenticated user",
		Action: user,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "json",
				Aliases:  []string{"j"},
				Value:    false,
				Required: false,
			},
		},
	}
}
