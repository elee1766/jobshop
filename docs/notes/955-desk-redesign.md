# First-party Desk and live API client

2026-11-07. PR #955. Decision: #569.

Desk is a first-party application, not a browser-plugin host. It remains an
optional HTTP process and an admin-socket client. The controller owns mutations.
`packages/jobshop-client` is a private, framework-neutral JavaScript package
with generated RPC types, injected transports, abort support and no global UI
or connection state. Remove `desk.plugins` from configuration. Pack-provided
browser assets are no longer loaded. Controller-side `desk_commands` remain.

## Live state

`api/live` registers its controller subscription before reading initial state.
Commits trigger coalesced snapshots of attention, runs, chats, improvements,
system state and the watched Trace. Unchanged sections are suppressed. Tick
measurements are delivered directly after a completed scheduler pass. Browser
consumers do not poll or fetch state again in response to invalidations.

Dashboard reconnects take fresh snapshots. Ordered controller events and
session journals use sequence cursors, SSE IDs and Last-Event-ID replay.
Active chat reconnects keep the last journal sequence and deduplicate frames.
Transient view commands are audited on the chat, but are not replayed as old
navigation instructions. Heartbeat comments do not query state.

Session streams watch the journal directory for appends. A missing watchable
directory has a bounded polling fallback. Watch removal or failure closes the
stream so reconnect can resume. Push attention reads are event-driven too,
with a timer only to retry failed reads or reconnect a dropped subscription.
External providers still need whatever event or polling mechanism they support.

## Controller and UI

The layout has Waiting, Not-moving and Inflight windows, a persistent chat rail,
Trace, Improvements and System. Named chats support unique names, archive and
restore. Restore retains the journal, session ID and ACP conversation ID, opens
the session socket and registers its suspended runner. The next message loads
the same ACP conversation. Attached references become model-only context.

Attention actions are recomputed at answer time and reject stale option titles.
View commands are validated, scoped to the originating chat and cannot mutate
work. Trace and its selected ticket inspector update through the live stream.
System shows pool capacity, resources and holders, sessions, ingest workers and
last-pass timings. Push excludes non-decision attention and clears resolved tags.

## Bounds and remaining handoff work

Overview reports truncation beyond 100 open runs. Improvements reads at most
500 learnings and edits. Trace uses the controller's bounded ticket tree.

This change establishes the new application and transport, not every proposed
provider-derived projection in the handoff. Remaining work includes PR-head
push timestamps from ingest, dedicated PR/CI Trace nodes, automatic promotion
of permanent stalls, configurable stall thresholds, sweep-cycle grouping and
provider cost attribution. Since-away currently filters in-flight changes, not
a cross-window resolved-work digest. Team/age filtering and non-inflight filter
commands are not supported. These must not be represented as available data.
Real Pixel Fold/Gboard geometry and actual phone notification clearing still
need device acceptance. Browser viewport checks are not substitutes for those.
