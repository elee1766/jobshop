# Learning loop

End to end check of capture (#264, #272), curate (#273) and the gate
(#274) on the host. `docs/verify/curator.md` covers the curate pass in
more detail.

Capture
1. Close a step `-r failed` on a scratch run: `jobshop learn list
   --since 10m` shows one `failure`, source `controller`, scope the run's
   repo.
2. `jobshop learn correction "x" --fix "y"` from a session: listed with
   source `agent` and the session's ticket.

Curate
3. `jobshop trigger list` shows `curate`; once it fires, the curate run
   ends with a `Curate:` PR and its learnings have `pr` set.
4. After that PR merges, the next pass runs `jobshop learn update <ids>
   --status applied --merged-at <t>`. `jobshop learn edits` shows one
   `pending` edit with that PR, the learning ids, its pools and a
   `baseline` whose `from` is 7 days before `merged_at`.

Gate
5. `jobshop trigger list` shows `evaluate` with `1 pending memory edits,
   none due` until 20 runs touch its pools or 72h pass, then one
   `evaluate` run with key `evaluate`.
6. The evaluate step's comment names each edit as `helped` or
   `not_helped`. `jobshop learn edits` shows `state`, `after` and
   `reason`; `jobshop learn list` shows `gate` on the cited learnings.
7. For a `not_helped` edit: a `Revert curator edit` issue, a revert PR
   through `pr-work`, and the cited learnings `rejected` with the
   original PR. The next curate pass does not propose the same edit.
8. To force a decision on a test edit: `jobshop learn evaluate <id>`.
