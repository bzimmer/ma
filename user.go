package ma

import (
	"github.com/bzimmer/smugmug"
	"github.com/urfave/cli/v2"
)

func user(c *cli.Context) error {
	user, err := client(c).User.AuthUser(c.Context, smugmug.WithExpansions("Node"))
	if err != nil {
		return err
	}
	return encoder(c).Encode("user", map[string]interface{}{
		"nickname": user.NickName,
		"uri":      user.URI,
		"nodeID":   user.Node.NodeID,
	})
}

func CommandUser() *cli.Command {
	return &cli.Command{
		Name:   "user",
		Usage:  "query the authenticated user",
		Action: user,
	}
}
