# Stall trigger and stall-check (#143)

2026-11-02. Built for #82 after three idle stretches on gasserver that
only the operator noticed.

**Trigger.** `when.stall` is a list of `{signal, after}` or `{signal,
check}`. A stall trigger has no other condition and needs `run:`; it
sets the run key itself to `stall:<signal>:<subject>`, so one stall has
at most one open run. Run vars may use `${stall_signal}`,
`${stall_subject}` and `${stall_detail}`. `cooldown` counts from the last
run of the same key closing; `max_runs` (default 3) caps runs per key,
after which the last run's escalation stands. At most 3 runs of one stall
trigger are open at a time, so a backlog is worked a few at a time.
`trigger list` shows the keys that would fire (`stalls`) or why none do.

**Store signals** (no shell, computed each tick). "Activity" is the
newest of creation, history rows, comments and `defer_until`; heartbeats
write no history, so a held ticket nobody works on still ages.

| Signal | Subject | Holds when |
| --- | --- | --- |
| `ready_unclaimed` | ticket | ready in a configured, non-chat pool, no activity for `after`. Tickets labeled `needs-human` or `blocked` are left to their own signals |
| `pool_quarantined` | `<pool>@<pause end>` | pool paused for crash-looping; one stall per pause |
| `step_stalled` | step | first unfinished step of a running run is held, no activity for `after` |
| `session_silent` | session | open session holds a ticket; newest of journal tail, claim activity and session creation older than `after` |
| `needs_human` | run or ticket | `needs_human` run, or open `needs-human` ticket that is not a run escalation, no activity for `after` |
| `blocked` | ticket | open ticket labeled `blocked`, no activity for `after` |

Any other signal name runs its `check` (`sh -c` in the root dir, bounded
by `when.check_timeout`): exit 0 with one stall per stdout line,
`<subject> <detail>`, a leading `#` stripped. For `issue_no_run` the
controller drops issues that an open run carries in `meta.github.issue`.

Runs, steps and escalation tickets of the trigger's own workflow are
never subjects, so stall-check does not chase itself.

**Pack.** `packs/github` adds the `stall` trigger and the `stall-check`
workflow (one `diagnose` step in `plan`, 1 retry, `prompts/stall.md`),
and drops `blocker-check`. gh checks cover `issue_no_run` (open work
issues older than 30m), `pr_unmerged` (non-draft, all checks green,
untouched 2h) and `no_merge` (no commit on `ci_branch` for 6h while PRs
are open; subject is the last commit). Thresholds are pack vars
`stall_after` (1h), `stall_step_after` (3h), `stall_human_after` (4h).
The plan pool gets `session-list`, `session-restart` and `trigger-list`
grants for the diagnosis. The planner labels blocked claims `blocked`
instead of starting blocker-check; a claim only on an issue is started
by hand with `--key stall:blocked:<n>`.

Escalation goes through the existing path: closing `diagnose` failed sets
`needs_human`, files "Escalate run <id>" with the step history, and the
gateway opens a Matrix thread for it (#125).

Drill: `docs/verify/stall-drill.md`.
