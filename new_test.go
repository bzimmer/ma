package ma_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
	"github.com/bzimmer/smugmug"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		err      string
		counters map[string]int
	}{
		{
			name: "new with no parent",
			args: []string{"ma", "new", "album"},
			err:  "Required flag \"parent\" not set",
		},
		{
			name: "new with no arguments",
			args: []string{"ma", "new", "--parent", "QWERTY0", "album"},
			err:  "expected one or two arguments",
		},
		{
			name: "new with invalid privacy",
			args: []string{"ma", "new", "--privacy", "garbage", "--parent", "QWERTY0", "album", "0YTREWQ"},
			err:  "privacy one of",
		},
		{
			name: "new with invalid url name",
			args: []string{"ma", "new", "--parent", "QWERTY0", "album", "0YTREWQ", "lower-case"},
			err:  "node url name must start with a capital letter",
		},
		{
			name: "new album",
			args: []string{"ma", "new", "--parent", "QWERTY0", "album", "2021-03-17 A Big Day"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)

			type response struct {
				Response struct {
					Node *smugmug.Node `json:"Node"`
				} `json:"Response"`
				Expansions map[string]*json.RawMessage `json:"Expansions,omitempty"`
				Code       int                         `json:"Code"`
				Message    string                      `json:"Message"`
			}

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				enc := json.NewEncoder(w)
				switch r.Method {
				case http.MethodGet, http.MethodPost:
					res := &response{
						Code:       200,
						Message:    "OK",
						Expansions: make(map[string]*json.RawMessage),
					}
					res.Response.Node = &smugmug.Node{
						NodeID: "FGHRYD",
						URIs: smugmug.NodeURIs{
							User:           &smugmug.APIEndpoint{URI: "/user/foo/bar"},
							Album:          &smugmug.APIEndpoint{URI: "/album/foo/bar"},
							HighlightImage: &smugmug.APIEndpoint{URI: "/highlightimage/bar/foo"},
						},
						Album: &smugmug.Album{AlbumKey: "123456"},
					}
					a.NoError(enc.Encode(res))
				}
			}))
			defer svr.Close()

			app := NewTestApp(t, tt.name, ma.CommandNew(), smugmug.WithBaseURL(svr.URL))
			err := app.RunContext(context.TODO(), tt.args)
			switch tt.err == "" {
			case true:
				a.NoError(err)
			case false:
				a.Error(err)
				a.Contains(err.Error(), tt.err)
			}

			for key, value := range tt.counters {
				counter, err := findCounter(app, key)
				a.NoError(err)
				a.Equalf(value, counter.Count, key)
			}
		})
	}
}
