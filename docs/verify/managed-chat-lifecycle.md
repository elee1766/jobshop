# Managed chat lifecycle

2026-11-07

The chat picker includes named Desk chats, escalation chats and attention chats.
Its type column distinguishes them. Other gateway sessions remain excluded.

- `archive` on a named chat, or `close` on a managed chat, stops its session and
  moves it out of recent chats. The linked ticket stays open.
- `resolve` on an escalation chat confirms before closing its linked ticket with
  resolution `done`, then closes the session. It does not resume a run or
  unquarantine a pool. Only the escalation's current session can resolve it.
- Archived conversations remain available in the archived picker and can be
  restored with their transcript and original chat identity.
- The attention prestart pass does not recreate an explicitly closed chat.
  An explicit discuss action can start another conversation for that item.

Run the controller regression with:

```sh
go test ./pkg/orchestration/control -run 'TestManagedChatDismissal|TestNamedDeskChatLifecycle' -count=1
```

Against an isolated controller with an `operator` chat pool, an escalation
`hold_prompt` and attention prestart enabled, run the browser workflow:

```sh
cd ui
DESK_MUTATE=1 DESK_URL=http://127.0.0.1:7492 npm run test:e2e -- chat-dismiss.spec.js --workers=1
```

The browser check covers picker types, close, archive selection, restore,
cancelled and confirmed resolution, and reload on desktop and mobile.
