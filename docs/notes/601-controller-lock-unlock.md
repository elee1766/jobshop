# Controller lock released with LOCK_UN (#601)

2026-11-05. `lock` in pkg/control held the controller lock with flock and
released it by closing the file. A flock belongs to the open file
description: a child forked by another goroutine (sessions, hooks, git)
between lock and unlock keeps a copy of the descriptor until its exec
closes it, so the close alone could leave the lock held. A controller
restarted in the same process then failed with "another controller holds
controller.lock" (TestTriggers flaked this way in CI).

Unlock now calls `flock(LOCK_UN)` before close, which releases the lock
for every descriptor sharing it. TestLockReleasedWhileFdShared dups the
lock fd to stand in for the child and checks the lock can be taken again.
