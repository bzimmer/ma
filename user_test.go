package ma_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
)

func TestUser(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/!authuser", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/user_cmac.json"))
	})

	for _, tt := range []harness{
		{
			name: "authuser",
			args: []string{"ma", "user"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			harnessFunc(t, tt, mux, ma.CommandUser)
		})
	}
}

func TestUserIntegration(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
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
			ma := command(tt.args...)
			out, err := ma.Output()
			a.NoError(err)
			res := make(map[string]interface{})
			dec := json.NewDecoder(bytes.NewBuffer(out))
			a.NoError(dec.Decode(&res))
			a.NotEqual("", res["nickname"])
		})
	}
}
