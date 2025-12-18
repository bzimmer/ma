package ma_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

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
			args: []string{"user"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, mux, ma.CommandUser)
		})
	}
}

func TestUserError(t *testing.T) {
	a := assert.New(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/!authuser", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		enc := json.NewEncoder(w)
		a.NoError(enc.Encode(map[string]any{
			"Response": map[string]any{
				"Uri":            "/api/v2/!authuser?_pretty=true",
				"Locator":        "User",
				"LocatorType":    "Object",
				"UriDescription": "Node with the given id.",
				"EndpointType":   "Node"},
			"Code":    404,
			"Message": "Not Found",
		}))
	})

	tests := []harness{
		{
			name: "authuser",
			args: []string{"user"},
			err:  http.StatusText(http.StatusNotFound),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, mux, ma.CommandUser)
		})
	}
}
