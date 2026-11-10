package linearsvc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
)

// Linear outages (JOB-118). A Linear call that times out, fails to
// connect or gets a 5xx or 429 marks its identity down: further calls as
// that identity fail at once with a downError until a backoff passes,
// so a hung API costs one request timeout per backoff, not one per wait
// or poll. The outage is logged once when it starts and once when Linear
// answers again; callers do not log a downError (see Down).

const (
	// BackoffFirst is the first backoff; it doubles up to backoffCap.
	BackoffFirst = 30 * time.Second
	backoffCap   = 5 * time.Minute
)

// health is the outage state per identity.
type health struct {
	mu   sync.Mutex
	down map[string]*outage
}

type outage struct {
	since, retry time.Time
	fails        int
	err          error
}

// downError wraps a Linear outage error, or stands for one while the
// identity backs off.
type downError struct {
	identity string
	retry    time.Time
	err      error
}

func (e *downError) Error() string {
	return fmt.Sprintf("linear: identity %s unavailable until %s: %v", e.identity, e.retry.UTC().Format(time.RFC3339), e.err)
}

func (e *downError) Unwrap() error { return e.err }

// Down reports whether err is a Linear outage already logged.
func Down(err error) bool {
	_, ok := errors.AsType[*downError](err)
	return ok
}

// Call runs fn as identity name unless the identity is backing off, and
// records whether Linear answered. fn bounds its own requests.
func (s *Service) Call(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	h := &s.health
	now := s.host.Now()
	h.mu.Lock()
	if o := h.down[name]; o != nil && now.Before(o.retry) {
		err := &downError{identity: name, retry: o.retry, err: o.err}
		h.mu.Unlock()
		return err
	}
	h.mu.Unlock()

	err := fn(ctx)
	if errors.Is(err, context.Canceled) || ctx.Err() != nil {
		return err // our own shutdown or caller's deadline, not Linear's
	}
	now = s.host.Now()
	h.mu.Lock()
	defer h.mu.Unlock()
	o := h.down[name]
	if !linear.Unavailable(err) {
		if o != nil {
			delete(h.down, name)
			s.log().InfoContext(ctx, "linear recovered", "identity", name, "down_for", now.Sub(o.since).Round(time.Second), "failed_calls", o.fails)
		}
		return err
	}
	if h.down == nil {
		h.down = map[string]*outage{}
	}
	if o == nil {
		o = &outage{since: now}
		h.down[name] = o
	}
	o.fails++
	o.err = err
	o.retry = now.Add(min(BackoffFirst<<min(o.fails-1, 8), backoffCap))
	if o.fails == 1 {
		s.log().WarnContext(ctx, "linear unavailable", "identity", name, "retry_at", o.retry, "err", err)
	}
	return &downError{identity: name, retry: o.retry, err: err}
}
