package state

import (
	"errors"
	"os"
	"path/filepath"
)

type Work struct {
	Repo   string
	Issue  int
	Branch string
}

type Cache struct {
	Work []Work
}

func Path() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "pullkin", "state.json"), nil
}

func Load() (Cache, error) {
	return Cache{}, errors.New("not implemented")
}

func Save(c Cache) error {
	return errors.New("not implemented")
}
