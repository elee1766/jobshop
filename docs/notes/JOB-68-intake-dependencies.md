# Linear intake: stated dependencies block (JOB-68, 2026-11-06)

- `jobshop linear eligible` and the `linear_issue_no_run` stall signal
  treat an issue as blocked when it has an open Linear blocker (any team)
  or a description line starting `Depends on <owner/repo>#<n>` (also
  `repo#n` and a GitHub issue URL) whose GitHub issue is open. A closed
  issue does not block. Reason: `depends on open owner/repo#n`.
- A bare `repo#n` takes its owner from the cwd's git origin (CLI) or the
  first Linear-tracked workflow's repo (controller).
- The stall signal reads each matching issue's context once per poll,
  plus one GitHub read per stated dependency. A read error keeps the
  issue as a stall (logged), as before this check.
