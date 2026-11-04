package control

import (
	"context"
	"path/filepath"
	"time"

	host "github.com/elee1766/acphost"

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
	return c.store.Recall(ctx, store.RecallQuery{
		Text: p.Query, Project: proj, Kind: p.Kind, TicketID: p.TicketID, Limit: p.Limit,
	})
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
