# Command step sandbox and run worktree (#419)

2026-11-04. Built on command steps (#396).

- A command step whose `pool` is sandboxed (#262) runs under that pool's
  bubblewrap sandbox: the same ro/rw/hide mounts and env allowlist as a
  session, with `runs/<run>/` in place of the session dir as its own
  writable dir. The step env (`JOBSHOP_VAR_*`, `JOBSHOP_STEP_KEY`,
  `JOBSHOP_TICKET`) and the controller's PATH are set inside it.
- `worktree: true` (command steps with a pool whose `repo` is set) runs
  the step in `runs/<run>/work`, a locked detached worktree of the pool
  repo at the pool `ref`. All worktree steps of a run share it, so a
  later step sees an earlier one's files.
- The run dir and worktree are removed when the run closes (done or
  cancelled). The run sweep also removes dirs of closed or missing runs,
  for a close that raced an attempt or a restart in between.
- Steps without a pool, or on an unsandboxed pool without `worktree`,
  run as before in the pool dir, repo or root.
