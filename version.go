package ma

import (
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

var (
	// buildVersion of the package
	buildVersion = "development"
	// buildTime of the package
	buildTime = "now"
	// buildCommit of the package
	buildCommit = "snapshot"
	// buildBuilder of the package
	buildBuilder = "local"
)

func version(c *cli.Context) error {
	log.Info().
		Str("version", buildVersion).
		Str("datetime", buildTime).
		Str("builder", buildBuilder).
		Str("commit", buildCommit).
		Msg("version")
	return encoder(c).Encode(map[string]string{
		"version":  buildVersion,
		"datetime": buildTime,
		"commit":   buildCommit,
		"builder":  buildBuilder,
	})
}

func CommandVersion() *cli.Command {
	return &cli.Command{
		Name:   "version",
		Usage:  "version information",
		Action: version,
	}
}
