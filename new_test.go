package ma_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
	"github.com/bzimmer/smugmug"
)

func TestNew(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	type response struct {
		Response struct {
			Node *smugmug.Node `json:"Node"`
		} `json:"Response"`
		Expansions map[string]*json.RawMessage `json:"Expansions,omitempty"`
		Code       int                         `json:"Code"`
		Message    string                      `json:"Message"`
	}

	var newResponse = func() *response {
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
		return res
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/node/FGHRYD", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			enc := json.NewEncoder(w)
			a.NoError(enc.Encode(newResponse()))
		}
	})
	mux.HandleFunc("/node/QWERTY0!children", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			enc := json.NewEncoder(w)
			a.NoError(enc.Encode(newResponse()))
		}
	})

	for _, tt := range []harness{
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
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			harnessFunc(t, tt, mux, ma.CommandNew)
		})
	}
}
