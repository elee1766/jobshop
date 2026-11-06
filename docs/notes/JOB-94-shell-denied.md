# Agent shell that cannot start commands (JOB-94)

2026-11-06. On gasserver (02:00 to 02:52), 11 sessions lost their tool
shell after a controller restart: every bash call failed in 0ms with
`Permission denied (os error 13)`, even `true`, while read tools worked.
Each had its host respawned (often "refreshing host env" right after the
restart) and its jcode conversation reloaded with `session/load`. The
first shell call after the load failed, and none recovered. Sessions in
the same batch that loaded fine, and every session that started a new
conversation, kept working. The cause is inside the reloaded jcode
conversation, not in the controller's sockets or mounts.

The controller cannot run a command in the agent's shell, so it watches
the ACP tool updates. Two calls of kind `execute` in a row that fail with
only `Permission denied (os error 13)` mark the shell broken. The runner
then cancels the turn, stops the host and spawns a new one that starts a
new conversation seeded with the journal summary (as for JOB-65) instead
of loading the old one. Claims are kept and the held ticket is prompted
again. A shell that breaks again before any shell call works counts as a
crash, so a host that cannot run commands at all ends crashed after the
usual restart budget. Logs: "agent shell cannot start commands", span
`session shell_denied`.
