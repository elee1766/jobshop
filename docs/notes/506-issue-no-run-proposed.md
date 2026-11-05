# issue_no_run skips proposed issues (#506)

2026-11-05. The github pack's `issue_no_run` stall source and the
stall-check prompt's "start a missing run" skip list now exclude issues
labeled `proposed`. A planner proposal (#488) waits on a human and never
gets a pr-work run; before this it opened `stall:issue_no_run:<n>` checks
that wasted a session. `pkg/config` runs the rendered jq filter on a
fixture (when jq is on PATH) to show a proposed issue is not emitted.
