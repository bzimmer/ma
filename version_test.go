package ma_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
)

func TestVersion(t *testing.T) {
	a := assert.New(t)

	app := NewTestApp(t, harness{name: "version"}, ma.CommandVersion(), "")
	a.NoError(app.RunContext(context.TODO(), []string{"ma", "version"}))
}
