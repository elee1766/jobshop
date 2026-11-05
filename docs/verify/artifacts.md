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
3. Desk `?view=artifacts`: the nav shows 1 unread; the pane renders the
   markdown and shows `out.csv` inline. `curl -sI
   localhost:<port>/api/artifacts/<id>/files/out.csv` shows
   `X-Content-Type-Options: nosniff` and `Content-Disposition`.
4. Press `a`: the artifact closes `acknowledged`, the unread count drops,
   and `jobshop ticket show <id>` history has a `close` by admin.
5. Prune: with `artifacts.retention: 1m` in `jobshop.yaml` and a
   `jobshop reload`, within an hour the controller log has `artifacts
   pruned`, the dir is gone, and `jobshop artifact show <id>` lists the
   file with `pruned: true`. Restore the retention after.
6. With `gateway.<name>.artifact_room` set, step 1 posts one line in that
   room.
