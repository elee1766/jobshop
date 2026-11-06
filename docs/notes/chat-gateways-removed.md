# Matrix and Discord gateways removed

2026-11-06. The Matrix and Discord chat gateways are removed. They had
been paused since 2026-11-04, with Desk as the operator UI, and no host
ran one. What remains:

- Desk: web UI and HTTP API over the admin socket, including chat with
  operator and escalation sessions.
- The A2A gateway (#547): `jobshop gateway <name>` with `network: a2a`
  hands escalation chats to another agent.

Removed: `pkg/gateway` chat code (adapter, Matrix, Discord, escalation
threads, attention digest, artifact room posts), the `matrix` and
`discord` networks and their config keys (`homeserver`, `rooms`,
`senders`, `interrupt_marker`, `escalation_room`, `attention_digest`,
`artifact_room`), the matching NixOS module options, and the
`mautrix` and `discordgo` dependencies.

A config that still sets `network: matrix` or `discord` fails to load
and names the removal. Chat-only keys on a gateway fail strict parsing.
Escalation tickets keep any `meta.gateway` thread fields written before;
the A2A gateway preserves them when it writes its own state.
