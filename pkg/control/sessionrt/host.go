// Package sessionrt is the process and isolation mechanics under a session
// host (JOB-146): host liveness and stop, the processes an agent leaves
// behind, the shell probe, cgroup setup and limits, POSIX ACL grants for
// a mapped uid, and starting a detached host. It builds on acphost's proc,
// cgroup, userns and sandbox packages and holds no ticket, pool or
// session state: callers pass directories, ids and specs. The controller
// decides when to call it and keeps every durable transition.
package sessionrt

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"syscall"
	"time"

	"github.com/elee1766/acphost/pkg/proc"
)

// HostStopWait is how long StopDetached waits for a host after SIGTERM
// before it sends SIGKILL, as the controller calls it.
const HostStopWait = 10 * time.Second

// procWait is how long StopProcs waits for processes to exit after SIGTERM.
const procWait = 5 * time.Second

// A session host records itself as a proc supervisor (#904): its pid in
// supervisor.pid, and supervisor.lock flocked while it runs. acphost
// StartDetached and in-process hosts both write them.

// HostAlive reports whether a host runs for the session in dir.
func HostAlive(dir string) bool {
	d, err := proc.Open(dir)
	return err == nil && d.Running()
}

// HostPid is the pid of the host recorded in dir, running or not, or 0.
func HostPid(dir string) int {
	d, err := proc.Open(dir)
	if err != nil {
		return 0
	}
	return d.Pid
}

// StopDetached stops the host recorded in dir: SIGTERM through proc's
// Stop (the host forwards it to the agent's group), a wait on
// supervisor.lock, SIGKILL after wait. It reports whether a host was
// running.
func StopDetached(dir string, wait time.Duration) bool {
	d, err := proc.Open(dir)
	if err != nil || d.Pid <= 1 || !d.Running() {
		return false
	}
	_ = d.Stop()
	waitHostExit(dir, wait)
	if d.Running() {
		_ = syscall.Kill(d.Pid, syscall.SIGKILL)
	}
	return true
}

// waitHostExit blocks until the supervisor in dir releases supervisor.lock
// or wait passes. It is proc.Detached.Wait without the exit record, which
// StopDetached does not need, and without its race when the timeout wins:
// Wait closes the lock file while its flock goroutine still reads the fd
// (acphost#227). Here the goroutine owns the file and closes it when the
// lock is granted, which the supervisor's exit (or SIGKILL) always brings.
func waitHostExit(dir string, wait time.Duration) {
	f, err := os.Open(filepath.Join(dir, proc.LockFile))
	if err != nil {
		return
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() { _ = f.Close() }()
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_SH)
	}()
	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case <-done:
	case <-t.C:
	}
}

// StopProcs stops processes that still carry the session's env after its
// host is gone. Agents may start a detached daemon (jcode `acp --socket`
// runs tools in a `jcode serve` outside the host's process group); it
// keeps the env it started with, so without this a refreshed session
// still runs tools with the old PATH (#111).
func StopProcs(dir string) {
	if dir == "" {
		return
	}
	pids := sessionProcs(dir)
	for _, p := range pids {
		_ = syscall.Kill(p, syscall.SIGTERM)
	}
	deadline := time.Now().Add(procWait)
	for len(pids) > 0 && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
		pids = slices.DeleteFunc(pids, procGone)
	}
	for _, p := range pids {
		_ = syscall.Kill(p, syscall.SIGKILL)
	}
}

// procGone reports whether pid has exited. A zombie counts as gone:
// kill(pid, 0) succeeds on it until its parent reaps it, and a detached
// daemon's parent (init or a subreaper) may take a while, which held
// every env refresh for the full 5s deadline (#433).
func procGone(pid int) bool {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return syscall.Kill(pid, 0) != nil
	}
	// The state follows the parenthesized command, which may hold spaces.
	i := bytes.LastIndexByte(b, ')')
	return i >= 0 && i+2 < len(b) && (b[i+2] == 'Z' || b[i+2] == 'X')
}

// sessionProcs lists processes whose environ has the JOBSHOP_SOCKET the
// controller gives agents of the session in dir (unique per root, unlike
// the session id). Linux only; elsewhere it finds none.
func sessionProcs(dir string) []int {
	want := []byte("JOBSHOP_SOCKET=" + filepath.Join(dir, "cli.sock"))
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}
	self := os.Getpid()
	var out []int
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil || pid <= 1 || pid == self {
			continue
		}
		env, err := os.ReadFile(filepath.Join("/proc", e.Name(), "environ"))
		if err != nil {
			continue
		}
		for _, kv := range bytes.Split(env, []byte{0}) {
			if bytes.Equal(kv, want) {
				out = append(out, pid)
				break
			}
		}
	}
	return out
}
