package ma_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
)

func TestVersion(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	app := NewTestApp(t, ma.CommandVersion())
	a.NoError(app.RunContext(context.TODO(), []string{"ma", "version"}))
}
