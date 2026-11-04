# Command step sandbox (#419)

On a host with a sandboxed pool that has a repo, start a run of a
workflow with a `type: command`, `pool: <pool>`, `worktree: true` step
that runs `pwd; ls /var/lib/jobshop/jobshop.db`.

- `jobshop run show <run>`: the step's output is `<root>/runs/<run>/work`
  and the `ls` fails (the db is hidden), so the step result is failed.
- `git -C <repo> worktree list` shows the run worktree while the run is
  open and not after it closes; `<root>/runs/<run>` is gone too.
