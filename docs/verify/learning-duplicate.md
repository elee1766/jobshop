# Check: learning duplicates (JOB-199)

On a host with several open learnings about one problem (admin socket or a
worker session).

1. `jobshop learn list --status open --limit 500 | jq length`, and note
   three ids `K`, `A`, `B` that describe the same problem.
2. `jobshop learn update A B --duplicate-of K` prints both with status
   `duplicate` and `duplicate_of` K.
3. `jobshop learn list --status open --limit 500 | jq '.[] | select(.id==K)
   | .duplicates'` is 2, and the open count dropped by 2.
   `jobshop learn list --duplicate-of K` returns A and B.
4. These fail and change nothing: `learn update K --duplicate-of K` (itself),
   `learn update <other open id> --duplicate-of B` (B is a duplicate, not
   open) and `--duplicate-of` an applied learning.
5. `jobshop learn update K --status applied --pr memory:1` (or any
   outcome): `learn list --duplicate-of K` still returns A and B, still
   `duplicate`.
6. `jobshop learn update A --status open` reopens A with no `duplicate_of`;
   K's count (read with `--status applied`) is 1.
7. `jobshop memory write` citing B is refused; cite K.
8. Desk, learn: K shows `1 duplicate` next to its problem and B in its
   details. A and B are not rows under "awaiting curator".
9. The curate trigger: with `when.learnings.min` set to the open count
   after step 3, it does not fire for the two duplicates.
