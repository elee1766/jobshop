# Configurable ticket id generator (JOB-57)

2026-11-06. `ids:` in jobshop.yaml picks the generator for new ticket
ids; `projects.<name>.ids` overrides it per project.

```yaml
ids: {generator: sonyflake}        # default
projects:
  hq: {ids: {generator: words, words: 3, separator: "-"}}
```

| generator | example body | sorts by time as a string |
|---|---|---|
| `sonyflake` (default) | `0033gnxkhg012` | yes |
| `seq` | `12` | no (`js-10` < `js-9`) |
| `ulid` | `01jbz...` (26) | yes |
| `uuidv7` | `0193...-7...` | yes |
| `uuid` (v4) | random | no |
| `xid` | 20 chars base32hex | yes (second, then counter) |
| `words` | `brisk-otter-lamp` | no |

- `pkg/ids`: one `Generator` interface (`Next`, `Sorted`, `Match`), one
  type per generator. No new modules; words come from an embedded list.
- Bodies have disjoint shapes, so ids of different generators under one
  prefix cannot collide, and switching generator only changes new ids.
  No config is rejected for collisions: `seq` starts one past the
  highest numeric id already under the prefix, and `words` retries a
  taken id inside the insert transaction.
- `seq` is minted inside the insert transaction from an `id_seqs` table
  (SQLite migration 00026, Postgres 00011).
- Nothing relies on id string order alone: List and claim order by
  `created_at` with id as a tiebreak. `IsSession` checks the `-s-` tag
  against every generator's shape.
