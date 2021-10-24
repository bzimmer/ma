package ma_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
)

func decode(a *assert.Assertions, r io.Reader) map[string]interface{} {
	data := make(map[string]interface{})
	dec := json.NewDecoder(r)
	a.NoError(dec.Decode(&data))
	return data
}

func newPatchTestMux(a *assert.Assertions) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/image/GH8UQ9-0", func(w http.ResponseWriter, r *http.Request) {
		a.Fail("should not be called")
	})
	mux.HandleFunc("/image/B2fHSt7-0", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/image_B2fHSt7-0.json"))
	})
	mux.HandleFunc("/image/B2fHSt7-1", func(w http.ResponseWriter, r *http.Request) {
		data := decode(a, r.Body)
		a.Contains(data, "Latitude")
		a.NoError(copyFile(w, "testdata/image_B2fHSt7-0.json"))
	})
	mux.HandleFunc("/image/B2fHSt7-2", func(w http.ResponseWriter, r *http.Request) {
		a.Fail("should not be called")
	})
	mux.HandleFunc("/image/B2fHSt7-3", func(w http.ResponseWriter, r *http.Request) {
		data := decode(a, r.Body)
		a.Contains(data, "KeywordArray")
		a.Empty(data["KeywordArray"])
		a.NoError(copyFile(w, "testdata/image_B2fHSt7-0.json"))
	})
	mux.HandleFunc("/album/RM4BL2", func(w http.ResponseWriter, r *http.Request) {
		data := decode(a, r.Body)
		a.Contains(data, "Name")
		a.Contains(data, "UrlName")
		a.NoError(copyFile(w, "testdata/album_RM4BL2.json"))
	})
	mux.HandleFunc("/album/RM4BLQ", func(w http.ResponseWriter, r *http.Request) {
		data := decode(a, r.Body)
		a.Contains(data, "Name")
		a.Contains(data, "UrlName")
		a.Equal("Foo-Bar", data["UrlName"])
		a.NoError(copyFile(w, "testdata/album_RM4BL2.json"))
	})
	return mux
}

func TestPatch(t *testing.T) {
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
				"--name", "2021-07-04 Fourth of July", "--urlname", "2021-07-04-Fourth-of-July", "RM4BL2"},
			counters: map[string]int{"ma.patch.album": 1},
		},
		{
			name: "album with auto url naming",
			args: []string{"ma", "patch", "album", "--force",
				"--name", "foo bar", "--auto-urlname", "RM4BLQ"},
			counters: map[string]int{"ma.patch.album": 1},
		},
		{
			name: "invalid url name",
			args: []string{"ma", "patch", "album", "--force", "--urlname", "this-is-invalid", "RM4BL2"},
			err:  ma.ErrInvalidURLName.Error(),
		},
		{
			name: "more than one album key",
			args: []string{"ma", "patch", "album", "RM4BL2", "XM4BL2"},
			err:  "expected only one albumKey argument",
		},
		{
			name: "no album keys",
			args: []string{"ma", "patch", "album"},
			err:  "expected one albumKey argument",
		},
		{
			name: "both urlname and auto-urlname",
			args: []string{"ma", "patch", "album", "--urlname", "Foo-Bar", "--auto-urlname", "XM4BL2"},
			err:  "only one of `auto-urlname` or `urlname` may be specified",
		},
		{
			name: "auto-urlname without name",
			args: []string{"ma", "patch", "album", "--auto-urlname", "XM4BL2"},
			err:  "cannot specify `auto-urlname` without `name`",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			mux := newPatchTestMux(a)
			run(t, tt, mux, ma.CommandPatch)
		})
	}
}
