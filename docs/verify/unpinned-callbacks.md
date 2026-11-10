# Verify callback runs follow live callbacks (JOB-188)

2026-11-10. Run on a scratch controller (`docs/verify/` runbooks that start
one use the isolated-controller skill), never the live store. Needs a pack
with a callback workflow whose `completed` callback starts a second step
and whose prompt text you can edit, for example `pr-work` or a fixture with
`on_start`/`on_step_completed` in a local `hooks.star`.

## Default: the run follows the deploy

1. Start a run of the workflow and wait for its first step.
2. Edit the callback source (for example the second step's prompt) and
   `jobshop reload`.
3. Close the first step. Expected: the second step appears with the edited
   prompt, the run stays `running`, and `jobshop run show <run>` has one
   comment `Repinned to the live callback manifest (...)`.
4. Edit the callback so it fails on the run's step ids, reload, close the
   next step. Expected: the run goes `needs_human` with the callback's
   error, not `hook manifest changed`, and the escalation body does not
   mention `--repin`.
5. Fix the callback, reload, `jobshop run resume <run>` (no flag). Expected:
   the run continues.

## Pinned: the run stops

1. Set `pin_callbacks: true` on the workflow, reload, start a run.
2. Repeat step 2 and 3 above. Expected: the run goes `needs_human` with
   `hook manifest changed`, no `Repinned` comment, the escalation names
   `--repin`, and plain `jobshop run resume <run>` fails.
3. `jobshop run resume <run> --repin` on the admin socket. Expected: the
   run continues under the new callbacks.

## Config check

`pin_callbacks: true` on a workflow with `steps` and no `on_start` fails
`jobshop config check` with `pin_callbacks needs a callback workflow`.
