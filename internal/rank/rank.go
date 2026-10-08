package rank

import (
	"slices"
	"strings"

	"github.com/destinyamba/gh-pullkin/internal/issues"
	"github.com/destinyamba/gh-pullkin/internal/manifest"
)

type Candidate struct {
	Issue issues.Issue
	Dep   manifest.Dep
	Score int
}

func Rank(cs []Candidate) []Candidate {
	slices.SortStableFunc(cs, func(a, b Candidate) int {
		// rule 1: direct first
		if a.Dep.Direct != b.Dep.Direct {
			if a.Dep.Direct {
				return -1
			}
			return 1
		}
		// rule 2: good first issue first
		if isGoodFirst(a.Issue) != isGoodFirst(b.Issue) {
			if isGoodFirst(a.Issue) {
				return -1
			}
			return 1
		}
		// rule 3: newer UpdatedAt first
		return b.Issue.UpdatedAt.Compare(a.Issue.UpdatedAt)
	})
	return cs
}

func isGoodFirst(issue issues.Issue) bool {
	for _, iss := range issue.Labels {
		if strings.EqualFold(iss, "good first issue") || strings.EqualFold(iss, "good-first-issue") {
			return true
		}
	}
	return false
}
