package pipeline

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/destinyamba/gh-pullkin/internal/depsdev"
	"github.com/destinyamba/gh-pullkin/internal/filter"
	"github.com/destinyamba/gh-pullkin/internal/issues"
	"github.com/destinyamba/gh-pullkin/internal/manifest"
	"github.com/destinyamba/gh-pullkin/internal/rank"
	"golang.org/x/sync/errgroup"
)

type repoFinder interface {
	SourceRepo(ctx context.Context, dep manifest.Dep) (depsdev.Repo, error)
}

type issueSource interface {
	Info(ctx context.Context, repo depsdev.Repo) (issues.RepoInfo, error)
	Open(ctx context.Context, repo depsdev.Repo) ([]issues.Issue, error)
}

type Pipeline struct {
	repos  repoFinder
	issues issueSource
}

type Result struct {
	Deps       int
	Candidates []rank.Candidate
	Skipped    []error
}

func New() (*Pipeline, error) {
	gh, err := issues.New()
	if err != nil {
		return nil, fmt.Errorf("github client: %w", err)
	}
	return &Pipeline{repos: depsdev.New(), issues: gh}, nil
}

func (p *Pipeline) Run(ctx context.Context, dir string, now time.Time) (Result, error) {
	var res Result

	deps, err := manifest.Scan(dir)
	if err != nil {
		return res, err
	}

	res.Deps = len(deps)

	found := make([]depsdev.Repo, len(deps))
	errs := make([]error, len(deps))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(8)
	for i, dep := range deps {
		g.Go(func() error {
			found[i], errs[i] = p.repos.SourceRepo(gctx, dep)
			return nil
		})
	}
	g.Wait()

	byRepo := map[depsdev.Repo]manifest.Dep{}
	for i, dep := range deps {
		if errors.Is(errs[i], depsdev.ErrNoRepo) || errors.Is(errs[i], depsdev.ErrNotFound) {
			continue
		}
		if errs[i] != nil {
			return res, errs[i]
		}
		repo := found[i]
		if old, ok := byRepo[repo]; ok && (old.Direct || !dep.Direct) {
			continue
		}
		byRepo[repo] = dep
	}

	repos := make([]depsdev.Repo, 0, len(byRepo))
	for repo := range byRepo {
		repos = append(repos, repo)
	}
	slices.SortFunc(repos, func(a, b depsdev.Repo) int {
		return cmp.Or(cmp.Compare(a.Owner, b.Owner), cmp.Compare(a.Name, b.Name))
	})

	for _, repo := range repos {
		info, err := p.issues.Info(ctx, repo)
		if err != nil {
			res.Skipped = append(res.Skipped, err)
			continue
		}
		if filter.Dead(info, now) {
			continue
		}

		open, err := p.issues.Open(ctx, repo)
		if err != nil {
			res.Skipped = append(res.Skipped, err)
			continue
		}
		for _, issue := range open {
			if filter.Taken(issue, now) || filter.Stale(issue, now) {
				continue
			}
			res.Candidates = append(res.Candidates, rank.Candidate{Issue: issue, Dep: byRepo[repo]})
		}
	}

	res.Candidates = rank.Rank(res.Candidates)
	return res, nil
}
