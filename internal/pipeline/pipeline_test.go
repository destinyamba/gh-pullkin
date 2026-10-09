package pipeline

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/destinyamba/gh-pullkin/internal/depsdev"
	"github.com/destinyamba/gh-pullkin/internal/issues"
	"github.com/destinyamba/gh-pullkin/internal/manifest"
)

var now = time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)

type fakeRepos map[string]depsdev.Repo

func (f fakeRepos) SourceRepo(_ context.Context, dep manifest.Dep) (depsdev.Repo, error) {
	repo, ok := f[dep.Name]
	if !ok {
		return depsdev.Repo{}, depsdev.ErrNoRepo
	}
	return repo, nil
}

type fakeIssues struct {
	info   map[depsdev.Repo]issues.RepoInfo
	open   map[depsdev.Repo][]issues.Issue
	broken map[depsdev.Repo]bool
	calls  []depsdev.Repo
}

func (f *fakeIssues) Info(_ context.Context, repo depsdev.Repo) (issues.RepoInfo, error) {
	if f.broken[repo] {
		return issues.RepoInfo{}, errors.New("boom")
	}
	return f.info[repo], nil
}

func (f *fakeIssues) Open(_ context.Context, repo depsdev.Repo) ([]issues.Issue, error) {
	f.calls = append(f.calls, repo)
	return f.open[repo], nil
}

func TestRun(t *testing.T) {
	dir := t.TempDir()
	goMod := []byte(`module example.com/test

go 1.26.0

require (
	example.com/alive/sub v1.0.0 // indirect
	example.com/alive v1.0.0
	example.com/dead v1.0.0
	example.com/broken v1.0.0
	example.com/indirect v1.0.0 // indirect
	example.com/nowhere v1.0.0
)
`)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), goMod, 0o644); err != nil {
		t.Fatalf("file: %v", err)
	}

	alive := depsdev.Repo{Owner: "o", Name: "alive"}
	dead := depsdev.Repo{Owner: "o", Name: "dead"}
	broken := depsdev.Repo{Owner: "o", Name: "broken"}
	indirect := depsdev.Repo{Owner: "o", Name: "indirect"}

	repos := fakeRepos{
		"example.com/alive":     alive,
		"example.com/alive/sub": alive,
		"example.com/dead":      dead,
		"example.com/broken":    broken,
		"example.com/indirect":  indirect,
	}

	recent := now.AddDate(0, 0, -1)
	gh := &fakeIssues{
		info: map[depsdev.Repo]issues.RepoInfo{
			alive:    {PushedAt: recent},
			dead:     {Archived: true, PushedAt: recent},
			indirect: {PushedAt: recent},
		},
		open: map[depsdev.Repo][]issues.Issue{
			alive: {
				{Repo: alive, Number: 1, Labels: []string{"help wanted"}, UpdatedAt: recent},
				{Repo: alive, Number: 2, HasOpenPR: true, UpdatedAt: recent},
				{Repo: alive, Number: 3, UpdatedAt: now.AddDate(-2, 0, 0)},
			},
			dead: {
				{Repo: dead, Number: 4, UpdatedAt: recent},
			},
			indirect: {
				{Repo: indirect, Number: 5, Labels: []string{"good first issue"}, UpdatedAt: recent},
			},
		},
		broken: map[depsdev.Repo]bool{broken: true},
	}

	p := &Pipeline{repos: repos, issues: gh}
	res, err := p.Run(context.Background(), dir, now)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	var got []int
	for _, c := range res.Candidates {
		got = append(got, c.Issue.Number)
	}
	if want := []int{1, 5}; !slices.Equal(got, want) {
		t.Errorf("expected issues %v, found %v", want, got)
	}

	if !res.Candidates[0].Dep.Direct {
		t.Errorf("expected shared repo to keep the direct dep")
	}

	if res.Deps != 6 {
		t.Errorf("expected 6 dependencies, found %d", res.Deps)
	}

	if len(res.Skipped) != 1 {
		t.Errorf("expected 1 skipped repo, found %d", len(res.Skipped))
	}

	if slices.Contains(gh.calls, dead) {
		t.Errorf("expected no issue fetch for a dead repo")
	}
	if n := len(gh.calls); n != 2 {
		t.Errorf("expected 2 issue fetches, found %d: %v", n, gh.calls)
	}
}

func TestRunStopsOnRepoLookupError(t *testing.T) {
	dir := t.TempDir()
	goMod := []byte("module example.com/test\n\ngo 1.26.0\n\nrequire example.com/a v1.0.0\n")
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), goMod, 0o644); err != nil {
		t.Fatalf("file: %v", err)
	}

	p := &Pipeline{repos: failingRepos{}, issues: &fakeIssues{}}
	if _, err := p.Run(context.Background(), dir, now); err == nil {
		t.Fatal("expected an error, found nil")
	}
}

type failingRepos struct{}

func (failingRepos) SourceRepo(context.Context, manifest.Dep) (depsdev.Repo, error) {
	return depsdev.Repo{}, errors.New("network down")
}
