package control

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/gfx-labs/jobshop/pkg/conversation/attention"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// attentionActions describes an item's choices (pkg/conversation/attention).
// It only reads, so supervisor reads cannot mutate work.
func (c *Controller) attentionActions(ctx context.Context, it rpc.AttentionItem) []attention.Action {
	return attention.Actions(ctx, attentionSource{c: c}, it)
}

// executeAttention carries out the picked choice. Every write an answer
// makes goes through the controller's own methods here, under sc.
func (c *Controller) executeAttention(ctx context.Context, sc scope, a attention.Action, note string) error {
	var err error
	switch a.Verb {
	case attention.VerbQuestion:
		_, err = c.answerQuestion(ctx, sc, rpc.AnswerParams{ID: a.Target, Option: a.Option.ID, Text: note})
	case attention.VerbDecision:
		_, err = c.decisionResolve(ctx, sc, rpc.DecisionResolveParams{ID: a.Target, Action: a.Step})
	case attention.VerbResume:
		_, err = c.runResume(ctx, sc, rpc.RunResumeParams{ID: a.Target, Step: a.Step})
	case attention.VerbCancel:
		_, err = c.runCancel(ctx, sc, rpc.RunCancelParams{ID: a.Target})
	case attention.VerbUnquarantine:
		_, err = c.poolUnquarantine(ctx, sc, rpc.PoolParams{Pool: a.Target})
	case attention.VerbProbe:
		err = c.probeOutageNow(a.Target)
	case attention.VerbClear:
		_, err = c.waitClear(ctx, sc, rpc.WaitClearParams{ID: a.Target, Reason: "operator decision"})
	case attention.VerbGuard:
		_, err = c.closeTicketIn(ctx, c.config(), sc, rpc.CloseParams{ID: a.Target, Resolution: a.Step})
	case attention.VerbExtend:
		current, getErr := c.store.Get(ctx, a.Target)
		if getErr != nil {
			return getErr
		}
		w := waitOf(current)
		if w == nil {
			return fmt.Errorf("%w: wait already cleared", domain.ErrConflict)
		}
		w.Timeout, w.TimedOut = c.now().Add(time.Hour), false
		fields, fieldErr := withWait(current, w)
		if fieldErr != nil {
			return fieldErr
		}
		_, err = c.store.Update(ctx, sc.actor, current.ID, store.Update{Fields: fields, IfRevision: current.Revision})
		if err == nil {
			err = store.RemoveLabel(ctx, c.store, sc.actor, current.ID, needsHumanLabel)
		}
	}
	return err
}

func (c *Controller) attentionAnswer(ctx context.Context, sc scope, p rpc.AttentionAnswerParams) (any, error) {
	if err := c.requireGrant(sc, "attention-answer", "attention answer"); err != nil {
		return nil, err
	}
	if len(p.Note) > 32<<10 {
		return nil, fmt.Errorf("%w: note is too long", domain.ErrInvalid)
	}
	items, err := c.attentionItems(ctx, c.config())
	if err != nil {
		return nil, err
	}
	index := slices.IndexFunc(items, func(it rpc.AttentionItem) bool { return it.Kind == p.Kind && it.ID == p.ID })
	if index < 0 {
		return nil, fmt.Errorf("%w: item is no longer waiting", domain.ErrConflict)
	}
	it := items[index]
	if !sc.admin && it.Project != sc.project {
		return nil, forbidden("attention item is outside this session's project")
	}
	actions := c.attentionActions(ctx, it)
	choice := slices.IndexFunc(actions, func(a attention.Action) bool { return a.Option.ID == p.Option })
	if choice < 0 {
		return nil, fmt.Errorf("%w: unknown option", domain.ErrInvalid)
	}
	a := actions[choice]
	if p.ExpectedTitle != "" && p.ExpectedTitle != a.Option.Title {
		return nil, fmt.Errorf("%w: choices changed, refresh before answering", domain.ErrConflict)
	}
	if !sc.admin {
		if t, err := c.store.Get(ctx, sc.session); err == nil {
			if f, err := domain.DecodeFields[domain.SessionFields](t); err == nil && f.Operator != "" {
				sc.actor = f.Operator
			}
		}
	}
	admin := scope{admin: true, actor: sc.actor}
	res, chatErr := c.ensureAttentionChat(ctx, c.config(), it, "", false)
	if err := c.executeAttention(ctx, admin, a, p.Note); err != nil {
		return nil, err
	}
	if chatErr == nil && res.Session != "" && a.Verb != attention.VerbKeep {
		c.humanWrote(ctx, res.Session)
		c.resumeSession(ctx, res.Session)
		_, err = c.sessionSend(ctx, admin, rpc.SessionSendParams{ID: res.Session, Text: fmt.Sprintf("Chose option %d: %s\n%s", p.Option, a.Option.Title, p.Note)})
		if err != nil {
			c.log.WarnContext(ctx, "attention follow-up", "session", res.Session, "err", err)
		}
	}
	if t, err := c.store.Get(ctx, it.ID); err == nil && t.Kind != domain.KindSession {
		if _, err := store.AddComment(ctx, c.store, sc.actor, t.ID, fmt.Sprintf("Attention answer: %s. %s", a.Option.Title, p.Note)); err != nil {
			c.log.WarnContext(ctx, "record attention answer", "ticket", t.ID, "err", err)
		}
	}
	return res, nil
}
