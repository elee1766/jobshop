# JOB-65: fall back to a new ACP session on "Unknown ACP session id"

2026-11-06.

When `session/prompt` fails with -32602 "Unknown ACP session id" (jcode
after a restart it did not reload), `runner.turn` opens a new
conversation on the same host with `session/new`, seeds the prompt with
the journal summary (the same path as a failed `session/load`), stores
the new id in `SessionFields.ACPSession` and retries the prompt once.

No host restart and no crash is counted, so the session does not reach
quarantine, and held tickets stay claimed. A second rejection in the same
turn is returned as an error as before. The lifecycle event is
`acp_session_unknown`.

The runner keeps the current ACP session id in memory (`runner.acpSession`)
so cancels, config sets and later turns use the replacement.

Test: `TestUnknownSessionFallback` (pkg/control).
