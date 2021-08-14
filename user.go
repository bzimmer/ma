package ma

import (
	"github.com/bzimmer/smugmug"
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
	enc, err := encoder(c)
	if err != nil {
		return err
	}
	return enc.Encode("user", map[string]interface{}{
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
