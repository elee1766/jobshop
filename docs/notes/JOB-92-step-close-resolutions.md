# Step close resolutions (JOB-92)

2026-11-06. `ticket.close` on a workflow step accepts only `done`,
`failed`, `rebase` and `cancelled`, plus `skipped` from an admin caller
(the controller skips steps whose `when` does not match). Anything else
is refused with `ErrInvalid` before the outcome is set, and the step
stays held.

Why: a reviewer closed a step with its review summary as the resolution.
The run engine finishes a step only on `done` or `skipped`
(`store.StepFinished`), so the join step never opened and the run held
a trigger's open slot for 90 minutes. Text belongs in a comment, a
verdict in `--outcome`. Other ticket kinds still take any resolution.
