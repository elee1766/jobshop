# Linear write commands (#280)

2026-11-04. Write half of "CLI parity" on target #258, after the read
commands (#279).

- `jobshop linear create --team T --title X [--description --state
  --label --assignee --parent]`, `update <KEY> [--title --state
  --assignee --add-label --remove-label --if-updated-at]`, `ask <KEY>
  <body> [--mention <user>] [--reply-to <comment id>]` and `relate <KEY>
  --blocks|--related <KEY>`. Names (state, label, user) are resolved
  against the team, case-insensitive; an unknown name fails and lists
  the valid ones. Output is JSON: `{op, id, identifier, url, replayed}`.
- `--dry-run` resolves names and prints the mutation (`op`, `query`,
  `variables`) without sending it or storing a key.
- `--key <k>`: the controller keeps `<project>/<k>` -> result in a new
  `write_keys` table (migration 00007, RPCs `writekey.claim` and
  `writekey.finish`). A key already done prints the stored result with
  `replayed: true`. A key is bound to its op. Created entities (issue,
  comment, relation) get a UUID derived from the key as their Linear
  `id`, so a retry whose first attempt landed but never reported back
  (the key left pending) reads that entity back instead of posting
  again. Without `--key` a write is sent once, unguarded.
- `update` reads `updatedAt` first: with `--if-updated-at` a mismatch
  fails before any write. After the mutation it reads the issue back
  and fails when the title, state, assignee or labels are not what was
  asked.
- `ask --mention` puts the user's profile URL at the start of the body:
  Linear turns it into a mention that notifies; a plain `@name` does
  not.
- Tokens come from `identity.token` as for the other linear commands
  and are never printed. Nothing here runs for `tracker: github`
  workflows: only these commands call Linear.
- Tests: `pkg/linear/write_test.go` on a fake GraphQL server (repeated
  key posts once, lost reply then retry posts once, dry run sends
  nothing, if-updated-at conflict, read back mismatch, mention URL),
  `pkg/store/sqlite/write_keys_test.go`.
