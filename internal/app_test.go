package internal_test

import (
	"testing"

	"github.com/bzimmer/ma/internal"
	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	app := internal.App()
	a := assert.New(t)
	err := app.Run(append([]string{app.Name}, "version"))
	a.NoError(err)
}
