package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/linearsvc"
	"github.com/elee1766/jobshop/pkg/control/threads"
	"github.com/elee1766/jobshop/pkg/control/triggers"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/integrations/linear"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/storage/store/writekey"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// Linear orchestration lives in pkg/control/linearsvc (JOB-142): the
// agent-session sweep, linear_issue_no_run intake, webhook handling, ask
// mirroring and the outage backoff. This file is the controller's side:
// the Host it hands in and the thin steps that call the service. The
// controller keeps the clients and their credentials (apiclients.go),
// every store write, the tick lock and the off-pass machinery.

// linearSvc is the Linear service, made on first use so controllers built
// without Run get one too.
func (c *Controller) linearSvc() *linearsvc.Service {
	c.linOnce.Do(func() { c.linSvc = linearsvc.New(linearHost{c}) })
	return c.linSvc
}

// sweepLinearAsks is the tick step that posts open questions and
// escalations on their Linear issues.
func (c *Controller) sweepLinearAsks(ctx context.Context, cfg *config.Config) {
	c.linearSvc().SweepAsks(ctx, cfg)
}

// sweepLinearAgentSessions is the tick step that ends open agent
// sessions. force skips the poll spacing.
func (c *Controller) sweepLinearAgentSessions(ctx context.Context, cfg *config.Config, force bool) {
	c.linearSvc().SweepSessions(ctx, cfg, force)
}

// signalLinearIssues is the linear_issue_no_run scan signal.
func (c *Controller) signalLinearIssues(ctx context.Context, cfg *config.Config, sig config.ScanSignal) ([]triggers.Finding, error) {
	found, pending, err := c.linearSvc().Signal(ctx, cfg, sig)
	if pending {
		return nil, triggers.ErrCheckPending
	}
	if err != nil {
		return nil, err
	}
	out := make([]triggers.Finding, len(found))
	for i, f := range found {
		out[i] = triggers.Finding{Signal: config.SignalLinearIssueNoRun, Subject: f.Subject, Detail: f.Detail}
	}
	return out, nil
}

// linearIntakeRPC is linear.intake (JOB-121): the states the
// linear_issue_no_run signals for p.Team take, so `linear create` can say
// when a new issue lands outside them. Open to every caller: it reads
// config only, and sessions are the ones creating issues.
func (c *Controller) linearIntakeRPC(_ context.Context, _ scope, p rpc.LinearIntakeParams) (any, error) {
	team := strings.TrimSpace(p.Team)
	if team == "" {
		return nil, fmt.Errorf("%w: linear.intake needs a team", domain.ErrInvalid)
	}
	watched, states := linearsvc.IntakeStates(c.config(), team)
	return rpc.LinearIntakeResult{Watched: watched, States: states}, nil
}

// linearAnswers takes replies to open questions' ask comments out of cs
// and answers those questions with them; the rest is returned for normal
// delivery. A failed answer leaves the reply in cs so it is not lost.
func (c *Controller) linearAnswers(ctx context.Context, th threads.Thread, cs []threads.Comment) []threads.Comment {
	if th.Linear == "" {
		return cs
	}
	replies := make([]linearsvc.Reply, len(cs))
	for i, cm := range cs {
		replies[i] = linearsvc.Reply{Parent: cm.Parent, Body: cm.Body, Author: threads.LinearAuthorPrefix + cm.Author}
	}
	taken := c.linearSvc().Answers(ctx, th.Linear, replies)
	var rest []threads.Comment
	for i, cm := range cs {
		if !taken[i] {
			rest = append(rest, cm)
		}
	}
	return rest
}

// linearHost is the controller as linearsvc.Host.
type linearHost struct{ c *Controller }

func (h linearHost) Config() *config.Config   { return h.c.config() }
func (h linearHost) Now() time.Time           { return h.c.now() }
func (h linearHost) Context() context.Context { return h.c.ctx }
func (h linearHost) Go(fn func())             { h.c.wg.Go(fn) }
func (h linearHost) CanMint() bool            { return h.c.minter != nil }

func (h linearHost) Client(name string) *linear.Client { return h.c.linearFor(name) }

func (h linearHost) ClientIn(cfg *config.Config, key string) (*linear.Client, error) {
	return h.c.linearClientIn(cfg, key)
}

func (h linearHost) OffGo(key, poller string, timeout time.Duration, fn func(ctx context.Context)) bool {
	return h.c.offGo(key, poller, timeout, false, func(ctx context.Context) any {
		fn(ctx)
		return nil
	})
}

func (h linearHost) OffTick(ctx context.Context, key, poller string, timeout time.Duration, run func(ctx context.Context) (any, error)) (any, bool, error) {
	res, err := triggers.OffTick(ctx, triggers.Need{
		Key: key, Poller: poller, Timeout: timeout,
		Run: func(ctx context.Context) triggers.Result {
			v, err := run(ctx)
			return triggers.Result{Val: v, Err: err}
		},
	})
	if errors.Is(err, triggers.ErrCheckPending) {
		return nil, true, nil
	}
	if err == nil {
		err = res.Err
	}
	return res.Val, false, err
}

func (h linearHost) List(ctx context.Context, f store.Filter) ([]domain.Ticket, error) {
	return h.c.store.List(ctx, f)
}

func (h linearHost) Get(ctx context.Context, id string) (domain.Ticket, error) {
	return h.c.store.Get(ctx, id)
}

func (h linearHost) OpenNeedsHuman(ctx context.Context, kind domain.Kind) ([]domain.Ticket, error) {
	return h.c.store.List(ctx, store.Filter{Kind: kind, Status: domain.StatusOpen, Label: needsHumanLabel, Limit: 200})
}

func (h linearHost) IsQuestion(t domain.Ticket) bool {
	_, ok := questionOf(t)
	return ok
}

func (h linearHost) IsEscalation(t domain.Ticket) bool {
	return metaString(t.Meta, runMetaPack, "escalation") != ""
}

func (h linearHost) SetMeta(ctx context.Context, id, pack string, v any) error {
	return h.c.setMetaPack(ctx, id, pack, v)
}

func (h linearHost) Comment(ctx context.Context, id, text string) error {
	_, err := store.AddComment(ctx, h.c.store, "controller", id, text)
	return err
}

func (h linearHost) Keys(project string) linear.KeyStore { return storeKeys{c: h.c, project: project} }

func (h linearHost) AnswerQuestion(ctx context.Context, id, text, from string) error {
	_, err := h.c.answerQuestion(ctx, scope{admin: true, actor: "controller"}, rpc.AnswerParams{ID: id, Text: text, From: from})
	return err
}

func (h linearHost) GitHubIssueOpen(ctx context.Context, cfg *config.Config, d linear.GitHubDep) (bool, error) {
	return h.c.githubIssueOpen(ctx, cfg, d)
}

func (h linearHost) IngestSecret(e config.IngestEndpoint) (ingest.Secret, error) {
	return ingest.SecretFor(h.c.root, h.c.config())(e)
}

func (h linearHost) Spool(ctx context.Context) (*ingest.Spool, error) { return h.c.queue().Spool(ctx) }

func (h linearHost) WakeDrain() { h.c.queue().WakeDrain() }

func (h linearHost) Log() *slog.Logger { return h.c.log }

// RunTriggers is the pass the tick runs, so a finding fires now. It takes
// the tick lock like any pass.
func (h linearHost) RunTriggers(ctx context.Context) {
	c := h.c
	unlock := c.lockTick(ctx, "webhook")
	defer unlock()
	if !c.draining() {
		c.passStep(ctx, c.config(), "run_triggers", c.runTriggers)
	}
}

func (h linearHost) PollThread(ctx context.Context, issue string) {
	h.c.pollSubscriptionThread(ctx, threads.Thread{Linear: issue})
}

// setMetaPack replaces meta[pack] on id, conditional on the revision read.
func (c *Controller) setMetaPack(ctx context.Context, id, pack string, v any) error {
	for attempt := 0; ; attempt++ {
		t, err := c.store.Get(ctx, id)
		if err != nil {
			return err
		}
		m := map[string]json.RawMessage{}
		if len(t.Meta) > 0 {
			_ = json.Unmarshal(t.Meta, &m)
		}
		if m[pack], err = json.Marshal(v); err != nil {
			return err
		}
		mb, err := json.Marshal(m)
		if err != nil {
			return err
		}
		_, err = c.store.Update(ctx, "controller", id, store.Update{Meta: mb, IfRevision: t.Revision})
		if errors.Is(err, domain.ErrConflict) && attempt < 5 {
			continue
		}
		return err
	}
}

// storeKeys is linear.KeyStore on the store, named like a session's keys
// of the ticket's project.
type storeKeys struct {
	c       *Controller
	project string
}

func (k storeKeys) Claim(ctx context.Context, key, op string) (string, bool, bool, error) {
	w, err := writekey.Claim(ctx, k.c.store, writeKeyName(scope{project: k.project}, key), op, "controller")
	if err != nil {
		return "", false, false, err
	}
	return w.Result, w.Done, !w.Done && !w.New, nil
}

func (k storeKeys) Finish(ctx context.Context, key, result string) (string, error) {
	w, err := writekey.Finish(ctx, k.c.store, writeKeyName(scope{project: k.project}, key), result)
	if err != nil {
		return "", err
	}
	return w.Result, nil
}

// githubIssueOpen reads a GitHub issue's state as the identity covering
// its repo. Tests replace it with c.ghIssueOpen.
func (c *Controller) githubIssueOpen(ctx context.Context, cfg *config.Config, d linear.GitHubDep) (bool, error) {
	if c.ghIssueOpen != nil {
		return c.ghIssueOpen(ctx, d)
	}
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	var v struct {
		State string `json:"state"`
	}
	if err := c.githubOrHost(cfg, "", d.Repo).Get(ctx, fmt.Sprintf("/repos/%s/issues/%d", d.Repo, d.Number), &v); err != nil {
		return false, err
	}
	return v.State == "open", nil
}
