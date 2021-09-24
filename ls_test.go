package ma_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma/internal"
)

func TestListIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()
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
			t.Parallel()
			a := assert.New(t)
			ma := internal.Command(tt.args...)
			out, err := ma.Output()
			a.NoError(err)
			res := make(map[string]interface{})
			dec := json.NewDecoder(bytes.NewBuffer(out))
			a.NoError(dec.Decode(&res))
			a.Equal(3, len(res))
			a.Equal("Folder", res["type"])
		})
	}
}
