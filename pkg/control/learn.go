package control

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/elee1766/jobshop/pkg/learning"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Learnings (#264). Agents record corrections, failures and workarounds
// with `jobshop learn`; the controller records the ones it sees itself
// (failed steps, review findings, escalations, answers, stalls). They live
// in their own store table, so no pool ever claims them; the curator reads
// them with `jobshop learn list`.

// learnSummaryMax bounds a controller-written summary.
const learnSummaryMax = 500

func (c *Controller) learnAdd(ctx context.Context, sc scope, p rpc.LearnParams) (any, error) {
	if err := store.CheckLearningKind(strings.TrimSpace(p.Kind), c.config().LearningKinds()); err != nil {
		return nil, err
	}
	l := store.Learning{
		Kind: p.Kind, Summary: p.Summary, Fix: p.Fix, Scope: p.Scope, Ticket: p.Ticket,
		Session: sc.session, Source: store.LearningSourceAgent,
	}
	if l.Ticket == "" && !sc.admin {
		// Default to the held ticket; holding none is fine.
		if id, err := c.current(ctx, sc, ""); err == nil {
			l.Ticket = id
		}
	}
	if l.Ticket != "" {
		if _, err := c.store.Get(ctx, l.Ticket); err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(l.Scope) == "" && l.Ticket != "" {
		if t, err := c.store.Get(ctx, l.Ticket); err == nil {
			l.Scope = c.learnScope(ctx, t)
		}
	}
	c.redactWrite(sc, "learning for", l.Ticket, &l.Summary, &l.Fix)
	return c.addLearning(ctx, l)
}

// addLearning stores l with its ticket's project, which recall filters on.
func (c *Controller) addLearning(ctx context.Context, l store.Learning) (store.Learning, error) {
	project := ""
	if l.Ticket != "" {
		if t, err := c.store.Get(ctx, l.Ticket); err == nil {
			project = t.Project
		}
	}
	return learning.Add(ctx, c.store, l, project)
}

func (c *Controller) learnList(ctx context.Context, _ scope, p rpc.LearnListParams) (any, error) {
	f := store.LearningFilter{Kind: p.Kind, Status: p.Status, Ticket: p.Ticket, Limit: p.Limit}
	// A glob kind (#524) is matched here: fetch without it, then filter.
	glob := strings.ContainsAny(p.Kind, "*?[")
	if glob {
		if _, err := path.Match(p.Kind, ""); err != nil {
			return nil, fmt.Errorf("%w: kind %q: %w", store.ErrInvalid, p.Kind, err)
		}
		f.Kind, f.Limit = "", store.DefaultLimit
	}
	if p.Scope != "" {
		s, err := store.ParseScope(p.Scope)
		if err != nil {
			return nil, err
		}
		f.Scope = s
	}
	if p.Since > 0 {
		f.Since = c.now().Add(-p.Since)
	}
	ls, err := learning.List(ctx, c.store, f)
	if err != nil || !glob {
		return ls, err
	}
	limit := p.Limit
	if limit <= 0 {
		limit = store.DefaultLimit
	}
	out := []store.Learning{}
	for _, l := range ls {
		if store.Glob(p.Kind, l.Kind) && len(out) < limit {
			out = append(out, l)
		}
	}
	return out, nil
}

// learnUpdate sets status and PR on learnings (#273). Open to every
// session: the curator marks the PR it opened, the refinery or the next
// curate pass marks the outcome. A rejected learning must name its PR, so
// the curator can see what was turned down.
func (c *Controller) learnUpdate(ctx context.Context, _ scope, p rpc.LearnUpdateParams) (any, error) {
	if len(p.IDs) == 0 {
		return nil, fmt.Errorf("%w: learn update needs at least one id", store.ErrInvalid)
	}
	if p.Status == "" && strings.TrimSpace(p.PR) == "" {
		return nil, fmt.Errorf("%w: learn update needs --status or --pr", store.ErrInvalid)
	}
	cur := make([]store.Learning, 0, len(p.IDs))
	for _, id := range p.IDs {
		l, err := learning.Get(ctx, c.store, id)
		if err != nil {
			return nil, err
		}
		st, pr := l.Status, l.PR
		if p.Status != "" {
			st = p.Status
		}
		if strings.TrimSpace(p.PR) != "" {
			pr = p.PR
		}
		if st == store.LearningRejected && strings.TrimSpace(pr) == "" {
			return nil, fmt.Errorf("%w: learning %d: rejected needs --pr", store.ErrInvalid, id)
		}
		l.Status, l.PR = st, pr
		cur = append(cur, l)
	}
	out := make([]store.Learning, 0, len(cur))
	merged := map[string][]int64{}
	var prs []string
	for i, l := range cur {
		u, err := learning.Update(ctx, c.store, l.ID, l.Status, l.PR)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
		// memory:<id> is a local write (#592), recorded by memory.write.
		if p.Status == store.LearningApplied && u.PR != "" && !strings.HasPrefix(u.PR, "memory:") {
			if _, ok := merged[u.PR]; !ok {
				prs = append(prs, u.PR)
			}
			merged[u.PR] = append(merged[u.PR], p.IDs[i])
		}
	}
	// An applied PR merged: record the memory edit and its baseline for
	// the curator gate (#274).
	for _, pr := range prs {
		if _, err := c.recordEdit(ctx, pr, merged[pr], p.MergedAt); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// learnScope is the narrowest scope a ticket names: its repo (from its
// meta or its run's), else its pool, else global.
func (c *Controller) learnScope(ctx context.Context, t store.Ticket) string {
	if repo := metaString(t.Meta, "github", "repo"); repo != "" {
		if s, err := store.ParseScope(store.RepoScope(repo)); err == nil {
			return s
		}
	}
	if t.ParentID != "" {
		if p, err := c.store.Get(ctx, t.ParentID); err == nil {
			if repo := metaString(p.Meta, "github", "repo"); repo != "" {
				if s, err := store.ParseScope(store.RepoScope(repo)); err == nil {
					return s
				}
			}
		}
	}
	if t.Pool != "" {
		if s, err := store.ParseScope(store.PoolScope(t.Pool)); err == nil {
			return s
		}
	}
	return store.ScopeGlobal
}

// record stores a controller learning about t. A failure only logs: a
// learning must never fail the action that caused it.
func (c *Controller) record(ctx context.Context, kind string, t store.Ticket, summary, fix string) {
	summary = clip(summary, learnSummaryMax)
	l := store.Learning{
		Kind: kind, Summary: summary, Fix: clip(fix, learnSummaryMax), Scope: c.learnScope(ctx, t),
		Ticket: t.ID, Source: store.LearningSourceController,
	}
	if t.Kind == store.KindStep || t.Kind == store.KindTask {
		l.Session = t.Holder
	}
	if _, err := c.addLearning(ctx, l); err != nil {
		c.log.Error("record learning", "ticket", t.ID, "kind", kind, "err", err)
	}
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

// stepName is "<step> of <workflow> run <id>" for a run's step, else the id.
func (c *Controller) stepName(ctx context.Context, t store.Ticket) string {
	sf, err := store.DecodeFields[store.StepFields](t)
	if err != nil || t.Kind != store.KindStep || sf.Step == "" {
		return t.ID
	}
	wf := ""
	if r, err := c.store.Get(ctx, t.ParentID); err == nil {
		if rf, err := store.DecodeFields[store.RunFields](r); err == nil {
			wf = rf.Workflow + " "
		}
	}
	return fmt.Sprintf("step %s (%s) of %srun %s", sf.Step, t.ID, wf, t.ParentID)
}

// recordStepFailure records a failed close or a release that counts as a
// step failure. reason is the release reason or the worker's last failed
// report note.
func (c *Controller) recordStepFailure(ctx context.Context, t store.Ticket, how, reason string) {
	if t.Kind != store.KindStep {
		return
	}
	sf, _ := store.DecodeFields[store.StepFields](t)
	s := fmt.Sprintf("%s %s (failure %d/%d)", c.stepName(ctx, t), how, sf.Failures, sf.Retries)
	if reason = strings.TrimSpace(reason); reason != "" {
		s += ": " + reason
	}
	c.record(ctx, store.LearningFailure, t, s, "")
}

// failedReason is the note of the worker's last failed or dropped report.
func failedReason(t store.Ticket) string {
	if rm, ok := reportOf(t); ok && (rm.State == rpc.ReportFailed || rm.State == rpc.ReportDropped) {
		return rm.Note
	}
	return ""
}

// recordReviewFinding records a resume of a finished step by a reviewer:
// the review found something the step should have done. The finding is
// the reviewer's last comment on its own ticket, when it holds one.
func (c *Controller) recordReviewFinding(ctx context.Context, sc scope, step store.Ticket) {
	s := "review resumed " + c.stepName(ctx, step)
	if !sc.admin {
		if id, err := c.current(ctx, sc, ""); err == nil && id != step.ID {
			if cs, err := c.store.Comments(ctx, id); err == nil && len(cs) > 0 {
				s += ": " + cs[len(cs)-1].Body
			}
		}
	}
	c.record(ctx, store.LearningCorrection, step, s, "")
}

// recordAnswer records a human's answer to a question as a correction.
func (c *Controller) recordAnswer(ctx context.Context, qt store.Ticket, about, from, text string) {
	t, err := c.store.Get(ctx, about)
	if err != nil {
		t = qt
	}
	q := strings.TrimPrefix(qt.Title, "Question: ")
	c.record(ctx, store.LearningCorrection, t, fmt.Sprintf("question %s on %s: %s", qt.ID, about, q), "answer from "+from+": "+text)
}

// recordStall records a stall the stall trigger fired on.
func (c *Controller) recordStall(ctx context.Context, s stall, run string) {
	t := store.Ticket{ID: run}
	if st, err := c.store.Get(ctx, s.subject); err == nil {
		t = st
	}
	c.record(ctx, store.LearningFailure, t, fmt.Sprintf("stall %s on %s: %s (stall-check run %s)", s.signal, s.subject, s.detail, run), "")
}
