package control

import (
	"context"

	"github.com/elee1766/jobshop/pkg/control/journalread"
	"github.com/elee1766/jobshop/pkg/journal"
	"github.com/elee1766/jobshop/pkg/learning"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// recall searches comments, session journals and learnings (#265). The
// project scope is decided here and the three sources are merged here;
// journal.Recall and learning.Recall query their own collections, and
// journalread.Summary builds what indexJournal writes (JOB-140).
func (c *Controller) recall(ctx context.Context, sc scope, p rpc.RecallParams) (any, error) {
	proj := p.Project
	if !sc.admin {
		var err error
		if proj, err = c.project(sc, p.Project); err != nil {
			return nil, err
		}
	}
	limit := p.Limit
	if limit <= 0 {
		limit = recallLimit
	}
	want := func(kind string) bool { return p.Kind == "" || p.Kind == kind }
	var sources [][]store.RecallHit
	if want(store.RecallComment) {
		hits, err := c.store.Recall(ctx, store.RecallQuery{Text: p.Query, Project: proj, TicketID: p.TicketID, Limit: limit})
		if err != nil {
			return nil, err
		}
		sources = append(sources, hits)
	}
	if want(store.RecallJournal) {
		hits, err := journal.Recall(ctx, c.store, p.Query, proj, p.TicketID, limit)
		if err != nil {
			return nil, err
		}
		sources = append(sources, hits)
	}
	if want(store.RecallLearning) {
		hits, err := learning.Recall(ctx, c.store, p.Query, proj, p.TicketID, limit)
		if err != nil {
			return nil, err
		}
		sources = append(sources, hits)
	}
	return mergeRecall(sources, limit), nil
}

// recallLimit is recall's limit when none is given.
const recallLimit = 10

// mergeRecall interleaves the hits of each source round robin, each in
// its own rank order: comments, journals and learnings have separate
// full-text indexes (#579) and their scores do not compare.
func mergeRecall(sources [][]store.RecallHit, limit int) []store.RecallHit {
	out := []store.RecallHit{}
	for i := 0; len(out) < limit; i++ {
		more := false
		for _, s := range sources {
			if i < len(s) && len(out) < limit {
				out = append(out, s[i])
				more = true
			}
		}
		if !more {
			break
		}
	}
	return out
}

// indexJournal replaces the session's record in the journals collection
// with a summary of its journal. One record per session keeps
// re-indexing after every turn idempotent across controller restarts.
func (r *runner) indexJournal(ctx context.Context, ticketID string) {
	text, at, ok := journalread.Summary(r.dir)
	if !ok {
		return
	}
	d := journal.Doc{Session: r.id, Ticket: ticketID, Project: r.project(), At: at, Text: text}
	if err := journal.Index(context.WithoutCancel(ctx), r.c.store, d); err != nil {
		r.log.WarnContext(ctx, "index journal for recall", "err", err)
	}
}
