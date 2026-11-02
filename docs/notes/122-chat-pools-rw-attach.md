# Chat pools keyed by thread, and rw attach (#122)

2026-11-02.

- A `type: chat` pool session is started by `session.send {pool, key}`
  (#121). The key is `SessionFields.Key`; the unique index on open
  session keys keeps one live session per key. Two keys get two
  sessions, the same key reuses one.
- Idle TTL: `idle_timeout` defaults to 1h for chat pools (5m for ticket
  pools). Each delivered message turn resets it. Waiting for messages
  is not stuck: the stuck count only grows on turns that leave a claimed
  ticket open, and chat sessions claim none. At the TTL the session
  closes `idle`; the next send for the key starts a new one.
- Controller restart: open sessions are driven again at startup, so a
  keyed session is adopted (live host) or respawned with `session/load`
  of the same ACP session. `session.send` by key finds it by its open
  ticket.
- `attach: rw` on a pool allows `jobshop session attach <id> --rw`
  (admin). It streams `session.subscribe` and sends each stdin line with
  `session.send {id, attach: true}`; a line starting with `!` interrupts.
  EOF or Ctrl-C detaches. Sends with `attach: true` are refused for
  sessions whose pool does not set `attach: rw`; an empty-text attach
  send is the up-front check and creates no message.
- Test: `TestChatPoolKeys`.
