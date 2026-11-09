# Message headers

2026-11-09. Run only against an isolated controller with the ACP test agent
(a `type: chat` pool whose agent is `control.test fake-agent`, see
`desk-chat.md` and `chat-responsiveness.md`). Rebuild the frontend and the Go
binary, restart Desk, then:

```sh
DESK_URL=http://127.0.0.1:7492 DESK_MUTATE=1 \
  npm --prefix ui run test:e2e -- message-headers.spec.js chat-responsiveness.spec.js --workers=1
```

Expected on desktop and mobile: each human message has a `you` bar above it and
no highlight background, an agent turn has one bar however many tool rows it
has, raw prompts have no bar, the read-only viewer shows the same bars, and the
history, tool spacing and live-follow checks in `chat-responsiveness.spec.js`
still pass.

Also look at a populated chat in light and dark themes at 1100 and 412 pixels: a
short reply between long agent messages, consecutive human messages, a queued
message while the agent is busy, and a streaming reply. The speaker must be clear
with the gutter colors ignored.
