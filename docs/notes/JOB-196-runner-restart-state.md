# Runner state across a controller restart

2026-11-10 (JOB-196).

A controller restart builds a new `runner` for each open session and adopts
its host. Before this, `delivered`, `backoffUntil`, `seed`, `resumed` and
`systemDue` lived only in memory. A chat that had talked for an hour took its
pool prompt turn again (and its `at: before/after` rules, `inject: always`
ones included), a rate limit backoff was forgotten and the next prompt hit
the limit again, and a respawn that was waiting for its first prompt lost
the recovery note and the journal summary.

Every field of `runner` now has a comment saying what a restart does with
it. In short:

| Field | After a restart |
| --- | --- |
| `injected`, `acpSession`, `traceParent`, `model`/`effort` | stored (`Injected`, `ACPSession`, `TraceParent`, `Model`/`Effort`) |
| `backoffUntil` | stored (`BackoffUntil`); the runner waits it out before claiming or prompting (`waitBackoff`) |
| `delivered` | stored (`Delivered`); zeroed with a new conversation |
| `lastClaim` | stored (`LastClaim`); the scaler reads it from the fields, so the least recently used order survives |
| `systemDue` | stored (`SystemDue`); set on a spawn, cleared by the turn that carries the rules |
| `resumed` | stored (`RecoveryDue`); cleared by the prompt that carries the note |
| `seed` | the mark is stored (`SeedDue`); the text is rebuilt from the host journal or the recall index |
| `agent`, `projectName`, `pool`, `dir`, `agentPid` | recomputed from the ticket, the config or the host's welcome |
| `unclosed`, `crashes`, `jobsSince`, `skipped`, `restartPending`, `cancelNow`, `inTurn`, `turning`, turn scratch (`turnRules`, `turnTicket`, `context`, `promptAttachments`, `promptPresentation`, `lastStop`, `traffic`) | reset on purpose |

Reset on purpose, and why: a controller restart is not a host crash
(`crashes`), the ticket's own history bounds an unclosed loop
(`releasedUnclosedBefore`), a restart request not yet served is asked again
by the operator (a stale env is caught by `staleEnv`), and the rest are per
turn and set and used inside one turn.

The idle window (`idleSince`) also restarts at adopt, so a chat gets a full
idle timeout after a controller restart.

`saveFields` keeps the stored values and the three marks (it already kept
`Injected`), since the runner's copy of the fields is older than what the
helpers wrote. A backoff cleared by a stopping controller (context done) is
not written back: the stored one is what the next controller must honor. A
stored backoff already past at adopt is dropped.

Checks: `TestRestartKeepsBackoff` (no prompt while the stored backoff runs,
then the ticket is retried and the field cleared),
`TestRestartKeepsChatPoolPrompt` (no pool prompt or `inject: always` before
rule on the second chat message after a restart),
`TestRestartKeepsRecoveryAndSummary` (a restart between a respawn with a
failed load and the next prompt keeps the recovery note and rebuilds the
summary). Without the matching reads in `startRunner` each of them fails.
