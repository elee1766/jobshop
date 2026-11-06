# Attention item actions show a pending state (JOB-60)

2026-11-06. After a canned action (Close ticket, Acknowledge, ...) succeeds,
in the Attention list or the chat panel's item bar, the row is dimmed and
shows "<action>…" instead of its buttons. Desk refreshes `api/attention`
at once, and again 1.5s later while the row is still listed. The row drops
when a load no longer lists it. If two loads after the action still list
it, the buttons come back with "still listed after <action>", so a no-op
is visible. A failed RPC keeps the error toast and leaves the row alone.

The state is client only (`ui/kit/attention.js`, a global shared by the
attention and chat plugin bundles); a reload forgets it.
