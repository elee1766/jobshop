# request_changes step outcome (JOB-98)

2026-11-06. A reviewing step sends findings back with one close:

    jobshop close -r request_changes --comment "<findings>"

(`--outcome request_changes` is the same close; `--comment -` reads
stdin.) It replaces "comment findings, `run resume <run> --step
deliver`, close done" in `packs/agent` review and `packs/github`
review-merge (case 4).

- Target: the step's `on_request_changes` (a new step field; it must be
  a step the reviewer needs directly), else the previous step when the
  reviewer needs it. Resolved at run start into `fields.on_request_changes`;
  runs started before this read the current config. A first step, or one
  with no target, is refused.
- One store transaction: the reviewer closes `request_changes`, counts
  `fields.review_rounds`, and reopens blocked behind the target; the
  target (must be closed) reopens with `fields.changes` = {step, ticket,
  round, max, findings}. No failure is counted on either step. The
  findings are also a comment on the reviewer's ticket.
- The reopened step's ticket prompt gets a `# Changes requested` section
  with the findings. A `done` close of the step clears `fields.changes`.
  The target moves to its `retry_pool`, and a review-finding correction
  is recorded, as `run resume` did.
- Cap: step field `max_review_rounds` (default 3). The close that makes
  round cap+1 leaves the reviewer closed `request_changes`, does not
  reopen the target, and the run goes to `needs_human` with an
  escalation ticket ("requested changes past its review round limit").
  `run resume` of that step resets rounds as it resets failures.
- Visible: resolution `request_changes` in `run show`, Desk and the
  `jobshop.closes` metric (`jobshop.resolution`); `run show`
  prints `review rounds n/max` and `changes requested by <step>`, Desk
  shows both as pills.

Tests: `TestRequestChanges` (in-process agent: one round then approve,
and the cap to needs_human), `TestRequestChangesRejected`, store
conformance `RequestChanges`.
