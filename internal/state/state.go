package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type Work struct {
	Repo       string
	Issue      int
	IssueID    string
	Title      string
	URL        string
	CommentURL string
	ClaimedAt  time.Time
	Branch     string
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
	path, err := Path()
	if err != nil {
		return Cache{}, err
	}
	data, err := os.ReadFile(path)
	if err == nil {
		var c Cache
		err = json.Unmarshal(data, &c)
		if err != nil {
			return Cache{}, fmt.Errorf("unmarshal cache: %w", err)
		}
		return c, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return Cache{}, fmt.Errorf("read cache: %w", err)
	}
	return Cache{}, nil
}

func Save(c Cache) error {
	path, err := Path()
	if err != nil {
		return err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}
	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}
	err = os.WriteFile(path, data, 0o644)
	if err != nil {
		return fmt.Errorf("write cache: %w", err)
	}
	return nil
}
