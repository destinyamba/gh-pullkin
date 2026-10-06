package manifest

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type Ecosystem string

const (
	Go  Ecosystem = "go"
	NPM Ecosystem = "npm"
)

type Dep struct {
	Ecosystem Ecosystem
	Name      string
	Version   string
	Direct    bool
}

func Scan(dir string) ([]Dep, error) {
	path := filepath.Join(dir, "go.mod")

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read go.mod: %w", err)
	}

	deps, err := ParseGoMod(data)
	if err != nil {
		return nil, fmt.Errorf("parse go.mod: %w", err)
	}

	return deps, nil
}
