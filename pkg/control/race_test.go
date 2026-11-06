package control_test

import (
	"os"
	"strings"
)

// noRaceExitSleep sets atexit_sleep_ms=0 in GORACE for the processes the
// tests start. The race runtime sleeps 1s at exit by default (to let
// racing goroutines report), and every host, agent and command step a
// test spawns is this test binary, so each child that exits cost a second
// of wall time under -race (JOB-82). Children inherit the env. A value
// already in GORACE is kept. A build without -race ignores GORACE.
func noRaceExitSleep() {
	v := os.Getenv("GORACE")
	if strings.Contains(v, "atexit_sleep_ms") {
		return
	}
	_ = os.Setenv("GORACE", strings.TrimSpace(v+" atexit_sleep_ms=0"))
}
