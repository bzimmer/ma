package ma_test

import (
	"net/http"
	"testing"

	"github.com/bzimmer/ma"
)

func TestUser(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/!authuser", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/user_cmac.json")
	})

	tests := []harness{
		{
			name: "authuser",
			args: []string{"ma", "user"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, mux, ma.CommandUser)
		})
	}
}
