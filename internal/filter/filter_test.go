package filter

import (
	"testing"
	"time"

	"github.com/destinyamba/gh-pullkin/internal/issues"
)

var now = time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

func daysAgo(n int) time.Time {
	return now.AddDate(0, 0, -n)
}

func TestTaken(t *testing.T) {
	tests := []struct {
		name  string
		issue issues.Issue
		want  bool
	}{
		{"no assignee, no PR", issues.Issue{}, false},
		{"open PR", issues.Issue{HasOpenPR: true}, true},
		{"open PR beats stale assignee", issues.Issue{HasOpenPR: true, Assignees: []string{"a"}, AssignedAt: daysAgo(90)}, true},
		{"fresh assignee", issues.Issue{Assignees: []string{"a"}, AssignedAt: daysAgo(3)}, true},
		{"assignee one day under the limit", issues.Issue{Assignees: []string{"a"}, AssignedAt: daysAgo(FreshAssigneeDays - 1)}, true},
		{"stale assignee", issues.Issue{Assignees: []string{"a"}, AssignedAt: daysAgo(FreshAssigneeDays + 1)}, false},
		{"assignee with unknown date", issues.Issue{Assignees: []string{"a"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Taken(tt.issue, now); got != tt.want {
				t.Errorf("expected %v, found %v", tt.want, got)
			}
		})
	}
}

func TestStale(t *testing.T) {
	tests := []struct {
		name  string
		issue issues.Issue
		want  bool
	}{
		{"updated recently", issues.Issue{UpdatedAt: daysAgo(10)}, false},
		{"updated just inside the limit", issues.Issue{UpdatedAt: now.AddDate(0, -StaleIssueMonths, 1)}, false},
		{"updated just outside the limit", issues.Issue{UpdatedAt: now.AddDate(0, -StaleIssueMonths, -1)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Stale(tt.issue, now); got != tt.want {
				t.Errorf("expected %v, found %v", tt.want, got)
			}
		})
	}
}

func TestDead(t *testing.T) {
	tests := []struct {
		name string
		repo issues.RepoInfo
		want bool
	}{
		{"recent push", issues.RepoInfo{PushedAt: daysAgo(10)}, false},
		{"archived with recent push", issues.RepoInfo{Archived: true, PushedAt: daysAgo(1)}, true},
		{"push just inside the limit", issues.RepoInfo{PushedAt: now.AddDate(0, -DeadAfterMonths, 1)}, false},
		{"push just outside the limit", issues.RepoInfo{PushedAt: now.AddDate(0, -DeadAfterMonths, -1)}, true},
		{"never pushed", issues.RepoInfo{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Dead(tt.repo, now); got != tt.want {
				t.Errorf("expected %v, found %v", tt.want, got)
			}
		})
	}
}
