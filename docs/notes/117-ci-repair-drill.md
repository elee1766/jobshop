# CI repair drill (#117)

Added 2026-11-02. `packs/github` var `ci_branch` (default `master`, kept
by selfdev) is the branch the `master-red` check passes to `gh run list
--branch`. The drill (`docs/verify/ci-repair-drill.md`) starts a
`ci-repair` run by hand for a red SHA on a throwaway `drill/red-ci`
branch, so master never goes red. Test: `TestGithubPackLoads` overrides
`ci_branch`.
