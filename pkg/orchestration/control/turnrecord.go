package control

import (
	"context"
	"errors"
	"time"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Turn rows (JOB-184). Whether a session runs a turn is otherwise only the
// runner's memory, and past turns only the session journal, which is too big
// to replay per trace. So a turn leaves a turn_start and a turn_end history
// row on every ticket the session holds in progress. They are written with
// store.Record: the ticket's revision does not move, so they never supersede
// a callback's read set or an if_revision write. A session holding nothing
// (a chat) records nothing.

// turnEndWrite bounds the turn_end write: it must not hang a runner that is
// finishing a turn.
const turnEndWrite = 10 * time.Second

// recordTurnStart writes turn_start on each ticket the session holds in
// progress and returns their ids, for recordTurnEnd.
func (r *runner) recordTurnStart(ctx context.Context) []string {
	held, err := r.c.store.List(ctx, store.Filter{Holder: r.id, Status: domain.StatusInProgress})
	if err != nil {
		if ctx.Err() == nil {
			r.log.WarnContext(ctx, "turn start: list held tickets", "err", err)
		}
		return nil
	}
	var ids []string
	for _, t := range held {
		if err := r.c.store.Record(ctx, r.id, t.ID, domain.OpTurnStart, struct{}{}); err != nil {
			if ctx.Err() == nil && !errors.Is(err, domain.ErrNotFound) {
				r.log.WarnContext(ctx, "turn start: record", "ticket", t.ID, "err", err)
			}
			continue
		}
		ids = append(ids, t.ID)
	}
	return ids
}

// recordTurnEnd writes turn_end with the stop reason on the tickets that
// got the turn_start, also those the turn closed or released. A lost
// connection (ctx done) writes nothing: the turn did not end, the
// controller did, and the timeline bounds it by the next turn or the claim.
func (r *runner) recordTurnEnd(ctx context.Context, ids []string, stop string) {
	if len(ids) == 0 || ctx.Err() != nil {
		return
	}
	wctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), turnEndWrite)
	defer cancel()
	for _, id := range ids {
		if err := r.c.store.Record(wctx, r.id, id, domain.OpTurnEnd, domain.TurnDiff{Stop: stop}); err != nil && !errors.Is(err, domain.ErrNotFound) {
			r.log.WarnContext(ctx, "turn end: record", "ticket", id, "err", err)
		}
	}
}
