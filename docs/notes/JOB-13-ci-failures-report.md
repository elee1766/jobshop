# CI failures in the planner's incident step (JOB-13)

2026-11-05. `jobshop report ci-failures --repo owner/name` lists the last
`--limit` (5) failed runs of workflow `--workflow` (ci) on `--branch`
(master) with gh, downloads each run's `failed-tests` artifact (#510) and
prints JSON: `runs`, `no_artifact` (failed runs without the artifact, e.g.
a build error or another job failing) and `tests`, one entry per test with
`flaky` and `broken` counts, the newest `verdict`, `runs` and `last`. Most
frequent first.

It runs gh in the session, not through the controller: it only reads, and
the session's gh already has the repo token. The github planner prompt
calls it after `report incidents`. Code: `pkg/flaky/runs.go`.
