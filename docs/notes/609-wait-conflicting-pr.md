# #609: checks waits on a conflicting PR

2026-11-05. GitHub runs no `pull_request` CI on a PR whose `mergeable` is
false, so a `github.checks:pr=...` wait on it never resolved and the step
sat until `--timeout` (1h), with nothing visible in CI. The wait evaluator
now reads `mergeable` from the PR fetch it already makes and clears such a
wait as `invalidated: <repo>#<n> conflicts with its base`. `null` (GitHub
still computing) keeps waiting. The implement and merge prompts tell the
agent to rebase, force-push and set the wait again.
