# #433 pkg/control tests: hourly tick

2026-11-04. Test controllers from poolRootOpt tick every hour, not 50ms.
Tests that wait for supervisor work use passEventually (pass_test.go),
which runs a pass between checks. On gasserver (6 cores, load ~12)
`go test -race ./pkg/control` went from 195s to 105s. Reaching 15s also
needs fewer sqlite round trips per pass; that part of #433 is still open.
