package control_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// TestTriggerCheckOutsideTick (JOB-109): a trigger check that takes long
// does not hold the supervisor pass. With the check blocked, a pass
// finishes, a reload returns, and another trigger's check runs and fires;
// when the slow check ends, its trigger fires.
func TestTriggerCheckOutsideTick(t *testing.T) {
	root, start := poolRoot(t, "1m")
	fifo := filepath.Join(root.Dir, "hold")
	require.NoError(t, syscall.Mkfifo(fifo, 0o600))
	cfg := fmt.Sprintf(reloadPassYAML, os.Args[0], root.Dir, fmt.Sprintf(`  a-slow:
    when: {check: "read x < %s", check_timeout: 5m}
    ticket: {title: slow-fired, pool: dev}
  b-fast:
    when: {check: "true"}
    ticket: {title: fast-fired, pool: dev}
`, fifo))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	titles := func() []string {
		t.Helper()
		var ts []domain.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: domain.KindTask}, &ts))
		var out []string
		for _, tk := range ts {
			out = append(out, tk.Title)
		}
		slices.Sort(out)
		return out
	}

	// Before JOB-109 this pass ran a-slow's check and never finished.
	pass(t, root.Dir, 1)

	// Opening the fifo for writing returns once the check opened it to
	// read: it is running, and the pass above is already over.
	opened := make(chan *os.File, 1)
	go func() {
		f, err := os.OpenFile(fifo, os.O_WRONLY, 0)
		if err != nil {
			f = nil
		}
		opened <- f
	}()
	var w *os.File
	select {
	case w = <-opened:
		require.NotNil(t, w, "open fifo")
	case <-time.After(dialWait):
		t.Fatal("the slow check never started")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	require.NoError(t, admin.Call(ctx, rpc.MethodReload, struct{}{}, nil), "reload returns while a check runs")
	passUntil(t, root.Dir, func() bool { return slices.Contains(titles(), "fast-fired") }, "the fast trigger fires while the slow check runs")
	require.NotContains(t, titles(), "slow-fired", "the slow check has not ended")

	_, err := w.WriteString("go\n")
	require.NoError(t, err)
	require.NoError(t, w.Close())
	passUntil(t, root.Dir, func() bool { return slices.Contains(titles(), "slow-fired") }, "the slow trigger fires once its check ends")
	require.Equal(t, []string{"fast-fired", "slow-fired"}, titles())
}

// TestTriggerCheckRunsOnce (JOB-109): while a trigger's check runs, later
// passes do not start it again.
func TestTriggerCheckRunsOnce(t *testing.T) {
	root, start := poolRoot(t, "1m")
	fifo := filepath.Join(root.Dir, "hold")
	require.NoError(t, syscall.Mkfifo(fifo, 0o600))
	runs := filepath.Join(root.Dir, "runs")
	cfg := fmt.Sprintf(reloadPassYAML, os.Args[0], root.Dir, fmt.Sprintf(`  slow:
    when: {check: "echo x >> %s; read x < %s; exit 1", check_timeout: 5m}
    ticket: {title: never, pool: dev}
`, runs, fifo))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	stop := start()
	defer stop()

	pass(t, root.Dir, 1)
	opened := make(chan *os.File, 1)
	go func() {
		f, err := os.OpenFile(fifo, os.O_WRONLY, 0)
		if err != nil {
			f = nil
		}
		opened <- f
	}()
	var w *os.File
	select {
	case w = <-opened:
		require.NotNil(t, w, "open fifo")
	case <-time.After(dialWait):
		t.Fatal("the check never started")
	}
	pass(t, root.Dir, 3)
	b, err := os.ReadFile(runs)
	require.NoError(t, err)
	require.Equal(t, "x\n", string(b), "three passes while the check ran started it once")

	_, err = w.WriteString("go\n")
	require.NoError(t, err)
	require.NoError(t, w.Close())
}
