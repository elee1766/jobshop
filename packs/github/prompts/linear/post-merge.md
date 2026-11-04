Step `post-merge` of a `pr-work` run for Linear issue {{.issue}}.

The PR for {{.issue}} is merged. Find the merge commit in the comments of
the run's `review-merge` step (`jobshop run show <run>` with the run id
from this ticket's `parent_id`), or from
`gh pr list --state merged --search "{{.issue}}" --json number,mergeCommit`.

Long commands are moved to the background, so do not wait inside the
turn. Check once; if CI or the deploy is pending, comment what you saw
and `jobshop release --until 20m` (a deferral, not a failed attempt;
a plain release is one).

1. Master CI on the merge commit:
   `gh run list --branch master --commit <sha> --json status,conclusion,databaseId`.
   Pending: release --until as above. Red: unless a `ci-repair` run for
   that SHA is open (`jobshop run list`), file a priority-0 dev ticket
   for it, comment it here, and `jobshop close -r failed`.
2. Deploy: is the merge commit live? How to tell: {{.deploy_check}}
   Run the check fresh on every loop. Do not reuse an earlier result or
   version, the deploy may have moved since. Not yet: release --until
   and check again. Still not after an hour of
   green CI: comment what you saw and `jobshop close -r failed`.
3. Both good: comment on the issue that it shipped in <sha> (and is
   deployed, if step 2 applies):
   `jobshop linear comment {{.issue}} "Shipped in <sha>"`, move it to
   {{.linear_done_state}} with
   `jobshop linear state {{.issue}} "{{.linear_done_state}}"`, then
   `jobshop close`. That finishes the run.
