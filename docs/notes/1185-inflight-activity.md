# In flight: activity apart from run state (#1185)

2026-11-10

A run in In flight stays visible while it is `running`. The row now says two
things: the `run` column is the run's lifecycle state, and the `activity`
column is what its current step is doing now.

Activity, in the order it is decided (`ui/src/lib/activity.ts`):

- `waiting on a person`: the run is `needs_human`, or the step has a
  `needs-human`, `question` or `proposed` label.
- A step with a holder: `in a turn` only when the live session (the system
  view, not the retained holder) is busy, with its running tool. Otherwise
  `held, between turns`, `held, suspended`, `host lost` (the controller's host
  is gone) or plain `held` when the session state is unknown.
- `deferred`: a wait on times alone, or a release until a time. Shows the
  reason and the resume time.
- `waiting on an event`: a wait with other conditions, with its reason and the
  unmet conditions.
- `waiting for a slot`: a full resource. `gated`: held for earlier steps.
  `queued`: nothing else applies.

`jobshop release --until` now records its reason on a step as
`fields.defer {reason, until}`. The reason is shown only while `until` equals
the step's `defer_until`, because other paths (quarantine, retries) move
`defer_until` without a reason. Without a recorded reason the row shows the
time alone.

Check: `docs/verify/inflight-activity.md`.
