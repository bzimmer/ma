package ma_test

import (
	"net/http"
	"testing"

	"github.com/bzimmer/ma"
)

func TestFind(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/!authuser", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/user_cmac.json")
	})
	mux.HandleFunc("/album!search", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_search_marmot.json")
	})
	mux.HandleFunc("/node!search", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/node_search_marmot.json")
	})

	for _, tt := range []harness{
		{
			name: "find",
			args: []string{"find", "Marmot"},
			counters: map[string]int{
				"find.album": 10,
				"find.node":  250,
			},
		},
		{
			name: "find nodes",
			args: []string{"find", "-n", "Marmot"},
			counters: map[string]int{
				"find.node": 250,
			},
		},
		{
			name: "find albums",
			args: []string{"find", "-a", "Marmot"},
			counters: map[string]int{
				"find.album": 10,
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, mux, ma.CommandFind)
		})
	}
}
