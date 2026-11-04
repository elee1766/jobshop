package control_test

import (
	"sync"
	"testing"
	"time"

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
