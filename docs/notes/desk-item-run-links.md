# Run links on Linear and GitHub items (2026-11-10)

Follow-up to JOB-163. "linked work" on a Desk Linear issue or GitHub PR lists
the Jobshop runs for it, and each row opens the run's trace.

Before this, a run was listed only when a ticket's `meta.github` or
`meta.linear` named the item. Runs started by triggers and callbacks have no
such meta, so most PRs showed no runs.

A run now links to an item when any of these name it (`activityLinks` in
`pkg/orchestration/control/activityhost.go`):

- `meta.github.repo` with `issue` or `pr`, or `meta.linear.issue`, on the
  run, one of its steps or a task.
- The run's `pr` var, with its `repo` var or its workflow's `repo`.
- A GitHub issue/PR URL or Linear issue URL in the ticket title. Run titles
  link their source (JOB-133, JOB-135).

`browse.show` also lists the runs of the work an item is tied to, after its
own, with the relation `via <ref>`: a Linear issue's attached PRs, and the
issues a PR closes or mentions (`browseRuns` in `browsehost.go`).

The activity feed's ticket links use the same map, so feed items gain the
same runs.

Not covered: a run that names its PR only in a `subject` var
(`<n>@<sha>`) and has no link in its title.
