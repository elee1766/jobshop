# Unknown ACP session: towa's -32002 (JOB-214)

2026-11-13. `session.UnknownSession` decides when a failed prompt means
the agent lost the ACP session, so the runner opens a new conversation
seeded with the journal summary instead of failing the turn
([JOB-65 note](JOB-65-acp-session-fallback.md)). It now recognises two
shapes:

- jcode: `-32602` with "Unknown ACP session id" in the message.
- towa: `-32002` ("Resource not found", ACP's code) with data
  `{"sessionId": "<id>"}`. towa builds this one error for a session that
  is unknown, owned by another connection, or closed.

The `-32002` case needs the `sessionId` key in the data object (decoded
as `map[string]any`, not matched as text). The code alone only says some
resource is missing; an agent may use it for something else, and a new
conversation for that would drop context for nothing.

Not covered: towa's other prompt errors (rate limit, overload, network,
auth as `-32603`/`-32000` with `data.towa.stop.kind`). `RateLimited` and
`ProviderUnreachable` still match message text and were not checked
against towa.

Tests: `TestPromptErrors`, `TestUnknownSessionFallback/{jcode,towa}`.
