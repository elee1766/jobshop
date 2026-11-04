# Unknown pool admin methods are warnings (#325)

2026-11-04. A deploy rolled gasserver back from 6403fd1 to 8148950, which
did not know the `session-config` admin method the operator pack grants
(#306). The pack failed, the last good config (#227) had been saved by
the newer binary and failed the same way, and the controller crash-looped
for 30 minutes. With no controller, no session `cli.sock` existed, so
agents could not comment, close or release.

Now an admin method the binary does not know is dropped from the pool
with a load warning (logged by the controller, printed by
`jobshop config check`). The grant is not given, and everything else in
the config still loads. Methods that were never grantable, such as
`shutdown`, are dropped the same way.

`TestControllerRestart` also dials the session's `cli.sock` after the
restart.
