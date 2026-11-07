package control

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// chatManage keeps lifecycle and name uniqueness in the controller.
func (c *Controller) chatManage(ctx context.Context, sc scope, p rpc.ChatManageParams) (any, error) {
	if err := requireAdmin(sc, "chat manage"); err != nil {
		return nil, err
	}
	c.chatMu.Lock()
	defer c.chatMu.Unlock()
	name := strings.Join(strings.Fields(p.Name), " ")
	if p.Action == "new" || p.Action == "rename" {
		if name == "" || utf8.RuneCountInString(name) > maxDisplayName {
			return nil, fmt.Errorf("%w: give a chat name of 1 to %d characters", store.ErrInvalid, maxDisplayName)
		}
		list, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Limit: math.MaxInt32})
		if err != nil {
			return nil, err
		}
		for _, t := range list {
			f, _ := store.DecodeFields[store.SessionFields](t)
			if t.ID != p.ID && (f.Named || strings.HasPrefix(f.Key, "desk:")) && strings.EqualFold(f.DisplayName, name) {
				return nil, fmt.Errorf("%w: chat name %q is already used", store.ErrConflict, name)
			}
		}
	}
	if p.Action == "new" {
		if strings.TrimSpace(p.Text) == "" || len(p.Text) > 32<<10 {
			return nil, fmt.Errorf("%w: first message must be 1 to 32768 bytes", store.ErrInvalid)
		}
		sess, _, err := c.keyedSession(ctx, c.config(), sc, p.Pool, "desk:"+strings.ToLower(rand.Text()[:12]))
		if err != nil {
			return nil, err
		}
		if err := c.updateSessionFields(ctx, sess.ID, func(f *store.SessionFields) { f.Named = true; f.DisplayName = name }); err != nil {
			return nil, err
		}
		if _, err := c.sessionSend(ctx, sc, rpc.SessionSendParams{ID: sess.ID, Text: p.Text}); err != nil {
			return nil, err
		}
		return c.sessionInfo(ctx, sess.ID)
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	f, err := store.DecodeFields[store.SessionFields](t)
	if err != nil || t.Kind != store.KindSession || !strings.HasPrefix(f.Key, "desk:") {
		return nil, fmt.Errorf("%w: not a named Desk chat", store.ErrInvalid)
	}
	switch p.Action {
	case "rename":
		if err := c.setDisplayName(ctx, sc.actor, t.ID, name); err != nil {
			return nil, err
		}
	case "archive":
		if t.Status != store.StatusClosed {
			if _, err := c.sessionCloseRPC(ctx, sc, rpc.SessionCloseParams{ID: t.ID, Reason: "chat archived"}); err != nil {
				return nil, err
			}
		}
		at := c.now()
		if err := c.updateSessionFields(ctx, t.ID, func(f *store.SessionFields) { f.Named = true; f.Archived = &at }); err != nil {
			return nil, err
		}
	case "restore":
		if t.Status != store.StatusClosed {
			return c.sessionInfo(ctx, t.ID)
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
		if err := c.updateSessionFields(ctx, t.ID, func(f *store.SessionFields) { f.Named = true; f.Archived = nil; f.Ended = nil; f.Suspended = true }); err != nil {
			return nil, err
		}
		if _, err := store.Reopen(ctx, c.store, sc.actor, t.ID); err != nil {
			return nil, err
		}
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
	if err != nil || !strings.HasPrefix(f.Key, "desk:") {
		return nil, fmt.Errorf("%w: desk commands require a named chat", store.ErrInvalid)
	}
	allowed := map[string][]string{
		"navigate": {"view", "ref"}, "filter": {"window", "repo", "step", "team", "waiting_on", "age"},
		"select": {"refs"}, "attach": {"refs"}, "expand": {"window"}, "collapse": {"window"},
	}
	keys, ok := allowed[p.Action]
	if !ok {
		return nil, fmt.Errorf("%w: unknown view action %q", store.ErrInvalid, p.Action)
	}
	for k, v := range p.Args {
		if !slices.Contains(keys, k) || len(v) > 2048 {
			return nil, fmt.Errorf("%w: invalid %s argument %q", store.ErrInvalid, p.Action, k)
		}
	}
	if p.Action == "navigate" && !slices.Contains([]string{"home", "trace", "improvements", "system"}, p.Args["view"]) {
		return nil, fmt.Errorf("%w: unknown view", store.ErrInvalid)
	}
	if slices.Contains([]string{"filter", "expand", "collapse"}, p.Action) && !slices.Contains([]string{"waiting", "not-moving", "inflight"}, strings.TrimPrefix(p.Args["window"], "/")) {
		return nil, fmt.Errorf("%w: unknown window", store.ErrInvalid)
	}
	if slices.Contains([]string{"select", "attach"}, p.Action) && strings.TrimSpace(p.Args["refs"]) == "" {
		return nil, fmt.Errorf("%w: refs are required", store.ErrInvalid)
	}
	ev := rpc.Event{Kind: "desk", Op: p.Action, Session: t.ID, Args: p.Args, Time: c.now(), Actor: sc.actor}
	if err := c.store.Record(ctx, sc.actor, t.ID, "desk_command", p); err != nil {
		return nil, err
	}
	c.events.publish([]rpc.Event{ev})
	return ev, nil
}
