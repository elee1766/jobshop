package control

import (
	"context"
	"time"

	"github.com/elee1766/jobshop/pkg/control/threads"
	"github.com/elee1766/jobshop/pkg/control/waitengine"
	"github.com/elee1766/jobshop/pkg/store"
)

// The controller side of targeted wait wakes (#679). waitengine.Wakes
// queues what an event can change; the pump below decides when to run a
// pass and applies the decisions under waitMu.

// wakeWaits asks the wait pump for a full pass now.
func (c *Controller) wakeWaits() {
	c.waitWakes.All()
	c.signalWaits()
}

// wakeWaitsSubjects asks the pump to evaluate the waits on subjects.
func (c *Controller) wakeWaitsSubjects(subjects ...string) {
	if c.waitWakes.Subjects(subjects...) {
		c.signalWaits()
	}
}

// wakeWaitsFor wakes the waits new inbox items name.
func (c *Controller) wakeWaitsFor(items []store.InboxItem) {
	if c.waitWakes.Items(items) {
		c.signalWaits()
	}
}

func (c *Controller) signalWaits() {
	if c.waitWake == nil {
		return
	}
	select {
	case c.waitWake <- struct{}{}:
	default:
	}
}

// pumpWaits evaluates waits: all of them on the interval, at the
// earliest time condition and on a full wake; only the named ones on a
// targeted wake.
func (c *Controller) pumpWaits() {
	full := true
	var subj map[string]bool
	for {
		if full {
			c.pollWaits(c.ctx, c.waitSrc(), c.threadSrc())
		} else {
			c.pollWaitsOn(c.ctx, c.waitSrc(), c.threadSrc(), subj)
		}
		next := c.untilNextTime(c.ctx)
		select {
		case <-c.ctx.Done():
			return
		case <-c.after(next):
			full, subj = true, nil
		case <-c.waitWake:
			w := c.waitWakes.Take()
			full, subj = w.All, w.Subjects
		}
	}
}

// pollWaitsOn evaluates the waits with a condition on one of subjects.
func (c *Controller) pollWaitsOn(ctx context.Context, src waitengine.Source, ths threads.Source, subjects map[string]bool) {
	if len(subjects) == 0 {
		return
	}
	c.waitMu.Lock()
	defer c.waitMu.Unlock()
	ts, err := c.waiting(ctx, "")
	if err != nil {
		if ctx.Err() == nil {
			c.log.WarnContext(ctx, "waits: list", "err", err)
		}
		return
	}
	pass := c.newWaitPass(src, ths)
	for _, t := range ts {
		if !waitengine.Names(waitOf(t), subjects) {
			continue
		}
		if err := c.evalWait(ctx, pass, t); err != nil && ctx.Err() == nil {
			c.log.WarnContext(ctx, "waits: evaluate", "ticket", t.ID, "err", err)
		}
	}
}

// untilNextTime is how long until the earliest unmet time condition or
// wait timeout, capped at the interval.
func (c *Controller) untilNextTime(ctx context.Context) time.Duration {
	ts, err := c.waiting(ctx, "")
	if err != nil {
		return c.waitEvery()
	}
	return c.nextWaitTime(ts)
}

// nextWaitTime is untilNextTime over ts.
func (c *Controller) nextWaitTime(ts []store.Ticket) time.Duration {
	var ws []store.Wait
	for _, t := range ts {
		if w := waitOf(t); w != nil {
			ws = append(ws, *w)
		}
	}
	return waitengine.NextDue(ws, c.now(), c.waitEvery())
}
