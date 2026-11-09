package control

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/elee1766/jobshop/pkg/conversation"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// chatManage keeps lifecycle and name uniqueness in the controller. What
// each kind of chat may do is in pkg/conversation (JOB-147); the
// writes, the chat mutex and the live session calls stay here. Archive
// and resolve close the session, resolve first closes the escalation
// ticket the chat answered for, and neither touches the journal, so the
// transcript stays for restore.
func (c *Controller) chatManage(ctx context.Context, sc scope, p rpc.ChatManageParams) (any, error) {
	if err := requireAdmin(sc, "chat manage"); err != nil {
		return nil, err
	}
	c.chatMu.Lock()
	defer c.chatMu.Unlock()
	var name string
	if conversation.NeedsName(p.Action) {
		var err error
		if name, err = conversation.NormalizeName(p.Name); err != nil {
			return nil, err
		}
		list, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Limit: math.MaxInt32})
		if err != nil {
			return nil, err
		}
		if err := conversation.CheckNameFree(list, p.ID, name); err != nil {
			return nil, err
		}
	}
	if p.Action == conversation.ActionNew {
		if err := conversation.ValidateAttachments(p.Attachments); err != nil {
			return nil, err
		}
		if err := conversation.CheckNew(p); err != nil {
			return nil, err
		}
		sess, _, err := c.keyedSession(ctx, c.config(), sc, p.Pool, conversation.DeskPrefix+strings.ToLower(rand.Text()[:12]))
		if err != nil {
			return nil, err
		}
		if err := c.updateSessionFields(ctx, sess.ID, func(f *domain.SessionFields) { f.Named = true; f.DisplayName = name }); err != nil {
			return nil, err
		}
		if _, err := c.sessionSend(ctx, sc, rpc.SessionSendParams{ID: sess.ID, Text: p.Text, Context: p.Context, Attachments: p.Attachments}); err != nil {
			return nil, err
		}
		return c.sessionInfo(ctx, sess.ID)
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	f, err := conversation.Target(t)
	if err != nil {
		return nil, err
	}
	switch p.Action {
	case conversation.ActionConvert:
		if err := conversation.CheckConvert(t, f); err != nil {
			return nil, err
		}
		// The stable routing key keeps pending messages and source references valid.
		if err := c.updateSessionFields(ctx, t.ID, func(f *domain.SessionFields) { conversation.Convert(f, name) }); err != nil {
			return nil, err
		}
	case conversation.ActionRename:
		if err := c.setDisplayName(ctx, sc.actor, t.ID, name); err != nil {
			return nil, err
		}
	case conversation.ActionArchive, conversation.ActionResolve:
		if p.Action == conversation.ActionResolve {
			id, err := conversation.ResolveTarget(f)
			if err != nil {
				return nil, err
			}
			target, err := c.store.Get(ctx, id)
			if err != nil {
				return nil, err
			}
			if err := conversation.CheckCurrent(target, t.ID); err != nil {
				return nil, err
			}
			if target.Status != domain.StatusClosed {
				if _, err := c.closeTicketIn(ctx, c.config(), sc, rpc.CloseParams{ID: id, Resolution: "done", Comment: conversation.ResolveComment}); err != nil && !errors.Is(err, domain.ErrClosed) {
					return nil, err
				}
			}
		}
		if t.Status != domain.StatusClosed {
			if _, err := c.sessionCloseRPC(ctx, sc, rpc.SessionCloseParams{ID: t.ID, Reason: "chat archived"}); err != nil {
				return nil, err
			}
		}
		at := c.now()
		if err := c.updateSessionFields(ctx, t.ID, func(f *domain.SessionFields) { f.Archived = &at }); err != nil {
			return nil, err
		}
	case conversation.ActionRestore:
		if t.Status != domain.StatusClosed {
			return c.sessionInfo(ctx, t.ID)
		}
		if pool, ok := c.config().Pools[t.Pool]; !ok || !pool.Chat() {
			return nil, fmt.Errorf("%w: chat pool %q is no longer configured", domain.ErrInvalid, t.Pool)
		}
		if _, ok, err := c.openKeyed(ctx, f.Key); err != nil {
			return nil, err
		} else if ok {
			return nil, fmt.Errorf("%w: another session already uses this chat key", domain.ErrConflict)
		}
		if err := c.humanSlot(ctx, c.config(), t.Pool); err != nil {
			return nil, err
		}
		// Keep the ACP conversation id and journal, then load on the next send.
		conversation.Restore(&f)
		fields, err := json.Marshal(f)
		if err != nil {
			return nil, err
		}
		reopened, err := c.store.Update(ctx, sc.actor, t.ID, store.Update{Reopen: true, Fields: fields, IfRevision: t.Revision})
		if err != nil {
			return nil, err
		}
		if err := c.openSession(reopened); err != nil {
			return nil, err
		}
		c.drive(c.config(), reopened)
	default:
		return nil, fmt.Errorf("%w: unknown chat action %q", domain.ErrInvalid, p.Action)
	}
	return c.sessionInfo(ctx, t.ID)
}

// deskCommand is view-only and can target only the caller's own chat.
func (c *Controller) deskCommand(ctx context.Context, sc scope, p rpc.DeskCommandParams) (any, error) {
	if err := c.requireGrant(sc, "desk-command", "desk command"); err != nil {
		return nil, err
	}
	if p.Session == "" {
		p.Session = sc.session
	}
	if !sc.admin && p.Session != sc.session {
		return nil, forbidden("desk command must target this session")
	}
	t, err := c.openSessionTicket(ctx, p.Session)
	if err != nil {
		return nil, err
	}
	f, err := domain.DecodeFields[domain.SessionFields](t)
	if err != nil {
		return nil, fmt.Errorf("%w: desk commands require a named chat", domain.ErrInvalid)
	}
	if err := conversation.CheckViewTarget(f); err != nil {
		return nil, err
	}
	if err := conversation.CheckView(p); err != nil {
		return nil, err
	}
	ev := rpc.Event{Kind: "desk", Op: p.Action, Session: t.ID, Args: p.Args, Time: c.now(), Actor: sc.actor}
	if err := c.store.Record(ctx, sc.actor, t.ID, "desk_command", p); err != nil {
		return nil, err
	}
	c.events.publish([]rpc.Event{ev})
	return ev, nil
}
