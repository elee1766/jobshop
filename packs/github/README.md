# github pack

Agents that work one GitHub repo through issues and PRs: a planner, dev
sessions and a refinery that merges. `packs/selfdev` is this pack with
Jobshop's values (`extends:`, `docs/notes/107-github-pack-extends.md`).

## Flow: one `pr-work` run per issue

The planner files an issue and starts
`jobshop run start pr-work --var issue=<n> --key issue-<n>`. The run has
three steps, each blocked on the one before (`docs/notes/89-runs.md`, `90-step-retries-escalation.md`):

1. `implement` (dev, 3 retries, `prompts/github/implement.md`): branch, PR,
   green CI, comment `PR <url> head <sha>`, close `done`.
2. `review-merge` (refinery, 2 retries, `prompts/github/merge.md`): review
   against the issue, merge with `--match-head-commit`. Findings: comment
   them, `jobshop run resume <run> --step implement`, and close this step
   `failed` (it reopens behind implement and counts an attempt).
3. `post-merge` (deploy-check, 2 retries, `prompts/github/post-merge.md`): master CI
   green on the merge commit and the deploy check (`deploy_check`), then
   a "shipped" comment on the issue.

The last step closing `done` closes the run. A step that runs out of
retries sets the run to `needs_human` and files "Escalate run <id>" in
`plan`; the planner labels the issue `needs-human` and mentions `@${human}`.
Dependent issues: an edge from the dependent run's `implement` step to the
dependency's `post-merge` step. `jobshop run show <run>` shows progress.

## Flow: `stall-check` on stalled work

The `stall` trigger (`when.stall`, `docs/notes/143-stall-check.md`)
starts one `stall-check` run per stall with key
`stall:<signal>:<subject>`: a ready ticket nobody claimed
(`stall_after`, 1h), a crash-looping pool, a held step with no progress
(`stall_step_after`, 3h), a session holding work with a silent journal
(1h), a `needs_human` run or ticket (`stall_human_after`, 4h), a ticket
labeled `blocked` (1h), and through `gh` checks: an open issue with no
run, a green PR nobody merged for 2h, no commit on `ci_branch` for 6h
while PRs are open. One step, `diagnose` (plan, 1 retry,
`prompts/stall.md`): it fixes what it can in place (start the missing
run, reroute or requeue, resume, rebase, cancel a stale drill, restart a
session, file a bug) and closes `done`, or closes `failed` with the
evidence, which escalates at once. Host config problems always
escalate. `max_runs: 2` per key, 30m cooldown, at most 3 open at once.
It replaces `blocker-check`: the planner labels a blocked claim
`blocked` and the trigger picks it up.

## Pools and triggers

- `dev` pool: works `implement` and `fix` steps. Tasks that are not run
  steps are turned into a run for their issue. Prompt: `prompts/github/dev.md`.
- `refinery` pool (`refinery_sessions`, default 1): works `review-merge` and `merge` steps.
  More than 1 needs branch protection with required, up-to-date checks. Prompt: `prompts/refinery.md`.
- `deploy-check` pool (`deploy_sessions`, default 3): works `post-merge`
  steps, which wait up to an hour for the deploy, so they never queue a
  review behind them (#167). Prompt: `prompts/deploy-check.md`.
- `plan` pool (1 session): reads the `target` issue, ticks verified items,
  files the next 2-4 issues and starts their runs, handles escalations.
  Prompt: `prompts/github/plan.md`.
- `replan` trigger: when `dev` has no ready tickets, creates a planning
  ticket in `plan` (10m cooldown after the last one closes).
- `master-red` trigger: when the latest completed master CI run failed,
  starts a `ci-repair` run keyed by the failing SHA (check command via
  `gh`; the controller needs `GH_TOKEN`).

## Flow: `ci-repair` per failing master SHA

`master-red` runs `jobshop run start ci-repair --var sha=<sha> --key
<sha>` itself (a `run:` trigger, `docs/notes/98-run-triggers.md`). Two steps:

1. `fix` (dev, 2 retries, `prompts/ci-fix.md`): read
   `gh run view --log-failed`, find or file the fix issue, open a PR,
   green CI, comment `PR <url> head <sha>`, close `done`.
2. `merge` (refinery, 2 retries, `prompts/ci-merge.md`): check the PR
   fixes the failure, merge with `--match-head-commit`, wait for master
   CI on the merge commit. Findings: `run resume <run> --step fix` and
   close `failed`.

The key makes one run per SHA: once it closes, the same SHA does not
start another. Two failed `fix` attempts set the run to `needs_human`
and file "Escalate run <id>" in `plan`.

## Vars

| Var | Meaning |
| --- | --- |
| `repo` | GitHub `owner/repo`; issues, PRs and the `master-red` check use it |
| `checkout` | Main checkout of the repo on the host (pool `repo:`); sessions get worktrees of it |
| `agent` | Agent name defined in the host's jobshop.yaml |
| `human` | GitHub user mentioned on escalations |
| `check_cmd` | Build and test command run before pushing, e.g. `make test` |
| `docs_rule` | Where implement puts design notes and runbook checks. Default: a new file per note, never a numbered section appended to a shared doc (parallel PRs collide on the number) |
| `ci_workflow` | GitHub Actions workflow name `master-red` watches. Default `ci` |
| `ci_branch` | Branch `master-red` watches. Default `master` |
| `dev_sessions` | `max_sessions` for the dev pool. Default 2 |
| `refinery_sessions` | `max_sessions` for the refinery. Default 1; more needs strict branch protection |
| `guide` | Where agents read the repo's conventions. Default: README, AGENTS.md, CONTRIBUTING.md |
| `deploy_check` | How `post-merge` and `stall-check` tell a commit is deployed. Default: no deploy, the check passes |
| `stall_after`, `stall_step_after`, `stall_human_after` | Stall thresholds (Go durations). Defaults 1h, 3h, 4h |
| `merge_method` | `squash` (default) or `merge`: `gh pr merge --<method> --match-head-commit <head>` |
| `required_approver` | Empty (default): no approval needed. A GitHub login: merge only when that login has an `APPROVED` review whose `commit_id` is the current head; otherwise the merge step closes `failed` and retries |
| `required_checks` | Empty (default): every check in `statusCheckRollup` must pass. Otherwise comma-separated check names that must be `SUCCESS` on the head |
| `tracker` | `github` (default) or `linear`: where `pr-work` issues live. Selects `prompts/<tracker>/` for the plan and dev pools and the pr-work steps (#232) |
| `linear_workspace`, `linear_team` | With `linear`: the workspace and the team key (ENG) this instance works. One instance per team and repo, with a `prefix:` |
| `linear_target` | With `linear`: the Linear issue the planner reads as its goal (ENG-1) |
| `linear_done_state` | With `linear`: state post-merge moves a shipped issue to. Default `Done` |
| `linear_intake_states`, `linear_intake_labels`, `linear_skip_labels`, `linear_poll` | With `linear_team`: the `linear_issue_no_run` stall starts stall-check for open issues in these states (default `Todo`), with one of these labels (any when empty), none of the skip labels, and no open run. Linear is read every `linear_poll` (5m) (#287) |

Approval check the refinery runs (prints the number of matching
approvals, above 0 to merge):

```sh
gh api --paginate repos/<repo>/pulls/<n>/reviews --jq \
  '[.[] | select(.user.login == "<required_approver>" and .state == "APPROVED" and .commit_id == "<headRefOid>")] | length'
```

```yaml
packs:
  - path: /var/lib/jobshop/src/jobshop/packs/github
    vars:
      repo: acme/widget
      checkout: /var/lib/jobshop/workspaces/widget
      agent: jcode
      human: wile
      check_cmd: make test
agents:
  jcode: {command: jcode, args: [acp]}
```

Pools use the host's `default_project`. Prompts are read when the
controller loads config, so a merged prompt change applies on restart.
Pool prompts get pack vars at load; step prompts get them at run start.
