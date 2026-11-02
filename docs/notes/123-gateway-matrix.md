# jobshop gateway, Matrix adapter (#123)

2026-11-02. Matrix only; Discord would be one more `gateway.Adapter`.

- `jobshop gateway <name>` runs `gateway.<name>` from jobshop.yaml: a
  separate process, an admin-socket client (always the root's admin
  socket, never `JOBSHOP_SOCKET`). Config: `network: matrix`,
  `homeserver`, `token_file` (relative to the root; the token is never
  in the yaml), `pool` (a `type: chat` pool), `rooms`, `senders`,
  `interrupt_marker` (default `!`). Rooms and senders are required;
  messages from other rooms or senders are ignored.
- Adapter interface (`pkg/gateway`): `Network`, `Run` (receive messages
  with thread, id, sender, text), `Post` (reply in thread), `React`.
- Key: `matrix:<room id>:<thread root event id>`. A message outside a
  thread roots its own thread; the reply goes into that thread. Edits
  and non-text events are dropped; events before the first sync are
  skipped, so a restart does not replay history. Unencrypted rooms only.
- Each message is `session.send` by pool and key (interrupt when
  prefixed with the marker) and gets a 👀 reaction. One
  `session.subscribe` per session posts each turn's text as an
  `m.notice` in the thread at `turn_end`. Turns that started before the
  gateway first saw the session are not reposted. A send error is posted
  in the thread.
- The bot joins the configured rooms on start (invite it first).
- NixOS: `services.jobshop.gateway.<name>` writes `settings.gateway.<name>`
  and runs `jobshop-gateway-<name>.service` (Restart=always).
- Tests: `pkg/gateway` (fake adapter and controller: allowlist, key,
  marker, one watcher, old turns), `TestParseGateway`. Live check:
  `docs/verify/matrix-gateway.md`.
