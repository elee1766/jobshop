# Linear end to end: README and controller test (#298)

2026-11-04. Closes the tests-and-docs gaps of #258.

- `packs/github/README.md` has a `## Linear instance` section: a full
  host config (linear-app identity, `on: intake` hook, one prefixed
  instance per team and repo with the `linear_*` vars) and the flow of
  one issue from intake to `linear_done_state`.
- `TestLinearEndToEnd` (pkg/control) runs one issue through the pieces
  against one fake Linear: the `linear_issue_no_run` stall fires once
  and starts stall-check; the intake hook denies the run start and the
  issue gets `needs-human` and an attention item; closing it approves,
  and stall-check's and the planner's starts with `linear-<KEY>` give
  one run that hides the issue from intake; a worker question is posted
  with the `linear_human` mention and `Blocked`, and a reply under it
  answers the question and removes `Blocked`.
- `docs/verify/linear-pack.md`: a step 7 for the question and answer
  path (#293), and the App section now reruns steps 2 to 7. The commands
  it names match the merged CLI.
