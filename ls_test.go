package ma_test

import (
	"net/http"
	"testing"

	"github.com/bzimmer/ma"
)

func TestList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/!authuser", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/user_cmac.json")
	})
	mux.HandleFunc("/node/zx4Fx", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/node_zx4Fx.json")
	})
	mux.HandleFunc("/image/B2fHSt7-0", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/image_B2fHSt7-0.json")
	})
	mux.HandleFunc("/album/RM4BL2", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_RM4BL2.json")
	})
	mux.HandleFunc("/album/qety", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/node/VsQ7zr", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/node_VsQ7zr.json")
	})
	mux.HandleFunc("/album/TDZWbg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_TDZWbg.json")
	})
	mux.HandleFunc("/album/TDZWbg!images", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/album_TDZWbg_images.json")
	})

	tests := []harness{
		{
			name:     "album",
			args:     []string{"ls", "album", "RM4BL2"},
			counters: map[string]int{"ls.album": 1},
		},
		{
			name:     "node",
			args:     []string{"ls", "node"},
			counters: map[string]int{"ls.node": 1},
		},
		{
			name:     "image with version",
			args:     []string{"ls", "image", "B2fHSt7-0"},
			counters: map[string]int{"ls.image": 1},
		},
		{
			name:     "image with autoversioning",
			args:     []string{"ls", "image", "B2fHSt7"},
			counters: map[string]int{"ls.image": 1},
		},
		{
			name: "invalid album",
			args: []string{"ls", "album", "qety"},
			err:  "Not Found",
		},
		{
			name: "node type album with no album flag",
			args: []string{"ls", "node", "VsQ7zr"},
			counters: map[string]int{
				"ls.node": 1,
			},
		},
		{
			name: "node recurse and image",
			args: []string{"ls", "node", "-R", "-i", "VsQ7zr"},
			counters: map[string]int{
				"ls.node":  1,
				"ls.image": 1,
			},
		},
		{
			name: "album recurse and image",
			args: []string{"ls", "album", "-i", "TDZWbg"},
			counters: map[string]int{
				"ls.album": 1,
				"ls.image": 1,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, &tt, mux, ma.CommandList)
		})
	}
}
