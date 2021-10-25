package ma_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
)

func TestList(t *testing.T) {
	a := assert.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/!authuser", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/user_cmac.json"))
	})
	mux.HandleFunc("/node/zx4Fx", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/node_zx4Fx.json"))
	})
	mux.HandleFunc("/image/B2fHSt7-0", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/image_B2fHSt7-0.json"))
	})
	mux.HandleFunc("/album/RM4BL2", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/album_RM4BL2.json"))
	})
	mux.HandleFunc("/album/qety", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		a.NoError(copyFile(w, "testdata/album_qety_404.json"))
	})
	mux.HandleFunc("/node/VsQ7zr", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/node_VsQ7zr.json"))
	})
	mux.HandleFunc("/album/TDZWbg", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/album_TDZWbg.json"))
	})
	mux.HandleFunc("/album/TDZWbg!images", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/album_TDZWbg_images.json"))
	})

	tests := []harness{
		{
			name:     "album",
			args:     []string{"ma", "ls", "album", "RM4BL2"},
			counters: map[string]int{"ma.ls.album": 1},
		},
		{
			name:     "node",
			args:     []string{"ma", "ls", "node"},
			counters: map[string]int{"ma.ls.node": 1},
		},
		{
			name:     "image with version",
			args:     []string{"ma", "ls", "image", "B2fHSt7-0"},
			counters: map[string]int{"ma.ls.image": 1},
		},
		{
			name:     "image with auto-versioning",
			args:     []string{"ma", "ls", "image", "--zero-version", "B2fHSt7"},
			counters: map[string]int{"ma.ls.image": 1},
		},
		{
			name: "image with no version and no auto-versioning",
			args: []string{"ma", "ls", "image", "B2fHSt7"},
			err:  "no version specified",
		},
		{
			name: "invalid album",
			args: []string{"ma", "ls", "album", "qety"},
			err:  "Not Found",
		},
		{
			name: "node type album with no album flag",
			args: []string{"ma", "ls", "node", "VsQ7zr"},
			counters: map[string]int{
				"ma.ls.node": 1,
			},
		},
		{
			name: "node recurse and image",
			args: []string{"ma", "ls", "node", "-R", "-i", "VsQ7zr"},
			counters: map[string]int{
				"ma.ls.node":  1,
				"ma.ls.image": 1,
			},
		},
		{
			name: "album recurse and image",
			args: []string{"ma", "ls", "album", "-i", "TDZWbg"},
			counters: map[string]int{
				"ma.ls.album": 1,
				"ma.ls.image": 1,
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
