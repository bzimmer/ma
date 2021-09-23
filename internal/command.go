package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Root finds the root of the source tree by recursively ascending until 'go.mod' is located
func Root() string {
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
			return x
		}
	}
	panic("unable to find go.mod")
}

func Command(args ...string) *exec.Cmd {
	return exec.Command(filepath.Join(Root(), "dist", "ma"), args...)
}
