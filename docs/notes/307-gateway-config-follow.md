# #307 Gateway follows config without a restart

2026-11-04. `jobshop gateway <name>` reads its own entry from the
controller's active config (`config.show` on the admin socket) every 30s
and applies changes to `rooms`, `senders`, `interrupt_marker` and
`escalation_room`. So a `config.reload` (a NixOS switch) reaches the
gateway within 30s with no restart.

- Matrix joins a newly added room. Removed rooms and senders are refused
  from the next message.
- An escalation room set after start starts the escalation follower (and
  the digest) then. Clearing it does not stop them; new threads need a
  room, so clear it with a restart.
- `network`, `homeserver`, `token_file`, `pool` and `attention_digest`
  still need a restart: they pick the adapter, the session keys' pool and
  the digest goroutine.
- A failed fetch (controller down) keeps the current settings.

Polling, not a config-changed event: the gateway already dials per call,
and a 30s delay is fine for allowlist changes.
