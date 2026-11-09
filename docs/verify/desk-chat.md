# Chat in Desk (#203)

Passed 2026-11-03 against a throwaway root on gasserver, with the
gasserver controller down (#223). The root has a `type: chat`,
`attach: rw` pool `operator` whose agent is the control test binary's
`fake-agent`, which echoes the last line of each message. 20 of 20
Playwright tests passed on desktop and Pixel 7, plus the escalation
send on desktop.

1. `go test -c -o control.test ./pkg/orchestration/control`, and a root whose
   jobshop.yaml has that pool (agent `command: <path>/control.test`,
   `args: [fake-agent]`).
2. `jobshop --root <root> serve`, then
   `jobshop --root <root> desk --listen 127.0.0.1:7483`.
3. For the escalation panel, play the gateway's part: create a
   `needs-human` ticket, `session.send` to pool `operator` with key
   `matrix:<room>:<root>`, and set `meta.gateway` to
   `{gateway, room, root, key, session}`, as `pkg/transport/gateway/escalate.go`
   does.
4. In `ui/` (browsers from nixpkgs, as in `docs/verify/desk.md`):
   `DESK_CHAT=1 DESK_CHAT_ESC_SEND=1 DESK_URL=http://127.0.0.1:7483 npx playwright test`.

Pass:
- "new operator session" shows the first prompt as a turn and the streamed
  reply (`echo: PONG`), then a composer follow-up gets `echo: PING`.
- The Sessions view shows a composer for `attach: rw` pools and a
  "Read-only:" reason for the others.
- The escalation drawer streams the gateway's session (its first turn is
  the Matrix-side message). A Desk reply is one more turn in that same
  session, and its echo appears there.

On the live controller, run without `DESK_CHAT_ESC_SEND` first: the chat
test starts one real operator session (agent cost), and sending into a
real escalation thread posts in its Matrix thread.
