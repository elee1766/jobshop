# Repin a callback run after a deploy (#1012)

2026-11-08. A callback workflow run pins its callbacks (entries, code,
grants, instance, reference scope) at start. A deploy that rewrites the
pack (#1005) replaces the manifest, so open runs fail their next callback
with `hook manifest changed`, go `needs_human`, and `run resume` refuses
because the old manifest is gone.

Built: `jobshop run resume <run> --repin` (RPC `run.resume`, `repin: true`).

- Admin socket only: it changes the code and authority the run executes
  under. Worker sessions get `forbidden`.
- Sets the run's `automation` entries, digest and reference scope to the
  live ones and comments `Repinned to the live callback manifest (<new>,
  was <old>) by <actor>.` on the run. Applied batches and completed steps
  are kept, so the run continues where it stopped.
- Then resumes as usual (retries the failed invocation, or reopens the
  first unfinished step). A run already on the live manifest is unchanged.
- Fails for a workflow that is no longer a loaded callback workflow:
  cancel that run.
- The escalation ticket body for a callback run whose pin is stale names
  `--repin`.

Not done, on purpose: no automatic repin. A plain resume and a reload
still refuse a changed manifest, so a code bug in a pinned callback is not
swapped out without an operator choosing it. A callback run is only
closed by its `finish_run` action, so a run whose steps are all done but
whose final callback could not run still needs one `--repin` resume.

Test: `TestHookRunResumeRepin` (pkg/control/automation_test.go).
