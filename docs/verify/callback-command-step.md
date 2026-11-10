# Callback command step

2026-11-10.

`go test ./pkg/orchestration/control -run '^TestDynamicCommand' -count=1`
covers the deferred and interrupted restart policies, arguments passed
as written, a sandboxed worktree and the session update check.

On a host, load a pack whose `on_start` callback returns
`start_step(id="gate", type="command", pool="<pool>", run=["sh", "-c",
"printf 'outcome: ready\\n'"], retries=1)`.

1. `jobshop run show <run>`: step `gate` runs and closes with outcome
   `ready`. Its ticket has no pool, and its fields have `command_spec`
   with the argv and pool.
2. Make the command exit 75 once. The step is deferred by `retry_after`
   with no failure counted. Restart the controller while it is deferred:
   the step runs the recorded command again.
3. From a session socket, `jobshop ticket update <step> --fields` with a
   changed `command_spec` is refused as forbidden. The same update on the
   admin socket is accepted.
4. A callback returning `run=[]`, a blank executable, or `retry_pool` or
   `prompt` with `type="command"` fails the invocation with the
   constructor message (`jobshop automation list --run <run>`).
