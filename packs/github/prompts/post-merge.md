Step `post-merge` of a `pr-work` run for issue #${issue}.

The PR for #${issue} is merged. Find the merge commit in the comments of
the run's `review-merge` step (`jobshop run show <run>` with the run id
from this ticket's `parent_id`), or from
`gh pr list --state merged --search "#${issue}" --json number,mergeCommit`.

Releasing this step counts as a failed attempt, so wait inside your turn
(`gh run watch`, `sleep 300` between checks) rather than releasing.

1. Master CI on the merge commit:
   `gh run list --branch master --commit <sha> --json status,conclusion,databaseId`.
   Pending: `gh run watch <id>`. Red: file a priority-0 fix issue
   (`gh issue list --search` first) and a dev ticket for it, comment
   both here, and `jobshop close -r failed`.
2. Deploy: is the merge commit live? How to tell: ${deploy_check}
   Not yet: sleep 5 minutes and check again. Still not after an hour of
   green CI: comment what you saw and `jobshop close -r failed`.
3. Both good: comment on the issue (`gh issue comment ${issue}`) that it
   shipped in <sha> (and is deployed, if step 2 applies), then
   `jobshop close`. That finishes the run.
