# Curator (#273)

2026-11-04. Last item ("Curator") of target #247 B.

- Trigger condition `when.learnings: {min, every}`: holds when at least
  `min` open learnings were recorded after the trigger's last fire (a
  trigger that never fired counts every open one), or when `every`
  passed since the last fire and at least one was. Like other triggers,
  an open ticket or run it created holds it, so with a `run:` and a key
  only one run is open at a time.
- selfdev pack: pool `learn` (`learn_sessions`, default 1,
  `prompts/learn.md`), workflow `curate` (one step `curate` in `learn`,
  1 retry, escalates to `plan`, `prompts/curate.md`), trigger `curate`
  (`min: ${curate_min_learnings}` default 10, `every: ${curate_every}`
  default 24h, run key `curate`). In selfdev, not operator: the edits
  land in this repo's `memory/` and `packs/` through its `pr-work`.
- The pass: settle earlier PRs, read rejected learnings, group open ones,
  make bounded memory/skill/prompt edits (`jobshop memory check` must
  pass), file a `Curate:` issue and PR listing learning ids per edit,
  start `pr-work` and comment `PR <url> head <sha>` on its implement step
  (#249), then `jobshop learn update <ids> --pr <url>`.
- `jobshop learn update <id>... [--status open|applied|rejected]
  [--pr <url>]` (`learn.update`). Learnings gained a `pr` column
  (migration 00005). A learning stays `open` while its PR is pending;
  the next pass sets `applied` (merged, not reverted) or `rejected`
  (closed unmerged or reverted). `rejected` needs a PR.
- Tests: `TestTriggerLearnings`, `TestLearnUpdate`,
  `TestSelfdevPackLoads`, `TestParseTriggers`.
- Live check: `docs/verify/curator.md`.
