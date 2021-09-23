package ma_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/rs/zerolog/log"

	"github.com/bzimmer/ma/internal"
)

func TestMain(m *testing.M) {
	root := internal.Root()
	if err := os.Chdir(root); err != nil {
		log.Error().Err(err).Msg("failed to change to root directory")
		os.Exit(1)
	}
	task := exec.Command("task", "build")
	if err := task.Run(); err != nil {
		log.Error().Err(err).Msg("failed to build `ma` binary")
		os.Exit(1)
	}
	os.Exit(m.Run())
}
