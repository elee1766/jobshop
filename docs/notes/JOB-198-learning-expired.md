# Learning status `expired` (JOB-198, 2026-11-10)

Open learnings only left the open list when a curator marked them applied or
rejected, and `rejected` needs a `--pr`. Ones the curator skips (auto-recorded
controller learnings, one-offs) stayed open for good and `learn list --status
open --limit 500` stopped reaching the oldest. A fourth status, `expired`,
takes them out of the curator's view.

- `domain.LearningExpired`; `domain.ValidLearningStatus` is the one check used
  by `ValidLearning` and `knowledge.UpdateLearning`. The record state is free
  text, so there is no migration.
- `jobshop learn update <ids> --status expired` needs no `--pr`; `--status open`
  reopens. `--status rejected` still needs one.
- `jobshop learn expire --older-than <duration> [--scope S] [--kind K|glob]`
  (`learn.expire`, admin socket only, a worker session is refused) sets
  `expired` on open learnings created before now minus the duration that have
  no `pr`. It prints `{"expired": n}`. A learning with a `pr` is left open: a
  fix is proposed. Each record is re-checked under its revision, so one that
  gained a `pr` after listing is skipped.
- Unchanged on purpose: the `curate` trigger's `when.learnings` counts `open`
  only, so expired ones drop out of the count. `Measure`/`EditMetrics` read
  learnings by `created_at` and ignore the status, so a gate window measures
  the same before and after expiry. Recall still finds expired learnings.
- Desk learn view: an `expired` section, closed by default, after `rejected`.
  `stageOf` maps an expired learning to it (`ui/src/lib/learn.test.ts`,
  `ui/e2e/learn.spec.js`).
- `packs/selfdev/prompts/curate.md` says expired learnings are not gathered.
- Not done: expiring automatically. The backlog grows again unless the command
  is run from time to time.
