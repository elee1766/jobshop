# Desk builds without the UI

2026-11-10.

`pkg/transport/desk` embeds `dist`, which Vite generates and Git ignores.
Before this, a checkout with no UI build did not compile: the embed
pattern matched nothing, so the package and everything importing it
(including `cmd/jobshop`) failed, and Go-only changes needed a UI build.

- `dist/.keep` is tracked (the rest of `dist` stays ignored), so the
  embed always matches. `ui/public/.keep` is the same file: Vite empties
  `dist` on build and copies `public` back in, which keeps the tracked
  file in place and the tree clean after a build.
- A binary built without the UI serves a short "Desk UI is not built"
  page with the build command instead of the app. The API routes work as
  before. `jobshop desk` logs a warning at start. `desk.Built()` reports
  which case the binary is.
- Tests that read the bundle (`TestStatic`, `TestBasePath`, `TestFavicon`,
  `TestStaticCompression`) skip without it. CI sets
  `JOBSHOP_REQUIRE_DESK_UI=1`, which turns the skip into a failure, so
  they still run on every PR.
- Release builds check `dist/index.html` exists before `go build`. The
  Nix build always copies the built UI in.
