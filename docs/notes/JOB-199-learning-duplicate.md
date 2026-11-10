# Learning status `duplicate`

2026-11-10 (JOB-199).

The curator groups learnings about the same problem but could only close a
group by applying or rejecting it. A group left for a later pass was read
again, row by row, on every pass. A learning can now be marked a duplicate
of the open learning kept for its group, so the group reads as one learning
with a count.

## Behavior

- Status `duplicate` and field `duplicate_of` (the kept learning's id),
  indexed on the learnings collection. `domain.ValidLearning` requires the
  two together.
- `jobshop learn update <ids> --duplicate-of <id>` sets both. It fails,
  changing nothing, when the target does not exist, is one of the ids, or
  is not `open`, and when an id is neither open nor already a duplicate, or
  is itself kept by duplicates (no chains: move those first). A duplicate
  can be moved to another kept learning. It cannot be combined with
  `--status` or `--pr`, and `--status duplicate` is refused. Same access as
  `learn update`.
- `learn update <id> --status open` reopens a duplicate and clears
  `duplicate_of`. Any other status or a `--pr` on a duplicate is refused:
  its outcome is its kept learning's, read through `duplicate_of`.
- `learn list` gives each learning a `duplicates` count (omitted when 0).
  `learn list --duplicate-of <id>` lists the duplicates of one learning.
  `--status duplicate` lists all of them. The count is computed from one
  read of the duplicate rows, not stored.
- No cascade: applying, rejecting or expiring the kept learning leaves its
  duplicates as `duplicate`.
- `memory write --learnings` refuses a duplicate id (it would be marked
  applied and detach from its group): cite the kept learning.
- Counts do not change. The `curate` trigger's `when.learnings` counts
  `open` only, so duplicates leave it. The gate's date-window reads
  (`LearningsBetween`, `CountLearnings` without a status) have no status
  filter, so each duplicate still counts as an occurrence, and
  `EditMetrics` gives the same result before and after marking.
- Not included: detecting duplicates when a learning is recorded.

## Desk

A duplicate whose kept learning is in the loaded list is not a row. The
kept learning shows `N duplicates` after its problem and lists them in its
details, in whatever section it sits (also when it is applied). A search
for a duplicate's id finds the kept learning. A duplicate whose kept
learning is not loaded (older than the 500 learnings Desk loads) lands in
a `duplicates` section, closed by default. Duplicates are not under
"awaiting curator".

## Curator

`packs/selfdev/prompts/curate.md` step 3: a group not acted on this pass
keeps its most informative learning open and marks the rest
`--duplicate-of` it. A group's size is the kept learning's `duplicates`
plus one and counts toward the two-learning rule. Apply or reject a group
by passing the kept learning's id.

## Checks

`TestLearnDuplicate` (real store: two duplicates of a third, list counts,
`--duplicate-of` list, rejected targets, no chains, apply leaves the group,
reopen drops the count, trigger count, gate window and `EditMetrics`
unchanged). `ui/src/lib/learn.test.ts` for `stageOf` and the count, and the
learn e2e spec. Runbook: `docs/verify/learning-duplicate.md`.
