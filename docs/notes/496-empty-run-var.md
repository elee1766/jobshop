# Empty run vars rejected (#496, 2026-11-05)

`run start` (RPC and run: triggers) fails when the workflow's `issue:` or
`pr:` var is given an empty or blank value. Before, `--var issue=` started
a pr-work run titled "Implement #" with key `issue-`, and every later empty
start returned that same broken run. Other vars may still be empty: stall
triggers pass an empty `detail`.
