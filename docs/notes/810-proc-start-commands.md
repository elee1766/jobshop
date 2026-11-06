# #810: controller commands run through acphost proc.Start

2026-11-06.

Hooks, admission hooks, prompt_rules commands, Desk exec, trigger checks
and command steps now start through acphost `proc.Start`
(`runIsolated` in `pkg/control/isolatedcmd.go`) instead of `os/exec`.
Every one runs in its own process group, which is killed on timeout and
when the controller dies. `proc.Exit` gives the result; a `SetupError`
(bwrap, user namespace or cgroup failure) is an error, so a hook fails
closed.

Which pool's isolation applies:

- hooks and prompt_rules: the session's pool (the event's `pool`). A pool
  with a sandbox, an isolation user or limits runs the command inside
  them, in a scratch dir under `<root>/tmp` that is removed after.
- command steps: the step's pool, with its run dir bound rw. Before this
  they got only the bwrap argv, with no user mapping or cgroup.
- events with no pool, Desk exec, trigger checks and admission hooks: no
  sandbox, root dir as before.

Test: `TestHookSandboxed` (needs working bwrap; skips otherwise).
