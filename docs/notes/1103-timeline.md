# Timeline view of a trace (#1103)

2026-11-09

`ticket.timeline` (params like `ticket.graph`: `id`, `limit`) returns the
graph around a ticket as bars on one time axis. Desk shows it as
`?view=trace&ref=<id>`, the only trace view since #1179 (the graph view and
its toggle are gone). The Desk trace snapshot
carries `timeline` only when the stream or `GET /api/trace/<ref>` is opened
with `timeline=1`, since it replays the history of every node.

## Data

`rows` come in tree order. A row's parent is its `parent_id`, else the ticket
whose holder created it (the `created` rule of `ticket trace`, docs/notes/53).
Each row has `segments` (contiguous, oldest first, the last has no `end` while
the ticket is open), `markers` (comment ticks, the PR a step announced with a
`PR <url>` comment, a session taking over), `started` (first claim), and
`partial`.

Segment states, replayed from the history rows:

| state | when |
|---|---|
| `queued` | open, not claimed, no wait, no deferral |
| `working` | claimed (actor is the session), or a run or batch that is open |
| `waiting` | `fields.wait` set. `wait` lists each condition with its label, URL or ticket target and state; `ended` says how a finished wait ended (met, invalidated, cleared, ended) |
| `human` | label `needs-human`, `question` or `proposed`, or a run in `needs_human` |
| `deferred` | released with an `until`, or a wait on a time alone |
| `retry` | a claim that ended in a failed attempt (release or close that counted a failure, or `lease_expired`) |
| `unknown` | before the oldest history row the store kept |

A closed ticket's bar ends at its close. A step closed `failed`, `rebase` or
`request_changes` that was reopened continues after it.

`arrows` run from the prerequisite to the dependent. `blocks` and `waits_for`
are `met` (the blocker closed before the dependent started), `blocking` (the
blocker is open and the dependent has not started) or `conflict` (the
dependent started before the blocker closed, or the blocker is still open).
`tracks` has no state. `wait` is a wait on a ticket or run, with the wait's own
state.

## Answers to the open questions on the issue

1. History is complete enough. Every transition already has a row: `create`,
   `claim`, `release`, `lease_expired`, `handoff`, `close`, `reopen`,
   `label_add`, `label_remove`, and `update` with the `fields.wait`,
   `fields.state` and `defer_until` diffs. No durable transition table is
   added. History is not compacted anywhere today. Ephemeral tickets and any
   ticket whose first row is not `create` show an `unknown` span and
   `partial`. Heartbeats write nothing and are not needed.
2. Wait set and clear and attention open and close already write rows:
   a wait is an `update` with the `fields` diff, a `needs-human` or `question`
   label is `label_add` and `label_remove`. One gap was closed: `create` now
   records the labels the ticket was born with (`diff.labels`), because a
   question is created with its labels and nothing else says when they began.
   Older tickets without it start `queued` and take the label from a later
   `label_add`.
3. Placement: a toggle on the trace, as preferred.

## Limits

Colors are the Jira palette from the issue, as `--tl-*` tokens (light and
dark). Every state also has a pattern or label, so color is never the only cue.
Stripes mark waiting, deferred and unknown. A bar that is open ends at the
dashed "now" line, which moves every 15 seconds and with each live update.
A thread comment tick is shown for the newest 40 comments of a ticket. The
node limit is the graph's (2000).
