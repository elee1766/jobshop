# Incident report for planner passes (#156)

2026-11-02. Part of #145. Planner passes file reliability issues from
evidence instead of memory.

`jobshop report incidents [--since 24h]` calls `report.incidents` (admin
socket, or a pool grant `report-incidents`; the github pack grants it to
`plan`). The result has:

- `logs`: controller WARN/ERROR lines grouped by level and message, most
  frequent first, with count, first and last time, and the attrs of the
  last line as `sample`.
- `sessions`: session tickets closed `crashed` or `stuck` in the window,
  with the last comment (the controller's reason).
- `steps`: run steps updated in the window with `fields.failures > 0`,
  with status, retries and the last comment as `last_error`.

Log lines come from an in-memory buffer (`pkg/control/incidents.go`): a
slog handler around the controller logger keeps the last 5000 WARN+
records. It is lost on controller restart, so after a deploy the window
covers only the time since start. Sessions and steps come from the store
and survive restarts. A persisted log was not needed for a daily pass;
the journal on gasserver still has the full history.

The planner prompt (`packs/github/prompts/plan.md`, step 4) runs the
report each pass and files one deduped issue per recurring cause.
