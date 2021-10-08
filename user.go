package ma

import (
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func user(c *cli.Context) error {
	user, err := client(c).User.AuthUser(c.Context, smugmug.WithExpansions("Node"))
	if err != nil {
		return err
	}
	metric(c).IncrCounter([]string{"user", "user"}, 1)
	log.Info().Str("nickname", user.NickName).Str("uri", user.URI).Str("nodeID", user.Node.NodeID).Msg("user")
	return encoder(c).Encode(user)
}

func CommandUser() *cli.Command {
	return &cli.Command{
		Name:     "user",
		HelpName: "user",
		Usage:    "query the authenticated user",
		Action:   user,
	}
}
