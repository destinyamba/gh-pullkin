package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/destinyamba/gh-pullkin/internal/issues"
	"github.com/destinyamba/gh-pullkin/internal/pipeline"
	"github.com/destinyamba/gh-pullkin/internal/state"
	"github.com/destinyamba/gh-pullkin/internal/tui"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 0 && args[0] == "nudge" {
		return nudge()
	}
	if len(args) > 0 && args[0] == "list" {
		return list()
	}
	p, err := pipeline.New()
	if err != nil {
		return err
	}
	gh, err := issues.New()
	if err != nil {
		return err
	}
	load := func() (pipeline.Result, error) {
		return p.Run(context.Background(), ".", time.Now())
	}
	saved, err := state.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "pullkin: could not read saved work:", err)
	}
	claim := func(issue issues.Issue, body string) (string, error) {
		url, err := gh.Comment(context.Background(), issue.ID, body)
		if err != nil {
			return "", err
		}
		saved.Add(state.Work{
			Repo:       issue.Repo.Owner + "/" + issue.Repo.Name,
			Issue:      issue.Number,
			IssueID:    issue.ID,
			Title:      issue.Title,
			URL:        issue.URL,
			CommentURL: url,
			ClaimedAt:  time.Now(),
		})
		if err := state.Save(saved); err != nil {
			return url, fmt.Errorf("could not save it: %w", err)
		}
		return url, nil
	}
	_, err = tea.NewProgram(tui.New(load, claim, saved.Work)).Run()
	return err
}

func nudge() error {
	saved, err := state.Load()
	if err != nil {
		return err
	}
	fmt.Println(saved.Summary(time.Now()))
	return nil
}

func list() error {
	p, err := pipeline.New()
	if err != nil {
		return err
	}

	res, err := p.Run(context.Background(), ".", time.Now())
	if err != nil {
		return err
	}

	for _, err := range res.Skipped {
		fmt.Fprintln(os.Stderr, "skipped:", err)
	}

	if res.Deps == 0 {
		fmt.Println("pullkin: no go.mod or package.json in this folder. Run it where your project's manifest is.")
		return nil
	}

	if len(res.Candidates) == 0 {
		fmt.Println("pullkin: no fixable issues found in your dependencies")
		return nil
	}

	for i, c := range res.Candidates {
		kind := "indirect"
		if c.Dep.Direct {
			kind = "direct"
		}
		fmt.Printf("%2d. %s/%s#%d  %s  (%s)\n    %s\n",
			i+1, c.Issue.Repo.Owner, c.Issue.Repo.Name, c.Issue.Number, c.Issue.Title, kind, c.Issue.URL)
	}
	return nil
}
