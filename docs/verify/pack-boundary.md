# Pack boundary check (JOB-172)

Run in the source checkout. No host state is touched.

```sh
go run ./pkg/tooling/archcheck/cmd/archcheck
go list -deps ./pkg/pack/cue | grep gfx-labs/jobshop      # only pkg/pack/cue
go list -deps ./pkg/pack | grep gfx-labs/jobshop          # pack, pack/cue, pack/tmpl
go test ./pkg/pack/... ./pkg/config/... ./pkg/orchestration/hooks/...
go run ./cmd/jobshop config check --file pkg/config/testdata/cue-starlark-example/jobshop.yaml
```

Expect archcheck `ok`, no `pkg/config`, `pkg/store`, `pkg/automation` or
`pkg/rpc` in either dependency list, and the same `config check` output as
before the change. A deploy keeps `config.last-good.yaml` readable: the
snapshot format did not change.
