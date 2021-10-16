package ma_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
)

func TestFind(t *testing.T) {
	a := assert.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/!authuser", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/user_cmac.json"))
	})
	mux.HandleFunc("/album!search", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/album_search_marmot.json"))
	})
	mux.HandleFunc("/node!search", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/album_search_marmot.json"))
	})

	for _, tt := range []harness{
		{
			name: "find",
			args: []string{"ma", "find", "Marmot"},
			counters: map[string]int{
				"ma.find.album": 10,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			harnessFunc(t, tt, mux, ma.CommandFind)
		})
	}
}
