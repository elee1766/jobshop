# In-flight PR links for issue-started runs (JOB-131, 2026-11-08)

`run.show` links used to come only from run vars and meta, so a run started
from an issue showed no PR in Desk's in-flight table even after its step
opened one.

- `run.show` now adds the PRs the run's steps record, in step order
  (`pkg/control/ticketgraph.go`, `runLinks`). Two sources, both written by
  the step itself:
  - the PR in a step's own `wait` conditions (`repo` and `pr`), which
    appears as soon as the step parks on CI;
  - a comment on the step that starts a line with `PR <github pull url>`,
    the handoff line the pr-work prompts already ask for.
- Never inferred: an issue number, a PR merely mentioned in a comment, or a
  comment on another ticket does not link. No PR recorded means no link, and
  the table shows a dash.
- Several PRs, and PRs in another repo, are all listed. The label is
  `repo#n` taken from the link, not from the run's repo.
- The in-flight table renders every PR link (wide column, and under the
  title in the compact layout) and the search box matches them. The overview
  is a live snapshot, so a new PR appears without a reload.
- The pr-work implement prompts now ask for `jobshop comment "PR <url>"` on
  the step right after the PR is opened, before the CI wait.

Not done: a separate "unavailable" state. The run carries no signal that
distinguishes "no PR yet" from "association data missing", so both show a
dash.

Checks: `TestRunShowStepPRLinks`; `ui/e2e/run-links.spec.js` (needs
`DESK_ISSUE_WORKFLOW`, see `docs/verify/inflight-pr-links.md`).
