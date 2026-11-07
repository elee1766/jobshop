# Jobcity migration examples and parity inventory

2026-11-07. JOB-112. Source snapshot: `elee1766/jobcity` commit `7b04817`,
read locally on November 7, 2026. This is a source inventory, not a claim
about live deployed state and not a deployment instruction. Re-read the
operator's current approved baseline before a migration. Jobcity policy,
credentials and operational instructions remain in its own repository.

Use [migration](migration.md) for the cutover procedure and
[callbacks](callbacks.md) for lifecycle contracts and [automation](automation.md)
for built-ins. The tested demo is
intentionally smaller than a production jobcity pack. Do not replace the
production flows with a shortened example that omits guards or wakeups.

## Review-request reconciliation

`packs/gfx/pack.yaml` defines one `review-requests` trigger per enabled
repository instance. Preserve these six registrations:

| Source | Event | Action | Additional filter |
| --- | --- | --- | --- |
| GitHub | `pull_request` | `review_requested` | Instance repository |
| GitHub | `pull_request` | `labeled` | Instance repository and label `review:elee-agent` |
| GitHub | `pull_request` | `synchronize` | Instance repository |
| GitHub | `pull_request` | `ready_for_review` | Instance repository |
| GitHub | `issue_comment` | `created` | Instance repository |
| GitHub | `pull_request_review_comment` | `created` | Instance repository |

The fallback is `15m`, check timeout `2m`, cooldown `1m`, `max_runs: 1`
per stall identity and `max_open: 10`. Each event is only a wakeup. The
legacy check re-reads authoritative GitHub state with:

```text
deno run -A "$JOBCITY_ROOT/tools/main.ts" pr review-requests --short --pr "{{.event.pr}}" "{{.repo}}"
```

A targeted event supplies the PR number. The fallback supplies an empty PR
and lists all applicable requests. Keep that distinction in the Starlark
version. Pass values as argv, not interpolated shell source. Give the hook
explicit executable/root paths and only the environment it needs. Shell
hooks do not inherit an agent identity or its sandbox automatically.

The discovered subject is the PR/head identity consumed by `pr-review`.
The legacy `review_requested` stall uses run key
`stall:review_requested:<subject>`, plus `meta.trigger.stall` equal to
`review_requested:<subject>` and trigger-scoped metadata key
`<instance-prefix>review-requests/stall:review_requested:<subject>`.
Preserve the workflow identity, instance prefix and subject spelling as
well as the textual key. Starting `review:<subject>` would not be a
compatible migration.

The full `pr-review` workflow also routes review risk, obtains exact-head
browser/device evidence, joins the selected review result and conditionally
merges approved third-party work. The intake callback must not replace
those phases with a single implementation or merge action. An
`issue_comment` payload does not contain the full PR. Fetch missing state
explicitly and preserve the screening/delegation policy.

## Delegated Linear intake

`packs/city/pack.yaml` has a host-wide unprefixed `linear-delegated`
trigger, enabled by the root instance. It wakes on Linear `Issue` `update`,
uses a `5m` fallback, `1m` check timeout and `max_runs: 1`. Its authoritative
check is `jcx linear delegated --short`. The source webhook is not proof
that an issue was delegated, nor permission to start arbitrary issues.

The trigger starts `linear-intake` with `issue=<subject>` and legacy key
`stall:delegated:<subject>`. That workflow's command step runs
`bin/intake-delegated`, which retains eligibility, routing and delegation
checks before starting an instance's `implement`. The deliberate
`bin/start-linear` entry point similarly checks eligibility and starts
`<instance>-implement` with `issue=<KEY>` and `key=linear-<KEY>`.
Preserve both layers' dedup identities. Do not collapse the intake run and
implementation run into one differently keyed action without checking
existing work.

No new work is created just because a human creates, edits or comments on
an untracked Linear issue. Preserve allowed delegators and the existing
application-user delegation gate. Keep current source filters and
intervals, not a generic all-issues poll.

## Stuck work, scheduled work and learnings

The gfx `stall` trigger has these signals:

| Signal | Age threshold |
| --- | --- |
| `ready_unclaimed` | `1h` |
| `pool_quarantined` | No age field |
| `step_stalled` | `3h` |
| `session_silent` | `1h` |
| `needs_human` | `4h` |
| `blocked` | `1h` |

It starts `stall-check` with `signal`, `subject` and `detail`, has a `30m`
cooldown and `max_runs: 2`. Preserve per-signal subject discovery,
existing-run lookup and escalation. A generic timer alone is not a
replacement for these predicates.

The host-wide city `improvement-sweep` uses cron `17 * * * *`, workflow
`improvement-sweep` and run key `sweep`. Keep it separate from the city
maintenance trigger `prune-testcontainers` (`*/30 * * * *`, key `prune`),
which has its own enable flag. Per-instance `review-lessons` uses
`47 8,20 * * *`. Per-instance `curate` uses the explicit learnings source:
minimum 5, interval `24h`, scope `repo:<repo>`, workflow `curate`, key
`curate`, gated by `learn_enabled`. The default-off flag remains meaningful
even when definitions are present.

Do not rename learnings wakeups to stuck-work signals. They select
different persisted state and need their own mapping and acceptance test.

## Implementation routing

The existing `implement` begins with agent `triage` (one retry), whose
recorded outcomes select `implement-luna` for `mechanical`,
`implement-sonnet` for `scoped`, or `implement-opus` for `primary` (three
retries each). `done` represents no-code/not-work. A command join named
`implemented` chooses the route that ran. A `post-merge` command follows
only a `merged` outcome, with two retries, `5m` retry delay and `2m`
timeout. These names and conditions are persisted compatibility inputs.

A Starlark route can replace gated sibling selection with a dictionary and
branch, but an in-flight old run still has the old sibling step IDs and
join. Drain or explicitly version the workflow as described in
[migration](migration.md#6-preserve-existing-runs-and-deduplication).
Do not delete the post-merge step or make triage a merge authority. Keep
agent triage during the format change. Jev adoption is separate.

## Instance and host boundaries

The snapshot keeps separate operator and city packs unprefixed, plus gfx
instances for retail, trade, account, sap, uniswapgo, oku-cli and
uniswap-sdks. Canoe remains disabled. Preserve each feature flag, not only
a common `enabled` flag. Examples of nonuniform values include Retail's
`main` branch and Android evidence, Trade's `master` default and browser
evidence, SAP/uniswapgo approver overrides, Account's merge policy and
public-repository author screening. Do not copy one instance's policy onto
all repositories as part of CUE unification.

Keep shared Android/iOS/browser resources and the operator escalation pool
available when any enabled flow requires them. Preserve instance-scoped
prompt-rule matching. Authentication, secret sources, origin-scoped
capability handles and service runtime security stay host-owned. A pack
variable cannot expand the host grant.
