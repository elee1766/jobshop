# Chat pool rw attach

With a chat pool that sets `attach: rw` (e.g. `operator`):

```sh
JOBSHOP_SOCKET= jobshop session new --agent jcode --pool operator --key verify:rw  # or a gateway send
JOBSHOP_SOCKET= jobshop session attach <id> --rw  # type a line; the reply streams back
```

- Typing a line shows `> ...` then the agent's reply and `[turn end: end_turn]`.
- A line starting with `!` during a long turn shows `[turn end: cancelled]`
  and then the new message's turn.
- Ctrl-C detaches; `session list` still shows the session open.
- `session attach <id> --rw` on a ticket-pool session fails with
  `does not allow rw attach`.
- Sending the same key again reaches the same session id, also after
  `systemctl restart jobshop`.
- Left alone past the pool's `idle_timeout`, the session closes `idle`.
