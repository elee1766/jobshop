# github.checks check= (JOB-200)

2026-11-10. `github.checks` takes `check=<name or glob>`, alias `name=`
(both in one condition is an error), on a commit (`repo=,sha=`) or a PR
(`pr=`). The wait then follows only the checks of that name, not every
check on the head. On a PR the name is the job name (`gh pr checks`), on a
commit a check run name or `<workflow> (workflow)`.

The name matches exactly or as a `path.Match` glob, so `go (*)` covers a
matrix. One name per condition: for several, add one condition each (the
default mode needs all of them).

The matched checks go through the same rules as an unfiltered wait:
pending keeps waiting, `state=failure` is any failed one, cancelled with
nothing pending or failed invalidates the wait, and on a commit success
needs at least one passed check.

| matched checks | note |
| -- | -- |
| none yet | not met, `<check> not reported on <sha>` |
| green | `<check> green on <sha>` |
| failed | `<check> failed on <sha>` |

Validation: `check` applies only to `github.checks`, does not combine with
`workflow=` and does combine with `required=true` (filters the required
list). A malformed glob is rejected. The graph label gets ` (<check>)`.
