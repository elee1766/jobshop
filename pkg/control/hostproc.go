package control

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"time"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/acphost/pkg/proc"
)

// hostStopWait is how long stopHost waits for a host after SIGTERM before
// it sends SIGKILL.
const hostStopWait = 10 * time.Second

// A session host records itself as a proc supervisor (#904): its pid in
// supervisor.pid, and supervisor.lock flocked while it runs. acphost
// StartDetached and in-process hosts both write them.

// hostAlive reports whether a host runs for the session in dir.
func hostAlive(dir string) bool {
	d, err := proc.Open(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return legacyHostAlive(dir)
	}
	return err == nil && d.Running()
}

// hostPid is the pid of the host recorded in dir, running or not, or 0.
func hostPid(dir string) int {
	d, err := proc.Open(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return legacyHostPid(dir)
	}
	if err != nil {
		return 0
	}
	return d.Pid
}

// stopHost asks a host to stop (it forwards SIGTERM to the agent's group),
// waits for it on its lock, and kills it if it is still running after
// hostStopWait. Then it stops what is left of the session's processes.
func (c *Controller) stopHost(dir string) {
	if c.opt.Hosts != nil && c.opt.Hosts.Stop(dir) {
		return
	}
	if stopDetachedHost(dir, hostStopWait) {
		stopSessionProcs(dir)
	}
}

// stopDetachedHost stops the host recorded in dir: SIGTERM through
// proc's Stop, a wait on supervisor.lock, SIGKILL after wait. It reports
// whether a host was running.
func stopDetachedHost(dir string, wait time.Duration) bool {
	d, err := proc.Open(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return stopLegacyHost(dir, wait)
	}
	if err != nil || d.Pid <= 1 || !d.Running() {
		return false
	}
	_ = d.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	// The exit record may be missing (a host killed before writing it):
	// only whether it still runs matters here.
	_, _ = d.Wait(ctx)
	cancel()
	if d.Running() {
		_ = syscall.Kill(d.Pid, syscall.SIGKILL)
	}
	return true
}

// Hosts started by a jobshop built with acphost before v0.17 have no
// supervisor.pid, only host.pid and host.lock. One may still run after
// the deploy that switched to StartDetached (#904); these keep it
// visible and stoppable. Drop them once no such host is left.

func legacyHostAlive(dir string) bool { return host.Alive(dir) }

func legacyHostPid(dir string) int {
	b, err := os.ReadFile(filepath.Join(dir, host.PidFile))
	if err != nil {
		return 0
	}
	var pid int
	if _, err := fmt.Sscan(string(b), &pid); err != nil {
		return 0
	}
	return pid
}

// stopLegacyHost is stopDetachedHost for a legacy host: SIGTERM, poll
// host.lock, SIGKILL after wait. It reports whether a host was running.
func stopLegacyHost(dir string, wait time.Duration) bool {
	pid := legacyHostPid(dir)
	if pid <= 1 || !host.Alive(dir) {
		return false
	}
	_ = syscall.Kill(pid, syscall.SIGTERM)
	deadline := time.Now().Add(wait)
	for host.Alive(dir) && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if host.Alive(dir) {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
	return true
}
