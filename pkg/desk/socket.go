package desk

import (
	"context"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// Socket is a Controller that dials the admin socket per call, so a
// controller restart costs one failed request, not a dead Desk.
type Socket struct{ Path string }

// Call implements Controller.
func (s Socket) Call(ctx context.Context, method string, p, out any) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	c, err := rpc.Dial(ctx, s.Path)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	return c.Call(ctx, method, p, out)
}

// Events implements Controller.
func (s Socket) Events(ctx context.Context, p rpc.EventsSubscribeParams, fn func(rpc.Event)) (rpc.EventsSubscribeResult, error) {
	c, err := rpc.Dial(ctx, s.Path)
	if err != nil {
		return rpc.EventsSubscribeResult{Seq: p.SinceSeq}, err
	}
	defer func() { _ = c.Close() }()
	return c.Events(ctx, p, fn)
}

// Subscribe implements Controller.
func (s Socket) Subscribe(ctx context.Context, p rpc.SessionSubscribeParams, fn func(rpc.SessionEvent)) (rpc.SessionSubscribeResult, error) {
	c, err := rpc.Dial(ctx, s.Path)
	if err != nil {
		return rpc.SessionSubscribeResult{Seq: p.SinceSeq}, err
	}
	defer func() { _ = c.Close() }()
	return c.Subscribe(ctx, p, fn)
}
