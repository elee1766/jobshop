# Linear team scope in prompts

2026-11-05. Operator request: agents only ever look at the Linear teams
the host configured.

- linear/dev.md, linear/plan.md, stall.md (`linear_issue_no_run`) and
  operator.md say: only the instance's team (operator: the teams in the
  host's identities); never read or write another team, even when an
  issue links there; no other Linear tools or tokens (linearis, raw
  API); escalate work that needs another team.
- The controller already enforces it: `jobshop linear` gets its token
  through `identity.token`, which refuses a team no identity of the
  project covers, and the Linear webhook drops events for uncovered
  teams. The prompts close the gap of an agent reaching Linear some
  other way.
