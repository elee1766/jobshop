# Write keys: remote key first, local 48h cache (#415)

2026-11-04.

- `write_keys` rows are pruned on the controller tick (every 10m) once
  older than `write_keys.retention` (default 48h, Go duration or `Nd`),
  done or pending. The count is logged as `write keys pruned`.
- A done local row replays its result without a remote call. A missing
  or pending row checks the remote first.
- Linear: creates already use `linear.KeyID(key)` as the entity id. If
  the mutation fails (Linear rejects a reused id, or the reply was lost)
  `Do` looks the id up and returns the existing entity as a replay.
- GitHub: `jobshop gh ... --jobshop-key <k>` on `pr|issue comment`,
  `pr review` and `issue create` appends the hidden marker
  `<!-- jobshop-key: <project>/<k> -->` to the body. Before posting it
  searches the PR/issue comments, reviews, or issue search for the marker
  and prints the match instead of posting. GitHub does not render HTML
  comments, so posted text reads the same. The flag is stripped before gh
  runs; it is not `--key` so it never shadows a gh flag.
- Issue search is eventually consistent: a retry seconds after the first
  post may miss it, but that retry still has the local row.
- Other targets (Linear update, no remote key): a retry after the row
  expires can write again. Updates are mostly idempotent by content.
