package ma

import (
	"github.com/bzimmer/smugmug"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func user(c *cli.Context) error {
	user, err := runtime(c).Smugmug().User.AuthUser(c.Context, smugmug.WithExpansions("Node"))
	if err != nil {
		return err
	}
	runtime(c).Metrics.IncrCounter([]string{"user", "user"}, 1)
	log.Debug().Str("nickname", user.NickName).Str("uri", user.URI).Str("nodeID", user.Node.NodeID).Msg("user")
	return runtime(c).Encoder.Encode(user)
}

func CommandUser() *cli.Command {
	return &cli.Command{
		Name:        "user",
		HelpName:    "user",
		Usage:       "Query the authenticated user",
		Description: "Query the authenticated user",
		Action:      user,
	}
}
