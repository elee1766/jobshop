# Triggers (#26)

`triggers:` entries have `when`, a `ticket` template and an optional
`cooldown`. Each tick the controller fires a trigger when all hold: no open
ticket it created (`meta.trigger.name`), cooldown passed since its last
ticket closed (a last ticket read with no `closed_at`, because it closed
mid-evaluation, waits for the next tick; #47, 2026-11-01), cron slot due, every count threshold met. Thresholds are
`{ready,open,in_progress}_{below,above}` over `when.pool`, `when.kind`
(default task) and `when.label`. State is read back from the store, so
restarts change nothing. Cron is 5-field, UTC unless `when.timezone`;
missed slots collapse into one fire, and a new trigger does not fire for
past slots. `jobshop trigger list` shows whether each would fire and why.
`when.check` (#34) is a `sh -c` command run in the root dir after the
store conditions hold; exit 0 holds. Non-empty trimmed stdout is a dedup
key (`meta.trigger.key`, prefixed with the trigger name): one ticket per
key, ever. `check_timeout` defaults to 1m; a timeout logs an error.
A trigger may start a workflow run instead of filing a ticket (`98-run-triggers.md`).
