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
	var all []Dep

	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err == nil {
		deps, err := ParseGoMod(data)
		if err != nil {
			return nil, fmt.Errorf("parse go.mod: %w", err)
		}
		all = append(all, deps...)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("read go.mod: %w", err)
	}

	data, err = os.ReadFile(filepath.Join(dir, "package.json"))
	if err == nil {
		deps, err := ParsePackageJSON(data)
		if err != nil {
			return nil, fmt.Errorf("parse package.json: %w", err)
		}
		all = append(all, deps...)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("read package.json: %w", err)
	}

	return all, nil
}
