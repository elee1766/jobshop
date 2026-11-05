package control

import (
	"context"
	"path/filepath"
	"time"

	host "github.com/elee1766/acphost"

	"github.com/elee1766/jobshop/pkg/learning"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// recallJournalBytes bounds the journal text indexed per session. Summarize
// keeps the first prompt and the newest items when a journal is longer.
const recallJournalBytes = 64 << 10

// recall searches comments, session journals and learnings (#265).
func (c *Controller) recall(ctx context.Context, sc scope, p rpc.RecallParams) (any, error) {
	proj := p.Project
	if !sc.admin {
		var err error
		if proj, err = c.project(sc, p.Project); err != nil {
			return nil, err
		}
	}
	q := store.RecallQuery{Text: p.Query, Project: proj, Kind: p.Kind, TicketID: p.TicketID, Limit: p.Limit}
	if q.Limit <= 0 {
		q.Limit = recallLimit
	}
	var idx, ls []store.RecallHit
	if p.Kind != store.RecallLearning {
		var err error
		if idx, err = c.store.Recall(ctx, q); err != nil {
			return nil, err
		}
	}
	if p.Kind == "" || p.Kind == store.RecallLearning {
		var err error
		if ls, err = learning.Recall(ctx, c.store, p.Query, proj, p.TicketID, q.Limit); err != nil {
			return nil, err
		}
	}
	return mergeRecall(idx, ls, q.Limit), nil
}

// recallLimit is recall's limit when none is given.
const recallLimit = 10

// mergeRecall interleaves comment and journal hits with learning hits,
// each in its own rank order: learnings have their own full-text index
// (#579) and the two scores do not compare.
func mergeRecall(a, b []store.RecallHit, limit int) []store.RecallHit {
	out := make([]store.RecallHit, 0, min(limit, len(a)+len(b)))
	for i := 0; len(out) < limit && (i < len(a) || i < len(b)); i++ {
		if i < len(a) {
			out = append(out, a[i])
		}
		if i < len(b) && len(out) < limit {
			out = append(out, b[i])
		}
	}
	return out
}

// indexJournal replaces the session's journal document in the recall
// index with a summary of its journal. One document per session keeps
// re-indexing after every turn idempotent across controller restarts.
func (r *runner) indexJournal(ctx context.Context, ticketID string) {
	frames, _, err := host.ReadJournal(filepath.Join(r.dir, host.JournalFile), 0)
	if err != nil || len(frames) == 0 {
		return
	}
	doc := store.RecallDoc{
		Kind: store.RecallJournal, Ref: r.id, TicketID: ticketID, Project: r.project(),
		At:   time.UnixMilli(frames[len(frames)-1].T),
		Body: host.Summarize(frames, recallJournalBytes),
	}
	if err := r.c.store.IndexRecall(context.WithoutCancel(ctx), doc); err != nil {
		r.log.Warn("index journal for recall", "err", err)
	}
}
