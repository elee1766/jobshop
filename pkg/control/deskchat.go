package control

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/elee1766/jobshop/pkg/control/chatlife"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// chatManage keeps lifecycle and name uniqueness in the controller. What
// each kind of chat may do is in pkg/control/chatlife (JOB-147); the
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
	if chatlife.NeedsName(p.Action) {
		var err error
		if name, err = chatlife.NormalizeName(p.Name); err != nil {
			return nil, err
		}
		list, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Limit: math.MaxInt32})
		if err != nil {
			return nil, err
		}
		if err := chatlife.CheckNameFree(list, p.ID, name); err != nil {
			return nil, err
		}
	}
	if p.Action == chatlife.ActionNew {
		if err := validateAttachments(p.Attachments); err != nil {
			return nil, err
		}
		if err := chatlife.CheckNew(p); err != nil {
			return nil, err
		}
		sess, _, err := c.keyedSession(ctx, c.config(), sc, p.Pool, chatlife.DeskPrefix+strings.ToLower(rand.Text()[:12]))
		if err != nil {
			return nil, err
		}
		if err := c.updateSessionFields(ctx, sess.ID, func(f *store.SessionFields) { f.Named = true; f.DisplayName = name }); err != nil {
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
	f, err := chatlife.Target(t)
	if err != nil {
		return nil, err
	}
	switch p.Action {
	case chatlife.ActionConvert:
		if err := chatlife.CheckConvert(t, f); err != nil {
			return nil, err
		}
		// The stable routing key keeps pending messages and source references valid.
		if err := c.updateSessionFields(ctx, t.ID, func(f *store.SessionFields) { chatlife.Convert(f, name) }); err != nil {
			return nil, err
		}
	case chatlife.ActionRename:
		if err := c.setDisplayName(ctx, sc.actor, t.ID, name); err != nil {
			return nil, err
		}
	case chatlife.ActionArchive, chatlife.ActionResolve:
		if p.Action == chatlife.ActionResolve {
			id, err := chatlife.ResolveTarget(f)
			if err != nil {
				return nil, err
			}
			target, err := c.store.Get(ctx, id)
			if err != nil {
				return nil, err
			}
			if err := chatlife.CheckCurrent(target, t.ID); err != nil {
				return nil, err
			}
			if target.Status != store.StatusClosed {
				if _, err := c.closeTicketIn(ctx, c.config(), sc, rpc.CloseParams{ID: id, Resolution: "done", Comment: chatlife.ResolveComment}); err != nil && !errors.Is(err, store.ErrClosed) {
					return nil, err
				}
			}
		}
		if t.Status != store.StatusClosed {
			if _, err := c.sessionCloseRPC(ctx, sc, rpc.SessionCloseParams{ID: t.ID, Reason: "chat archived"}); err != nil {
				return nil, err
			}
		}
		at := c.now()
		if err := c.updateSessionFields(ctx, t.ID, func(f *store.SessionFields) { f.Archived = &at }); err != nil {
			return nil, err
		}
	case chatlife.ActionRestore:
		if t.Status != store.StatusClosed {
			return c.sessionInfo(ctx, t.ID)
		}
		if pool, ok := c.config().Pools[t.Pool]; !ok || !pool.Chat() {
			return nil, fmt.Errorf("%w: chat pool %q is no longer configured", store.ErrInvalid, t.Pool)
		}
		if _, ok, err := c.openKeyed(ctx, f.Key); err != nil {
			return nil, err
		} else if ok {
			return nil, fmt.Errorf("%w: another session already uses this chat key", store.ErrConflict)
		}
		if err := c.humanSlot(ctx, c.config(), t.Pool); err != nil {
			return nil, err
		}
		// Keep the ACP conversation id and journal, then load on the next send.
		chatlife.Restore(&f)
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
		return nil, fmt.Errorf("%w: unknown chat action %q", store.ErrInvalid, p.Action)
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
	f, err := store.DecodeFields[store.SessionFields](t)
	if err != nil {
		return nil, fmt.Errorf("%w: desk commands require a named chat", store.ErrInvalid)
	}
	if err := chatlife.CheckViewTarget(f); err != nil {
		return nil, err
	}
	if err := chatlife.CheckView(p); err != nil {
		return nil, err
	}
	ev := rpc.Event{Kind: "desk", Op: p.Action, Session: t.ID, Args: p.Args, Time: c.now(), Actor: sc.actor}
	if err := c.store.Record(ctx, sc.actor, t.ID, "desk_command", p); err != nil {
		return nil, err
	}
	c.events.publish([]rpc.Event{ev})
	return ev, nil
}
