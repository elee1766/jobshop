# Linear intake trigger (#287)

2026-11-04. Part of #258. Linear issues ready for work get a run
without the planner reading the target.

**Signal.** `linear_issue_no_run` is a stall signal the controller
computes (no `check`, `repo` or `after`). Fields on the stall entry:

| Field | Meaning |
| --- | --- |
| `team` | Linear team key. Empty turns the signal off |
| `states` | Comma-separated state names; empty matches any open state |
| `labels` | Comma-separated; the issue needs one of them, empty matches any |
| `skip_labels` | Comma-separated; any of them drops the issue |
| `poll` | How often Linear is read (default 5m) |

Each poll lists the team's issues (up to 8 pages of 250) with the token
of the identity covering the team (`identity.token`, #200/#278) and
keeps open ones (state type not completed or canceled) that match. Names
compare case-insensitively. Between polls the cached list is reused.
Every tick the controller drops issues that an open run carries in
`meta.linear.issue`. The subject is the issue key, so the stall-check
run key is `stall:linear_issue_no_run:ENG-12`. A Linear error is logged
(`linear_issue_no_run` WARN) and finds nothing.

**Pack.** packs/github lists the signal in its `stall` trigger with
`team: ${linear_team}`, so GitHub-tracker instances keep it off. New
vars: `linear_intake_states` (default `Todo`), `linear_intake_labels`
(empty), `linear_poll` (5m); `linear_skip_labels` is shared with `linear
eligible`. The stall prompt runs `jobshop linear eligible` and, if
eligible, `jobshop run start <prefix>pr-work --var issue=<KEY> --key
linear-<KEY>`; a repeated start returns the open run.

Tests: `TestLinearIntake` (fake GraphQL server: one matching issue fires
once; skip-labeled, wrong-state, closed, unlabeled and already-running
issues do not; one read per poll), `TestLinearIntakeOff`,
`TestLinearIntakeMatch`. Live check: `docs/verify/linear-intake.md`.
