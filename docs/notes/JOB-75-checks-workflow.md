# github.checks workflow= (JOB-75)

2026-11-06. `github.checks` takes `workflow=<name or file>` (`ci`,
`ci.yml` or `.github/workflows/ci.yml`), on a commit (`repo=,sha=`) or a
PR (`pr=`). The wait then follows the newest run of that workflow on the
commit (`actions/runs?head_sha=`, highest run id; a PR uses its head)
instead of every check run. Without it, one finished workflow (release)
could make a commit read green before `ci` had queued its jobs.

| newest run of the workflow | result | note |
| -- | -- | -- |
| none | not met | `<wf> not started on <sha>` |
| queued, in progress | not met | `<wf> in progress on <sha>` |
| success | `state=success` met | `<wf> green on <sha>` |
| failure | `state=failure` met | `<wf> failed on <sha>` |
| cancelled | invalidated, the agent resumes now | `<wf> cancelled on <sha>` |
| skipped | invalidated | `<wf> skipped on <sha>` |

`required` does not combine with `workflow=` (validation error).

Without `workflow=`, a cancelled check is no longer a failure on a PR
either (#837 did this for commits): when nothing is pending or failed
and something was cancelled, the wait is invalidated with `checks
cancelled on <sha>`. Skipped checks count for neither side.

The implement prompts (PR head) and the post-merge prompts (master
commit) pass `workflow={{.ci_workflow}}`, default `ci`.
