# Runner live connection and ACP traffic (JOB-154, 2026-11-08)

Slice 4 of JOB-146. JOB-146 stays open. No behavior change. The coupling
analysis and boundary are the dated JOB-154 comment on #569.

## Moved into `pkg/control/sessionrt`

| File | Owns |
|---|---|
| `connect.go` | `Connect`: the host dial poll (stops early on a spawn failure or an exited host) |
| `errors.go` | `UnknownSession`, `RateLimited`, `ProviderUnreachable`, `AgentBusy`: classify ACP prompt errors |
| `client.go` | `Client` (the ACP client side: permission answer, tool and text updates, turn-end signal, unsupported file and terminal methods) and `Traffic` (open tool calls and the agent text tail of the running turn) |

## Adapters and synchronization

- `Client.Refuse` is how the controller denies tool kinds to a read-only
  attention chat (`runner.client` builds it from `isReadOnly` and
  `chatlife.ReadOnlyTool`). `Client` has no store, ticket or pool access.
- `Client.TurnEnd` is the runner's size-1 `turnEnd` channel, written
  without blocking.
- `Traffic` has its own mutex. ACP callbacks run on the SDK reader
  goroutine and no longer take `runner.mu`. `runner.turn` calls
  `Traffic.Reset` before the prompt (after clearing `lastStop` under
  `mu`); `session list` reads `Running`, `deliver` and the stuck reason
  read `Said`. No code holds `runner.mu` and the `Traffic` lock together.

## Stays in `session.go`

Ticket claims and `afterTurn`, prompt composition and rule dedupe, crash
budget, `finish`, every durable transition, `applyConfig`, and the retry
loop of `turn` and `session()`, which interleave ACP calls with `stopHost`
and errRefresh. Candidates for a later slice: a connection supervisor that
owns the three watcher goroutines and the shutdown join, and the cancel
plumbing of `interruptible`.
