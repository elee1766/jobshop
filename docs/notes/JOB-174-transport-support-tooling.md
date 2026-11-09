# JOB-174: transport adapters, support primitives and tooling

2026-11-09. Child of JOB-164, after JOB-165 (map) and JOB-166 (domain). One PR
moves ten packages and removes the last two `control-is-top` exceptions.

## Moves

| From | To |
|---|---|
| `pkg/rpc`, `pkg/rpc/rpcschema` | `pkg/transport/rpc`, `pkg/transport/rpc/rpcschema` |
| `pkg/desk` | `pkg/transport/desk` |
| `pkg/cli` | `pkg/transport/cli` |
| `pkg/gateway` | `pkg/transport/gateway` |
| `pkg/clock`, `pkg/redact`, `pkg/logctx`, `pkg/telemetry` | `pkg/support/{clock,redact,logctx,telemetry}` |
| `pkg/runbook`, `pkg/flaky` | `pkg/tooling/{runbook,flaky}` |

`transport`, `support` and `tooling` are directories, not Go packages. There is
no `support` API, container or helper package: each child stays a separate leaf.
No wrapper or forwarding package was left at an old path.

## Decisions

- `logctx` stays a package. Three callers in two families (control, cli, and
  telemetry, which reads what they set), so it earns its own leaf.
- `runbook` and `flaky` are not folded into the CLI. Each has one caller, but
  `flaky` also backs `cmd/ci-flaky` and both are tested alone with fixtures.
  Folding them would put a markdown parser and a `go test -json` classifier
  into the largest package for no dependency gain. They stay leaf tools.
- The CLI no longer imports `control`. `cli.Main(args, run)` takes a
  `RunController` and `jobshop serve` calls it with the flags, the process
  telemetry and the version. `cmd/jobshop` builds it from `control.Run` and adds
  `sessionrt.DefaultShellProbe`. A binary that passes nil gets an error from
  `serve`, which suits tests that only run client commands. `cmd/jobshop` is the
  one new exception in `control-is-top`: the composition root.
- Desk and CLI tests no longer build `control.Options`. `controltest.StartWith`
  and `controltest.WithVersion` start the controller for an adapter test; the
  CLI's two `control.Run` tests now use `controltest`, which also drops their
  retry loops (the harness waits for the socket).
- Transport schemas stay in `rpc`. Shared business types are already in `domain`
  (JOB-166). `rpc` still imports `store` for a few result shapes and acphost
  `cgroup`, `sandbox` and `userns` for session spec wire fields. Those are
  JOB-173 and JOB-167 questions and are not changed here.

## Rules added (`pkg/tooling/archcheck`)

- `rpc-imports-no-adapter`, `desk-not-from-cli-or-gateway`,
  `gateway-not-from-cli-or-desk`: the adapters are siblings. The CLI hosts the
  `desk` and `gateway` commands, so it is the one package allowed to import them.
- `transport-reaches-no-runtime` (closure): rpc, Desk and the gateway reach
  neither `control`, `sessionrt`, `hooks`, a store backend, a database driver
  nor the session host.
- `telemetry-takes-only-logctx`, and `runbook` and `flaky` added to
  `independent-leaves`.
- The `control-is-top` exceptions for `cli` and `desk` are gone. Nothing in
  `known.go` names JOB-174 any more.

## Path audit

Updated for the moves: `go:generate` in `pkg/transport/cli`, `deskgen` and its
header in `packages/jobshop-client/rpc.d.ts`, `ui/vite.config.ts` (`outDir`
`../pkg/transport/desk/dist`), `.gitignore`, `.github/workflows/ci.yml` and
`release.yml` (artifact paths, `ci-flaky`, the `-X ...cli.version` ldflags),
`nix/package.nix` (fileset difference, ldflags, the go2nix `packageOverrides`
key, the `deskUi` copy), `llm/` and the verify runbooks. Tests that read repo
files by relative path (`llm/cli.md`, `docs/verify`, `deploy/grafana`,
`rpc.d.ts`) got one more `..`. The `-X` flags matter: a wrong package path
makes `jobshop version` print `dev (unknown)` without failing the build.
Older `docs/notes/*` keep the paths they were written with.

## Compatibility

No flag, command, JSON field, config key or stored value changed. `llm/cli.md`
differs only in the `go generate` path. `go run ./pkg/transport/rpc/rpcschema/cmd/deskgen`
produces the same declarations apart from the header comment. No `go.mod`,
`go.sum` or `vendorHash` change.
