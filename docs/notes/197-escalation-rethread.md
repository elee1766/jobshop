# Gateway: replace an escalation thread whose root is gone (#197)

2026-11-03. A Matrix post rejected with M_NOT_FOUND, or M_INVALID_PARAM
naming the event id ("Invalid event_id in thread message: M_NOT_FOUND"),
is `ErrThreadGone`; a retry would fail the same way. The post WARN now
logs the root. If the thread is an open escalation this gateway owns,
`rethread` opens a new thread in `escalation_room`, writes its room and
root to `meta.gateway` (conditional on the old root, so it happens once),
moves session watches to it and posts the dropped text there. `key` is
unchanged, so the same session continues; replies in the new thread map
to it through the recorded key, including after a gateway restart.
Posts to non-escalation threads are only logged.
