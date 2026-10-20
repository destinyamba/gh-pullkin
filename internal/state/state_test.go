package state

import (
	"fmt"
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

func TestAddReplacesSameIssue(t *testing.T) {
	var c Cache
	c.Add(Work{IssueID: "I_1", CommentURL: "old"})
	c.Add(Work{IssueID: "I_2", CommentURL: "b"})
	c.Add(Work{IssueID: "I_1", CommentURL: "new"})

	if len(c.Work) != 2 {
		t.Fatalf("expected 2 items, found %d", len(c.Work))
	}
	if got := c.Claimed()["I_1"]; got != "new" {
		t.Errorf("expected I_1 to be updated, found %q", got)
	}
}

func TestSummary(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)

	if got := (Cache{}).Summary(now); got != "pullkin: nothing in progress" {
		t.Errorf("empty: found %q", got)
	}

	one := Cache{Work: []Work{{Repo: "cli/cli", Issue: 9789, Title: "hangs", ClaimedAt: now.AddDate(0, 0, -3)}}}
	want := "pullkin: 1 issue in progress\n  cli/cli#9789  hangs  (claimed 3 days ago)"
	if got := one.Summary(now); got != want {
		t.Errorf("one:\nexpected %q\nfound    %q", want, got)
	}

	var many Cache
	for i := range 5 {
		many.Add(Work{IssueID: fmt.Sprint(i), Repo: "o/r", Issue: i, ClaimedAt: now})
	}
	got := many.Summary(now)
	if !strings.HasPrefix(got, "pullkin: 5 issues in progress") || !strings.HasSuffix(got, "and 2 more") {
		t.Errorf("many: found %q", got)
	}
}

func TestAgo(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		t    time.Time
		want string
	}{
		{now.Add(-2 * time.Hour), "today"},
		{now.AddDate(0, 0, -1), "1 day ago"},
		{now.AddDate(0, 0, -10), "10 days ago"},
	}
	for _, tt := range tests {
		if got := Ago(tt.t, now); got != tt.want {
			t.Errorf("expected %q, found %q", tt.want, got)
		}
	}
}
