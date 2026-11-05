# Work that ends without a PR (JOB-10)

2026-11-05. Linear and GitHub issues are not all code changes; a run had
no correct exit for a task, a check, a question or a tracking issue.

- `pr-work` and `pr-work-linear` (packs/github): `review-merge` and
  `post-merge` have `when: {step: implement, outcome: pr}`
  (docs/notes/385-step-needs-when.md). Implement closes `--outcome pr`
  when it opened a PR. With `--outcome done`, or no outcome, both later
  steps are skipped and the run closes done.
- `prompts/linear/implement.md` and `prompts/github/implement.md` start
  by classifying the issue: code change (`--outcome pr`), task with no
  code (do it, post the result, move to `linear_done_state` or close the
  GitHub issue, `--outcome done`), needs a person (`linear ask` /
  `report waiting`), not work (tracking, parent, target, duplicate:
  comment why, skip label or Backlog, `--outcome done`). The last issue
  comment names the path taken.
- A plain `jobshop close` on implement now skips review. The prompts say
  to close a PR with `--outcome pr`.

Test: `TestNonPRWork` (in-process host, fake agent closing with an
outcome): `done` skips both steps and the run closes done; `pr` runs all
three. Moving the issue to Done is the prompt's job, not the
controller's, so the test does not use a fake Linear.

Live: after deploy, move JOB-7 back to Todo; it should run the isolation
verify, comment the result and close without a PR.
