package control

import (
	"context"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/journalread"
	"github.com/elee1766/jobshop/pkg/control/journalship"
	"github.com/elee1766/jobshop/pkg/journal"
	"github.com/elee1766/jobshop/pkg/redact"
	"github.com/elee1766/jobshop/pkg/store"
)

// shipper is the journal shipper (#530, pkg/control/journalship), made on
// first use so controllers built without Run get one too.
func (c *Controller) shipper() *journalship.Shipper {
	c.shipOnce.Do(func() {
		c.ship = journalship.New(shipHost{c}, c.log, c.metrics)
	})
	return c.ship
}

// pumpJournal ships journals until the controller stops.
func (c *Controller) pumpJournal() { c.shipper().Run(c.ctx) }

// shipHost is the controller as journalship.Host.
type shipHost struct{ c *Controller }

func (h shipHost) Config() *config.Config                 { return h.c.config() }
func (h shipHost) Now() time.Time                         { return h.c.now() }
func (h shipHost) After(d time.Duration) <-chan time.Time { return h.c.after(d) }
func (h shipHost) SessionDir(id string) string            { return h.c.root.SessionDir(id) }
func (h shipHost) Redactor(session string) *redact.Redactor {
	return h.c.redactor(journalread.Env(h.c.root.SessionDir(session)))
}

func (h shipHost) List(ctx context.Context, f store.Filter) ([]store.Ticket, error) {
	return h.c.store.List(ctx, f)
}

func (h shipHost) SaveCursor(ctx context.Context, session string, seq int64) error {
	return h.c.updateSessionFields(ctx, session, func(f *store.SessionFields) { f.ShippedSeq = max(f.ShippedSeq, seq) })
}

// Work is a ticket's run, workflow and step, for row fields.
func (h shipHost) Work(id string) journal.Work {
	ctx := context.WithoutCancel(h.c.ctx)
	t, err := h.c.store.Get(ctx, id)
	if err != nil || t.Kind != store.KindStep {
		return journal.Work{}
	}
	sf, err := store.DecodeFields[store.StepFields](t)
	if err != nil {
		return journal.Work{}
	}
	w := journal.Work{Run: sf.Run, Step: sf.Step}
	if run, err := h.c.store.Get(ctx, sf.Run); err == nil {
		if rf, err := store.DecodeFields[store.RunFields](run); err == nil {
			w.Workflow = rf.Workflow
		}
	}
	return w
}
