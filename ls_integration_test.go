//go:build integration

package ma_test

import (
	"testing"

	"github.com/bzimmer/smugmug"
	"github.com/stretchr/testify/assert"
)

func TestListIntegration(t *testing.T) {
	a := assert.New(t)
	for _, tt := range []harnessIntegration{
		{
			name: "ls",
			args: []string{"-j", "ls", "node"},
			after: func(res map[string]interface{}) {
				a.Equal(smugmug.TypeFolder, res["Type"])
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			harnessIntegrationFunc(t, tt)
		})
	}
}
