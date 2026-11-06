# JOB-70: idle sessions with background jobs

2026-11-06

A session that releases its ticket while its own background commands
run looks idle, and ending it (scale down, idle timeout) kills them
with the host's process tree and removes the worktree.

- Controller guard (chosen: keep, bounded). Before an idle session ends
  scaled_down, drained or idle, the runner walks the agent pid's
  descendants in /proc. Anything that is not the agent's own
  executable, the host binary or a sandbox helper (bwrap, pasta) counts
  as a job; between turns an agent runs no tools. While jobs run the
  session stays open, for at most one lease from when they were first
  seen. Past that it ends anyway, with a WARN and a comment on the
  session ticket naming the killed commands.
- `jobshop release` and `jobshop handoff` refuse when the current git
  branch has commits no remote has, unless `--force`. Detached HEADs
  and non-git dirs are not checked.
- The worker prompt says never to release while own background
  commands run, and to push work in progress first.

Tests: `TestScaleDownWaitsForBackgroundJobs`, `TestCheckPushed`.
