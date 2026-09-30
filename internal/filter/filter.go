package filter

import "github.com/destinyamba/gh-pullkin/internal/issues"

func Taken(issue issues.Issue) bool {
	return false
}

func Dead(lastMerge int64) bool {
	return false
}
