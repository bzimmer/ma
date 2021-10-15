//go:build integration

package ma_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserIntegration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "auth user",
			args: []string{"-j", "user"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			ma, err := command(tt.args...)
			a.NoError(err)
			out, err := ma.Output()
			a.NoError(err)
			res := make(map[string]interface{})
			dec := json.NewDecoder(bytes.NewBuffer(out))
			a.NoError(dec.Decode(&res))
			a.NotEqual("", res["nickname"])
		})
	}
}
