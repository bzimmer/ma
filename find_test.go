package ma_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
	"github.com/bzimmer/smugmug"
)

func TestFind(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var filename string
		switch r.URL.Path {
		case "/!authuser":
			filename = "testdata/user_cmac.json"
		default:
			filename = "testdata/album_search_marmot.json"
		}
		a.NotEmpty(filename)
		fp, err := os.Open(filename)
		a.NoError(err)
		defer fp.Close()
		_, err = io.Copy(w, fp)
		a.NoError(err)
	}))
	defer svr.Close()

	app := NewTestApp(t, ma.CommandFind(), smugmug.WithBaseURL(svr.URL))
	a.NoError(app.RunContext(context.TODO(), []string{"ma", "find", "Marmot"}))

	counter, err := findCounter(app, "ma.find.album")
	a.NoError(err)
	a.Equal(10, counter.Count)
	counter, err = findCounter(app, "ma.find.node")
	a.Error(err)
	a.Empty(counter)
}
