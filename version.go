package ma

import (
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

var (
	// buildVersion of the package
	buildVersion = "development" //nolint
	// buildTime of the package
	buildTime = "now" //nolint
	// buildCommit of the package
	buildCommit = "snapshot" //nolint
	// buildBuilder of the package
	buildBuilder = "local" //nolint
)

func version(c *cli.Context) error {
	log.Info().
		Str("version", buildVersion).
		Str("datetime", buildTime).
		Str("builder", buildBuilder).
		Str("commit", buildCommit).
		Msg("version")
	return runtime(c).Encoder.Encode(map[string]string{
		"version":  buildVersion,
		"datetime": buildTime,
		"commit":   buildCommit,
		"builder":  buildBuilder,
	})
}

func CommandVersion() *cli.Command {
	return &cli.Command{
		Name:        "version",
		HelpName:    "version",
		Usage:       "Show the version information of the binary",
		Description: "Show the version information of the binary",
		Action:      version,
	}
}
