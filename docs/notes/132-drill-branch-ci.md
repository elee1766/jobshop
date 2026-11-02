# CI on the drill branch

2026-11-02 (#132). `ci.yml` runs only on pushes to master and on pull
requests, so `drill/red-ci` from `docs/verify/ci-repair-drill.md` never
got a CI run and step 2 waited forever (seen in the #117 drill, worked
around with draft PR #130). The runbook now opens a draft PR for the
branch and closes it at cleanup. Adding `drill/**` to the push trigger
was the other option, but agent tokens cannot push workflow changes.
