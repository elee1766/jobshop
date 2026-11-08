# Desk assets are generated at build time

2026-11-08. PR #1051. Decision: #569. Supersedes the committed-output build
instructions in the older Desk notes.

`pkg/desk/dist` is ignored output. UI source and public assets remain in `ui/`.
Source builds run `npm --prefix ui ci` and `npm --prefix ui run build` before Go.
Plain `go install ...@version` cannot generate these assets; use a release binary,
Nix, or build from a checkout and run `go install ./cmd/jobshop`.

Nix builds `desk-ui` separately from UI and shared-client sources with a pinned
npm dependency hash. Its output is copied into the Go source before package
discovery, so both the go2nix plugin and fallback builder see the embedded files.
Local generated output is excluded from this source. Go-only edits do not
invalidate the frontend derivation.

CI passes one frontend artifact to Go checks. Releases build the frontend from
the release tag and pass it to both architecture builds. The release uploader
selects only binary artifacts. Binaries still embed Desk and require neither
Node nor external UI files at runtime.

Acceptance: `docs/verify/generated-desk-assets.md`.
