# Curator

Check that the curate trigger starts one pass and that learnings move
through their states (#273).

1. `JOBSHOP_SOCKET= jobshop trigger list` shows `curate` with
   `open learnings N is below 10` while fewer than 10 open learnings
   exist since its last run.
2. `jobshop learn list --status open` reaches 10 (or 24h pass with at
   least one): `jobshop run list` shows one `curate` run with key
   `curate`; `trigger list` shows it held by that run. More learnings do
   not start a second run.
3. The pass ends with a `Curate:` issue, a PR whose body names learning
   ids per edit, and a `pr-work` run whose implement step has a
   `PR <url> head <sha>` comment. `jobshop learn list` shows those
   learnings with `pr` set and status `open`.
4. After the PR merges, the next pass sets them `applied`; for a closed
   unmerged PR, `rejected` with the PR.
5. `jobshop memory log` lists the pass's local writes, and `jobshop
   memory check` on `<root>/memory` exits 0 (#592).
