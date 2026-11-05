# Stall chain across instances (JOB-17)

2026-11-05.

- Builds on #704 (stall-check runs are never `needs_human` subjects, one
  check per stall key across triggers, later firings held while it or
  its escalation is open).
- `needs_human` runs belong to the pack instance that declares the run's
  workflow: a stall trigger only checks runs whose workflow comes from
  the same `packs` entry as its own `run.workflow`. Workflows declared in
  jobshop.yaml itself match any trigger.
- `jobshop decision approve|reject` works on proposals with
  `meta.linear.issue`. Approve adds `target`, removes `proposed`, moves
  open sub-issues to the first state of the team's
  `linear_issue_no_run` `states` (Todo when none) and comments. Reject
  comments and moves the issue to the team's first canceled state.
