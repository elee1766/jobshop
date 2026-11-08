# Desk learn view

2026-11-08. Run on an isolated controller (see `desk-redesign.md`). Never
seed a live instance.

1. Seed a throwaway root with learnings in each state: open, an open PR
   (`jobshop learn update <id> --pr <url>`), rejected (`--status
   rejected --pr <url>`), applied with `--merged-at`, and a long summary.
   Create runs of a workflow so steps exist, then `jobshop learn
   evaluate <edit id>` to get `helped` and `not_helped` edits.
2. Serve Desk from a build of the branch (the UI is embedded) and open
   `?view=improvements` at 1000 and 390 px, light and dark.
3. Expected: the summary line counts items and names `not helped` when
   there are any. Sections run not helped, fix proposed, evaluating,
   awaiting curator, then collapsed history. A not helped row shows the
   problem, `attempted fix <text>`, the gate's reason and a PR link. A
   learning with no fix reads `fix not recorded`. Details show the cited
   learnings and the before and after table. No horizontal overflow, long
   titles clamp to two lines.
4. `DESK_URL=<desk> npm --prefix ui run test:e2e -- learn.spec.js desk.spec.js`
   passes on both projects, also with `DESK_PREFIX=1` (Desk with
   `--trusted-proxy 127.0.0.1`).
