# Message headers in chat and session transcripts (JOB-137)

2026-11-09. `ui/src/screens/Transcript.tsx`, shared by the chat panel and the
read-only session viewer, labels who is speaking.

- A human message has a `you` label bar (tag background, thin rule) above it and
  a black gutter rule beside it. The text has no highlight background. A sender
  other than the operator (anything but `admin` or `desk:*`) shows its own name.
- An agent turn has one muted label bar on its first row, then its tool rows and
  replies follow under it with a pale gutter rule. The label is the session's
  pool in the session viewer (from the trace's session info), else `agent`.
- Consecutive human messages each get their own bar. Queued and sending messages
  use the same `you` bar plus the delivery line.
- `controller context`, `raw model prompt` and `legacy prompt` stay muted
  disclosures with no header, placed under the human message they came with.
  A turn with no human message shows only its disclosures.
- Labels and structure carry the meaning. The black and pale gutter colors and
  the tag background are secondary and come from the theme tokens, so dark
  works without overrides. Same text column for both speakers, no bubbles.
- Row keys, measurement, history prepend anchoring and end-following are
  unchanged: the headers are part of each virtualized row.

Check: `docs/verify/message-headers.md`.
