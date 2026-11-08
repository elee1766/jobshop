# REST PR reads and revert rules in pkg/github (JOB-159, 2026-11-08)

Part of JOB-139. No behavior change.

## What moved to pkg/github (pulls.go)

- `Pull` (was `control.ghPR`): number, title, body, html_url, merged,
  merge commit sha and mergeable_state of GET /repos/{r}/pulls/{n}.
- `Client.PullRequest` and `Client.CommitPulls`, with the `PullSource`
  interface they satisfy (was `control.revertSource`). The trace route is
  still `/repos/{repo}/activity`.
- `IsRevert` and `RevertedPR` with the two regexes (were `isRevert`,
  `revertedPR`). Pure apart from the `PullSource` reads.

## What stays in pkg/control

- `liveActivity` picks the repo's client (`githubOrHost`, identity token)
  and calls `Client.PullRequest`/`CommitPulls`.
- `checkReverts` (learning writes, once per revert), `prWorkRun` and
  `rebaseResolution` (the rebase decision; the state rule is
  `github.NeedsRebase`).

## Two readers, on purpose

`Pull` is the REST view (identity token per repo). `PRInfo` is the gh-CLI
view (host token) the waits read. Token scopes are unchanged
(see JOB-143-github-observe.md).
