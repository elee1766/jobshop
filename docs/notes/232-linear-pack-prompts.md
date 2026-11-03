# Linear tracker prompts in packs/github (#232)

2026-11-03. packs/github takes `tracker` (default `github`). With
`tracker: linear` the `plan` and `dev` pool prompts and the `pr-work`
step prompts come from `prompts/linear/` instead of `prompts/github/`,
and `pr-work` gets `tracker: linear` and `linear_workspace`, so its
`issue` var is a key like ENG-7 (#201).

Vars: `linear_workspace`, `linear_team` (team key), `linear_target`
(the Linear issue the planner reads as its goal), `linear_done_state`
(default `Done`). Team to repo mapping is one pack instance per team,
with a `prefix:` when a host has several (#184).

The prompts use `jobshop linear issue|comment|state`: implement reads
the spec, moves it to In Progress and comments the PR link; review-merge
reads acceptance criteria from Linear; post-merge comments the shipped
sha and moves the issue to `linear_done_state`. The planner has no
list or create command, so it starts runs for keys the target names
and proposes new issues as target comments for a human to file.
Out-of-scope findings become dev tickets, not GitHub issues.

Unchanged: refinery, ci-repair, stall-check and their prompts still use
GitHub (they are about PRs and master CI). `issue_no_run` only sees
GitHub issues. Test: `TestGithubPackLinearTracker`. Live check:
`docs/verify/linear-pack.md`, waiting on a Linear token.
