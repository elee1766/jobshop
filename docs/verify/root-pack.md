# Verify a root pack

2026-11-07. JOB-122.

1. Run `jobshop init ./root-pack-check`. Confirm `jobshop.yaml` selects
   `pack: pack.yaml` and `pack.yaml` contains the project. A second init
   must fail without changing either file.
2. Put shared agents, resources and work definitions in the root pack.
   Move the old service `packs:` list into it without changing child vars
   or prefixes. Keep storage, credentials, listeners and runtime security
   defaults in the service file. Preserve each relative path's target if
   the root pack lives in a different directory.
3. Before changing a running root, run `jobshop config check --help` and
   check the proposed service file with `jobshop config check --file <file>`.
   Compare resolved projects, pools, workflows, resources and source paths
   with the old configuration. Resolve any mixed service/work warnings.
4. With `reload.watch: true` in the service file, edit a root-pack definition,
   then a child prompt. Check `jobshop config show` reflects each change.
   Add a child pack to the root pack, then edit that child's prompt to
   verify the watcher follows the new directory.
5. Make the root pack invalid. The controller must report invalid config
   while retaining its last working configuration. Restore the file and
   confirm the alert clears. A host-only setting such as `store:` in the
   root pack must be rejected.

Automated coverage: `go test ./pkg/config ./pkg/transport/cli` and
`go test ./pkg/control -run '^TestConfigWatch'`. These cover composition,
inheritance, namespacing, provenance, linked configs, deployment path
rewrites, startup fallback, initialization collisions and event-driven reload.
