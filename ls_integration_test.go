//go:build integration

package ma_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/bzimmer/smugmug"
	"github.com/stretchr/testify/assert"
)

func TestListIntegration(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "ls",
			args: []string{"-j", "ls", "node"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			ma, err := command(tt.args...)
			a.NoError(err)
			out, err := ma.Output()
			a.NoError(err)
			res := make(map[string]interface{})
			dec := json.NewDecoder(bytes.NewBuffer(out))
			a.NoError(dec.Decode(&res))
			a.Equal(smugmug.TypeFolder, res["Type"])
		})
	}
}
