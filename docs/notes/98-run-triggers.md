# Run triggers (#98)

A trigger has exactly one of `ticket:` and `run: {workflow, vars, key}`.
Load fails on an unknown workflow, a var the workflow lacks, or a
workflow var set by neither the trigger nor pack vars. `vars` and `key`
are templates whose only var is `${check_output}`, the first line of the
check's trimmed stdout (allowed only with `when.check`); pack vars are
substituted at load as elsewhere. Firing calls the `run start` code path
as the controller, with the key, so an open run with that key is reused
instead of a second one. The run ticket carries `meta.trigger`, so the
open run holds the trigger, its close starts the cooldown, and the check
key dedups exactly as for ticket triggers. Steps do not carry the
trigger meta.
