# Stall checks do not cascade (#704)

2026-11-05.

- A run started by a stall trigger (`meta.trigger.key` is
  `<trigger>/stall:...`) is never a `needs_human` stall subject. When a
  stall check needs a human it has already filed its escalation, and a
  check on it only adds another one.
- One stall check per stall key (`stall:<signal>:<subject>`) across all
  triggers: a stall does not fire while any trigger's run for that key is
  open, or while the escalation of any such run is open. The first trigger
  to fire wins. Per-trigger cooldown and max_runs still apply on top.
