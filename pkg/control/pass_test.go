package control_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/control"
)

// passChans maps a root dir to its controller's Options.Pass channel.
var passChans sync.Map

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
// each check instead of waiting for a ticker. Host events (a turn ending)
// still arrive on their own; the deadline only matters on failure.
func passEventually(t *testing.T, key any, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(dialWait)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal(msg)
		}
		pass(t, key, 1)
		// Gives host events time to land without spinning passes.
		<-time.After(20 * time.Millisecond)
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
