# #626 Session display names

2026-11-05. Sessions have a `display_name` (session fields), a label for
Desk, `session list` and other lists. It is not unique and never used for
lookup or dedup; `key` stays the identifier.

- Set at spawn: `session new --name`, or the pool's `display_name` config
  field for every session the pool starts.
- Unset, lists show a default: `<pool>: <title of the held ticket>`, else
  the session ticket's title.
- `jobshop session rename "<name>"` renames the calling session; any
  session may rename itself, also while read-only. `--id` renames another
  session and needs admin or the `session-rename` pool grant. An empty name
  restores the default. Whitespace is collapsed, at most 120 characters.
- The operator chat prompt tells the chat to name itself after its topic.
- Desk's Sessions table shows the name under the id.
