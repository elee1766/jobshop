# Pool interrupt permission (JOB-206)

2026-11-12. A pool's `interrupt` setting decides whether an interrupt
message (`jobshop send --interrupt`, `session.send {interrupt}`, Desk chat
interrupt, a subscribed comment on a held ticket) may cancel a running
turn of the pool's sessions.

- `allow` (default, also when unset): the #16 behavior. The runner's
  interrupt poll sends `session/cancel` and delivers at once.
- `queue`: interrupt sends are accepted, but the poll ignores them. The
  message is delivered after the turn, like any other message.
- `refuse`: as `queue`, and the controller rejects interrupt sends
  addressed to the pool (`pool:<name>`, a session id or a session key of
  one of its sessions) with a forbidden error. Plain sends still work.
  Interrupt messages the controller files itself (subscription
  comments) are not refused; they wait for the turn to end.

Any other value fails config load. The runner reads the setting on every
message poll, so a reload applies to a running turn. A message to a key
with no open session is not refused at send time (no pool to check); the
poll gate still keeps it from cancelling a `queue` or `refuse` pool's turn.
