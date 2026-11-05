# Linear intake skips tracking issues (#670)

2026-11-05.

The github pack's default `linear_skip_labels` is now
`Blocked,tracking,target,proposed` (target and proposed from JOB-4). Without it, a Linear tracking issue in Todo with
no open run (JOB-3) got its own pr-work-linear run. GitHub intake
already skipped `tracking` issues. A host that sets
`linear_skip_labels` itself has to add `tracking` to its value.
