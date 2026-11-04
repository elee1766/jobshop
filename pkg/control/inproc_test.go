package control_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	host "github.com/elee1766/acphost"

	"github.com/elee1766/jobshop/pkg/hostspec"
)

// inProcHosts runs session hosts as goroutines of the test process
// (control.Options.Hosts), so controller-logic tests skip the host
// process spawn the race build pays about a second for (#361). The agent
// is still the fake-agent child acphost starts. Tests about real process
// trees (restart and adopt, sandbox, env refresh, stuck and busy) opt
// out with realHosts, see poolRootOpt. Agent crashes work in-process
// through Exited (#467).
type inProcHosts struct {
	mu   sync.Mutex
	runs map[string]chan struct{} // dir -> closed when Main returns
}

func newInProcHosts() *inProcHosts {
	return &inProcHosts{runs: map[string]chan struct{}{}}
}

func (h *inProcHosts) Start(dir string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if done, ok := h.runs[dir]; ok {
		select {
		case <-done:
		default:
			return fmt.Errorf("in-process host for %s still running", dir)
		}
	}
	done := make(chan struct{})
	h.runs[dir] = done
	opts := hostspec.HostOptions(dir)
	go func() {
		defer close(done)
		_ = host.Main(dir, opts...)
	}()
	return nil
}

// Stop sends the host its stop method, as the controller's SIGTERM would
// make it do, and waits for Main to return.
func (h *inProcHosts) Stop(dir string) bool {
	h.mu.Lock()
	done, ok := h.runs[dir]
	h.mu.Unlock()
	if !ok {
		return false
	}
	select {
	case <-done:
		return true
	default:
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		if nc, _, _, err := host.Connect(ctx, dir); err == nil {
			b, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "method": host.MethodStop})
			_, _ = nc.Write(append(b, '\n'))
			_ = nc.Close()
		}
		select {
		case <-done:
			return true
		case <-ctx.Done():
			return true
		case <-time.After(50 * time.Millisecond):
		}
	}
}

// Exited is closed when the host for dir returns (control.HostExits).
func (h *inProcHosts) Exited(dir string) <-chan struct{} {
	h.mu.Lock()
	defer h.mu.Unlock()
	if done, ok := h.runs[dir]; ok {
		return done
	}
	return nil
}

// stopAll stops every host still running, for test cleanup.
func (h *inProcHosts) stopAll() {
	h.mu.Lock()
	dirs := make([]string, 0, len(h.runs))
	for d := range h.runs {
		dirs = append(dirs, d)
	}
	h.mu.Unlock()
	for _, d := range dirs {
		h.Stop(d)
	}
}

// selfPid is skipped by cleanup that signals host pids: an in-process
// host writes the test process's pid.
var selfPid = os.Getpid()
