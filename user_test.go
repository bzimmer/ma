package ma_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bzimmer/ma"
	"github.com/bzimmer/smugmug"
)

func TestUser(t *testing.T) {
	t.Parallel()
	a := assert.New(t)

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.Equal("/!authuser", r.URL.Path)
		fp, err := os.Open("testdata/user_cmac.json")
		a.NoError(err)
		defer fp.Close()
		_, err = io.Copy(w, fp)
		a.NoError(err)
	}))
	defer svr.Close()

	app := NewTestApp(t, "user", ma.CommandUser(), smugmug.WithBaseURL(svr.URL))
	a.NoError(app.RunContext(context.TODO(), []string{"ma", "user"}))
}

func TestUserIntegration(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "auth user",
			args: []string{"-j", "user"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			ma := Command(tt.args...)
			out, err := ma.Output()
			a.NoError(err)
			res := make(map[string]interface{})
			dec := json.NewDecoder(bytes.NewBuffer(out))
			a.NoError(dec.Decode(&res))
			a.NotEqual("", res["nickname"])
		})
	}
}
