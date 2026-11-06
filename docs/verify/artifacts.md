# Artifacts (#508)

After a deploy that includes #508, on gasserver:

1. In a live session of a repo pool, write `report.md` and `out.csv` in
   the worktree and run `jobshop artifact deliver "Verify artifacts"
   --file report.md --attach out.csv`. It prints the artifact with
   `meta.artifact.ticket` set to the held ticket and one file.
   `ls /var/lib/jobshop/artifacts/<id>/` shows `out.csv`.
2. `jobshop artifact deliver x --attach /etc/hostname` from the same
   session is refused (outside the session dir), and nothing new is in
   `/var/lib/jobshop/artifacts/`.
3. `curl -s localhost:<port>/api/artifacts?unread=1` lists it. `curl -sI
   localhost:<port>/api/artifacts/<id>/files/out.csv` shows
   `X-Content-Type-Options: nosniff` and `Content-Disposition`. (Desk has
   no built-in Artifacts view since JOB-80.)
4. `jobshop artifact ack <id>`: the artifact closes `acknowledged`, it
   leaves the unread list, and `jobshop ticket show <id>` history has a
   `close` by admin.
5. Prune: with `artifacts.retention: 1m` in `jobshop.yaml` and a
   `jobshop reload`, within an hour the controller log has `artifacts
   pruned`, the dir is gone, and `jobshop artifact show <id>` lists the
   file with `pruned: true`. Restore the retention after.
6. With `gateway.<name>.artifact_room` set, step 1 posts one line in that
   room.
