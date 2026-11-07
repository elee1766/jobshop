package control

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

type attentionAction struct {
	option               rpc.QuestionOption
	action, target, step string
}

// attentionActions only describes choices, so supervisor reads cannot mutate work.
func (c *Controller) attentionActions(ctx context.Context, it rpc.AttentionItem) []attentionAction {
	var actions []attentionAction
	add := func(title, description, action, target, step string) {
		actions = append(actions, attentionAction{rpc.QuestionOption{ID: len(actions) + 1, Title: title, Description: description}, action, target, step})
	}
	t, _ := c.store.Get(ctx, it.ID)
	runID, pool := metaString(t.Meta, runMetaPack, "id"), metaString(t.Meta, poolMetaPack, "name")
	if it.Kind == rpc.AttentionRun {
		runID = it.ID
	}
	if it.Kind == rpc.AttentionPool {
		pool = it.ID
	}
	switch {
	case it.Kind == rpc.AttentionQuestion && len(it.Options) > 0:
		for _, opt := range it.Options {
			actions = append(actions, attentionAction{option: opt, action: "question", target: it.ID})
		}
	case it.Kind == rpc.AttentionDecision:
		for _, action := range []string{rpc.DecisionApprove, rpc.DecisionReject, rpc.DecisionDismiss} {
			add(strings.ToUpper(action[:1])+action[1:], "Resolve the proposal as "+action, "decision", it.ID, action)
		}
	case it.Kind == rpc.AttentionWait:
		add("Clear wait", "Make the ticket eligible again", "clear", it.ID, "")
		add("Extend one hour", "Keep waiting with a new timeout", "extend", it.ID, "")
		if runID != "" {
			add("Cancel run", "Stop this workflow run", "cancel", runID, "")
		}
	case metaString(t.Meta, hookMetaPack, "key") != "":
		add("Allow", "Resolve the guardrail approval", "guard", it.ID, "done")
		add("Deny", "Resolve the guardrail approval", "guard", it.ID, "denied")
	case runID != "":
		add("Resume", "Retry the first unfinished step", "resume", runID, "")
		if run, err := c.store.Get(ctx, runID); err == nil {
			if rr, err := c.runResult(ctx, run); err == nil {
				for _, st := range rr.Steps {
					sf, _ := store.DecodeFields[store.StepFields](st)
					if !store.StepFinished(st) || len(actions) >= 4 {
						break
					}
					add("Resume at "+sf.Step, "Reopen this earlier step", "resume", runID, sf.Step)
				}
			}
		}
		add("Cancel run", "Stop this workflow run", "cancel", runID, "")
	case pool != "":
		add("Unquarantine after fix", "Let the pool accept work again", "unquarantine", pool, "")
		add("Keep quarantined", "Leave the pool stopped", "keep", pool, "")
	default:
		add("Discuss", "Open the resolving chat with context", "discuss", it.ID, "")
		add("Leave waiting", "Keep this item open for a later decision", "keep", it.ID, "")
	}
	if recommended := metaString(t.Meta, deskMetaPack, "attention_recommend"); recommended != "" {
		for i := range actions {
			actions[i].option.Recommended = recommended == strconv.Itoa(actions[i].option.ID)
		}
	}
	return actions
}

func (c *Controller) executeAttention(ctx context.Context, sc scope, a attentionAction, note string) error {
	var err error
	switch a.action {
	case "question":
		_, err = c.answerQuestion(ctx, sc, rpc.AnswerParams{ID: a.target, Option: a.option.ID, Text: note})
	case "decision":
		_, err = c.decisionResolve(ctx, sc, rpc.DecisionResolveParams{ID: a.target, Action: a.step})
	case "resume":
		_, err = c.runResume(ctx, sc, rpc.RunResumeParams{ID: a.target, Step: a.step})
	case "cancel":
		_, err = c.runCancel(ctx, sc, rpc.IDParams{ID: a.target})
	case "unquarantine":
		_, err = c.poolUnquarantine(ctx, sc, rpc.PoolParams{Pool: a.target})
	case "clear":
		_, err = c.waitClear(ctx, sc, rpc.WaitClearParams{ID: a.target, Reason: "operator decision"})
	case "guard":
		_, err = c.closeTicketIn(ctx, c.config(), sc, rpc.CloseParams{ID: a.target, Resolution: a.step})
	case "extend":
		current, getErr := c.store.Get(ctx, a.target)
		if getErr != nil {
			return getErr
		}
		w := waitOf(current)
		if w == nil {
			return fmt.Errorf("%w: wait already cleared", store.ErrConflict)
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
		return nil, fmt.Errorf("%w: note is too long", store.ErrInvalid)
	}
	items, err := c.attentionItems(ctx, c.config())
	if err != nil {
		return nil, err
	}
	index := slices.IndexFunc(items, func(it rpc.AttentionItem) bool { return it.Kind == p.Kind && it.ID == p.ID })
	if index < 0 {
		return nil, fmt.Errorf("%w: item is no longer waiting", store.ErrConflict)
	}
	it := items[index]
	if !sc.admin && it.Project != sc.project {
		return nil, forbidden("attention item is outside this session's project")
	}
	actions := c.attentionActions(ctx, it)
	choice := slices.IndexFunc(actions, func(a attentionAction) bool { return a.option.ID == p.Option })
	if choice < 0 {
		return nil, fmt.Errorf("%w: unknown option", store.ErrInvalid)
	}
	a := actions[choice]
	if p.ExpectedTitle != "" && p.ExpectedTitle != a.option.Title {
		return nil, fmt.Errorf("%w: choices changed, refresh before answering", store.ErrConflict)
	}
	if !sc.admin {
		if t, err := c.store.Get(ctx, sc.session); err == nil {
			if f, err := store.DecodeFields[store.SessionFields](t); err == nil && f.Operator != "" {
				sc.actor = f.Operator
			}
		}
	}
	admin := scope{admin: true, actor: sc.actor}
	res, chatErr := c.ensureAttentionChat(ctx, c.config(), it, "", false)
	if err := c.executeAttention(ctx, admin, a, p.Note); err != nil {
		return nil, err
	}
	if chatErr == nil && res.Session != "" && a.action != "keep" {
		c.humanWrote(ctx, res.Session)
		c.resumeSession(ctx, res.Session)
		_, err = c.sessionSend(ctx, admin, rpc.SessionSendParams{ID: res.Session, Text: fmt.Sprintf("Chose option %d: %s\n%s", p.Option, a.option.Title, p.Note)})
		if err != nil {
			c.log.WarnContext(ctx, "attention follow-up", "session", res.Session, "err", err)
		}
	}
	if t, err := c.store.Get(ctx, it.ID); err == nil && t.Kind != store.KindSession {
		if _, err := store.AddComment(ctx, c.store, sc.actor, t.ID, fmt.Sprintf("Attention answer: %s. %s", a.option.Title, p.Note)); err != nil {
			c.log.WarnContext(ctx, "record attention answer", "ticket", t.ID, "err", err)
		}
	}
	return res, nil
}
