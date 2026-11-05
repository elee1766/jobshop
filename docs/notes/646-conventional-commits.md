# Conventional commit PR titles (#646)

2026-11-05. Master is squash-only and the PR title is the commit
subject, which release-please (#644) reads, so titles are conventional
commits: `<type>(<scope>): <summary> (#<issue>)`.

- `.github/workflows/pr-title.yml` (job `pr-title`) runs
  `amannn/action-semantic-pull-request@v5` on `pull_request_target`
  (opened, edited, synchronize, reopened). Types: feat, fix, perf,
  refactor, docs, test, ci, build, chore, revert. Scope is free-form and
  optional; the summary must not start uppercase. It reads only the
  title and checks out no code. The operator adds `pr-title` to the
  master ruleset's required checks after it first runs.
- packs/github implement prompts (github and linear trackers) give the
  title format, the scope convention and `!` / `BREAKING CHANGE:` for
  breaking CLI, config, protocol or store schema changes. ci-fix uses
  `fix(<scope>): ...`. The merge prompts (github, linear, ci-merge) fix a
  non-conforming title with `gh pr edit --title` before merging.
- `CONTRIBUTING.md` (linked from README "Develop") states the rule for
  human contributors.
