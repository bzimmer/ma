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

	log.Info().Str("nickname", user.NickName).Str("uri", user.URI).Msg("user")

	return nil
}

func CommandUser() *cli.Command {
	return &cli.Command{
		Name:    "user",
		Aliases: []string{"u", "authuser"},
		Usage:   "query the authenticated user",
		Action:  user,
	}
}
