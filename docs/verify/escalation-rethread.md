# Escalation rethread (#197)

After a `gateway: post` WARN with `Invalid event_id in thread message` in
the escalation room, the journal shows `gateway: escalation rethreaded`
with `old_root` and `root`, the room has a new "continues here" thread
with the session's reply, and `jobshop ticket show <escalation>` has
`meta.gateway.root` equal to the new root and the old `key`. A reply in
the new thread gets an answer from the same session.
