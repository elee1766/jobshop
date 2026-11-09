# JOB-166: shared domain package

2026-11-09. Second child of JOB-164, after the JOB-165 baseline
(`docs/notes/JOB-165-package-boundaries.md`). Nothing user visible changes.

## What moved

`pkg/domain` holds the types and invariants more than one family needs. It
imports the standard library only. Every move below is a whole type or file,
with its validation, so a rule travels with the type it checks.

| File | From | Holds |
|---|---|---|
| `ticket.go` | store | `Ticket`, `Edge`, `Comment`, `Status`, `EdgeKind`, `OpenStatuses`, the `Err*` values |
| `kinds.go` | store | `Kind`, the typed fields of each kind (`TaskFields`, `SessionFields`, `MessageFields`, `RunFields`, `StepFields`, `BatchFields`), run and step vocabulary, `NormalizeFields`, `DecodeFields`, `StepFailure`, `StepExhausted`, `ReviewCapped` |
| `wait.go` | store | `Wait`, `WaitCond`, their validation and `WaitOf` |
| `learning.go` | store | `Learning`, its kinds, sources, statuses, scope parsing, kind patterns and `Glob` |
| `memedit.go` | store (`gate.go`) | `MemoryEdit`, `PoolMetrics` and the edit states |
| `resource.go` | store | `ResourceUse`, `ResourceWait`, `Slot`, `SlotDemand`, `SlotEvent`, `ResourceReport`, `ResourceConfig` |
| `inbox.go` | store | `InboxItem` and the inbox states |
| `meta.go` | store | `MetaString`, `MetaMap`, the built-in meta pack names |
| `guard.go`, `question.go` | rpc | `GuardAllow`, `GuardDeny`, `GuardAskHuman`, `QuestionOption`, `MaxQuestionOptions` |

## What stayed, and why

- `store`: the `Store` interface, `Filter`, `Update`, `Graph`, `Outcome`,
  `ClaimNext`, `LearningFilter`, `InboxFilter`, `InboxUpdate`, `RecallQuery`,
  `RecallHit`, `Thread` and `Parts`, collections (`Record`, `CollectionSpec`,
  `Collection`), `PollCursor`, `WriteKey`, the tracing wrapper and the helpers
  over `List`, `Apply` and `Update`. They describe how to ask a backend, not
  what a ticket is, and a collection spec is SQL serialization.
- `rpc`: every `*Params` and `*Result` envelope, method names, event types,
  `GuardParams` and `GuardResult` (the wire shape of a guard call), the types
  that wrap `acphost` cgroup, sandbox and user values, and `ToolTarget`.
- `config` types, provider values, session runtime types, and the `ids`
  generator. `ids` has three callers that need only the generator and its
  shape check (`IsSession`, `SessionTag`), none needs it for a type, and
  domain must not take sonyflake, so no identifier kind moved.
- `Record` stays in `store`: `hooks` still uses collections for its own
  durable state, which is persistence, not a shared concept.

## Decisions worth knowing

- Error values moved with the types, and their messages keep the `store:`
  prefix. They reach users and logs, and `errors.Is` still matches across
  every layer because there is one value. `store` does not alias them: the
  old internal paths were removed, not forwarded.
- `rpc.AutomationResult` and `rpc.AutomationListParams` stay wire types. To
  take `hooks` off `rpc`, `hooks.Engine.List` takes a `hooks.ListFilter` and
  `Show` returns a `hooks.Audit`. The controller's `hookaudit.go` maps them to
  the same JSON as before. That removed the only reason `hooks` reached the ACP
  SDK, so the `command-paths-no-acp` exception for hooks is gone.
- Schema keys: `rpcschema` names a JSON schema def by the last element of the
  package path, so `GET /api/rpc` now reports `domain.Ticket` where it reported
  `store.Ticket`. The TypeScript client names types by bare name and
  `packages/jobshop-client/rpc.d.ts` is byte-identical. The one assertion on a
  def key (`pkg/desk/rpc_test.go`) was updated.
- No aliases or wrappers: about 370 files were rewritten from `store.X` and
  `rpc.X` to `domain.X` by an AST pass that edits selectors by offset, and
  imports were regrouped by the repo's formatter settings.

## Dependency evidence

Production imports of `store` or `rpc` removed by the move, from
`archcheck -inventory` before and after:

- `config` no longer reaches `store` or `rpc` at all, so loading configuration
  no longer links the SQL builders or the ACP connection.
- `control/question` imports `domain` only. `control/scaleplan` imports `config`
  and `domain`.
- `control/waitengine` no longer imports `store`.
- `hooks` no longer reaches `rpc`, so it no longer reaches the ACP SDK or the
  acphost cgroup, sandbox and userns packages through it.

New rules in `pkg/tooling/archcheck`: `domain-is-a-leaf` (no other package of
the module, test files included), the `domain` family (standard library only),
`types-not-from-store-or-rpc` and `hooks-not-from-rpc` (closure rules that fail
when those consumers reach `store` or `rpc` again). The `hooks -> rpc ->
acp-go-sdk` baseline exception was deleted.

## Compatibility checked

- JSON tags, stored vocabulary, error text, `NormalizeFields` strictness, wait
  validation and step failure counting are pinned by `pkg/domain/domain_test.go`
  against the values they had before the move.
- The store conformance suite is unchanged apart from the package name. The
  SQLite run passed locally. The Postgres run skips without a server, so it is
  CI's (`JOBSHOP_TEST_POSTGRES`, and the skip check in `ci.yml` fails a skip).
  No migration, no table, no stored JSON changed.
- `go run ./pkg/rpc/rpcschema/cmd/deskgen` produces no diff.
- No `go.mod` or `go.sum` change, and `nix build .#jobshop.goModules` is the
  same closure, so `vendorHash` is unchanged.

## Left for later children

`config -> store` is gone for types, but `config` still imports `automation`
(JOB-172). `rpc` still imports `store` for query shapes in a few results, and
imports acphost cgroup, sandbox and userns for session spec wire fields
(JOB-167, JOB-174). Other consumers (`control/attention`, `control/chatlife`,
`control/threads`, `journalship`) import `store` because they read tickets
through the interface, which is correct.
