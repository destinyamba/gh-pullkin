package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/destinyamba/gh-pullkin/internal/depsdev"
	"github.com/destinyamba/gh-pullkin/internal/issues"
	"github.com/destinyamba/gh-pullkin/internal/manifest"
	"github.com/destinyamba/gh-pullkin/internal/pipeline"
	"github.com/destinyamba/gh-pullkin/internal/rank"
	"github.com/destinyamba/gh-pullkin/internal/state"
)

func cand(n int, title string) rank.Candidate {
	return rank.Candidate{
		Issue: issues.Issue{ID: fmt.Sprintf("I_%d", n), Repo: depsdev.Repo{Owner: "o", Name: "r"}, Number: n, Title: title, URL: "https://x/" + title},
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
	case "esc":
		msg = tea.KeyPressMsg{Code: tea.KeyEscape}
	default:
		msg = tea.KeyPressMsg{Code: rune(key[0]), Text: key}
	}
	next, _ := m.Update(msg)
	return next.(Model)
}

func TestLoadingThenList(t *testing.T) {
	m := New(nil, nil, nil)
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
	m := loaded(New(nil, nil, nil), pipeline.Result{Deps: 1, Candidates: []rank.Candidate{cand(1, "a"), cand(2, "b")}}, nil)

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
	noManifest := loaded(New(nil, nil, nil), pipeline.Result{}, nil).render()
	if !strings.Contains(noManifest, "No go.mod or package.json") {
		t.Errorf("expected no-manifest message, found:\n%s", noManifest)
	}

	noIssues := loaded(New(nil, nil, nil), pipeline.Result{Deps: 3}, nil).render()
	if !strings.Contains(noIssues, "No fixable issues") {
		t.Errorf("expected no-issues message, found:\n%s", noIssues)
	}

	failed := loaded(New(nil, nil, nil), pipeline.Result{}, errors.New("not logged in")).render()
	if !strings.Contains(failed, "not logged in") {
		t.Errorf("expected error message, found:\n%s", failed)
	}
}

func run(m Model, cmd tea.Cmd) Model {
	if cmd == nil {
		return m
	}
	next, _ := m.Update(cmd())
	return next.(Model)
}

func pressCmd(m Model, key string) (Model, tea.Cmd) {
	next, cmd := m.Update(tea.KeyPressMsg{Code: rune(key[0]), Text: key})
	return next.(Model), cmd
}

func TestClaimPostsAfterYes(t *testing.T) {
	var gotID, gotBody string
	claim := func(issue issues.Issue, body string) (string, error) {
		gotID, gotBody = issue.ID, body
		return "https://x/comment", nil
	}
	m := loaded(New(nil, claim, nil), pipeline.Result{Deps: 1, Candidates: []rank.Candidate{cand(1, "a"), cand(2, "b")}}, nil)
	m = press(m, "down")

	m = press(m, "c")
	if !m.confirming || !strings.Contains(m.render(), issues.ClaimMessage) {
		t.Fatalf("expected confirm box with draft, found:\n%s", m.render())
	}

	m, cmd := pressCmd(m, "y")
	m = run(m, cmd)

	if gotID != "I_2" || gotBody != issues.ClaimMessage {
		t.Errorf("expected claim on I_2 with the claim message, found %q %q", gotID, gotBody)
	}
	if m.confirming || !strings.Contains(m.render(), "✓ claimed") {
		t.Errorf("expected claimed tag after posting, found:\n%s", m.render())
	}

	m = press(m, "c")
	if m.confirming || !strings.Contains(m.status, "already claimed") {
		t.Errorf("expected no second claim, found status %q", m.status)
	}
}

func TestClaimCancel(t *testing.T) {
	called := false
	claim := func(issues.Issue, string) (string, error) { called = true; return "", nil }

	for _, key := range []string{"n", "esc"} {
		m := loaded(New(nil, claim, nil), pipeline.Result{Deps: 1, Candidates: []rank.Candidate{cand(1, "a")}}, nil)
		m = press(press(m, "c"), key)
		if m.confirming {
			t.Errorf("%s: expected confirm box closed", key)
		}
	}
	if called {
		t.Errorf("expected no comment posted")
	}
}

func TestClaimError(t *testing.T) {
	claim := func(issues.Issue, string) (string, error) { return "", errors.New("forbidden") }
	m := loaded(New(nil, claim, nil), pipeline.Result{Deps: 1, Candidates: []rank.Candidate{cand(1, "a")}}, nil)

	m, cmd := pressCmd(press(m, "c"), "y")
	m = run(m, cmd)

	if !strings.Contains(m.status, "forbidden") || strings.Contains(m.render(), "✓ claimed") {
		t.Errorf("expected error status and no claimed tag, found:\n%s", m.render())
	}
}

func TestClaimedFromStateShowsTag(t *testing.T) {
	saved := []state.Work{{IssueID: "I_1", Repo: "o/r", Issue: 1, Title: "a", CommentURL: "https://x/c"}}
	m := loaded(New(nil, nil, saved), pipeline.Result{Deps: 1, Candidates: []rank.Candidate{cand(1, "a")}}, nil)
	if !strings.Contains(m.render(), "✓ claimed") {
		t.Errorf("expected claimed tag from saved state, found:\n%s", m.render())
	}
	m = press(m, "c")
	if m.confirming {
		t.Errorf("expected no second claim for a saved issue")
	}
}

func TestClaimPostedButNotSaved(t *testing.T) {
	claim := func(issues.Issue, string) (string, error) { return "https://x/c", errors.New("could not save") }
	m := loaded(New(nil, claim, nil), pipeline.Result{Deps: 1, Candidates: []rank.Candidate{cand(1, "a")}}, nil)

	m, cmd := pressCmd(press(m, "c"), "y")
	m = run(m, cmd)

	if !strings.Contains(m.render(), "✓ claimed") || !strings.Contains(m.status, "could not save") {
		t.Errorf("expected claimed tag and save warning, found:\n%s", m.render())
	}
}

func TestInProgressSection(t *testing.T) {
	m := New(nil, nil, nil)
	if strings.Contains(m.render(), "In progress") {
		t.Errorf("expected no section with nothing claimed")
	}

	saved := []state.Work{{IssueID: "I_9", Repo: "cli/cli", Issue: 9789, Title: "hangs"}}
	m = New(nil, nil, saved)
	if out := m.render(); !strings.Contains(out, "In progress") || !strings.Contains(out, "cli/cli#9789") {
		t.Errorf("expected section on the loading screen, found:\n%s", out)
	}

	claim := func(issues.Issue, string) (string, error) { return "https://x/c", nil }
	m = loaded(New(nil, claim, nil), pipeline.Result{Deps: 1, Candidates: []rank.Candidate{cand(1, "fresh")}}, nil)
	m, cmd := pressCmd(press(m, "c"), "y")
	m = run(m, cmd)
	if out := m.render(); !strings.Contains(out, "In progress") || !strings.Contains(out, "o/r#1") {
		t.Errorf("expected new claim in the section, found:\n%s", out)
	}
}
