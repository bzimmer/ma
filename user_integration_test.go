//go:build integration

package ma_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserIntegration(t *testing.T) {
	a := assert.New(t)
	for _, tt := range []harnessIntegration{
		{
			name: "auth user",
			args: []string{"-j", "user"},
			after: func(res map[string]interface{}) {
				a.NotEqual("", res["nickname"])
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			harnessIntegrationFunc(t, tt)
		})
	}
}
