// Package apicall is what the Linear and GitHub clients share: the token
// source and the per-call options. A client asks its source for a token on
// every request and keeps no per-call state, so one client per identity
// serves every caller, and a rotated or renewed token is picked up without
// rebuilding it. What differs per call (a timeout, an observer) rides on
// the context.
package apicall

import (
	"context"
	"time"
)

// Source supplies the token for each request.
type Source interface {
	// Token returns the token for the next request.
	Token(ctx context.Context) (string, error)
	// Renew answers a 401 for stale. Returning stale itself means there is
	// nothing newer, and the client keeps the refused response.
	Renew(ctx context.Context, stale string) (string, error)
}

// Static is a fixed token, such as LINEAR_API_KEY or a test's.
type Static string

// Token returns the token.
func (s Static) Token(context.Context) (string, error) { return string(s), nil }

// Renew returns the same token: there is nothing to renew.
func (s Static) Renew(context.Context, string) (string, error) { return string(s), nil }

// Observer watches a client's requests for one caller: the Starlark
// built-ins redact the tokens sent and journal whether a request reached
// the API. Any field may be nil.
type Observer struct {
	Token func(token string) // the token about to be sent
	Sent  func()             // a request is about to go out
	Done  func()             // its response was read in full
}

type observerKey struct{}

type timeoutKey struct{}

// WithObserver attaches o to the requests made with ctx.
func WithObserver(ctx context.Context, o Observer) context.Context {
	return context.WithValue(ctx, observerKey{}, o)
}

// Sending reports token and the start of a request to ctx's observer.
func Sending(ctx context.Context, token string) {
	o, _ := ctx.Value(observerKey{}).(Observer)
	if o.Token != nil {
		o.Token(token)
	}
	if o.Sent != nil {
		o.Sent()
	}
}

// Received reports a fully read response to ctx's observer.
func Received(ctx context.Context) {
	if o, _ := ctx.Value(observerKey{}).(Observer); o.Done != nil {
		o.Done()
	}
}

// WithRequestTimeout bounds each request made with ctx by d, below the
// client's default: a poller's call has many requests and fails one fast.
func WithRequestTimeout(ctx context.Context, d time.Duration) context.Context {
	return context.WithValue(ctx, timeoutKey{}, d)
}

// Request returns the context for one request: ctx bounded by its request
// timeout, if any.
func Request(ctx context.Context) (context.Context, context.CancelFunc) {
	if d, ok := ctx.Value(timeoutKey{}).(time.Duration); ok && d > 0 {
		return context.WithTimeout(ctx, d)
	}
	return ctx, func() {}
}
