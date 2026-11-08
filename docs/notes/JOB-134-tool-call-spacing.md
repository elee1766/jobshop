# Tighter spacing for runs of chat tool calls

2026-11-08. JOB-134.

In `SessionTranscript.svelte` (chat and the read-only session viewer) a tool
call row directly after another tool call gets `tool-next` and no top
margin, so a run of tool calls stacks without gaps. The first tool of a run
keeps `--space-3` above it, and the first non-tool row after a run gets
`after-tools` with `--space-3` padding, so runs stay apart from human and
agent messages.

Spacing sits on the following row only: a newly appended row never changes
the height of a row the virtualizer already measured. Tool summaries keep
`--space-1` vertical padding (`--space-2` on coarse pointers) as the hit
target, and an expanded tool's output keeps `--space-2` below it.

The fake ACP agent in `pkg/control` tests replies with four tool calls when
the message contains "tool run". `chat-responsiveness.spec.js` uses it to
check gaps, focus, hit target and remeasuring after expansion in both
transcripts on desktop and mobile (see `docs/verify/chat-responsiveness.md`
for the isolated setup).
