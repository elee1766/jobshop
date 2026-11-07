# Structured transcripts and the Trace workspace

2026-11-07

Controller-originated ACP `session/prompt` frames carry a versioned
`_meta["jobshop/presentation"]` object. It records original message IDs,
senders and text separately from pool instructions, recovery notes, summaries
and attached controller context. The actual prompt sent to the agent is unchanged.
The session journal remains the source of truth. Normalized session events expose
`messages`, `context` and `raw_prompt` for both live delivery and older-history paging.

Desk renders original messages normally and collapses controller context and raw
prompts. Older frames without this metadata remain available under `legacy prompt`.
Desk does not guess provenance from phrases or strip matching text from user input.

`chat.manage` with `action: convert` and a name converts an open escalation chat
into a named chat. Its session ID, routing key, ACP conversation and journal stay
in place. The source escalation is not resolved. Later source closure does not
send completion instructions into the converted conversation. The chat can be
archived independently, but cannot use the escalation-only `resolve` action.

Trace separates a ticket's parent, children and dependency directions. History
summarizes actions, step names, reasons and references, with raw diffs collapsed.
Claimed tickets and session references can open a read-only transcript beside the
details on desktop or in a dismissible bottom sheet on mobile. This viewer uses
the existing session SSE stream and paginated event endpoint, never a write API.

Home uses canonical run links for PR and Linear issue navigation. Run links use
explicit metadata or workflow variables plus configured Linear identity context.
No Linear workspace is invented from an issue key alone.
