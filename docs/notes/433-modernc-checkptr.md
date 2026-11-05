# CI race tests skip checkptr in modernc (#433)

2026-11-05. `-race` turns on checkptr for every package. modernc.org/sqlite
is C translated to Go with unsafe pointer arithmetic, and checkptr on it was
about a third of pkg/control's race-test CPU. CI now runs
`go test -race -gcflags='modernc.org/...=-d=checkptr=0' ./...`: our packages
keep checkptr, and the race detector still instruments modernc.

pkg/control -race on a 6-CPU sandbox: 21.8s to 15.9s wall, ~69s to ~47s
CPU. Run the same flags locally to match CI timing.
