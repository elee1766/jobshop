# Gateway, Discord adapter (#182)

2026-11-03. Unit tests only; no live bot token yet.

- Config: `network: discord`, `token_file` (bot token, without `Bot `),
  `pool`, `rooms` = channel ids (text or forum), `senders` = user ids,
  `escalation_room` = a channel id from `rooms`, usually a forum.
  `homeserver` is matrix only. The bot needs the Message Content intent.
- Key: `discord:<channel id>:<thread id>`. A message in a text channel
  starts a thread on itself (the thread gets the message's id), and the
  reply goes there. A message in a thread or forum post keys by the
  thread's parent channel and the thread id. DMs and other channel types
  are dropped.
- Dropped: bot and own messages, system messages, empty content. Edits
  are `MESSAGE_UPDATE` and are not handled.
- Replies split at 2000 characters, preferring a line break, then a
  space. The 👀 reaction on a thread starter message goes to the parent
  channel.
- Escalations: `Open` in a forum channel creates a new post; in a text
  channel it sends a message and starts a thread on it.
- discordgo reconnects and resumes the gateway session
  (`ShouldReconnectOnError`); `SyncEvents` keeps messages in order.
- NixOS: `network = "discord"`; `homeserver` now defaults to empty and is
  only written when set.
- Tests: `pkg/gateway/discord_test.go` (fake session: thread start,
  allowlist, key, react, forum and text escalation, splitting) and
  `TestParseGateway`.
