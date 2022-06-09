package ma_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
)

func decode(a *assert.Assertions, r io.Reader) map[string]any {
	data := make(map[string]any)
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
		http.ServeFile(w, r, "testdata/image_B2fHSt7-0.json")
	})
	mux.HandleFunc("/image/B2fHSt7-1", func(w http.ResponseWriter, r *http.Request) {
		data := decode(a, r.Body)
		a.Contains(data, "Latitude")
		http.ServeFile(w, r, "testdata/image_B2fHSt7-0.json")
	})
	mux.HandleFunc("/image/B2fHSt7-2", func(w http.ResponseWriter, r *http.Request) {
		a.Fail("should not be called")
	})
	mux.HandleFunc("/image/B2fHSt7-3", func(w http.ResponseWriter, r *http.Request) {
		data := decode(a, r.Body)
		a.Contains(data, "KeywordArray")
		a.Empty(data["KeywordArray"])
		http.ServeFile(w, r, "testdata/image_B2fHSt7-0.json")
	})
	mux.HandleFunc("/image/B2fHSt7-4", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		enc := json.NewEncoder(w)
		a.NoError(enc.Encode(map[string]any{
			"Response": map[string]any{
				"Uri":            "/api/v2/image/B2fHSt7-4?_pretty=true",
				"Locator":        "Image",
				"LocatorType":    "Object",
				"UriDescription": "Image by key",
				"EndpointType":   "Image",
			},
			"Code":    404,
			"Message": "Not Found",
		}))
	})
	mux.HandleFunc("/album/RM4BL3", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		enc := json.NewEncoder(w)
		a.NoError(enc.Encode(map[string]any{
			"Response": map[string]any{
				"Uri":            "/api/v2/album/RM4BL3?_pretty=true",
				"Locator":        "Album",
				"LocatorType":    "Object",
				"UriDescription": "Album by key",
				"EndpointType":   "Album",
			},
			"Code":    404,
			"Message": "Not Found",
		}))
	})
	mux.HandleFunc("/album/RM4BL2", func(w http.ResponseWriter, r *http.Request) {
		data := decode(a, r.Body)
		a.Contains(data, "Name")
		a.Contains(data, "UrlName")
		http.ServeFile(w, r, "testdata/album_RM4BL2.json")
	})
	mux.HandleFunc("/album/RM4BLQ", func(w http.ResponseWriter, r *http.Request) {
		data := decode(a, r.Body)
		a.Contains(data, "Name")
		a.Contains(data, "UrlName")
		a.Equal("Foo-Bar", data["UrlName"])
		http.ServeFile(w, r, "testdata/album_RM4BL2.json")
	})
	return mux
}

func TestPatch(t *testing.T) {
	for _, tt := range []harness{
		{
			name:     "dry run",
			args:     []string{"ma", "patch", "image", "--dryrun", "--keyword", "foo", "GH8UQ9-0"},
			counters: map[string]int{"ma.patch.image.dryrun": 1},
		},
		{
			name:     "keyword",
			args:     []string{"ma", "patch", "image", "--keyword", "foo", "B2fHSt7-0"},
			counters: map[string]int{"ma.patch.image": 1},
		},
		{
			name:     "latitude",
			args:     []string{"ma", "patch", "image", "--latitude", "48.4321", "B2fHSt7-1"},
			counters: map[string]int{"ma.patch.image": 1},
		},
		{
			name: "no patches image",
			args: []string{"ma", "patch", "image", "B2fHSt7-2"},
		},
		{
			name: "no patches album",
			args: []string{"ma", "patch", "album", "RM4BLQ"},
		},
		{
			name:     "empty keywords",
			args:     []string{"ma", "patch", "image", "--dryrun", "--keyword", "", "B2fHSt7-3"},
			counters: map[string]int{"ma.patch.image.dryrun": 1},
		},
		{
			name: "album",
			args: []string{"ma", "patch", "album",
				"--name", "2021-07-04 Fourth of July", "--urlname", "2021-07-04-Fourth-of-July", "RM4BL2"},
			counters: map[string]int{"ma.patch.album": 1},
		},
		{
			name:     "album with auto url naming",
			args:     []string{"ma", "patch", "album", "--name", "foo bar", "--auto", "RM4BLQ"},
			counters: map[string]int{"ma.patch.album": 1},
		},
		{
			name: "invalid url name",
			args: []string{"ma", "patch", "album", "--urlname", "this-is-invalid", "RM4BL2"},
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
			name: "both urlname and auto urlname",
			args: []string{"ma", "patch", "album", "--urlname", "Foo-Bar", "--auto", "XM4BL2"},
			err:  "only one of `auto` or `urlname` may be specified",
		},
		{
			name: "invalid urlname",
			args: []string{"ma", "patch", "album", "--name", "bar baz", "--urlname", "foo bar", "XM4BL2"},
			err:  "node url name must start with a number or capital letter",
		},
		{
			name: "auto urlname without name",
			args: []string{"ma", "patch", "album", "--auto", "XM4BL2"},
			err:  "cannot specify `auto` without `name`",
		},
		{
			name: "patch album 404",
			args: []string{"ma", "patch", "album", "--name", "bar foo", "RM4BL3"},
			err:  http.StatusText(http.StatusNotFound),
		},
		{
			name: "patch image 404",
			args: []string{"ma", "patch", "image", "--title", "something bar foo", "B2fHSt7-4"},
			err:  http.StatusText(http.StatusNotFound),
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			mux := newPatchTestMux(a)
			run(t, &tt, mux, ma.CommandPatch)
		})
	}
}
