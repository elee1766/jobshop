# Memory moved out of the repo (#592)

2026-11-05. Operator decision: agent memory stays on the Jobshop host,
lasting guidance goes in pack prompts.

- `memory/` is gone from the repo. Its lines moved to prompts: generic
  process lines (PR checkout by `pull/<n>/head`, CI waits only after the
  pushed sha shows, `comment -t`, `close -r failed` takes no text, no
  `--body-file -`, umbrella issues get no run, `run resume --step`,
  stale `mergeStateStatus` and `merge-tree`) to
  `packs/github/prompts/github/{dev,implement,merge,plan}.md` and
  `stall.md`; Jobshop lines (sqlc via `go generate`, migration
  renumbering, deskgen after rebase, golden, pack_test, bwrap/pasta PATH,
  npm via nix shell, `-race` under load) to
  `packs/selfdev/prompts/guide.md`, rule `selfdev-guide` on
  elee1766/jobshop turns. Lines about plan.md conflicts were dropped.
- `memory_dir` defaults to `memory`, relative to the root dir, so
  `<root>/memory`, in selfdev and operator. `""` still turns it off. The
  NixOS module creates `${dataDir}/memory` for the service user.
- `jobshop memory write <scope> [--learnings ids] [--from-file f]`
  (`memory.write`, admin or pool grant `memory-write`; the learn pool has
  it): scope `global`, `repo/<owner>__<name>`, `pool/<name>` or
  `skills/<name>`. It refuses a memory file over the cap, writes the
  file, records a `memory_edits` row with `path`, `before` and
  `existed` (pr empty), measures the baseline then, and marks the cited
  learnings applied with pr `memory:<edit id>`.
- `learn evaluate` restores `before` on `not_helped` (removes the file
  when the write created it) and marks the learnings rejected. A newer
  write to the same path is kept; the old edit is only marked.
- `jobshop memory show [scope]` and `jobshop memory log` (= `learn
  edits`) list files and edits. Migration sqlite 00018, postgres 00003:
  `pr` may be empty, unique when set.
- Prompt fixes stay PRs through `pr-work` (`curate.md`, `evaluate.md`).
- Tests: `TestMemoryWrite*` (pkg/control), `pkg/memory` write tests,
  `TestRuleMemory`, pack tests. Live: `docs/verify/memory-skills.md`.
