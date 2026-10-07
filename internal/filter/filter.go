package filter

import (
	"time"

	"github.com/destinyamba/gh-pullkin/internal/issues"
)

const (
	FreshAssigneeDays = 28
	DeadAfterMonths   = 3
	StaleIssueMonths  = 12
)

func Stale(issue issues.Issue, now time.Time) bool {
	return issue.UpdatedAt.Before(now.AddDate(0, -StaleIssueMonths, 0))
}

func Taken(issue issues.Issue, now time.Time) bool {
	if issue.HasOpenPR {
		return true
	}
	if len(issue.Assignees) == 0 {
		return false
	}
	if issue.AssignedAt.IsZero() {
		return true
	}
	return issue.AssignedAt.After(now.AddDate(0, 0, -FreshAssigneeDays))
}

func Dead(repo issues.RepoInfo, now time.Time) bool {
	if repo.Archived {
		return true
	}
	return repo.PushedAt.Before(now.AddDate(0, -DeadAfterMonths, 0))
}
