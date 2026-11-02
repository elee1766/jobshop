# Attach and session admin (#17)

- `jobshop session attach <id> [--since <seq>] [--json] [--no-follow]`
  renders the journal file and follows it; works while the controller is
  down.
- `session list` joins session tickets with their dirs: `alive`, pids,
  current ticket, last journal seq, `exit`, `crashes`, `backoff_until`,
  orphans. `--state <resolution>` and `--since <dur>` (#52) list only
  closed sessions matching both, without orphans:
  `session list --state crashed --since 24h` is the crash evidence.
- `session reap` stops hosts with no open session; `--closed` also deletes
  old closed session dirs.
- `shutdown` leaves sessions running; `shutdown --all` ends them first.
