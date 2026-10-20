package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const MaxShown = 3

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

func (c *Cache) Add(w Work) {
	for i, old := range c.Work {
		if old.IssueID == w.IssueID {
			c.Work[i] = w
			return
		}
	}
	c.Work = append(c.Work, w)
}

func (c Cache) Summary(now time.Time) string {
	if len(c.Work) == 0 {
		return "pullkin: nothing in progress"
	}
	noun := "issues"
	if len(c.Work) == 1 {
		noun = "issue"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "pullkin: %d %s in progress", len(c.Work), noun)
	for i, w := range c.Work {
		if i == MaxShown {
			fmt.Fprintf(&b, "\n  and %d more", len(c.Work)-MaxShown)
			break
		}
		fmt.Fprintf(&b, "\n  %s#%d  %s  (claimed %s)", w.Repo, w.Issue, w.Title, Ago(w.ClaimedAt, now))
	}
	return b.String()
}

func Ago(t, now time.Time) string {
	days := int(now.Sub(t).Hours() / 24)
	switch {
	case days <= 0:
		return "today"
	case days == 1:
		return "1 day ago"
	default:
		return fmt.Sprintf("%d days ago", days)
	}
}

func (c Cache) Claimed() map[string]string {
	m := make(map[string]string, len(c.Work))
	for _, w := range c.Work {
		m[w.IssueID] = w.CommentURL
	}
	return m
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
