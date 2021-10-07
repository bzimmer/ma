package ma_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
	"github.com/bzimmer/smugmug"
)

func TestPatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		counter string
		count   int
	}{
		{
			name:    "no force",
			args:    []string{"ma", "patch", "--keyword", "foo", "GH8UQ9-0"},
			counter: "ma.patch.patched.dryrun",
			count:   1,
		},
		{
			name:    "force",
			args:    []string{"ma", "patch", "--force", "--keyword", "foo", "B2fHSt7-0"},
			counter: "ma.patch.patched",
			count:   1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasPrefix(r.URL.Path, "/image/B2fHSt7-0"):
					fp, err := os.Open("testdata/image_B2fHSt7-0.json")
					a.NoError(err)
					defer fp.Close()
					_, err = io.Copy(w, fp)
					a.NoError(err)
				default:
					a.Failf("expected call for image", r.URL.Path)
				}
			}))
			defer svr.Close()

			app := NewTestApp(t, tt.name, ma.CommandPatch(), smugmug.WithBaseURL(svr.URL))

			_, err := findCounter(app, tt.counter)
			a.Error(err)

			a.NoError(app.RunContext(context.TODO(), tt.args))

			counter, err := findCounter(app, tt.counter)
			a.NoError(err)
			a.Equal(tt.count, counter.Count)
		})
	}
}
