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
`plan`; the planner labels the issue `needs-human` and mentions `@<human>`.
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
- `stall` pool (1 session, `stall_model`, `stall_effort`): works `stall-check`
  `diagnose` steps, apart from the planner.
- CI waits: steps that wait on CI park with `jobshop wait add
  'github.checks:...' --any --resume` (#381) and resume within a minute
  of CI finishing, instead of a fixed `release --until` (#555). Deploys
  and tags park on `deploy:` and `github.tag:` waits (JOB-11); a timed
  release for an event is refused without `--time-wait`.
- `plan` pool (1 session): reads the `target` issue, ticks verified items,
  files the next 2-4 issues and starts their runs, handles escalations.
  Prompt: `prompts/github/plan.md`.
- `replan` trigger: when `dev` has no ready tickets, creates a planning
  ticket in `plan` (10m cooldown after the last one closes).
- `master-red` trigger: when the latest completed master CI run failed,
  starts a `ci-repair` run keyed by the failing SHA (check command via
  `gh`; the controller needs `GH_TOKEN`).
- `release-pr` trigger: a green release PR (label `release_pr_label`)
  starts a `release-merge` run keyed `release-pr-<n>`: one `merge` step in
  `refinery`, `prompts/release-merge.md` (#698).
- `pr-unmerged` trigger: a green non-draft PR untouched for 2h (minus the
  `pr_unmerged` skips) starts a `pr-review` run keyed `pr-<n>`:
  `review-merge` then `post-merge`, prompts in `prompts/pr-review/`. A
  `pr-work` run that opens a PR in another instance's repo starts that
  instance's `<prefix>pr-review` with the same key and its review-merge
  waits on that run (JOB-33).

Prompt rules `recall` and `learn-nudge` are in `prompts/rules/`. Trigger
tickets take `body_file` like pools take `prompt_file`, with pack vars
substituted.

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

## Linear instance

Issues live in Linear, code goes through GitHub PRs. One instance per
Linear team and repo, each with its own `prefix:` (#184). Host
jobshop.yaml:

```yaml
identities:
  eng-app:
    kind: linear-app                 # OAuth client_credentials (#278)
    client_id_file: /run/secrets/linear-client-id
    client_secret_file: /run/secrets/linear-client-secret
    workspace: acme
    teams: [ENG]
    # scopes default: read, write, issues:create, comments:create
hooks:
  - on: intake                       # checks Linear issue text (#288)
    command: jev decide
packs:
  - path: /src/jobshop/packs/github
    prefix: eng-
    vars:
      repo: acme/app
      checkout: /src/app
      agent: claude
      human: wile                    # GitHub user
      check_cmd: make test
      tracker: linear
      linear_workspace: acme
      linear_team: ENG
      linear_target: ENG-1           # planner goal issue
      linear_human: Wile             # Linear user mentioned on questions
      linear_intake_states: Todo
      linear_intake_labels: agent    # empty: any label
      linear_skip_labels: Blocked,target,proposed,tracking
      linear_done_state: Done
```

How one issue flows:

1. Intake: every `linear_poll`, the `linear_issue_no_run` stall signal
   lists ENG issues in `linear_intake_states` with an intake label and
   no skip label. One with no open run starts `eng-stall-check` (#287).
   The planner lists the same filter on its own passes (#292).
2. stall-check or the planner runs `jobshop linear eligible` and starts
   `eng-pr-work` with key `linear-<KEY>`. Either start returns the other's
   run. The `intake` hook reads the issue text first; a refusal labels
   the issue `needs-human` and files a ticket on `jobshop attention`,
   and closing that ticket `done` lets the next start through.
3. implement moves the issue to In Progress, opens a PR on `repo`, and
   comments its link. With `linear.ask_on_issue: true` (off by default,
   JOB-114) a worker question is posted on the issue mentioning
   `linear_human`, with label `Blocked`; a reply under it answers (#293).
4. The refinery merges; post-merge comments the sha and moves the issue
   to `linear_done_state`.

Optional push (#300): with `linear.webhook_listen` set and the
identity's `webhook_secret_file`, Linear's Issue and Comment webhooks
run steps 1 and 3's reply handling right away instead of at the next
poll. Linear must reach the address, so it needs a reverse proxy or
tunnel. Polling still runs.

Live checks: `docs/verify/linear-pack.md`, `docs/verify/linear-intake.md`.

## Vars

| Var | Meaning |
| --- | --- |
| `repo` | GitHub `owner/repo`; issues, PRs and the `master-red` check use it |
| `checkout` | Main checkout of the repo on the host (pool `repo:`); sessions get worktrees of it |
| `agent` | Agent name defined in the host's jobshop.yaml |
| `human` | GitHub user mentioned on escalations |
| `propose_targets` | `true` (default): when targets run out the planner files one issue labeled `proposed` (`Proposed target: ...`, created if missing) and a `proposed` ticket that `jobshop attention` lists as a decision. A human approves by relabeling it `target`. `false`: only humans file targets (#488) |
| `escalation_chat_pool` | Empty (default): an escalation is only a ticket. A `type: chat` pool (e.g. the operator pack's `operator`): the controller opens a chat session there, key `escalation:<ticket id>`, whose first message (model context) is picked by cause, see Escalation chat prompts below, and records it in `meta.escalation`. Desk's chat pool should be the same (#379) |
| `authors` | Comma-separated GitHub logins whose issues start work besides `human` and App-filed issues. Others wait until someone adds the `accepted` label (#340) |
| `stall_pr_skip_authors` | Comma-separated logins whose green PRs `pr_unmerged` does not flag, besides `human` (#509) |
| `stall_pr_skip_labels` | Comma-separated PR labels `pr_unmerged` does not flag. Default `hold` (#509) |
| `release_pr_label` | Label of release-please PRs. Default `autorelease: pending`. A green, non-draft open PR with it gets one `release-merge` run (refinery merges it without review), and `pr_unmerged` does not flag it. Empty turns the `release-pr` trigger off (#698) |
| `check_cmd` | Build and test command run before pushing, e.g. `make test` |
| `docs_rule` | Where implement puts design notes and runbook checks. Default: a new file per note, never a numbered section appended to a shared doc (parallel PRs collide on the number) |
| `ci_workflow` | GitHub Actions workflow name `master-red` watches, and the workflow the implement and post-merge `github.checks` waits follow (`workflow=`). Default `ci` |
| `ci_branch` | Branch `master-red` watches. Default `master` |
| `dev_sessions` | `max_sessions` for the dev pool. Default 2 |
| `dev_model`, `dev_effort` | dev pool `model` and `effort`. Default empty: the agent's |
| `dev_sonnet_sessions`, `dev_sonnet_model`, `dev_sonnet_effort` | The `dev-sonnet` pool for easier implement steps (JOB-59). Defaults 1, `claude-sonnet-5-5`, `high` |
| `dev_pool` | Default pool of pr-work's `implement` step, also a run var (`--var dev_pool=dev-sonnet`). A failed attempt or review resume moves the step to `dev` (`retry_pool`). Default `{{.pack_prefix}}dev`, the instance's own dev pool |
| `refinery_sessions` | `max_sessions` for the refinery. Default 1; more needs strict branch protection |
| `refinery_model`, `refinery_effort` | Refinery pool `model` and `effort` (#304). Default empty: the agent's |
| `stall_model`, `stall_effort` | `stall` pool `model` and `effort`. Default empty: the agent's |
| `guide` | Where agents read the repo's conventions. Default: README, AGENTS.md, CONTRIBUTING.md |
| `deploy_check` | How `post-merge` and `stall-check` tell a commit is deployed. Default: no deploy, the check passes |
| `stall_after`, `stall_step_after`, `stall_human_after` | Stall thresholds (Go durations). Defaults 1h, 3h, 4h |
| `merge_method` | `squash` (default) or `merge`: `gh pr merge --<method> --match-head-commit <head>` |
| `required_approver` | Empty (default): no approval needed. A GitHub login: merge only when that login has an `APPROVED` review whose `commit_id` is the current head; otherwise the merge step closes `failed` and retries |
| `required_checks` | Empty (default): every check in `statusCheckRollup` must pass. Otherwise comma-separated check names that must be `SUCCESS` on the head |
| `tracker` | `github` (default) or `linear`: where `pr-work` issues live. Selects `prompts/<tracker>/` for the plan and dev pools and the pr-work steps (#232) |
| `linear_workspace`, `linear_team` | With `linear`: the workspace and the team key (ENG) this instance works. One instance per team and repo, with a `prefix:`. With `github` (#652): set both to also take Linear issues of that team; they run as `pr-work-linear` (Linear prompts, same pools, PRs on GitHub) and the planner lists them too. Empty `linear_team`: GitHub only |
| `linear_target` | With `linear`: the Linear issue the planner reads as its goal (ENG-1) |
| `linear_done_state` | With `linear`: state post-merge moves a shipped issue to. Default `Done` |
| `linear_human` | With `linear`: the Linear user worker questions and run escalations mention when posted on the run's issue (#293). Replies under that comment answer the question |
| `linear_intake_states`, `linear_intake_labels`, `linear_skip_labels`, `linear_poll` | With `linear_team`: the `linear_issue_no_run` stall starts stall-check for open issues in these states (default `Todo`), with one of these labels (any when empty), none of the skip labels (default `Blocked`, set while a question waits, #293, `target` and `proposed`, planner parents, JOB-4, and `tracking`; set one with `jobshop linear label <KEY> add <label>`), and no open run. Linear is read every `linear_poll` (5m) (#287). The planner lists the same filter (#292) |
| `linear_new_state` | State of Linear issues agents file (out-of-scope finds, review findings, planner work). Default `Todo`, an intake state, so they are worked; Backlog is left to humans |
| `linear_intake_assignee` | Planner `linear list --assignee` filter: a name, `me`, or empty for any (#292) |

### Escalation chat prompts

With `escalation_chat_pool` set, a hold that needs a human opens a chat
session keyed `escalation:<ticket id>` (#379). Its first message is
context for the model, not a visible message: the session answers with a
summary. A workflow's `escalate` block picks the prompt by cause:
`prompt` (run escalation, `prompts/escalation.md`), `pool_prompt` (pool
quarantine, `prompts/escalation-pool.md`) and `hold_prompt` (a wait past
its timeout, a guard ask_human, any task or step labeled `needs-human`,
`prompts/escalation-hold.md`). Questions get no chat. Template vars,
besides the pack vars, always set and empty when the cause has none:

| Var | Meaning |
| --- | --- |
| `ticket`, `title`, `body` | The escalation ticket |
| `cause` | `run`, `pool`, `wait`, `guard` or `label` |
| `reason` | Short cause text: the failed step and its last reason, the quarantine error, the wait or guard reason, the last comment |
| `held`, `held_title` | The held ticket: the failed step, the waiting or labeled ticket, the guard session's ticket |
| `worker` | The session that raised the hold (last holder of the step, who added the label, who asked the guard) |
| `run`, `workflow`, `step` | The run, its workflow and the failed step name (run escalations) |
| `issue_ref` | How to read the issue (`gh issue view ...` or Linear) |
| `pool`, `quarantine` | The quarantined pool and its error (pool escalations) |

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
  # jcode's own memory is off: learnings come from jobshop (#320).
  jcode: {command: jcode, args: [acp], env: {JCODE_MEMORY_ENABLED: "0"}}
```

Pools use the host's `default_project`. Prompts are read when the
controller loads config, so a merged prompt change applies on restart.
Pool prompts get pack vars at load; step prompts get them at run start.
