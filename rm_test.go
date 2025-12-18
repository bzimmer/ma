package ma_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/bzimmer/smugmug"
	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
)

func TestRemove(t *testing.T) {
	a := assert.New(t)

	type response struct {
		Response struct {
			Image *smugmug.Image `json:"Image"`
		} `json:"Response"`
		Code    int    `json:"Code"`
		Message string `json:"Message"`
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/album/TDZWbg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_TDZWbg.json")
	})
	mux.HandleFunc("/album/TDZWbg!images", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_TDZWbg_images.json")
	})
	mux.HandleFunc("/album/TDZWbg/image/TL4PJfh-0", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			enc := json.NewEncoder(w)
			a.NoError(enc.Encode(&response{
				Code:    200,
				Message: "OK",
			}))
		}
	})
	mux.HandleFunc("/album/QWERTY0/image/743XwH7-0", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			enc := json.NewEncoder(w)
			a.NoError(enc.Encode(&response{
				Code:    200,
				Message: "OK",
			}))
		}
	})

	for _, tt := range []harness{
		{
			name: "rm image id with no serial number",
			args: []string{"rm", "image", "--album", "TDZWbg", "TL4PJfh"},
			counters: map[string]int{
				"rm.image.attempt": 1,
				"rm.image.success": 1,
			},
		},
		{
			name: "rm image id with no serial number dryrun",
			args: []string{"rm", "image", "--album", "TDZWbg", "--dryrun", "TL4PJfh"},
			counters: map[string]int{
				"rm.image.dryrun": 1,
			},
		},
		{
			name: "rm an image with a non-existent id",
			args: []string{"rm", "image", "--album", "QWERTY0", "743XwH7-4"},
			err:  http.StatusText(http.StatusNotFound),
			counters: map[string]int{
				"rm.image.attempt": 1,
				"rm.image.failure": 1,
			},
		},
		{
			name: "rm image with an explicit serial number",
			args: []string{"rm", "image", "--album", "QWERTY0", "743XwH7-0"},
			counters: map[string]int{
				"rm.image.attempt": 1,
				"rm.image.success": 1,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, mux, ma.CommandRemove)
		})
	}
}
