# Chat responsiveness and history

2026-11-07

Use an isolated controller with the ACP test agent, an operator chat pool and
escalation chats. Do not send test messages to existing operator conversations.
Build the frontend before the Go binary, then restart the owned Desk process.

```sh
npm --prefix ui run check
npm --prefix ui run build
go build -o "$JCODE_SCRATCH_DIR/jobshop-chat-check" ./cmd/jobshop
DESK_URL=http://127.0.0.1:7492 DESK_MUTATE=1 \
  DESK_ROOT="$JCODE_SCRATCH_DIR/desk-workspace-root" \
  JOBSHOP_BIN="$JCODE_SCRATCH_DIR/jobshop-chat-check" \
  npm --prefix ui run test:e2e -- \
  chat-responsiveness.spec.js trace-session.spec.js --workers=1
```

- A held agent turn remains visibly working. The next human message appears
  before the HTTP acknowledgement, changes from sending to queued, and appears
  exactly once after the agent takes it. The test delays a real HTTP response,
  not the controller's persistence or the agent's output.
- Scrolling near the top requests older raw journal pages automatically.
  Prepending retains the same message and screen position. Both transcript
  surfaces reach the first human message and stop requesting earlier pages.
- The virtualizer measures variable-height rows and renders only the nearby
  window. Expanded prompt/tool state survives leaving that window. Images and
  expanded content are remeasured. Following live output applies only near the
  end, not while reading history.
- Single Markdown newlines remain visible. Human messages retain their original
  text and generated controller context remains collapsed.
- A selected-chat send validates its ID and managed key with ticket.show rather
  than listing all sessions. Arbitrary session attachment still requires rw
  permission. Subscription state includes the initial raw history cursor,
  including frames that have no visible transcript content.
