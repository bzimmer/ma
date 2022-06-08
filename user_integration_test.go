//go:build integration

package ma_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserIntegration(t *testing.T) {
	a := assert.New(t)
	tests := []harnessIntegration{
		{
			name: "auth user",
			args: []string{"-j", "user"},
			after: func(res map[string]any) {
				a.NotEqual("", res["nickname"])
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
