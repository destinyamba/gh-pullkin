package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/destinyamba/gh-pullkin/internal/pipeline"
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
	_, err := tea.NewProgram(tui.New()).Run()
	return err
}

func nudge() error {
	fmt.Println("pullkin: nothing in progress")
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
