//go:build integration

package ma_test

import (
	"testing"

	"github.com/bzimmer/smugmug"
	"github.com/stretchr/testify/assert"
)

func TestListIntegration(t *testing.T) {
	a := assert.New(t)
	tests := []harnessIntegration{
		{
			name: "ls",
			args: []string{"-j", "ls", "node"},
			after: func(res map[string]any) {
				a.Equal(smugmug.TypeFolder, res["Type"])
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			runIntegration(t, tt)
		})
	}
}
