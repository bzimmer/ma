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

	enc := encoder(c, "user")
	msg := map[string]interface{}{
		"nickname": user.NickName,
		"uri":      user.URI,
	}
	return enc.Encode(msg)
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
