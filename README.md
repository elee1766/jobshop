# jobshop

multi agent orchestration on a sqlite ticket store. replaces gas city for how we actually use it.

## what it does

- tickets in sqlite, every write goes through the controller
- pools of agents (acp, e.g. `jcode acp`) claim tickets, work them, close them
- one detached host process per session, survives controller restarts
- triggers make tickets from store conditions, cron, or a check command
- packs bundle pools, prompts and triggers. `packs/selfdev` is the one that builds jobshop itself

## run

```
nix build .#jobshop
jobshop serve            # controller, in a dir with jobshop.yaml
jobshop ticket create "do the thing" --pool dev
jobshop session list
jobshop session attach <id>
```

config is `jobshop.yaml`, see `packs/selfdev/README.md` for an example.

## dev

```
go build ./... && go vet ./... && go test ./...
```

if you touch go.mod update `vendorHash` in `nix/package.nix`. see AGENTS.md.

design and decisions live in `docs/`.
