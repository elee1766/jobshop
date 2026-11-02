# Runs (#89)

`jobshop run start <workflow> -v k=v [--key k]` (admin, or a session in
the workflow's project) checks vars (every workflow var set, no others),
fills step titles and prompts from pack vars plus run vars, and calls
`CreateGraph` once: a `run` ticket (no pool, project of the first step's
pool, `fields {workflow, pack, vars, key, state: running}`), one `step`
ticket per step (`parent_id` = run, step pool, body = prompt, `fields
{run, step, index, retries, failures: 0}`), and `blocks` edges step i+1 on
step i. With `issue`, run and steps get `meta.github.issue` (a number when
numeric). Ids come from `Store.NewID` so edges are wired before insert.
With `--key`, an open run of the same workflow and key is returned
(`existing: true`); starts are serialized in the controller. `step` is a
claimable kind, so pool sessions claim steps like tasks; `ticket create`
rejects `run` and `step`.

When a step closes through `ticket close` the controller closes its run
`done` (state `done`) if every step is closed `done`; each tick also
sweeps running runs for this. `run show <id>` prints the run line, vars,
and one line per step (index, step, id, pool, status, resolution, holder,
failures/retries); `--json` prints the raw result. `run list [--state]`
lists runs newest first. `run cancel` closes open steps and the run
`cancelled`. `run resume [--step s]` reopens the step (default: the
first not closed `done`), resets its failures, comments on it, and sets
the run back to `running`. A step that is closed `failed` or still
unfinished keeps the next step blocked (`90-step-retries-escalation.md`).
