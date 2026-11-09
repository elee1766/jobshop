# Jobshop for agents

For an agent operating a Jobshop root, writing a pack, or changing this
repo. This page links to the docs rather than copying them. Humans start
at the [README](../README.md).

## Changing this repo

1. [AGENTS.md](../AGENTS.md): rules, the check command, docs policy.
2. [docs/plan.md](../docs/plan.md) sections 3 (architecture) and 6
   (decisions).
3. [docs/layout.md](../docs/layout.md): packages and dependencies.
4. Generate Desk assets: `npm --prefix ui ci && npm --prefix ui run build`.
   Then check: `go build ./... && go vet ./... && go test ./... && golangci-lint run ./...`.
   After changing a command's help, run `go generate ./pkg/transport/cli` to
   refresh [cli.md](cli.md).

## Writing or editing a pack

- [docs/configuration](../docs/configuration/README.md) covers CUE packs, Starlark hooks and migration.
- [docs/packs.md](../docs/packs.md) covers legacy YAML packs.
- [packs/github](../packs/github/README.md) is the worked example.
- `jobshop config show` prints the resolved config with the source of
  each entry. `jobshop reload` applies edits and keeps the old config if
  the new one is invalid.

## Working as a session inside Jobshop

The `jobshop` CLI in a session talks to that session's own `cli.sock`
and acts as the session: its ticket, pool and project. From there you
can:

- read and update your ticket: `jobshop ticket show`, `jobshop comment`,
  `jobshop close`, `jobshop release [--until]`;
- report state: `jobshop report progress|blocked|waiting|failed`;
- deliver a report to Desk: `jobshop artifact deliver`;
- search past work: `jobshop recall`, and record lessons with
  `jobshop learn`;
- get a short-lived scoped token: `jobshop identity token <owner/repo>`;
- use trackers as the project's identity: `jobshop gh ...`,
  `jobshop linear ...`.

Out of reach: the admin socket, `jobshop.db`, `secrets/` and other
sessions. See [docs/isolation.md](../docs/isolation.md). Every command
and flag: `jobshop --help` or [cli.md](cli.md).

## Operating a root

With the admin socket:

- `jobshop attention`: what waits on a human.
- `jobshop run show <run>`, `jobshop run resume <run> --step <id>`.
- `jobshop session list`, `jobshop session attach <id> [--rw]`.
- `jobshop ticket update <id> --defer <time>`.
- `jobshop reload` after editing yaml or packs.

More in [docs/running.md](../docs/running.md).
