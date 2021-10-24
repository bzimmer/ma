package ma_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
)

func TestPatch(t *testing.T) {
	a := assert.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/image/GH8UQ9-0", func(w http.ResponseWriter, r *http.Request) {
		a.Fail("should not be called")
	})
	mux.HandleFunc("/image/B2fHSt7-0", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/image_B2fHSt7-0.json"))
	})
	mux.HandleFunc("/image/B2fHSt7-1", func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]interface{})
		dec := json.NewDecoder(r.Body)
		a.NoError(dec.Decode(&data))
		a.Contains(data, "Latitude")
		a.NoError(copyFile(w, "testdata/image_B2fHSt7-0.json"))
	})
	mux.HandleFunc("/image/B2fHSt7-2", func(w http.ResponseWriter, r *http.Request) {
		a.Fail("should not be called")
	})
	mux.HandleFunc("/image/B2fHSt7-3", func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]interface{})
		dec := json.NewDecoder(r.Body)
		a.NoError(dec.Decode(&data))
		a.Contains(data, "KeywordArray")
		a.Empty(data["KeywordArray"])
		a.NoError(copyFile(w, "testdata/image_B2fHSt7-0.json"))
	})
	mux.HandleFunc("/album/RM4bL2", func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]interface{})
		dec := json.NewDecoder(r.Body)
		a.NoError(dec.Decode(&data))
		a.Contains(data, "Name")
		a.Contains(data, "UrlName")
		a.NoError(copyFile(w, "testdata/album_RM4bL2.json"))
	})

	for _, tt := range []harness{
		{
			name:     "no force",
			args:     []string{"ma", "patch", "image", "--keyword", "foo", "GH8UQ9-0"},
			counters: map[string]int{"ma.patch.image.dryrun": 1},
		},
		{
			name:     "force",
			args:     []string{"ma", "patch", "image", "--force", "--keyword", "foo", "B2fHSt7-0"},
			counters: map[string]int{"ma.patch.image": 1},
		},
		{
			name:     "force",
			args:     []string{"ma", "patch", "image", "--force", "--latitude", "48.4321", "B2fHSt7-1"},
			counters: map[string]int{"ma.patch.image": 1},
		},
		{
			name: "no patches",
			args: []string{"ma", "patch", "image", "B2fHSt7-2"},
		},
		{
			name:     "empty keywords",
			args:     []string{"ma", "patch", "image", "--keyword", "", "B2fHSt7-3"},
			counters: map[string]int{"ma.patch.image.dryrun": 1},
		},
		{
			name: "album",
			args: []string{"ma", "patch", "album", "--force",
				"--name", "2021-07-04 Fourth of July", "--urlname", "2021-07-04-Fourth-of-July", "RM4bL2"},
			counters: map[string]int{"ma.patch.album": 1},
		},
		{
			name: "invalid url name",
			args: []string{"ma", "patch", "album", "--force", "--urlname", "this-is-invalid", "RM4bL2"},
			err:  ma.ErrInvalidURLName.Error(),
		},
		{
			name: "more than one album key",
			args: []string{"ma", "patch", "album", "RM4bL2", "XM4bL2"},
			err:  "expected only one albumKey argument",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			run(t, tt, mux, ma.CommandPatch)
		})
	}
}
