# Desk e2e preflight (JOB-213)

2026-11-14. Desk PRs broke e2e specs their authors did not run (message
headers, turn phase, graph cards). `npm --prefix ui run e2e:affected` maps the
files a branch changed to the specs that cover them and runs just those on an
isolated controller.

- `ui/e2e/affected.mjs` takes the diff against the merge base with
  `origin/master` plus uncommitted and untracked files (`--base REF` changes
  it, `--files path...` gives repo-relative paths). Without `--run` it only
  lists the specs. With it, it calls `run-isolated.mjs <specs>`, so it needs
  `JOBSHOP_BIN` and `FAKE_AGENT_BIN` like that script.
- The table in `affected.mjs` is regexes over source paths. Shared code (CSS,
  `App`, live state, `lib/api`, the build files, `packages/jobshop-client`, the
  Go Desk handlers) runs every spec. A `ui/src`, `kit` or `e2e` file with no
  rule also runs every spec.
- `affected.test.mjs` (in `npm run test:unit`, so in the `ui` CI job) fails for
  a source file nothing maps and for a spec the table names that
  `groups.mjs` does not list. That is the CI check on sample diffs; the CI
  runner has no merge base to diff.
- `run-isolated.mjs` now takes spec names, and `trace` for the trace group.
  `groups.mjs` lists them. The trace group (desk, live, search, shell,
  title-links, trace-axis, trace-detail, trace-whole) was never in CI, which is
  how graph-card changes broke specs unseen. It runs as a second `e2e` step on
  its own controller: on a shared one, `pool-controls` and `turn-phase`'s
  mobile cases timed out on RPCs.
- Not run by either: chat-dismiss, chat-responsiveness, home-sections,
  run-links and trace-session failed in a trial run on the isolated controller
  (on a loaded host, with timeouts on chat and run RPCs; not diagnosed), and
  attachments and chat need a real agent. A change to them is not covered by
  the preflight.
