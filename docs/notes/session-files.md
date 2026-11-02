# Session files and restart

Files in `sessions/<id>/`: `spec.json`, `host.lock` (flock; held means
alive), `host.pid`, `host.log`, `acp.sock`, `journal.ndjson`, `exit.json`,
and `work/` (git worktree when the pool sets `repo`).

- The host rewrites request ids so a new controller never collides with an
  old one, re-sends unanswered agent requests to a new controller, and
  reports a prompt that finished while detached as `_jobshop/turn_end`.
- Backpressure (decided 2026-11-01, #39): agent stdout is read and
  journaled without waiting on the controller. Frames are then queued per
  connection and one writer goroutine drains the queue, so a controller
  that stops reading blocks only that goroutine. The queue holds at most
  8 MiB or 4096 frames. Past that, agent notifications are dropped from
  memory only: the writer replays them from the journal in seq order
  before the next live frame and logs the count (`replayed frames from
  journal`; the host has no metrics). Responses, agent requests and host
  messages are never dropped. At most 256 agent requests wait for a
  controller answer; more get a JSON-RPC error back to the agent. With no
  controller nothing is queued. Optional `Hello.since` replays agent
  notifications after a seq on connect (protocol unchanged).
- Stopping a host: SIGTERM (forwarded to the agent's group, SIGKILL after
  5s); the controller SIGKILLs the host after 10s.
- Leak guards: the agent gets `Pdeathsig: SIGKILL` and its own process
  group; the host is a child subreaper; the host exits with its agent and
  after 30 minutes with no controller.
- Controller start: open session tickets are adopted (live host) or
  respawned (`session/load`); hosts with no open session are stopped.
- Load fallback (#43): if the agent lacks `LoadSession` or `session/load`
  fails, the controller calls `session/new` and builds a summary from
  `journal.ndjson` (`host.Summarize`: prompts, agent text, tool call titles,
  latest plan; at most 16 KiB, keeping the first prompt and the newest
  items). The summary is prepended to the next turn's prompt, ended by
  `[end of summary]`, with no extra turn; a later summary skips text
  before that marker so summaries do not nest. The log says which path
  (`load` or `new+summary`) was taken. The pending summary is memory only:
  a controller restart before the next turn loses it.
- Prompts use `context.WithoutCancel`, and stopping the controller closes
  the connection instead of cancelling requests, so turns keep running in
  the host.
- Git: `fetch` then `worktree add --detach <ref>` per session, removed when
  the session ends; git calls on one repo are serialized.
- A worktree deleted under a live session (seen on gasserver 2026-11-01:
  every tool call failed for five hours, tickets claimed and dropped) makes
  the runner respawn the host before its next claim. Spawning prunes stale
  worktree registrations and adds a fresh worktree at `ref`; the
  conversation and claims are kept, uncommitted work in the old one is lost.
- `${SESSION_DIR}` expands in agent args.
