# pr-work route step

2026-11-10. `go test ./packs -run TestPRWorkRoutes` covers both routes
and an unknown outcome.

On a host running `packs/github` or `packs/selfdev`, after the deploy:

1. `jobshop run list --state running`: any run that reports `hook manifest
   changed` was open across the deploy. Cancel and restart it, or
   `jobshop run resume <run> --repin`.
2. `jobshop run start <prefix>pr-work --var issue=<n> --key issue-<n>`.
   `jobshop run show <run>` lists one step, `route`, in `<prefix>triage`.
3. When it closes, the issue has a `route=<word>: <reason>` comment and
   the run has an `implement` step in `<prefix>dev-sonnet` (`scoped`) or
   `<prefix>dev` (`primary`). There is no second implement step.
4. `jobshop run start <prefix>pr-work --var issue=<n> --var dev_pool=dev`
   is refused: the var is gone.
