package control

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/knowledge"
	"github.com/gfx-labs/jobshop/pkg/orchestration/triggers"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Learnings (#264). Agents record corrections, failures and workarounds
// with `jobshop learn`; the controller records the ones it sees itself
// (failed steps, review findings, escalations, answers, findings). They live
// in their own store table, so no pool ever claims them; the curator reads
// them with `jobshop learn list`.

// learnSummaryMax bounds a controller-written summary.
const learnSummaryMax = 500

func (c *Controller) learnAdd(ctx context.Context, sc scope, p rpc.LearnParams) (any, error) {
	if err := domain.CheckLearningKind(strings.TrimSpace(p.Kind), c.config().LearningKinds()); err != nil {
		return nil, err
	}
	l := domain.Learning{
		Kind: p.Kind, Summary: p.Summary, Fix: p.Fix, Scope: p.Scope, Ticket: p.Ticket,
		Session: sc.session, Source: domain.LearningSourceAgent,
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
func (c *Controller) addLearning(ctx context.Context, l domain.Learning) (domain.Learning, error) {
	project := ""
	if l.Ticket != "" {
		if t, err := c.store.Get(ctx, l.Ticket); err == nil {
			project = t.Project
		}
	}
	return knowledge.AddLearning(ctx, c.store, l, project)
}

func (c *Controller) learnList(ctx context.Context, _ scope, p rpc.LearnListParams) (any, error) {
	f := store.LearningFilter{Kind: p.Kind, Status: p.Status, Ticket: p.Ticket, DuplicateOf: p.DuplicateOf, Limit: p.Limit}
	// A glob kind (#524) is matched here: fetch without it, then filter.
	glob := strings.ContainsAny(p.Kind, "*?[")
	if glob {
		if _, err := path.Match(p.Kind, ""); err != nil {
			return nil, fmt.Errorf("%w: kind %q: %w", domain.ErrInvalid, p.Kind, err)
		}
		f.Kind, f.Limit = "", store.DefaultLimit
	}
	if p.Scope != "" {
		s, err := domain.ParseScope(p.Scope)
		if err != nil {
			return nil, err
		}
		f.Scope = s
	}
	if p.Since > 0 {
		f.Since = c.now().Add(-p.Since)
	}
	ls, err := knowledge.ListLearnings(ctx, c.store, f)
	if err != nil || !glob {
		return ls, err
	}
	limit := p.Limit
	if limit <= 0 {
		limit = store.DefaultLimit
	}
	out := []domain.Learning{}
	for _, l := range ls {
		if domain.Glob(p.Kind, l.Kind) && len(out) < limit {
			out = append(out, l)
		}
	}
	return out, nil
}

// learnUpdate sets status and PR on learnings (#273), or marks them
// duplicates of an open one (JOB-199). Open to every session: the curator
// marks the PR it opened, the refinery or the next curate pass marks the
// outcome. A rejected learning must name its PR, so the curator can see
// what was turned down. A duplicate takes no outcome of its own: it can
// only be reopened, and applying or rejecting a group goes through the
// learning kept.
func (c *Controller) learnUpdate(ctx context.Context, _ scope, p rpc.LearnUpdateParams) (any, error) {
	if len(p.IDs) == 0 {
		return nil, fmt.Errorf("%w: learn update needs at least one id", domain.ErrInvalid)
	}
	if p.DuplicateOf != 0 {
		if p.Status != "" || strings.TrimSpace(p.PR) != "" {
			return nil, fmt.Errorf("%w: --duplicate-of cannot be combined with --status or --pr", domain.ErrInvalid)
		}
		return knowledge.MarkDuplicates(ctx, c.store, p.IDs, p.DuplicateOf)
	}
	if p.Status == domain.LearningDuplicate {
		return nil, fmt.Errorf("%w: use --duplicate-of <id> to mark duplicates", domain.ErrInvalid)
	}
	if p.Status == "" && strings.TrimSpace(p.PR) == "" {
		return nil, fmt.Errorf("%w: learn update needs --status, --pr or --duplicate-of", domain.ErrInvalid)
	}
	cur := make([]domain.Learning, 0, len(p.IDs))
	for _, id := range p.IDs {
		l, err := knowledge.GetLearning(ctx, c.store, id)
		if err != nil {
			return nil, err
		}
		if l.Status == domain.LearningDuplicate && p.Status != domain.LearningOpen {
			return nil, fmt.Errorf("%w: learning %d is a duplicate of %d: update %d instead, or reopen it with --status open", domain.ErrInvalid, id, l.DuplicateOf, l.DuplicateOf)
		}
		st, pr := l.Status, l.PR
		if p.Status != "" {
			st = p.Status
		}
		if strings.TrimSpace(p.PR) != "" {
			pr = p.PR
		}
		if st == domain.LearningRejected && strings.TrimSpace(pr) == "" {
			return nil, fmt.Errorf("%w: learning %d: rejected needs --pr", domain.ErrInvalid, id)
		}
		l.Status, l.PR = st, pr
		cur = append(cur, l)
	}
	out := make([]domain.Learning, 0, len(cur))
	merged := map[string][]int64{}
	var prs []string
	for i, l := range cur {
		u, err := knowledge.UpdateLearning(ctx, c.store, l.ID, l.Status, l.PR)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
		// memory:<id> is a local write (#592), recorded by memory.write.
		if p.Status == domain.LearningApplied && u.PR != "" && !knowledge.IsMemoryRef(u.PR) {
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

// learnExpire sets expired on open learnings recorded before the cutoff
// that have no PR (JOB-198), so curator passes stop reading them. Admin
// only: it changes many records in one call. An expired learning is not
// deleted, recall still finds it, and `learn update --status open` reopens
// it.
func (c *Controller) learnExpire(ctx context.Context, sc scope, p rpc.LearnExpireParams) (any, error) {
	if err := requireAdmin(sc, "learn expire"); err != nil {
		return nil, err
	}
	if p.OlderThan <= 0 {
		return nil, fmt.Errorf("%w: learn expire needs --older-than, a positive duration", domain.ErrInvalid)
	}
	if strings.ContainsAny(p.Kind, "*?[") {
		if _, err := path.Match(p.Kind, ""); err != nil {
			return nil, fmt.Errorf("%w: kind %q: %w", domain.ErrInvalid, p.Kind, err)
		}
	}
	scopeName := ""
	if p.Scope != "" {
		s, err := domain.ParseScope(p.Scope)
		if err != nil {
			return nil, err
		}
		scopeName = s
	}
	ls, err := knowledge.ExpireLearnings(ctx, c.store, c.now().Add(-p.OlderThan), scopeName, p.Kind)
	if err != nil {
		return nil, err
	}
	c.log.InfoContext(ctx, "learnings expired", "count", len(ls), "older_than", p.OlderThan, "by", sc.actor)
	return rpc.LearnExpireResult{Expired: len(ls)}, nil
}

// learnScope is the narrowest scope a ticket names: its repo (from its
// meta or its run's), else its pool, else global.
func (c *Controller) learnScope(ctx context.Context, t domain.Ticket) string {
	if repo := metaString(t.Meta, "github", "repo"); repo != "" {
		if s, err := domain.ParseScope(domain.RepoScope(repo)); err == nil {
			return s
		}
	}
	if t.ParentID != "" {
		if p, err := c.store.Get(ctx, t.ParentID); err == nil {
			if repo := metaString(p.Meta, "github", "repo"); repo != "" {
				if s, err := domain.ParseScope(domain.RepoScope(repo)); err == nil {
					return s
				}
			}
		}
	}
	if t.Pool != "" {
		if s, err := domain.ParseScope(domain.PoolScope(t.Pool)); err == nil {
			return s
		}
	}
	return domain.ScopeGlobal
}

// record stores a controller learning about t. A failure only logs: a
// learning must never fail the action that caused it.
func (c *Controller) record(ctx context.Context, kind string, t domain.Ticket, summary, fix string) {
	summary = clip(summary, learnSummaryMax)
	l := domain.Learning{
		Kind: kind, Summary: summary, Fix: clip(fix, learnSummaryMax), Scope: c.learnScope(ctx, t),
		Ticket: t.ID, Source: domain.LearningSourceController,
	}
	if t.Kind == domain.KindStep || t.Kind == domain.KindTask {
		l.Session = t.Holder
	}
	if _, err := c.addLearning(ctx, l); err != nil {
		c.log.ErrorContext(ctx, "record learning", "ticket", t.ID, "kind", kind, "err", err)
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
func (c *Controller) stepName(ctx context.Context, t domain.Ticket) string {
	sf, err := domain.DecodeFields[domain.StepFields](t)
	if err != nil || t.Kind != domain.KindStep || sf.Step == "" {
		return t.ID
	}
	wf := ""
	if r, err := c.store.Get(ctx, t.ParentID); err == nil {
		if rf, err := domain.DecodeFields[domain.RunFields](r); err == nil {
			wf = rf.Workflow + " "
		}
	}
	return fmt.Sprintf("step %s (%s) of %srun %s", sf.Step, t.ID, wf, t.ParentID)
}

// recordStepFailure records a failed close or a release that counts as a
// step failure. reason is the release reason or the worker's last failed
// report note.
func (c *Controller) recordStepFailure(ctx context.Context, t domain.Ticket, how, reason string) {
	if t.Kind != domain.KindStep {
		return
	}
	sf, _ := domain.DecodeFields[domain.StepFields](t)
	s := fmt.Sprintf("%s %s (failure %d/%d)", c.stepName(ctx, t), how, sf.Failures, sf.Retries)
	if reason = strings.TrimSpace(reason); reason != "" {
		s += ": " + reason
	}
	c.record(ctx, domain.LearningFailure, t, s, "")
}

// failedReason is the note of the worker's last failed or dropped report.
func failedReason(t domain.Ticket) string {
	if rm, ok := reportOf(t); ok && (rm.State == rpc.ReportFailed || rm.State == rpc.ReportDropped) {
		return rm.Note
	}
	return ""
}

// recordReviewFinding records a resume of a finished step by a reviewer:
// the review found something the step should have done. The finding is
// the reviewer's last comment on its own ticket, when it holds one.
func (c *Controller) recordReviewFinding(ctx context.Context, sc scope, step domain.Ticket) {
	s := "review resumed " + c.stepName(ctx, step)
	if !sc.admin {
		if id, err := c.current(ctx, sc, ""); err == nil && id != step.ID {
			if cs, err := store.Comments(ctx, c.store, id); err == nil && len(cs) > 0 {
				s += ": " + cs[len(cs)-1].Body
			}
		}
	}
	c.record(ctx, domain.LearningCorrection, step, s, "")
}

// recordAnswer records a human's answer to a question as a correction.
func (c *Controller) recordAnswer(ctx context.Context, qt domain.Ticket, about, from, text string) {
	t, err := c.store.Get(ctx, about)
	if err != nil {
		t = qt
	}
	q := strings.TrimPrefix(qt.Title, "Question: ")
	c.record(ctx, domain.LearningCorrection, t, fmt.Sprintf("question %s on %s: %s", qt.ID, about, q), "answer from "+from+": "+text)
}

// recordFinding records a finding the scan trigger fired on.
func (c *Controller) recordFinding(ctx context.Context, s triggers.Finding, run string) {
	t := domain.Ticket{ID: run}
	if st, err := c.store.Get(ctx, s.Subject); err == nil {
		t = st
	}
	c.record(ctx, domain.LearningFailure, t, fmt.Sprintf("finding %s on %s: %s (triage run %s)", s.Signal, s.Subject, s.Detail, run), "")
}
