package rank

import (
	"slices"
	"testing"
	"time"

	"github.com/destinyamba/gh-pullkin/internal/issues"
	"github.com/destinyamba/gh-pullkin/internal/manifest"
)

func TestRank(t *testing.T) {
	tests := []struct {
		name string
		in   []Candidate
		want []int
	}{
		{
			name: "direct before indirect",
			in: []Candidate{
				cand(1, false, "help wanted", 1),
				cand(2, true, "help wanted", 1),
			},
			want: []int{2, 1},
		},
		{
			name: "good first issue before others",
			in: []Candidate{
				cand(4, false, "good-first-issue", 16),
				cand(1, true, "help wanted", 1),
				cand(2, true, "good first issue", 1),
			},
			want: []int{2, 1, 4},
		},
		{
			name: "newer updated at first",
			in: []Candidate{
				cand(4, false, "help wanted", 16),
				cand(1, true, "help wanted", 1),
				cand(2, true, "help wanted", 2),
			},
			want: []int{2, 1, 4},
		},
		{
			name: "label check ignores case",
			in: []Candidate{
				cand(1, true, "help wanted", 1),
				cand(2, true, "Good First Issue", 1),
			},
			want: []int{2, 1},
		},
		{
			name: "all rules together",
			in: []Candidate{
				cand(1, false, "good first issue", 4),
				cand(2, true, "help wanted", 1),
				cand(3, true, "Good First Issue", 2),
				cand(4, true, "help wanted", 3),
			},
			want: []int{3, 4, 2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Rank(tt.in)

			gotNums := make([]int, len(got))
			for i, cand := range got {
				gotNums[i] = cand.Issue.Number
			}

			if !slices.Equal(gotNums, tt.want) {
				t.Errorf("Rank() = %v, want %v", gotNums, tt.want)
			}
		})
	}
}

func cand(num int, direct bool, label string, day int) Candidate {
	return Candidate{
		Issue: issues.Issue{
			Number:    num,
			Labels:    []string{label},
			UpdatedAt: time.Date(2026, 10, day, 0, 0, 0, 0, time.UTC),
		},
		Dep: manifest.Dep{Direct: direct},
	}
}
