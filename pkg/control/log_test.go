package control_test

import (
	"log/slog"
	"testing"
)

// testLog writes to the test's output, so controller logs show only for
// a failing test or with -v (JOB-49).
func testLog(t *testing.T) *slog.Logger {
	return slog.New(slog.NewTextHandler(t.Output(), nil))
}
