package ma_test

import (
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
