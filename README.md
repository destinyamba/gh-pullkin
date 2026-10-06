# pullkin

pullkin scans your project's dependencies, finds open issues in them that you could fix, and guides you to a merged PR. It uses no AI. Every answer comes from your files and GitHub data.

I built this because I struggled to find projects I wanted to contribute to or found intersting enough to be engaged and solve proactively. So I thought why not solve issues for project I already use daily or regularly.

## How pullkin picks issues

pullkin reads your `go.mod` and `package.json`, finds each dependency's GitHub repo, and looks for open issues labeled `good first issue` or `help wanted`. It then drops issues someone else is working on and repos that won't merge your PR.

### Taken issues

pullkin skips an issue if either is true:

- **It has a linked open pull request.** Someone is already working on it.
- **It was assigned in the last 28 days.** The assignee is likely still on it. If pullkin can't tell when the issue was assigned, it treats the assignee as active.

An issue assigned more than 28 days ago with no open PR is shown. The assignee may have moved on, so ask in the issue before you start.

### Dead repos

pullkin skips a repo if either is true:

- **It is archived.** Archived repos are read-only and can't accept PRs.
- **It has had no push in 3 months.** Your PR is unlikely to get a review. (Will review age of last push in future iterations)

These limits live in [`internal/filter/filter.go`](internal/filter/filter.go) as `FreshAssigneeDays` and `DeadAfterMonths`.
