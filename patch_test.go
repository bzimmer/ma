package ma_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
)

func TestPatch(t *testing.T) {
	a := assert.New(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/image/B2fHSt7-0", func(w http.ResponseWriter, r *http.Request) {
		a.NoError(copyFile(w, "testdata/image_B2fHSt7-0.json"))
	})

	for _, tt := range []harness{
		{
			name:     "no force",
			args:     []string{"ma", "patch", "--keyword", "foo", "GH8UQ9-0"},
			counters: map[string]int{"ma.patch.patched.dryrun": 1},
		},
		{
			name:     "force",
			args:     []string{"ma", "patch", "--force", "--keyword", "foo", "B2fHSt7-0"},
			counters: map[string]int{"ma.patch.patched": 1},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			harnessFunc(t, tt, mux, ma.CommandPatch)
		})
	}
}
