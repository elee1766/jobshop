package control

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"

	acp "github.com/coder/acp-go-sdk"

	"github.com/elee1766/jobshop/pkg/rpc"
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
func (r *runner) deliver(ctx context.Context, conn *acp.ClientSideConnection) error {
	defer func() { r.promptAttachments = nil }()
	var claimed []store.Ticket
	attachmentBytes := 0
	for _, m := range r.inbox(ctx, false) {
		f, _ := store.DecodeFields[store.MessageFields](m)
		size := 0
		for _, a := range f.Attachments {
			size += a.Size
		}
		if len(claimed) > 0 && attachmentBytes+size > rpc.MaxAttachmentBytes {
			break
		}
		t, cl, err := r.c.store.Claim(ctx, m.ID, r.id, r.c.leaseDur())
		if err != nil {
			continue // another session took it
		}
		if _, err := store.Close(ctx, r.c.store, "controller", t.ID, &cl, resolutionDelivered); err != nil && !errors.Is(err, store.ErrClosed) {
			r.log.WarnContext(ctx, "close delivered message", "message", t.ID, "err", err)
			continue
		}
		claimed = append(claimed, t)
		attachmentBytes += size
	}
	if len(claimed) == 0 {
		return nil
	}
	var b strings.Builder
	var msgCtx []string
	for _, m := range claimed {
		f, _ := store.DecodeFields[store.MessageFields](m)
		attached, content, err := r.messageAttachments(m, f)
		if err != nil {
			return err
		}
		r.promptAttachments = append(r.promptAttachments, content...)
		if f.Context != "" {
			msgCtx = append(msgCtx, f.Context)
		}
		if m.Title == "" && m.Body == "" && attached == "" {
			continue // context only
		}
		if b.Len() == 0 {
			b.WriteString(messagesHeader)
		}
		fmt.Fprintf(&b, "\n\nFrom %s (%s):\n%s", f.From, m.ID, m.Title)
		if m.Body != "" {
			b.WriteString("\n" + m.Body)
		}
		b.WriteString(attached)
	}
	// Chat sessions claim no tickets, so the pool prompt would never reach
	// them; it leads their first message turn (again after a respawn).
	poolPrompt := r.c.config().Pools[r.pool].Chat() && r.delivered.IsZero()
	parts := r.useSend(r.c.assemble(ctx, r.turnSpec(turnMessage, nil, poolPrompt), PromptPart{Source: "messages", Text: b.String()}))
	text := joinParts(parts)
	if len(msgCtx) > 0 {
		// A message with context is background for the model: the pool
		// prompt, rules and the context go in one block the conversation
		// view does not show, and only the message text stays visible.
		text = joinParts(parts[len(parts)-1:])
		r.setContext(joinParts(append(parts[:len(parts)-1:len(parts)-1], PromptPart{Text: strings.Join(msgCtx, "\n\n")})))
	}
	r.log.InfoContext(ctx, "delivering messages", "messages", ids(claimed))
	// A message turn can be interrupted too (a chat user sending a
	// correction); the interrupting message is then delivered at once.
	interrupted, err := r.interruptible(ctx, conn, text)
	// The turn that took an answer is the asker's reaction (JOB-52).
	r.c.answerDelivered(ctx, claimed)
	r.mu.Lock()
	r.delivered = r.c.now()
	said := r.lastSaid.String()
	r.mu.Unlock()
	if err == nil && r.c.isReadOnly(r.id) {
		// The first turn of an attention chat is its summary (#606, #607).
		r.c.recordAttentionSummary(ctx, r.id, said)
	}
	if err != nil || !interrupted {
		return err
	}
	return r.deliver(ctx, conn)
}

// interruptible runs a turn while watching for interrupt messages. On one
// it sends session/cancel; the turn then ends and the caller delivers.
// `session restart --now` cancels the turn the same way.
func (r *runner) interruptible(ctx context.Context, conn *acp.ClientSideConnection, text string) (bool, error) {
	var interrupted atomic.Bool
	// A cancel sent before this turn is not for it.
	select {
	case <-r.cancelNow:
	default:
	}
	r.inTurn.Store(true)
	defer r.inTurn.Store(false)
	wctx, stop := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		tick := r.c.clk().NewTicker(r.c.opt.MessagePoll)
		defer tick.Stop()
		for {
			select {
			case <-wctx.Done():
				return
			case <-r.restartNow:
				r.log.InfoContext(ctx, "restart --now: cancelling turn")
				_ = conn.Cancel(context.WithoutCancel(wctx), acp.CancelNotification{SessionId: r.acpID()})
				return
			case <-r.cancelNow:
				r.log.InfoContext(ctx, "session cancel: cancelling turn")
				_ = conn.Cancel(context.WithoutCancel(wctx), acp.CancelNotification{SessionId: r.acpID()})
				return
			case <-tick.C():
			}
			if len(r.inbox(wctx, true)) > 0 {
				r.log.InfoContext(ctx, "interrupt message: cancelling turn")
				interrupted.Store(true)
				_ = conn.Cancel(context.WithoutCancel(wctx), acp.CancelNotification{SessionId: r.acpID()})
				return
			}
		}
	}()
	err := r.turn(ctx, conn, text)
	stop()
	<-done
	return interrupted.Load(), err
}
