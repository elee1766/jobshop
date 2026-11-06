# github.checks workflow= matching (JOB-88)

2026-11-06. `workflow=` on a `github.checks` wait matches a run
case-insensitively by workflow name, path, file name or file stem, as
trigger matching does. `workflow=ci` now finds a workflow named `CI` in
`.github/workflows/ci.yml`; before, only the exact name or file matched
and the wait sat at `ci not started` while CI was green.
