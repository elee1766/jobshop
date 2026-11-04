# Jobshop: repo layout and build

Companion to `plan.md`.

## Principles

- One binary, `jobshop`, with subcommands. The controller, session host,
  CLI and (later) gateway are the same executable, so the host can be
  spawned from `/proc/self/exe` and there is one thing to build, pin and
  roll back.
- `pkg/` for all packages (decided 2026-11-01). Importable, but no API
  stability promise until something outside this repo depends on it.
- Packages follow the architecture boundaries in `plan.md`: store, control
  (controller), session host, rpc (sockets), config, telemetry. Packages
  import downward only; `cmd/` wires them together.
- No role names in Go. Agents, pools and prompts live in config.
- Plain `go` for development and CI. Nix builds the deployable package and
  the NixOS module.

## Tree

```
jobshop/
├── cmd/jobshop/              thin shim: calls cli.Main(os.Args) and exits
├── pkg/
│   ├── store/                Store interface, Ticket/Edge/Comment types, errors,
│   │   │                     engine kinds and their typed field structs
│   │   ├── storetest/        conformance suite every backend runs
│   │   └── sqlite/           modernc.org/sqlite backend
│   │       ├── migrations/   goose *.sql files, embedded
│   │       └── queries/      sqlc *.sql queries and generated Go
│   ├── rpc/                  JSON-RPC method names, request/response types,
│   │                         client over the acp-go-sdk connection
│   ├── control/              controller: tick loop, claims release, lease sweep,
│   │                         session supervision, admin and CLI socket handlers
│   ├── config/               YAML loading, validation, defaults
│   ├── hostspec/             jobshop's spec.json on acphost: conv_export
│   │                         field and the conversation exporter option
│   ├── gateway/              chat gateway: adapter interface, routing, Matrix, Discord
│   ├── desk/                 Jobshop Desk HTTP server over the admin API;
│   │                         dist/ is the built UI, embedded
│   ├── telemetry/            OTel providers, slog bridge, Prometheus exporter,
│   │                         conversation export
│   ├── cli/                  CLI subcommands (ticket, claim, assign, session, ...)
│   ├── cron/                 5-field cron parser for triggers (no dependency)
│   └── ids/                  ticket ids: per-project prefix + sonyflake base32
├── docs/                     plan.md, data-model.md, layout.md
├── ui/                       Desk UI source (Svelte 5, Vite); builds into pkg/desk/dist
├── examples/                 sample jobshop.yaml, agent prompts
├── nix/
│   ├── package.nix           go2nix build, buildGo127Module fallback
│   └── module.nix            NixOS module: user, slice, service
├── flake.nix
├── flake.lock
├── go.mod
├── go.sum
└── .github/workflows/ci.yml
```

Dependency direction (an arrow means "may import"):

```mermaid
flowchart TD
  cmd[cmd/jobshop] --> cli
  cli --> control & acphost & hostspec
  cli --> rpc & config
  config --> cron
  control --> store & rpc & config & telemetry & acp & acphost & hostspec & cron
  hostspec --> acphost & telemetry
  sqlite[store/sqlite] --> store
  storetest[store/storetest] --> store
  rpc --> store
```

`acphost` never imports jobshop: it has no ticket or policy logic
(plan section 8.3), which keeps it small and rarely changing.

The session host is the external module `github.com/elee1766/acphost`
(issue #256, 2026-11-04), extracted from the old `pkg/host` so other
projects can run ACP agents under a detachable host. It has no ticket or
policy logic (plan section 8.3). Conversation export is not part of its
spec: `pkg/hostspec` keeps the `conv_export` field in spec.json and hands
`acphost.Main` an exporter built on `telemetry.ConvExporter`. Default
method and GC root prefixes (`_jobshop/`, `jobshop-host-`) stay, so hosts
started by older builds are adopted.

Subcommands:

| Command | Runs |
|---|---|
| `jobshop serve` | controller (long-running) |
| `jobshop session-host <dir>` | session host, spawned by the controller only |
| `jobshop ticket ...`, `claim`, `assign`, `comment`, ... | CLI over a socket |
| `jobshop session list/attach/reap` | CLI over the admin socket |
| `jobshop migrate` | apply store migrations (also done on `serve` start) |

## Dependencies (decided 2026-11-01)

| Need | Module | Version at decision |
|---|---|---|
| SQLite | `modernc.org/sqlite` | v1.61.1 |
| Migrations | `github.com/pressly/goose/v3` (embedded `*.sql`) | v3.29.0 |
| JSON-RPC (sockets) and ACP | `github.com/coder/acp-go-sdk` (its `Connection` for the CLI/admin sockets; its types and agent/client sides for ACP) | v0.13.5 |
| CLI | `github.com/alecthomas/kong` | v1.16.1 |
| YAML config | `github.com/goccy/go-yaml` | v1.19.2 |
| IDs | `github.com/sony/sonyflake/v2`, Crockford base32 | v2.2.0 |
| Tracing, logs, metrics | `go.opentelemetry.io/otel` + sdk, OTLP HTTP exporters | v1.47.0 |
| Prometheus | `go.opentelemetry.io/otel/exporters/prometheus` | v0.69.0 |
| slog to OTel | `go.opentelemetry.io/contrib/bridges/otelslog` | v0.21.0 |
| Tests | `github.com/stretchr/testify` | v1.12.1 |
| Typed SQL (build-time codegen) | `github.com/sqlc-dev/sqlc` | v1.31.1 |
| Linux syscalls | `golang.org/x/sys/unix` | v0.49.0 |
| `/proc` parsing (PID start time) | `github.com/prometheus/procfs` | v0.23.0 |
| Matrix client-server API (gateway) | `maunium.net/go/mautrix` (2026-11-02, #123): the maintained Go Matrix client, used by the mautrix bridges; pure Go without its optional crypto, so `CGO_ENABLED=0` holds | v0.32.0 |
| Discord bot API (gateway) | `github.com/bwmarrin/discordgo` (2026-11-03, #182): the most used Go Discord library, REST plus gateway websocket with resume; pure Go | v0.29.0 |
| Goroutine groups | `golang.org/x/sync/errgroup` | v0.24.0 |
| Secret redaction (conversation export) | `github.com/zricethezav/gitleaks/v8` (`detect`) | v8.30.1 |
| Desk UI (build-time only, `ui/`) | Svelte 5, Vite, svelte-check, Playwright (2026-11-03, #202): Gas City Desk's stack; output is committed and embedded, so the Go module gains no dependency. @lucide/svelte for the activity view's icons (2026-11-04, #255); tree-shaken per icon | svelte 5.57.3, vite 8.3.5, @lucide/svelte 1.67.0 |
| Lint (tool, not a module) | golangci-lint, built by `nix/golangci-lint.nix` with Go 1.27 because nixpkgs builds it with Go 1.26, which refuses go 1.27 modules (2026-11-03). Same version in CI, the dev shell and agents' PATH | v2.15.0 |

Uses:
- **sqlc**: queries live as plain `.sql` files in `pkg/store/sqlite/queries/`
  and sqlc generates typed Go next to them. It is a build-time tool, run
  via `go tool sqlc` (Go 1.24+ `tool` directive in `go.mod`), so the
  version is pinned and nothing extra lands in the binary. Generated code
  is committed; CI checks it is up to date (`sqlc diff`). goose
  migrations are the schema source sqlc reads. Supports Postgres for the
  later backend.
- **x/sys/unix**: setsid, `Pdeathsig`, `PR_SET_CHILD_SUBREAPER`,
  `SO_PEERCRED` on sockets, flock for the controller lock.
- **procfs**: PID start time so leak reaping never kills a reused PID.
  Already an indirect dependency of the Prometheus exporter.
- **Prometheus exporter** (added 2026-11-01, issue #14): brings
  `github.com/prometheus/client_golang` for its registry and `promhttp`
  handler. `pkg/telemetry` uses a private registry per controller rather
  than the global one, so tests and restarts do not collide.
- **OTLP HTTP exporters and otelslog** (added 2026-11-01, issue #15):
  `otlptracehttp`, `otlploghttp` (v0.23.0) and `sdk/log` for
  `telemetry.otlp_endpoint`; the bridge sends controller logs with trace
  ids. They pull in `grpc-gateway` and `otlp` protos (types only, HTTP
  transport). Only imported by `pkg/telemetry`.
- **errgroup**: controller and host goroutine lifetimes and shutdown.
- **gitleaks**: its detector and default rule set redact secrets in
  conversation export before records are queued, plus launch-env secret
  values (exact match). Large dependency (about 74 modules in its
  `go.mod`); accepted for rule coverage and maintenance (decided
  2026-11-01). Only imported by `pkg/telemetry`.

Standard library by choice (decided 2026-11-01):
- `log/slog` for logging, `net/http` for `/metrics`, `text/tabwriter`
  for CLI tables, `encoding/base32` for IDs.
- `encoding/json/v2` for JSON (available in Go 1.27 without an
  experiment flag; verified 2026-11-01).
- `testing/synctest` for time-dependent tests (leases, ticks); no fake
  clock library.
- Small local helpers for retry/backoff.
- `os/exec` running `git`; no go-git. Agents use `gh`; the controller does
  not call GitHub.
- Host journal is hand-rolled (sequence-numbered append file plus index),
  so fsync and backpressure stay under our control.
- Kind field validation is `Validate()` methods on the Go structs; no
  validation-tag library.
- Config reload is explicit (`jobshop reload`); no file watcher.

Notes:
- JSON-RPC library (decided 2026-11-01, replacing `go.lsp.dev/jsonrpc2`):
  `coder/acp-go-sdk` for both ACP and the CLI/admin sockets. Its
  `Connection` is a general NDJSON JSON-RPC peer (method name plus raw
  params in, `*RequestError` out) with a normal per-request context, and
  it has no dependencies outside the standard library. Pre-1.0, so pin
  the version and read the changelog on upgrades. `go.lsp.dev/jsonrpc2`
  v1.0.1 was dropped: its server helper hardcodes LSP header framing, and
  its pooled per-request context panics (`cannot create context from nil
  parent`) when anything, such as `database/sql`'s `Rows.awaitDone`,
  calls `Done()` after the handler returns. Unfixed on its main branch as
  of 2026-11-01.
- Ticket IDs (decided 2026-11-01): `<project prefix>-<13 chars>`, e.g.
  `js-0031w02j0r012`. A Sonyflake 63-bit ID, big-endian, encoded in
  lowercase Crockford base32 (no `i l o u`). Fixed width, so IDs sort by
  creation time as strings. Settings: 1 ms time unit, 47 time bits
  (about 4,400 years from a 2026-01-01 epoch), 12 sequence bits (4,096
  IDs per ms), 4 machine-id bits fixed to 1. Only the controller
  generates IDs and the controller lock guarantees one writer, so the
  machine id never needs to vary. Chosen over xid (20 chars) for length;
  uniqueness is the same for a single writer.
- Cron (decided 2026-11-01, issue #26): `pkg/cron` is a small 5-field
  parser (about 150 lines), not `robfig/cron/v3`. Triggers need only
  `Parse` and `Next`. The controller tick does the scheduling, so the
  library's scheduler, seconds field and descriptors would be unused.
  Supported: `*`, numbers, ranges, lists, steps, day of week 0-7, and the
  Vixie rule that restricted day-of-month and day-of-week match if either
  does. Names (`MON`) and macros (`@daily`) are not supported.
- New direct dependencies need a reason recorded here.

## Build

**Development and CI: plain Go.**

```
go build ./...
go vet ./...
golangci-lint run
go test -race ./...
```

CI runs those on GitHub Actions, plus `nix build` to catch packaging
breakage. No Makefile unless a real need appears.

**Deployment: Nix flake.**

- Inputs: `nixpkgs` pinned to `nixos-26.05`, the same channel gasserver
  runs. It ships `buildGo127Module` (Go 1.27.3), matching `go.mod`.
- `packages.x86_64-linux.jobshop`: go2nix (per-package derivations, see
  `docs/notes/go2nix.md`) when its plugin is loaded, else
  `buildGo127Module`; `cmd/jobshop`, `CGO_ENABLED = 0` (pure-Go SQLite), version and
  commit via `-ldflags -X`, `vendorHash` pinned. Tests run in CI, not in
  the Nix build (`doCheck = false`), same as Gas City's flake.
- `nixosModules.jobshop`: the deployment unit (below).
- `devShells.default`: go, gopls, golangci-lint, sqlite.

## Root directory (decided 2026-11-01)

A Jobshop deployment is one directory, the root. The server and CLI find
it from `--root` / `JOBSHOP_ROOT`, else the nearest ancestor of the cwd
that contains `jobshop.yaml` (like git and `.git`). Every path derives
from the root; there are no separate config, data or socket path flags.

```
<root>/
├── jobshop.yaml          config; marks the directory as a root
├── jobshop.db            SQLite store (+ -wal, -shm)
├── controller.lock       flock held by the running controller
├── admin.sock            admin socket, 0700 dir keeps it private
└── sessions/<id>/        spec.json, host.{lock,pid,log}, acp.sock, cli.sock,
                          journal.ndjson, exit.json, work/ (git worktree)
```

Copying or backing up the directory moves the whole deployment, and
several roots can run side by side for testing. The NixOS module uses
`dataDir` as the root and links the generated config into it as
`jobshop.yaml`. Agents get `JOBSHOP_SOCKET` (their session socket), not
the root.

## Deployment on gasserver (decided 2026-11-01)

gasserver is an isolated OVH machine; Jobshop runs there next to Gas City
as its own user, not as `agent`.

- **User `jobshop`**: own home, root directory `/var/lib/jobshop`. GitHub
  access is the elee-agent App through agent-credd tokens (plan.md
  section 6).
- **Resource limit via its own slice** (decided 2026-11-01): `jobshop.slice`
  with `CPUQuota = 600%` (six cores' worth, placed by the scheduler) and
  `MemoryMax = 8G` (plus a lower `MemoryHigh`), `TasksMax`, and oomd
  pressure kill, mirroring how `modules/resources.nix` caps
  `user-1001.slice` for `agent`.
- **Network**: same egress policy as `agent` (public internet and
  loopback; no private, CGNAT/tailnet or link-local ranges), copied from
  `modules/agent/default.nix`.
- **Service**: `jobshop.service` runs `jobshop serve` as `jobshop` in
  `jobshop.slice`, with `KillMode=process` so restarting the controller
  does not kill session hosts (they are children of the controller but
  must outlive it, plan section 8.2). Session hosts started with `setsid`
  stay in the service's cgroup, so they inherit the slice limits.
- **Binary pinning**: the NixOS module points at a specific `jobshop`
  store path. Promote = bump the flake input on gasserver and switch;
  rollback = previous generation. This replaces the symlink scheme in
  `docs/plan.md`.
- **Where the module lives**: `nixosModules.jobshop` in this repo, imported
  by `gasserver.nix` as a flake input. Host-specific values (limits,
  secrets) stay in `gasserver.nix`.
- **Workspaces** (2026-11-02, #92): each repo the agents work on is
  `services.jobshop.workspaces.<name>`; a second repo needs only host
  values:

  ```nix
  services.jobshop.workspaces = {
    jobshop = {
      url = "https://github.com/elee1766/jobshop.git";
      livePacks = [ "packs/selfdev" ];
      livePackVars = { repo = "/var/lib/jobshop/workspaces/jobshop"; };
      selfDeploy = true; # the repo self-deploy follows; at most one
    };
    other = {
      url = "https://github.com/elee1766/other.git";
      livePacks = [ "packs/github" ];
    };
  };
  ```

  The old `services.jobshop.workspace` still works and means
  `workspaces.default`.
- **Restart vs full stop** (decided 2026-11-01): `systemctl restart` and a
  Nix switch restart only the controller; hosts keep running and are
  re-adopted. A full stop cleans up session hosts too. systemd's ExecStop
  cannot tell a restart from a stop, so full stop is an explicit command,
  `jobshop shutdown`, which asks the controller over the admin socket to
  end every session (stop agents, flush journals, release claims, reap
  hosts) and then exit. The NixOS module wires it to system shutdown.
- **Old builds** (decided 2026-11-01): a host keeps running the store path
  it started from. Each running host registers a Nix GC root for its
  binary (under `/nix/var/nix/gcroots/per-user/jobshop/`), removed when
  the host exits, so garbage collection never deletes a build a live
  session uses.
