# #365: control tests wait for the admin socket without a 5s cap

2026-11-04.

The `dial` helper in `pkg/control/control_test.go` used to poll the socket
for at most 5s. Under `go test -race -count=5` on a 6-core host, about 40
parallel controllers starting at once (migrate, config load, pack load)
took longer than that, and about ten tests failed in `dial`. The wait is
now `dialWait` (2m). It only matters when a test fails: `dial` returns as
soon as the socket answers, so passing runs do not get slower.

With dial fixed, the same run then failed TestTriggers, TestTicketTrace and
TestRateLimitBackoff on their own fixed 5s and 30s `Eventually` waits. Every
fixed-wall-time `Eventually` in `pkg/control` tests now uses `dialWait` too.

A readiness callback on `control.Options` was considered. It would need
all 21 `control.Run` call sites and session sockets to change, and the
test would still need a failure timeout, so it was not done.
