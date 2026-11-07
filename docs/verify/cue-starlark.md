# Verify CUE packs and Starlark hooks

2026-11-07. JOB-112. Read the [configuration guide](../configuration/README.md)
and [migration guide](../configuration/migration.md) first. This runbook
does not authorize deployment, capability grants or writes to a live store.

## Installed binary and fixture, read only

Set `JOBSHOP_SOURCE` in the operator environment to the checked-out source
revision matching the installed binary. This block reads configuration
without a controller, capability calls or file writes:

```sh readonly
: "${JOBSHOP_SOURCE:?set JOBSHOP_SOURCE to the matching source checkout}"
jobshop version
jobshop config check --file "$JOBSHOP_SOURCE/pkg/config/testdata/cue-starlark-example/jobshop.yaml"
jobshop config resolve --file "$JOBSHOP_SOURCE/pkg/config/testdata/cue-starlark-example/jobshop.yaml" | jq -e '.pools["demo-dev"] and .workflows["demo-change"]'
```

The fixture must load offline, select the default branch and apply the
`demo-` prefix. No agent starts. Missing source files, unsupported fields
or a mismatched binary are a failure, not permission to restart anything.

## Local repository checks

These commands build and write tool caches, so they are not unattended
read-only host blocks. Run in the source checkout:

```sh
go test ./pkg/config ./pkg/packcue ./pkg/automation/... ./pkg/hooks/... ./pkg/cli
go test ./pkg/control -run 'TestCUE|TestAutomation|TestHook|TestConfigWatch'
go run ./cmd/jobshop config check --file pkg/config/testdata/cue-starlark-example/jobshop.yaml
go run ./cmd/jobshop config resolve --file pkg/config/testdata/cue-starlark-example/jobshop.yaml | jq -e '.pools["demo-dev"].max_sessions == 0'
```

Require the tests that load this exact fixture, not only parallel inline
examples. Confirm loader tests cover imports, defaults, incomplete/conflict
errors, unknown fields, provenance, offline resolution and last-good
behavior. Runtime tests must load the example `.star` files and invoke
callbacks, not merely parse their text.

Before merging code changes also run the repository's full build, vet,
test, lint and Nix checks in [AGENTS.md](../../AGENTS.md). Record actual
commands and results in the PR, including any check not run. A regex that
matches no tests is not evidence of acceptance coverage.

| Boundary | Checked-in acceptance | What it does not establish |
| --- | --- | --- |
| Native CUE config | `TestCUEPack` in [cuepack_test.go](../../pkg/config/cuepack_test.go) loads the actual example module | A healthy controller or remote service |
| Callback API examples | `TestDocumentedHookExamples` in [examples_test.go](../../pkg/automation/examples_test.go) invokes the actual workflow/event files | Persisted scheduling, ingress or exactly-once effects |
| Capability examples | `TestDocumentedCapabilityExamples` in the same file exercises default-deny, local HTTP/Jev and an actual shell executable when present | Production credentials, public APIs or a service grant |
| Persisted controller transitions | Hook integration tests in [automation_test.go](../../pkg/control/automation_test.go) exercise recorded output routing, restart and whole-batch rejection | Production source-policy parity or operator rollout |

Keep the exact source-artifact tests separate from the controller tests:
one catches documentation drift, the other reaches persisted lifecycle
behavior. Report skipped shell execution or unmatched integration-test
names instead of counting those as passes.

## Current controller state, read only

Against an already authorized controller, capture the baseline without
changing triggers, sessions, cursors or runs:

```sh readonly
jobshop config show
jobshop trigger list
jobshop run list
jobshop attention
```

To inspect an authorized run's durable callbacks, set `JOBSHOP_ROOT` and
`RUN_ID` explicitly, then select an invocation's `Key` as `INVOCATION_ID`:

```sh readonly
jobshop --root "$JOBSHOP_ROOT" --socket "$JOBSHOP_ROOT/admin.sock" automation list --run "$RUN_ID" --limit 50
jobshop --root "$JOBSHOP_ROOT" --socket "$JOBSHOP_ROOT/admin.sock" automation show "$INVOCATION_ID"
```

Check the invocation and attempt `Value.actions`/`Value.targets`, recorded
input, terminal state and ordered calls. No immediate capability calls
means an empty call list, not missing callback execution. These audit
commands are read only but require the admin socket and can expose private
event data. See [audit field interpretation](../configuration/callbacks.md#inspect-and-resume-recorded-work).

Check the intended config revision and provenance, active/disabled names,
watch state, source registrations and existing run IDs. Compare with the
approved migration baseline. Do not publish secrets or private event data
in the acceptance record.

## Disposable-root acceptance, operator only

Use a copied example in a new disposable root/store. Never reuse the live
store or live event delivery destination. Configure an installed agent
only when deliberately testing step execution. The fixture's zero pool
capacity is intentional.

Pin both `--root "$DISPOSABLE_ROOT"` and
`--socket "$DISPOSABLE_ROOT/admin.sock"` on every controller-client
command. A root flag alone does not override an ambient `JOBSHOP_SOCKET`.
Verify the socket belongs to the disposable controller before creating or
closing any run. Stop only that controller when done.

1. Check and resolve before starting the disposable controller. Exercise
   the workflow start callback, one step completion, a structured output,
   invalid output and an unknown action reference. Check the recorded
   transition, error context and retry/escalation result. Use
   `automation list --run` followed by `automation show` to verify that the
   same invocation ID exposes attempt inputs, planned actions and resulting
   ticket targets. Do not count help parsing alone as audit acceptance.
2. Deliver one custom payload, a real GitHub-shaped wakeup and a timer
   wakeup. Verify the payload is not invented/enriched and repeated
   delivery preserves the same logical run/dedup identity. Verify existing
   cooldown and open caps against multiple subjects.
3. Invoke HTTP/shell/Jev without grants. Each must fail clearly. For a
   separately authorized local test endpoint, grant only what is needed.
   Check status handling, invalid JSON, stdout/stderr/nonzero exit,
   timeout, cancellation and output/execution ceilings.
4. Test origin-bound authentication with two different local origins.
   Verify a token reaches only its configured origin, authenticated
   cross-origin redirects fail, and caller headers do not cross an
   unauthenticated redirect. Keep keys out of tracebacks and print output.
5. Enable Jev with the actual Decisions endpoint and a secret reference.
   Check an allowed selection/reason, disabled behavior, invalid option,
   timeout/default behavior and no key in persisted inputs/diagnostics.
6. Reload a valid imported CUE change, a new same-package `.cue` file,
   an imported `.star` helper and a prompt. Break each in turn. The active
   configuration must remain the last valid revision. Repair and verify
   the error clears and watches follow the new imports.
7. Stop/restart only the disposable controller around a pending callback,
   accepted action and completed step. Re-deliver the event and verify
   deduplication, stale-state rejection and no extra transition. Reconcile
   any external call whose result could have been lost. A restart does not
   establish exactly-once shell/HTTP/model effects.
8. Compare before/after resolved static config. For jobcity flows complete
   the migration inventory, including all six review wakeups, 15-minute
   authoritative reconciliation, delegated Linear intake, stuck-work,
   improvement and learnings mappings. One successful demo does not certify
   those production policies.
9. Exercise rollback using the same disposable store and original run IDs.
   Restore the old selector/pack revision without resetting dedup, cursor
   or approval state. If an older binary cannot safely interpret new runs,
   verify the drain procedure instead of forcing a downgrade.

Record binary/config revisions, fixture tests, real event/action results,
old and new run IDs, failure diagnostics and remaining acceptance gaps.
These operator checks describe required evidence. They are not a claim
that this PR deployed or migrated jobcity.
