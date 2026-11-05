# enabled: false on pools, workflows and triggers (JOB-24)

2026-11-05. A pool, workflow or trigger with `enabled: false` (default
true) is parsed and validated like any other entry, then removed from
the loaded config. No trigger fires, no run starts, no pool spawns
sessions, and no stall check (`ready_unclaimed`, orphans) reports it.
The field is templatable, so a pack can expose it as a var
(`enabled: "{{.flow_on}}"`) and a host can set it per instance; the
host file can also set `enabled` on a pack's entry directly.

References: a live entry that needs a disabled one is a load error naming
both. Checked: trigger `when.pool`, `ticket.pool` and `run.workflow`;
workflow step pools (those filled from pack vars; a pool from a run var
fails at run start with "pool X is disabled"), `escalate_pool` and
`escalate.chat_pool`; hook `pools`, gateway pools, `escalation.chat_pool`
and `attention.prestart_pool`.

Runs in flight when a reload disables an entry: their steps wait. Open
tickets in a disabled pool stay ready and unclaimed until the pool is
enabled again, and are not reported as orphans. Steps of a disabled
workflow keep their tickets; controller-run (command, http) steps of a
disabled workflow do not run, since the workflow is not loaded. No new
run of a disabled workflow starts.

`config show` lists the entries under `disabled`, and `pool list` lists
disabled pools under `disabled`. Code: `pkg/config/disabled.go`.
