# Check: learn expire (JOB-198)

On a host with old open learnings (admin socket, not a worker session).

1. `jobshop learn list --status open --limit 500 | jq length` and note the count.
2. `jobshop learn expire --older-than 120h --kind 'review-*'` prints
   `{"expired": N}`. Run it again: `expired` is 0.
3. `jobshop learn list --status expired --limit 500 | jq length` is N, and
   `--status open` dropped by N. A learning with a `pr` is still open.
4. `jobshop recall <words from an expired summary>` still finds it.
5. `jobshop learn update <id> --status open` reopens one;
   `jobshop learn update <id> --status rejected` fails without `--pr`.
6. From a worker session `jobshop learn expire --older-than 1h` is refused.
7. Desk, learn: the `expired` section is closed and counts N; "awaiting curator"
   no longer lists them.
