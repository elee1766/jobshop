# Desk transcript: a controller turn shows its prompt once

2026-11-10.

A turn with no human message (a ticket prompt, a step prompt, a recovery
note) is all controller context. Its `jobshop/presentation` context is the
whole prompt, so the event's `context` and `raw_prompt` are the same text.
Desk listed both, as `controller context` and `raw model prompt`, and the
prompt read as sent twice.

It was sent once. In the journals of live sessions each such turn has one
`session/prompt` frame with one text block, and every context paragraph
occurs in it once.

`Transcript.tsx` now leaves out `raw model prompt` when it equals the
context. A turn with human messages keeps both: there the raw prompt is the
message envelope plus the context, which differs from either part.

Check: `ui/e2e/message-headers.spec.js` (`a ticket turn shows its prompt
once`), run by `ui/e2e/run-isolated.mjs`.
