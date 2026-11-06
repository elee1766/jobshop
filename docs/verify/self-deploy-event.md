# Self-deploy starts within a minute of CI success (#826)

On gasserver, with the repo's webhook sending `workflow_run` events:

1. Merge a PR to master. Note the time CI finishes.
2. Within a minute the controller log has `trigger fired
   trigger=self-deploy event=workflow_run`, and `journalctl -u
   jobshop-self-deploy` shows a run for the merge commit (no wait for
   the timer).
3. `jobshop ticket list --label trigger-command` shows the start as a
   closed message ticket.
4. Merge two PRs close together: the second deploys on the next start or
   timer run, and `deployed` ends at the second commit.
5. Break the path once (remove the polkit rule): the start ticket stays
   open with needs-human and the polkit error as its comment.
