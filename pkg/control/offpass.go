package control

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/support/telemetry"
)

// Off-pass calls (JOB-103). The supervisor pass holds tickMu, so a
// provider call, Linear read or git command inside it stopped scheduling
// for as long as it took. Trigger checks and intake hooks go through the
// check jobs (triggercheck.go); the pass's other calls go through
// offCalls: the pass starts the call in a goroutine, one in flight per
// key, and a later pass takes the result and does the store writes. A
// controller built without Run (internal tests) runs every call in line.

// offCalls is the calls running off the pass and their unclaimed results.
type offCalls struct {
	mu      sync.Mutex
	running map[string]bool
	done    map[string]offDone
}

type offDone struct {
	val any
	at  time.Time
}

// offResultKeep drops a result no pass took (its slot or ticket went
// away) after a while.
const offResultKeep = 10 * time.Minute

// offTake returns key's finished result and forgets it.
func (c *Controller) offTake(key string) (any, bool) {
	o := &c.off
	o.mu.Lock()
	defer o.mu.Unlock()
	d, ok := o.done[key]
	if ok {
		delete(o.done, key)
	}
	return d.val, ok
}

// offGo runs fn off the pass under timeout, unless a call for key is
// running or its result has not been taken. With keep, fn's value is
// kept for offTake; without, it is dropped. It reports whether fn
// started. Without background calls fn runs now and its value is kept.
func (c *Controller) offGo(key, poller string, timeout time.Duration, keep bool, fn func(ctx context.Context) any) bool {
	if !c.backgroundChecks {
		ctx, cancel := context.WithTimeout(c.ctx, timeout)
		v := fn(ctx)
		cancel()
		c.offFinish(key, keep, v)
		return true
	}
	o := &c.off
	o.mu.Lock()
	if _, ok := o.done[key]; ok || o.running[key] {
		o.mu.Unlock()
		return false
	}
	if o.running == nil {
		o.running, o.done = map[string]bool{}, map[string]offDone{}
	}
	o.running[key] = true
	o.mu.Unlock()
	c.wg.Go(func() {
		ctx, cancel := context.WithTimeout(c.ctx, timeout)
		v := fn(ctx)
		c.offTimedOut(ctx, poller, key, timeout)
		cancel()
		o.mu.Lock()
		delete(o.running, key)
		o.mu.Unlock()
		c.offFinish(key, keep, v)
	})
	return true
}

func (c *Controller) offFinish(key string, keep bool, v any) {
	if !keep {
		return
	}
	o := &c.off
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.done == nil {
		o.running, o.done = map[string]bool{}, map[string]offDone{}
	}
	now := c.now()
	for k, d := range o.done {
		if now.Sub(d.at) > offResultKeep {
			delete(o.done, k)
		}
	}
	o.done[key] = offDone{val: v, at: now}
	// A pass records it now, not at the next tick.
	if c.offWake != nil {
		select {
		case c.offWake <- struct{}{}:
		default:
		}
	}
}

// offTimedOut logs and counts a call whose ctx ran out its timeout.
func (c *Controller) offTimedOut(ctx context.Context, poller, key string, timeout time.Duration) {
	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return
	}
	c.log.WarnContext(c.ctx, "off-pass call exceeded its timeout", "poller", poller, "timeout", timeout, "key", strings.ReplaceAll(key, "\x00", " "))
	if c.metrics != nil {
		c.metrics.PollerTimeouts.Add(c.ctx, 1, metric.WithAttributes(telemetry.AttrPoller.String(poller)))
	}
}
