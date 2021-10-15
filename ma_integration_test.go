//go:build integration

package ma_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func binary() (string, error) {
	dir, err := root()
	if err != nil {
		return "", err
	}
	cmd := filepath.Join(dir, "dist", "ma")
	if _, err = os.Stat(cmd); err != nil {
		return "", err
	}
	return cmd, nil
}

func command(args ...string) (*exec.Cmd, error) {
	cmd, err := binary()
	if err != nil {
		return nil, err
	}
	return exec.Command(cmd, args...), nil
}
