package clock_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/clock"
)

var t0 = time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)

func fired(ch <-chan time.Time) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

func TestFakeAfter(t *testing.T) {
	f := clock.NewFake(t0)
	ch := f.After(time.Minute)
	f.Advance(59 * time.Second)
	assert.False(t, fired(ch), "not due yet")
	f.Advance(time.Second)
	assert.True(t, fired(ch))
	assert.Equal(t, t0.Add(time.Minute), f.Now())
	assert.Zero(t, f.Waiters(), "a fired After is gone")
	assert.True(t, fired(f.After(0)), "zero fires at once")
}

func TestFakeTicker(t *testing.T) {
	f := clock.NewFake(t0)
	tk := f.NewTicker(time.Second)
	f.Advance(time.Second)
	assert.True(t, fired(tk.C()))
	// A jump past several periods delivers one tick, like a slow reader.
	f.Advance(10 * time.Second)
	assert.True(t, fired(tk.C()))
	assert.False(t, fired(tk.C()))
	tk.Stop()
	f.Advance(time.Hour)
	assert.False(t, fired(tk.C()), "stopped")
	assert.Zero(t, f.Waiters())
}

func TestFakeSetBack(t *testing.T) {
	f := clock.NewFake(t0)
	ch := f.After(time.Minute)
	f.Set(t0.Add(-time.Hour))
	assert.False(t, fired(ch))
	assert.Equal(t, t0.Add(-time.Hour), f.Now())
}

func TestFakeWaitFor(t *testing.T) {
	f := clock.NewFake(t0)
	armed := make(chan struct{})
	go func() {
		ch := f.After(time.Minute)
		close(armed)
		<-ch
	}()
	require.True(t, f.WaitFor(1, make(chan struct{})))
	<-armed
	f.Advance(time.Minute)

	done := make(chan struct{})
	close(done)
	assert.False(t, f.WaitFor(5, done), "done closes the wait")
}
