package state

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func useTempCache(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CACHE_HOME", dir)
}

func TestPathUsesTempCache(t *testing.T) {
	useTempCache(t)

	path, err := Path()
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if !strings.HasPrefix(path, os.Getenv("HOME")) {
		t.Fatalf("expected path under the temp dir, found %s", path)
	}
}

func TestLoadMissingFile(t *testing.T) {
	useTempCache(t)

	c, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(c.Work) != 0 {
		t.Errorf("expected empty cache, found %d items", len(c.Work))
	}
}

func TestSaveThenLoad(t *testing.T) {
	useTempCache(t)

	want := Cache{Work: []Work{{
		Repo:       "cli/cli",
		Issue:      9789,
		IssueID:    "I_kwDODKw3uc6a4dHo",
		Title:      "gh auth login --with-token hangs",
		URL:        "https://github.com/cli/cli/issues/9789",
		CommentURL: "https://github.com/cli/cli/issues/9789#issuecomment-1",
		ClaimedAt:  time.Date(2026, 10, 20, 9, 30, 0, 0, time.UTC),
	}}}

	if err := Save(want); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %+v, found %+v", want, got)
	}
}

func TestLoadBadFile(t *testing.T) {
	useTempCache(t)

	path, err := Path()
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if _, err := Load(); err == nil {
		t.Fatal("expected an error for a broken file, found nil")
	}
}
