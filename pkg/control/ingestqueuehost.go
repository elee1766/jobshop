package control

import (
	"github.com/elee1766/jobshop/pkg/clock"
	"github.com/elee1766/jobshop/pkg/control/ingestqueue"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/ingest"
)

// queue is the ingest queue (JOB-141, pkg/control/ingestqueue), made on
// first use so controllers built without New (tests) get one too.
func (c *Controller) queue() *ingestqueue.Service {
	c.queueOnce.Do(func() {
		if c.ingestQ == nil {
			c.ingestQ = ingestqueue.New(queueHost{c}, c.log)
		}
	})
	return c.ingestQ
}

// queueHost is the controller as ingestqueue.Host.
type queueHost struct{ c *Controller }

func (h queueHost) Clock() clock.Clock               { return h.c.clk() }
func (h queueHost) Store() ingest.InboxStore         { return h.c.store }
func (h queueHost) SpoolPath() string                { return h.c.config().Ingest.SpoolPath(h.c.root) }
func (h queueHost) WakeWaits(its []domain.InboxItem) { h.c.wakeWaitsFor(its) }
func (h queueHost) Drained(its []domain.InboxItem)   { h.c.linearSvc().Delivered(its) }
