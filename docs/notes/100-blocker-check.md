# selfdev blocker-check (#100)

`blocker-check` (vars `[ticket, issue]`, `issue: issue`,
`escalate_pool: plan`) has one step, `assess` (plan, `retries: 1`,
`prompts/blocker.md`). The planner starts it with `--key blocker-<id>`
(`ticket=none`, `--key blocker-issue-<n>` when only the issue claims a
blocker). The step checks self-deploy (master vs `jobshop version`),
pools (`ticket list --orphaned`) and permissions or credentials, comments
the evidence on the issue, then either routes the work back (`ticket
update --pool`, `run resume`, drop `needs-human`) and closes `done`, or
closes `failed`, which escalates at once.
