package tui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/destinyamba/gh-pullkin/internal/depsdev"
	"github.com/destinyamba/gh-pullkin/internal/issues"
	"github.com/destinyamba/gh-pullkin/internal/manifest"
	"github.com/destinyamba/gh-pullkin/internal/pipeline"
	"github.com/destinyamba/gh-pullkin/internal/rank"
)

func cand(n int, title string) rank.Candidate {
	return rank.Candidate{
		Issue: issues.Issue{Repo: depsdev.Repo{Owner: "o", Name: "r"}, Number: n, Title: title, URL: "https://x/" + title},
		Dep:   manifest.Dep{Direct: true},
	}
}

func loaded(m Model, res pipeline.Result, err error) Model {
	next, _ := m.Update(loadedMsg{res: res, err: err})
	return next.(Model)
}

func press(m Model, key string) Model {
	var msg tea.KeyPressMsg
	switch key {
	case "down":
		msg = tea.KeyPressMsg{Code: tea.KeyDown}
	case "up":
		msg = tea.KeyPressMsg{Code: tea.KeyUp}
	}
	next, _ := m.Update(msg)
	return next.(Model)
}

func TestLoadingThenList(t *testing.T) {
	m := New(nil)
	if !strings.Contains(m.render(), "Finding issues") {
		t.Errorf("expected loading text")
	}

	m = loaded(m, pipeline.Result{Deps: 2, Candidates: []rank.Candidate{cand(1, "first"), cand(2, "second")}}, nil)
	out := m.render()
	if !strings.Contains(out, "2 fixable issues") || !strings.Contains(out, "r#1") || !strings.Contains(out, "second") {
		t.Errorf("expected list, found:\n%s", out)
	}
}

func TestCursorStaysInBounds(t *testing.T) {
	m := loaded(New(nil), pipeline.Result{Deps: 1, Candidates: []rank.Candidate{cand(1, "a"), cand(2, "b")}}, nil)

	m = press(m, "up")
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after up at top, found %d", m.cursor)
	}
	m = press(press(press(m, "down"), "down"), "down")
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 at bottom, found %d", m.cursor)
	}
}

func TestEmptyStates(t *testing.T) {
	noManifest := loaded(New(nil), pipeline.Result{}, nil).render()
	if !strings.Contains(noManifest, "No go.mod or package.json") {
		t.Errorf("expected no-manifest message, found:\n%s", noManifest)
	}

	noIssues := loaded(New(nil), pipeline.Result{Deps: 3}, nil).render()
	if !strings.Contains(noIssues, "No fixable issues") {
		t.Errorf("expected no-issues message, found:\n%s", noIssues)
	}

	failed := loaded(New(nil), pipeline.Result{}, errors.New("not logged in")).render()
	if !strings.Contains(failed, "not logged in") {
		t.Errorf("expected error message, found:\n%s", failed)
	}
}
