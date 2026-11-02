# selfdev pack

Jobshop agents that develop Jobshop through GitHub issues and PRs.

## Flow: one `pr-work` run per issue

The planner files an issue and starts
`jobshop run start pr-work --var issue=<n> --key issue-<n>`. The run has
three steps, each blocked on the one before (`docs/plan.md` 7.15, 7.16):

1. `implement` (dev, 3 retries, `prompts/implement.md`): branch, PR,
   green CI, comment `PR <url> head <sha>`, close `done`.
2. `review-merge` (refinery, 2 retries, `prompts/merge.md`): review
   against the issue, merge with `--match-head-commit`. Findings: comment
   them, `jobshop run resume <run> --step implement`, and close this step
   `failed` (it reopens behind implement and counts an attempt).
3. `post-merge` (refinery, 2 retries, `prompts/post-merge.md`): master CI
   green on the merge commit and `jobshop version` at that commit or later
   (self-deploy picked it up), then a "shipped" comment on the issue.

The last step closing `done` closes the run. A step that runs out of
retries sets the run to `needs_human` and files "Escalate run <id>" in
`plan`; the planner labels the issue `needs-human` and mentions @elee1766.
Dependent issues: an edge from the dependent run's `implement` step to the
dependency's `post-merge` step. `jobshop run show <run>` shows progress.

## Flow: `blocker-check` on blocked claims

When an open issue or ticket says it is blocked or waiting on an operator,
the planner starts `jobshop run start blocker-check --var ticket=<id>
--var issue=<n> --key blocker-<id>`. One step, `assess` (plan, 1 retry,
`prompts/blocker.md`), re-checks the claim against current state: is
self-deploy live (master vs `jobshop version`), does the pool exist
(`jobshop ticket list --orphaned`), do the permission or credential
exist. Stale: it comments the evidence, routes the ticket back to a pool
and closes `done`. Real: it closes `failed`, which escalates at once.

## Pools and triggers

- `dev` pool: works `implement` steps. Tasks that are not run steps (e.g.
  `master-red`) are turned into a run for their issue. Prompt:
  `prompts/dev.md`.
- `refinery` pool (1 session): works `review-merge` and `post-merge`
  steps, so merges happen one at a time. Prompt: `prompts/refinery.md`.
- `plan` pool (1 session): reads the `target` issue, ticks verified items,
  files the next 2-4 issues and starts their runs, handles escalations.
  Prompt: `prompts/plan.md`.
- `replan` trigger: when `dev` has no ready tickets, creates a planning
  ticket in `plan` (10m cooldown after the last one closes).
- `master-red` trigger: when the latest completed master CI run failed,
  creates one priority-0 `dev` ticket per failing SHA (check command via
  `gh`; the controller needs `GH_TOKEN`).

## Vars

| Var | Meaning |
| --- | --- |
| `repo` | Main checkout of github.com/elee1766/jobshop; sessions get worktrees of it |
| `agent` | Agent name defined in the host's jobshop.yaml |
| `dev_sessions` | `max_sessions` for the dev pool |

```yaml
packs:
  - path: /var/lib/jobshop/src/jobshop/packs/selfdev
    vars: {repo: /var/lib/jobshop/src/jobshop, agent: jcode, dev_sessions: 2}
agents:
  jcode: {command: jcode, args: [acp]}
```

Pools use the host's `default_project`. Prompts are read when the
controller loads config, so a merged prompt change applies on restart.
