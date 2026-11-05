# go2nix build

2026-11-04. `nix/package.nix`, `flake.nix`, CI `nix` job.

- buildGoModule compiled the whole binary in one derivation, so CI
  rebuilt every package on any change (about 90s).
- go2nix (numtide, flake input) compiles each Go package in its own
  derivation. Locally a one-file edit in `pkg/ids` rebuilt in 3.5s; a
  cold build was about 60s. Hestia caches each package derivation in CI.
- It needs its Nix plugin in the evaluator (`plugin-files`). The plugin
  runs `go list` with the Go it was built with, so the flake exports
  `go2nix-plugin`, built with `go_1_27`; upstream's uses Go 1.26 and
  refuses go.mod's `go 1.27.3`. Plugin ABI: Nix 2.34, so CI installs
  Nix 2.34.9.
- No lockfile: module hashes come from `go.sum` at evaluation.
- Without the plugin (`builtins ? resolveGoPackages` false), the package
  falls back to `buildGo127Module` with the pinned `vendorHash`, so
  gasserver and users without the plugin still build. CI builds the
  fallback's `goModules` to keep the pin honest. The two builds give
  different store paths.
- CI runs on GitHub-hosted `ubuntu-latest` (free for public repos), switched
  back from Ubicloud on 2026-11-05.
