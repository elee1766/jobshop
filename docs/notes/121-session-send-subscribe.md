# session.send and session.subscribe (#121)

Admin socket only.

- `session.send {id | pool+key, text, interrupt}` creates a message
  ticket to the session id (first line title, rest body; `16-messages.md` delivery).
  With pool and key and no open session for the key, a `type: chat` pool
  starts one (`started: true`); a ticket pool is an error. A key open in
  another pool is a conflict. Message turns are interruptible too.
- Chat pools claim no tickets and the supervisor starts no sessions for
  them. Idle TTL and restart by key are #122.
- `session.subscribe {id, since_seq}` stays open and sends
  `session.event` notifications on the same connection: `turn_start`
  (prompt), `text` (message chunk), `tool` (id, title, status),
  `turn_end` (stop reason or error), each with its journal seq, and
  `state` (status, resolution; no seq) first and on change. It returns
  `{seq, closed, resolution}` when the session closes or the client
  hangs up. Reconnect with the last seq seen; `rpc.Client.Subscribe`
  tracks it. Test: `TestSessionSendSubscribe`.
