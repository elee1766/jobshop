# 529: session.journal

2026-11-05.

`session.journal {id, since_seq, limit, kinds, from}` returns a page of a
session's ACP journal: `{frames: [{seq, t, from, msg}], next_seq, done}`
in seq order. It reads `sessions/<id>/journal.ndjson` with acphost's
reader, so it works for live and ended sessions while the dir exists.
There is no copy and no index.

- `limit` defaults to 500, max 5000. `next_seq` is the last seq looked
  at, including frames the filters skipped, so a filtered caller pages
  without rescanning. `done` means the page reached the end.
- `kinds` matches `msg.method` or `params.update.sessionUpdate`
  (`session/prompt`, `tool_call_update`, `agent_message_chunk`). `from`
  is `agent` or `controller`.
- Access: admin; a session for its own id; a pool with the
  `session-journal` admin grant for sessions in its own project.
- Every frame goes through `telemetry.RedactJSON`, the conversation
  export redaction: secret-named keys, secret values from the agent's
  env (controller env plus spec env), and gitleaks rules.
- It takes the session id only. A later multi-controller setup can
  forward it to the controller holding the host without an API change.
- CLI: `jobshop session journal <id> [--since] [--limit] [--kinds]
  [--from] [--follow]` prints JSON lines. `--follow` polls once a second
  rather than using `session.subscribe`, which is admin-only and sends
  summarized events, not frames.
- Desk: `GET /api/session/{id}/journal?since_seq=&limit=&kinds=a,b&from=`.
