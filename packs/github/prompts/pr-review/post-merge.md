Step `post-merge` of a `pr-review` run for PR #{{.pr}} in {{.repo}}.

The PR is merged. Find the merge commit in the comments of the run's
`review-merge` step (`jobshop run show <run>` with the run id from this
ticket's `parent_id`), or from
`gh pr view {{.pr}} --repo {{.repo}} --json mergeCommit,closingIssuesReferences,title,body`.

Long commands are moved to the background, so do not wait inside the
turn. Check once, then park on what is pending (a deferral, not a
failed attempt; a plain release is one).

1. Master CI on the merge commit:
   `gh run list --repo {{.repo}} --branch master --commit <sha> --json status,conclusion,databaseId`.
   Pending: `jobshop wait add 'github.checks:repo={{.repo}},sha=<sha>,state=success,workflow={{.ci_workflow}}' 'github.checks:repo={{.repo}},sha=<sha>,state=failure,workflow={{.ci_workflow}}' --any --resume --timeout 1h --reason "master CI on <sha>"`
   and end the turn. Red: file a priority-0 fix issue
   (`gh issue list --search` first) and a dev ticket for it
   (`--pool {{.pool_prefix}}dev -p 0`), comment
   both here, and `jobshop close -r failed`. Cancelled (the wait ends
   `invalidated: checks cancelled`): a newer push superseded that run,
   it is not red. Wait the same way on the newest master commit instead.
   A newer master commit that contains <sha> (`git merge-base
   --is-ancestor <sha> <newer>`) with green CI counts as green: the red
   was fixed after it (infra, a flake or a later fix). Go to step 2 with
   that commit. Red again with the fix already filed (an earlier comment
   here) and no green descendant: do not fail again, park on the fix,
   `jobshop wait add 'run:id=<fix ticket>,state=closed' --resume --timeout 6h --reason "fix for red master"`,
   then wait for CI on the newest master commit.
2. Deploy: is the merge commit live? How to tell: {{.deploy_check}}
   Run the check fresh each time. Not yet: comment what you saw,
   `jobshop wait add 'deploy:repo={{.repo}},sha=<sha>' --resume --timeout 1h --reason "deploy of <sha>"`
   and end the turn. Past the timeout, or when `wait add` says this
   controller reports no build commit, run the check by hand: still not
   live after an hour of green CI, comment what you saw and
   `jobshop close -r failed`.
3. All good: comment on the linked issue, if any, that it shipped in
   <sha> (`gh issue comment <n> --repo <repo>`, or `jobshop linear
   comment <KEY> "Shipped in <sha>"` and `jobshop linear state <KEY>
   "{{.linear_done_state}}"` for a Linear key), then `jobshop close`.
   That finishes the run; a `pr-work` run of another instance waiting on
   it (`run:id=<this run>,state=done`) resumes and finishes its own
   steps.
