# pullkin handoff

## What it is

An open source TUI that scans a developer's project dependencies, finds open issues in those dependencies they could fix, and guides them from "I use this package" to "my PR got merged". Goal: make contributing to open source low barrier.

## Why it exists

The author (an individual dev, SE2) wants meaningful work outside the day job, in big codebases, solving hard problems by hand. Past attempts failed at every step after finding an issue: no link to the project, fear of scope, the fork/pull flow was scary, couldn't build the repo, no reminder, forgot about it. So the product is a guide, not just a finder.

## Prior art

- OpenSauced (github.com/open-sauced): main `app` and `hot` repos archived in 2025; `ai` repo (find projects matching your skills) last updated Sep 2024.
- goodfirstissue.dev, up-for-grabs.net, CodeTriage: list issues, but not based on your dependencies, and no guidance after you pick one.

## Settled decisions

- **Name:** `pullkin`. On 2026-09-30, pullkin.com, pullkin.dev, the npm package and the GitHub org were all free. Nothing bought yet; check again before launch.
- **Main user:** an individual dev running it on their own project. Teams and maintainers come later, if ever.
- **Goal:** personal and portfolio first; wide adoption is the stretch goal; no business yet.
- **Form:** a nice TUI.
- **No AI anywhere.** Every answer comes from real sources (CI workflow files, CONTRIBUTING.md, GitHub data). "The no-AI way to learn open source" may become its identity.
- **The flow (all six steps):**
  1. **Pick**: only issues in deps you actually use.
  2. **Size**: flag small ones (few files touched, clear repro, maintainer replied).
  3. **Start**: one action forks, clones, branches, and handles claiming.
  4. **Setup**: show build and test steps, mainly read from `.github/workflows/*.yml` plus CONTRIBUTING.md.
  5. **Nudge**: remember what you're working on and remind you.
  6. **Ship**: walk through push, PR, and replying to review.
  - MVP focus: Pick, Start, Setup, Nudge. Size and Ship can be basic at first.
- **Pipeline:** one parser per manifest file, then a shared path: deps.dev (package → source repo) → GitHub issues (`good first issue` / `help wanted`) → filter out taken ones (assignee, linked PR, "I'm on it" comment) and dead repos (no recent merges) → rank (direct deps above transitive).
- **Later idea:** find what the user actually imports from each dep, and boost issues that touch that code.

## Open decisions (resolve these first, with the user)

Recommendations given but not yet confirmed:

1. **Language/TUI:** Go + Bubble Tea (Charm). Recommended: the author writes Go; single binary.
2. **Install/run:** a `gh` extension (`gh-pullkin`) for the MVP, reusing gh's login and API access; standalone binary later.
3. **Nudge:** show in-progress work on launch, plus a shell greeting (`pullkin nudge` in `.zshrc`). A weekly GitHub Action is for later.
4. **State:** GitHub is the source of truth (issues you commented on, fork branches, your PRs); a local file is only a cache.
5. **Ecosystems in MVP:** Go + npm.
6. **Claiming etiquette:** read CONTRIBUTING.md and the repo's habits, draft the right comment (or none), and post only when the user confirms.

Still unvisited after those: thresholds for "stale" and "taken", how Size is scored, TUI screen layout, MVP cut line and first milestone, license, README positioning.

## Working with the author

- Keep replies short and plain; answer first.
- No comments in code by default.
- Never commit unless asked; the author drives their own git history. No Claude co-author trailer.
- They want to write much of the code by hand. Coach, and don't just generate everything; ask how much they want written for them.
