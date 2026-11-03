# Work references carry the repo (#183)

2026-11-03. `meta.github` held a bare issue number, so a controller
working two repos could not tell `a/x#5` from `b/y#5`: the
`issue_no_run` stall check dropped an issue when any repo had a run for
that number, and escalation prompts said "the run's issue var" without
a repo.

Built:
- Workflows take `repo: owner/name` (required with `issue`) and an
  optional `pr` var (one of `issue` or `pr`). Runs, their steps and
  their escalation tickets get `meta.github = {repo, issue|pr}`.
  packs/github sets `repo: ${repo}`. Batches start runs through the
  same path, so their runs carry it too.
- The `issue_no_run` stall signal takes `repo:` (required) and only
  counts open runs whose `meta.github.repo` matches.
- The gateway's escalation prompt names `gh issue view <n> --repo
  <repo>` (or `gh pr view`) from `meta.github`.
- Migration: the database does not know the repo, so instead of SQL a
  controller pass at startup (`backfillGitHubRepo`) sets
  `meta.github.repo` on runs, steps and escalations that lack it, from
  the run's workflow config. Idempotent; runs of removed workflows are
  left alone.

Tests: `TestBackfillGitHubRepo`, `TestStallCheckSignal` (same number in
another repo), `TestGitHubRef`, workflow config cases. Live check:
`docs/verify/github-repo-refs.md`.
