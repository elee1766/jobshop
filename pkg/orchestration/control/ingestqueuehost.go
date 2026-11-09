package control

import (
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/orchestration/ingest"
	"github.com/elee1766/jobshop/pkg/orchestration/ingest/queue"
	"github.com/elee1766/jobshop/pkg/support/clock"
)

// queue is the ingest queue (JOB-141, pkg/orchestration/ingest/queue), made on
// first use so controllers built without New (tests) get one too.
func (c *Controller) queue() *queue.Service {
	c.queueOnce.Do(func() {
		if c.ingestQ == nil {
			c.ingestQ = queue.New(queueHost{c}, c.log)
		}
	})
	return c.ingestQ
}

// queueHost is the controller as queue.Host.
type queueHost struct{ c *Controller }

func (h queueHost) Clock() clock.Clock               { return h.c.clk() }
func (h queueHost) Store() ingest.InboxStore         { return h.c.store }
func (h queueHost) SpoolPath() string                { return h.c.config().Ingest.SpoolPath(h.c.root) }
func (h queueHost) WakeWaits(its []domain.InboxItem) { h.c.wakeWaitsFor(its) }
func (h queueHost) Drained(its []domain.InboxItem)   { h.c.linearSvc().Delivered(its) }
