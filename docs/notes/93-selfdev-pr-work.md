# selfdev pr-work (#93)

`packs/github/pack.yaml` (selfdev since #107) declares `pr-work` (vars `[issue]`, `issue:
issue`, `escalate_pool: plan`): `implement` (dev, 3 retries,
`prompts/implement.md`), `review-merge` (refinery, 2,
`prompts/merge.md`) and `post-merge` (refinery, 2,
`prompts/post-merge.md`). The planner runs `jobshop run start pr-work
--var issue=<n> --key issue-<n>` per new issue, and a dependency is an
edge from the dependent run's `implement` to the dependency's
`post-merge`. Pool prompts now only frame the role; the step body is the
instruction. Review findings: the refinery comments, runs `run resume
<run> --step implement`, and closes `review-merge` `failed`, so the step
reopens behind implement and counts an attempt. `post-merge` checks
master CI on the merge commit and that `jobshop version` reports that
commit or a descendant (self-deploy picked it up). Because a release
counts as a failed attempt (`90-step-retries-escalation.md`), the step prompts wait in-turn
(`gh run watch`, `sleep`) instead of deferring and releasing. A dev task
that is not a step (e.g. `master-red`) is turned into a run for its
issue. The refinery still handles older non-step hand-off tickets the
old way.
