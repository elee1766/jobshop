// Package clock is the time source the controller reads, so tests can
// run its timers on fake time (#362).
package clock

import (
	"sort"
	"sync"
	"time"
)

// Clock is the subset of package time the controller uses.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
	NewTicker(d time.Duration) Ticker
}

// Ticker is a time.Ticker behind an interface.
type Ticker interface {
	C() <-chan time.Time
	Stop()
}

// Real is the wall clock.
type Real struct{}

// Now is time.Now.
func (Real) Now() time.Time { return time.Now() }

// After is time.After.
func (Real) After(d time.Duration) <-chan time.Time { return time.After(d) }

// NewTicker wraps time.NewTicker.
func (Real) NewTicker(d time.Duration) Ticker { return realTicker{time.NewTicker(d)} }

type realTicker struct{ t *time.Ticker }

func (r realTicker) C() <-chan time.Time { return r.t.C }
func (r realTicker) Stop()               { r.t.Stop() }

// Fake is a clock that moves only when Advance or Set is called. Timers
// and tickers whose time has come fire during the call, in time order.
// Channels are buffered by one, like package time, so a slow reader
// misses ticks instead of blocking Advance.
type Fake struct {
	mu      sync.Mutex
	now     time.Time
	waiters []*waiter
	added   chan struct{} // closed and replaced when a waiter is added
}

type waiter struct {
	at    time.Time
	every time.Duration // 0 for After
	ch    chan time.Time
}

// NewFake returns a fake clock reading t.
func NewFake(t time.Time) *Fake { return &Fake{now: t, added: make(chan struct{})} }

// Now returns the fake time.
func (f *Fake) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

// After fires once d after the current fake time. d <= 0 fires at once.
func (f *Fake) After(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	f.mu.Lock()
	defer f.mu.Unlock()
	if d <= 0 {
		ch <- f.now
		return ch
	}
	f.addLocked(&waiter{at: f.now.Add(d), ch: ch})
	return ch
}

// NewTicker fires every d of fake time until stopped.
func (f *Fake) NewTicker(d time.Duration) Ticker {
	if d <= 0 {
		panic("clock: non-positive ticker interval")
	}
	w := &waiter{every: d, ch: make(chan time.Time, 1)}
	f.mu.Lock()
	defer f.mu.Unlock()
	w.at = f.now.Add(d)
	f.addLocked(w)
	return fakeTicker{f, w}
}

func (f *Fake) addLocked(w *waiter) {
	f.waiters = append(f.waiters, w)
	close(f.added)
	f.added = make(chan struct{})
}

// Advance moves the clock forward by d, firing what comes due.
func (f *Fake) Advance(d time.Duration) {
	f.mu.Lock()
	to := f.now.Add(d)
	f.mu.Unlock()
	f.Set(to)
}

// Set moves the clock to t. Moving back fires nothing.
func (f *Fake) Set(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !t.After(f.now) {
		f.now = t
		return
	}
	for {
		sort.Slice(f.waiters, func(i, j int) bool { return f.waiters[i].at.Before(f.waiters[j].at) })
		if len(f.waiters) == 0 || f.waiters[0].at.After(t) {
			break
		}
		w := f.waiters[0]
		f.now = w.at
		select {
		case w.ch <- w.at:
		default:
		}
		if w.every > 0 {
			// One tick per Set past the interval, like a ticker whose
			// reader was slow: it does not replay every missed period.
			w.at = t.Add(w.every)
		} else {
			f.waiters = f.waiters[1:]
		}
	}
	f.now = t
}

// Waiters returns how many timers and tickers are pending.
func (f *Fake) Waiters() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.waiters)
}

// WaitFor blocks until at least n timers and tickers are pending, so a
// test can Advance only after the code under test armed its timer.
// It returns false if done closes first.
func (f *Fake) WaitFor(n int, done <-chan struct{}) bool {
	for {
		f.mu.Lock()
		if len(f.waiters) >= n {
			f.mu.Unlock()
			return true
		}
		added := f.added
		f.mu.Unlock()
		select {
		case <-added:
		case <-done:
			return false
		}
	}
}

func (f *Fake) stop(w *waiter) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, x := range f.waiters {
		if x == w {
			f.waiters = append(f.waiters[:i], f.waiters[i+1:]...)
			return
		}
	}
}

type fakeTicker struct {
	f *Fake
	w *waiter
}

func (t fakeTicker) C() <-chan time.Time { return t.w.ch }
func (t fakeTicker) Stop()               { t.f.stop(t.w) }
