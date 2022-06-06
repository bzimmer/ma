package ma_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
	"github.com/bzimmer/smugmug"
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
			name: "rm invalid image",
			args: []string{"ma", "rm", "image", "--album", "QWERTY0", "743XwH7"},
			err:  "no version specified for image key {743XwH7}",
			counters: map[string]int{
				"ma.rm.image.attempt": 1,
				"ma.rm.image.failure": 1,
			},
		},
		{
			name: "rm invalid image",
			args: []string{"ma", "rm", "image", "--album", "QWERTY0", "743XwH7-4"},
			err:  http.StatusText(http.StatusNotFound),
			counters: map[string]int{
				"ma.rm.image.attempt": 1,
				"ma.rm.image.failure": 1,
			},
		},
		{
			name: "rm invalid image",
			args: []string{"ma", "rm", "image", "-0", "--album", "QWERTY0", "743XwH7"},
			counters: map[string]int{
				"ma.rm.image.attempt": 1,
				"ma.rm.image.success": 1,
			},
		},
		{
			name: "rm image",
			args: []string{"ma", "rm", "image", "--album", "QWERTY0", "743XwH7-0"},
			counters: map[string]int{
				"ma.rm.image.attempt": 1,
				"ma.rm.image.success": 1,
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, mux, ma.CommandRemove)
		})
	}
}
