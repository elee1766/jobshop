# Env refresh and session restart (#11, #37)

`spec.json` carries `env_hash`, a sha256 of the agent's full env (inherited
controller env minus per-start systemd variables, then the spec env). When
adopting a live idle host whose hash differs, the controller stops it and
respawns it at once with `session/load`; a busy session refreshes after its
turn. `jobshop session restart <id>` does the same on demand, keeping
claims; `--now` cancels the running turn first. Neither counts as a crash.
A respawned host's stale `exit.json` is removed before spawn, and an
adopted host whose agent never initialized is treated as fresh, so the
conversation is always loaded. A live session also re-checks the hash
before each claim, after every turn and on every idle poll (#106,
2026-11-02), so a session that stays connected across a deploy or a
reload that changes agent env refreshes too. Stopping a host also stops
any process left with the session's `JOBSHOP_SOCKET` in its environ
(#111): jcode `acp --socket` runs tools in a detached `jcode serve`
daemon outside the host's process group, so without this the refreshed
agent reconnected to a daemon that still had the old PATH.
