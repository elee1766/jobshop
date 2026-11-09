Step `post-merge` of a `pr-work` run for Linear issue {{.issue}}.

The PR for {{.issue}} is merged. Find the merge commit in the comments of
the run's `review-merge` step (`jobshop run show <run>` with the run id
from this ticket's `parent_id`), or from
`gh pr list --state merged --search "{{.issue}}" --json number,mergeCommit`.

Long commands are moved to the background, so do not wait inside the
turn. Check once, then park on what is pending (a deferral, not a
failed attempt; a plain release is one).

1. Master CI on the merge commit:
   `gh run list --branch master --commit <sha> --json status,conclusion,databaseId`.
   Pending: `jobshop wait add 'github.checks:repo={{.repo}},sha=<sha>,state=success,workflow={{.ci_workflow}}' 'github.checks:repo={{.repo}},sha=<sha>,state=failure,workflow={{.ci_workflow}}' --any --resume --timeout 1h --reason "master CI on <sha>"`
   and end the turn; it resumes here within a minute of CI finishing.
   Cancelled (the wait ends `invalidated: checks cancelled`): a newer
   push superseded that run, it is not red. Wait the same way on the
   newest master commit instead.
   Red: unless a `ci-repair` run for
   that SHA is open (`jobshop run list`), file a priority-0 dev ticket
   for it (`--pool {{.pack_prefix}}dev -p 0`), comment it here, and
   `jobshop close -r failed`.
   A newer master commit that contains <sha> (`git merge-base
   --is-ancestor <sha> <newer>`) with green CI counts as green: the red
   was fixed after it (infra, a flake or a later fix). Go to step 2 with
   that commit. Red again with the fix already filed (an earlier comment
   here) and no green descendant: do not fail again, park on the fix,
   `jobshop wait add 'run:id=<fix ticket>,state=closed' --resume --timeout 6h --reason "fix for red master"`,
   then wait for CI on the newest master commit.
2. Deploy: is the merge commit live? How to tell: {{.deploy_check}}
   Run the check fresh on every loop. Do not reuse an earlier result or
   version, the deploy may have moved since. Not yet: comment what you
   saw, `jobshop wait add 'deploy:repo={{.repo}},sha=<sha>' --resume --timeout 1h --reason "deploy of <sha>"`
   and end the turn; it resumes when the controller runs a build
   containing <sha>. Past the timeout, or when `wait add` says this
   controller reports no build commit, run the check by hand: still not
   live after an hour of green CI, comment what you saw and
   `jobshop close -r failed`.
3. Both good: re-read the issue (`jobshop linear issue {{.issue}}`). If its
   completion gate or acceptance list is not met by what has merged (this
   PR was a slice), do not move it to {{.linear_done_state}}: comment what
   shipped in <sha> and what remains, keep the issue open and the run's
   follow-up work tracked (a wait or a new run, per the issue), then
   `jobshop close`. Otherwise comment on the issue that it shipped in
   <sha> (and is deployed, if step 2 applies):
   `jobshop linear comment {{.issue}} "Shipped in <sha>"`, move it to
   {{.linear_done_state}} with
   `jobshop linear state {{.issue}} "{{.linear_done_state}}"`, then
   `jobshop close`. That finishes the run.
