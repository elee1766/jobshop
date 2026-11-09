# Generated Desk assets

2026-11-08. Run from a clean checkout. No live controller or credentials needed.

## Source build

```sh
npm --prefix ui ci
npm --prefix ui run check
npm --prefix ui run build
go build ./...
go test ./pkg/transport/desk
node --test packages/jobshop-client/index.test.js ui/kit/links.test.js
git check-ignore pkg/transport/desk/dist/index.html
```

`pkg/transport/desk/dist` must contain the generated HTML, assets, fonts and public files,
but `git ls-files pkg/transport/desk/dist` must be empty. Building the frontend again must
not produce a Git diff. Plain `go install ...@version` is not a supported build:
it does not generate the embedded frontend. Use release binaries, Nix or the
source build above followed by `go install ./cmd/jobshop`.

## Nix builders

```sh
nix build .#desk-ui --out-link result-desk-ui
nix build .#jobshop --out-link result-fallback
./result-fallback/bin/jobshop version
nix build .#go2nix-plugin --out-link result-plugin
nix --option plugin-files "$PWD/result-plugin/lib/nix/plugins/libgo2nix_plugin.so" \
  build .#jobshop --out-link result-plugin-jobshop
./result-plugin-jobshop/bin/jobshop version
nix flake check --no-build
```

Run the fallback command without an ambient go2nix plugin. Both binaries must
serve Desk's HTML and referenced JS, CSS, fonts, manifest and service worker
when started outside the source checkout, without Node installed at runtime.
Use a disposable root and loopback Desk port, not a live instance.

The frontend derivation includes UI source and the shared client, not Go source
or local `dist` output. Go-only changes must not rebuild it. With the same Node
version, compare `result-desk-ui` against the locally generated `pkg/transport/desk/dist`.
`nix build .#desk-ui --rebuild` checks reproducibility. Lockfile changes require
updating `npmDepsHash` in `nix/desk-ui.nix`.

## CI and releases

The UI CI job uploads `desk-ui`; Go checks download it before building. Nix
checks build both fallback and plugin variants independently. Release frontend
output is built from the release tag and shared by both binary architectures.
Verify release archives contain working self-contained binaries, and only
`jobshop-*` binary artifacts reach the release uploader, not the frontend
artifact. The generated asset directory is never committed.
