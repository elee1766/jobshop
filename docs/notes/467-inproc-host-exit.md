# 467: in-process hosts report agent exits

2026-11-04. Follow-up to #433.

An in-process test host (acphost.Main in a goroutine) returns when its
agent exits, but the connection it accepted from the controller stays
open: a host process would have its fds closed by exit. The controller
never saw the drop, so crash tests needed real host processes.

`control.HostExits` is an optional `HostRunner` method, `Exited(dir)`,
closed when that dir's host returns. The runner closes its host
connection on it, then sees the drop as a host exit as usual. Real
hosts do not implement it. TestQuarantine now runs in-process: about
15s to 2.4s alone under -race.

## 2026-11-09 update (#1133)

acphost v0.22.4 closes every accepted client connection when `Main`
returns, so a client reads EOF as it would from a real host exit. The
`control.HostExits` interface, the controller goroutine that closed the
connection and `InProcHosts.Exited` are removed. `docs/plan.md` is
frozen and still mentions `control.HostExits`.
