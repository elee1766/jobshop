# Guardrail hooks (#188)

2026-11-03.

Config `hooks:` lists commands run at fixed points. Default none, and
with none every point allows.

```yaml
hooks:
  - on: action            # claim, action or merge
    actions: [push, merge] # optional, action and merge only
    pools: [dev]           # optional
    command: jev decide    # sh -c in the root dir, JSON event on stdin
    timeout: 30s
```

- Points: `claim` runs in the controller before a session claims a
  ticket (admin claims and resuming held work skip it). `action` runs
  when a session calls `jobshop guard <action> [--target] [--detail]`
  before push, pr-create, merge, comment or a message to a human.
  `merge` is `jobshop guard merge`; `on: action` hooks see it too.
- Event on stdin: `on`, `action`, `target`, `detail`, `session`,
  `pool`, `project` and the ticket (the claimed one, or the session's
  held one for guard).
- Answer on stdout: `{"decision":"allow|deny|ask_human","reason":"..."}`.
  Empty stdout with exit 0 is allow. Non-zero exit, timeout or bad JSON
  is deny: a broken guardrail must not let actions through.
- Hooks run in config order; the first non-allow answer wins.
- Deny reaches the session as the reply: `jobshop guard` prints the
  answer and exits 1, a refused claim is a forbidden error with the
  reason. A claim-next pick the hook refuses is released with a 10m
  defer so the next claim gets other work.
- ask_human files one `needs-human` ticket (meta `hook.key` =
  session/action/target). Repeated guard calls read it: open is still
  ask_human, closed done allows, any other resolution denies. On claim,
  ask_human is a deny.
- Tests: `TestGuardHooks`, `TestGuardNoHooks`.
- Live check: `docs/verify/guardrail-hooks.md`.
