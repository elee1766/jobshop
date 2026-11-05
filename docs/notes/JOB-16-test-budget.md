# JOB-16: slowest tests (2026-11-05)

CI master run 42107670111 had four tests over 5s:

- `TestAdmissionTimeout` (5.04s each subtest): the hook ran `sleep 5`
  under `sh -c`. The context killed `sh`, but `sleep` kept stdout open,
  so `Output` waited the full 5s. Admission hooks now run in their own
  process group, killed whole on timeout, with a 1s `WaitDelay`. This
  was a real bug: a timed-out hook held admission until its children
  exited. Now 0.23s.
- `TestCLIDoc` (6.7s): built one kong parser per command serially.
  `cliDoc` renders them in parallel. Now under 1s.
- `TestCollectionIndexPlans` (5.9s on CI, under 1s locally): left alone;
  it seeds 300 rows and runs in parallel with the rest of the package.
