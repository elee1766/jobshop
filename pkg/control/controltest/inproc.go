package controltest

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	host "github.com/elee1766/acphost"

	"github.com/elee1766/jobshop/pkg/hostspec"
)

// InProcHosts runs session hosts as goroutines of the test process
// (control.Options.Hosts), so controller-logic tests skip the host
// process spawn the race build pays about a second for (#361). The agent
// is still the child process acphost starts. Tests about real process
// trees (restart and adopt, sandbox, env refresh, stuck and busy) use
// detached hosts instead. Agent crashes work in-process through Exited
// (#467).
//
// Not built on acphost's acphosttest (#723): the controller calls Start
// from its own goroutine, where acphosttest.Start's t.Fatalf is not
// allowed; acphosttest.Start rewrites the spec the controller already
// wrote; and it blocks until the socket listens, while the controller
// polls the socket itself. Stop and Exited here match acphosttest's
// Stop and Done.
type InProcHosts struct {
	mu   sync.Mutex
	runs map[string]chan struct{} // dir -> closed when Main returns
}

// NewInProcHosts returns a runner with no hosts.
func NewInProcHosts() *InProcHosts {
	return &InProcHosts{runs: map[string]chan struct{}{}}
}

// Start runs host.Main for dir in a goroutine (control.HostRunner).
func (h *InProcHosts) Start(dir string) error {
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
func (h *InProcHosts) Stop(dir string) bool {
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
func (h *InProcHosts) Exited(dir string) <-chan struct{} {
	h.mu.Lock()
	defer h.mu.Unlock()
	if done, ok := h.runs[dir]; ok {
		return done
	}
	return nil
}

// StopAll stops every host still running, for test cleanup.
func (h *InProcHosts) StopAll() {
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
