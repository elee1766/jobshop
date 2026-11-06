package control_test

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/clock"
	"github.com/elee1766/jobshop/pkg/control"
)

// passChans maps a root dir to its controller's Options.Pass channel.
var passChans sync.Map

// fakeClocks maps a root dir to the fake clock its controller runs on
// (poolRootOpt). Timers on it fire only when a test advances it.
var fakeClocks sync.Map

// clockStep is how far passEventually moves a fake clock per attempt: the
// poll, restart and message intervals poolRootOpt sets are all this long.
const clockStep = 50 * time.Millisecond

// clockBudget caps what one passEventually call advances, so a wait that
// never ends cannot spin the clock past a pool's idle_timeout (1m in most
// tests). It is far above the longest backoff chain a test waits through
// (JOB-66: 100+200+400+800ms).
const clockBudget = 30 * time.Second

// settle gives the controller a moment to arm its next timer after the
// clock moved: it returns when a timer is added, or after a short real
// wait when none comes (the controller is waiting on a host, not the
// clock). Only a bound: the loop is driven by the condition.
func settle(clk *clock.Fake) {
	done := make(chan struct{})
	stop := time.AfterFunc(5*time.Millisecond, func() { close(done) })
	defer stop.Stop()
	clk.WaitFor(clk.Waiters()+1, done)
}

// fakeClock returns the clock registered under dir, or nil.
func fakeClock(key any) *clock.Fake {
	if v, ok := fakeClocks.Load(key); ok {
		return v.(*clock.Fake)
	}
	return nil
}

// advance moves the fake clock of the controller registered under dir
// by d, firing the backoffs, restart delays and idle polls due in it.
func advance(t *testing.T, dir string, d time.Duration) {
	t.Helper()
	clk := fakeClock(dir)
	if clk == nil {
		t.Fatalf("no fake clock for %v", dir)
	}
	clk.Advance(d)
}

// withPass sets opt.Pass and registers it under dir for pass.
func withPass(dir string, opt control.Options) control.Options {
	ch := make(chan chan struct{})
	passChans.Store(dir, ch)
	opt.Pass = ch
	return opt
}

// passUntil runs supervisor passes until cond holds, failing after 50.
// For controllers on a fake clock, whose ticker never fires by itself.
func passUntil(t *testing.T, key any, cond func() bool, msg string) {
	t.Helper()
	for range 50 {
		if cond() {
			return
		}
		pass(t, key, 1)
	}
	require.True(t, cond(), msg)
}

// passEventually is require.Eventually for state the supervisor pass
// changes: test controllers tick hourly (#433), so it runs a pass before
// each check instead of waiting for a ticker. A controller on a fake clock
// (poolRootOpt) also gets the clock stepped per attempt, so a backoff or
// restart delay passes in fake time, not a real wait. Host events (a turn
// ending) still arrive on their own; the deadline only matters on failure.
func passEventually(t *testing.T, key any, cond func() bool, msg string) {
	t.Helper()
	clk := fakeClock(key)
	budget := clockBudget
	deadline := time.Now().Add(dialWait)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal(msg)
		}
		pass(t, key, 1)
		if clk != nil && budget > 0 {
			clk.Advance(clockStep)
			budget -= clockStep
			settle(clk)
		} else {
			runtime.Gosched()
		}
	}
}

// pass runs n whole supervisor passes on the controller registered under
// key (its root dir, or its admin client for runRoot) and
// returns when they are done. It replaces sleeping "several ticks" before
// a negative assertion (#359). The timeout only matters on failure.
func pass(t *testing.T, key any, n int) {
	t.Helper()
	v, ok := passChans.Load(key)
	if !ok {
		t.Fatalf("no pass channel for %v", key)
	}
	ch := v.(chan chan struct{})
	for range n {
		done := make(chan struct{})
		select {
		case ch <- done:
		case <-time.After(30 * time.Second):
			t.Fatal("controller took no pass")
		}
		select {
		case <-done:
		case <-time.After(30 * time.Second):
			t.Fatal("pass did not finish")
		}
	}
}
