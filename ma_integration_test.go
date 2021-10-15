//go:build integration

package ma_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// root finds the root of the source tree by recursively ascending until 'go.mod' is located
func root() (string, error) {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	paths := []string{string(os.PathSeparator)}
	paths = append(paths, strings.Split(path, string(os.PathSeparator))...)
	for len(paths) > 0 {
		x := filepath.Join(paths...)
		root := filepath.Join(x, "go.mod")
		if _, err := os.Stat(root); os.IsNotExist(err) {
			paths = paths[:len(paths)-1]
		} else {
			return x, nil
		}
	}
	return "", errors.New("unable to find go.mod")
}

func command(t *testing.T, args ...string) *exec.Cmd {
	dir, err := root()
	if err != nil {
		t.Error(err)
	}
	cmd := filepath.Join(dir, "dist", "ma")
	if _, err = os.Stat(cmd); err != nil {
		t.Error(err)
	}
	if err != nil {
		t.Error(err)
	}
	return exec.Command(cmd, args...)
}

type harnessIntegration struct {
	name  string
	args  []string
	exit  int
	after func(map[string]interface{})
}

func harnessIntegrationFunc(t *testing.T, tt harnessIntegration) {
	a := assert.New(t)
	ma := command(t, tt.args...)
	out, err := ma.Output()
	a.NoError(err)
	res := make(map[string]interface{})
	dec := json.NewDecoder(bytes.NewBuffer(out))
	a.NoError(dec.Decode(&res))
	if tt.after != nil {
		tt.after(res)
	}
}
