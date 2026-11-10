# run list --workflow and --key (JOB-183)

2026-11-10

`jobshop run list` takes `--workflow <name>` and `--key <key>` next to
`--project`, `--state` and `--limit`. Both are exact matches on
`fields.workflow` and `fields.key` (the `run.list` params `workflow` and
`key`), they combine with each other and with `--state`, and closed runs
are listed. Results stay newest first. A session socket is still limited
to its own project.

```sh
jobshop run list --workflow pr-work --state running
jobshop run list --key linear-APP-123
```
