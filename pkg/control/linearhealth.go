package control

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/linear"
)

// Linear outages (JOB-118). A Linear call that times out, fails to
// connect or gets a 5xx or 429 marks its identity down: further calls as
// that identity fail at once with errLinearDown until a backoff passes,
// so a hung API costs one request timeout per backoff, not one per wait
// or poll. The outage is logged once when it starts and once when Linear
// answers again; callers do not log errLinearDown.

const (
	linearBackoffFirst = 30 * time.Second
	linearBackoffCap   = 5 * time.Minute
)

// linearHealth is the outage state per identity.
type linearHealth struct {
	mu   sync.Mutex
	down map[string]*linearOutage
}

type linearOutage struct {
	since, retry time.Time
	fails        int
	err          error
}

// errLinearDown wraps a Linear outage error, or stands for one while the
// identity backs off.
type errLinearDown struct {
	identity string
	retry    time.Time
	err      error
}

func (e *errLinearDown) Error() string {
	return fmt.Sprintf("linear: identity %s unavailable until %s: %v", e.identity, e.retry.UTC().Format(time.RFC3339), e.err)
}

func (e *errLinearDown) Unwrap() error { return e.err }

// linearDown reports whether err is a Linear outage already logged.
func linearDown(err error) bool {
	_, ok := errors.AsType[*errLinearDown](err)
	return ok
}

// linearRequestTimeoutFor is the per-request bound for Linear calls made
// by pollers and wait passes; tests lower it.
func (c *Controller) linearRequestTimeoutFor() time.Duration {
	if c.linearTimeout > 0 {
		return c.linearTimeout
	}
	return linearRequestTimeout
}

// linearDo runs fn with identity name's shared client, each request
// bounded by linearRequestTimeoutFor, through linearCall.
func (c *Controller) linearDo(ctx context.Context, name string, fn func(context.Context, *linear.Client) error) error {
	return c.linearCall(ctx, name, func(ctx context.Context) error {
		return fn(c.pollerCtx(ctx), c.linearFor(name))
	})
}

// linearDoKey is linearDo as the identity covering key's team.
func (c *Controller) linearDoKey(ctx context.Context, key string, fn func(context.Context, *linear.Client) error) error {
	team, ok := linear.Team(key)
	if !ok {
		return fmt.Errorf("bad Linear issue key %q", key)
	}
	cfg := c.config()
	name, ok := cfg.IdentityFor(cfg.IdentityNames(), team)
	if !ok {
		return fmt.Errorf("no identity covers team %s", team)
	}
	return c.linearDo(ctx, name, fn)
}

// linearCall runs fn as identity name unless the identity is backing off,
// and records whether Linear answered. fn bounds its own requests.
func (c *Controller) linearCall(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	h := &c.linHealth
	now := c.now()
	h.mu.Lock()
	if o := h.down[name]; o != nil && now.Before(o.retry) {
		err := &errLinearDown{identity: name, retry: o.retry, err: o.err}
		h.mu.Unlock()
		return err
	}
	h.mu.Unlock()

	err := fn(ctx)
	if errors.Is(err, context.Canceled) || ctx.Err() != nil {
		return err // our own shutdown or caller's deadline, not Linear's
	}
	now = c.now()
	h.mu.Lock()
	defer h.mu.Unlock()
	o := h.down[name]
	if !linear.Unavailable(err) {
		if o != nil {
			delete(h.down, name)
			c.log.InfoContext(ctx, "linear recovered", "identity", name, "down_for", now.Sub(o.since).Round(time.Second), "failed_calls", o.fails)
		}
		return err
	}
	if h.down == nil {
		h.down = map[string]*linearOutage{}
	}
	if o == nil {
		o = &linearOutage{since: now}
		h.down[name] = o
	}
	o.fails++
	o.err = err
	o.retry = now.Add(min(linearBackoffFirst<<min(o.fails-1, 8), linearBackoffCap))
	if o.fails == 1 {
		c.log.WarnContext(ctx, "linear unavailable", "identity", name, "retry_at", o.retry, "err", err)
	}
	return &errLinearDown{identity: name, retry: o.retry, err: err}
}
