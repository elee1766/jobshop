package control

import (
	"context"
	"time"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/journal"
	"github.com/gfx-labs/jobshop/pkg/storage/journal/ship"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/support/redact"
)

// shipper is the journal shipper (#530, pkg/storage/journal/ship), made on
// first use so controllers built without Run get one too.
func (c *Controller) shipper() *ship.Shipper {
	c.shipOnce.Do(func() {
		c.ship = ship.New(shipHost{c}, c.log, c.metrics)
	})
	return c.ship
}

// pumpJournal ships journals until the controller stops.
func (c *Controller) pumpJournal() { c.shipper().Run(c.ctx) }

// shipHost is the controller as ship.Host.
type shipHost struct{ c *Controller }

func (h shipHost) Config() *config.Config                 { return h.c.config() }
func (h shipHost) Now() time.Time                         { return h.c.now() }
func (h shipHost) After(d time.Duration) <-chan time.Time { return h.c.after(d) }
func (h shipHost) SessionDir(id string) string            { return h.c.root.SessionDir(id) }
func (h shipHost) Redactor(session string) *redact.Redactor {
	return h.c.redactor(sessionEnv(h.c.root.SessionDir(session)))
}

func (h shipHost) List(ctx context.Context, f store.Filter) ([]domain.Ticket, error) {
	return h.c.store.List(ctx, f)
}

func (h shipHost) SaveCursor(ctx context.Context, session string, seq int64) error {
	return h.c.updateSessionFields(ctx, session, func(f *domain.SessionFields) { f.ShippedSeq = max(f.ShippedSeq, seq) })
}

// Work is a ticket's run, workflow and step, for row fields.
func (h shipHost) Work(id string) journal.Work {
	ctx := context.WithoutCancel(h.c.ctx)
	t, err := h.c.store.Get(ctx, id)
	if err != nil || t.Kind != domain.KindStep {
		return journal.Work{}
	}
	sf, err := domain.DecodeFields[domain.StepFields](t)
	if err != nil {
		return journal.Work{}
	}
	w := journal.Work{Run: sf.Run, Step: sf.Step}
	if run, err := h.c.store.Get(ctx, sf.Run); err == nil {
		if rf, err := domain.DecodeFields[domain.RunFields](run); err == nil {
			w.Workflow = rf.Workflow
		}
	}
	return w
}
