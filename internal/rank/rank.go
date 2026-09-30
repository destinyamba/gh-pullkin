package rank

import (
	"github.com/destinyamba/gh-pullkin/internal/issues"
	"github.com/destinyamba/gh-pullkin/internal/manifest"
)

type Candidate struct {
	Issue issues.Issue
	Dep   manifest.Dep
	Score int
}

func Rank(cs []Candidate) []Candidate {
	return cs
}
