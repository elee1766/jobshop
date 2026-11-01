package control

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	acp "github.com/coder/acp-go-sdk"

	"github.com/elee1766/jobshop/pkg/store"
)

const (
	resolutionDelivered = "delivered"
	messagesHeader      = "Messages for you:"
)

// inbox returns open messages addressed to this session by id or key, or
// to its pool. interruptOnly limits it to interrupt messages.
func (r *runner) inbox(ctx context.Context, interruptOnly bool) []store.Ticket {
	msgs, err := r.c.store.List(ctx, store.Filter{Kind: store.KindMessage, Status: store.StatusOpen})
	if err != nil {
		return nil
	}
	to := map[string]bool{r.id: true}
	if r.pool != "" {
		to["pool:"+r.pool] = true
	}
	if t, err := r.c.store.Get(ctx, r.id); err == nil {
		if f, err := store.DecodeFields[store.SessionFields](t); err == nil && f.Key != "" {
			to[f.Key] = true
		}
	}
	var out []store.Ticket
	for _, m := range msgs {
		f, err := store.DecodeFields[store.MessageFields](m)
		if err != nil || !to[f.To] || (interruptOnly && !f.Interrupt) {
			continue
		}
		out = append(out, m)
	}
	return out
}

// deliver takes waiting messages and sends them as one prompt. Each is
// claimed (the first session to claim a pool message wins) and closed as
// delivered before the prompt, so the agent never sees a message as its
// current ticket. Delivery is at most once: a crash mid-prompt loses it.
func (r *runner) deliver(ctx context.Context, conn *acp.ClientSideConnection, acpSession string) error {
	var claimed []store.Ticket
	for _, m := range r.inbox(ctx, false) {
		t, cl, err := r.c.store.Claim(ctx, m.ID, r.id, r.c.leaseDur())
		if err != nil {
			continue // another session took it
		}
		if _, err := r.c.store.CloseTicket(ctx, "controller", t.ID, &cl, resolutionDelivered); err != nil && !errors.Is(err, store.ErrClosed) {
			r.log.Warn("close delivered message", "message", t.ID, "err", err)
			continue
		}
		claimed = append(claimed, t)
	}
	if len(claimed) == 0 {
		return nil
	}
	var b strings.Builder
	b.WriteString(messagesHeader)
	for _, m := range claimed {
		f, _ := store.DecodeFields[store.MessageFields](m)
		fmt.Fprintf(&b, "\n\nFrom %s (%s):\n%s", f.From, m.ID, m.Title)
		if m.Body != "" {
			b.WriteString("\n" + m.Body)
		}
	}
	r.log.Info("delivering messages", "messages", ids(claimed))
	return r.turn(ctx, conn, acpSession, b.String())
}

// interruptible runs a turn while watching for interrupt messages. On one
// it sends session/cancel; the turn then ends and the caller delivers.
// `session restart --now` cancels the turn the same way.
func (r *runner) interruptible(ctx context.Context, conn *acp.ClientSideConnection, acpSession, text string) (bool, error) {
	var interrupted atomic.Bool
	wctx, stop := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		tick := time.NewTicker(r.c.opt.MessagePoll)
		defer tick.Stop()
		for {
			select {
			case <-wctx.Done():
				return
			case <-r.restartNow:
				r.log.Info("restart --now: cancelling turn")
				_ = conn.Cancel(context.WithoutCancel(wctx), acp.CancelNotification{SessionId: acp.SessionId(acpSession)})
				return
			case <-tick.C:
			}
			if len(r.inbox(wctx, true)) > 0 {
				r.log.Info("interrupt message: cancelling turn")
				interrupted.Store(true)
				_ = conn.Cancel(context.WithoutCancel(wctx), acp.CancelNotification{SessionId: acp.SessionId(acpSession)})
				return
			}
		}
	}()
	err := r.turn(ctx, conn, acpSession, text)
	stop()
	<-done
	return interrupted.Load(), err
}
