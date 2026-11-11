package control

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Derived tickets (JOB-195, docs/notes/JOB-195-derived-ticket-close.md).
// A ticket the controller files on behalf of another ticket records that
// cause in its meta (run escalations: meta.run.id, questions:
// meta.question.ticket) and closes when the cause ends or a newer ticket
// of the same kind replaces it. The close comment names why.

const (
	// resolutionSuperseded closes an escalation a later round, or a
	// resumed run, replaced.
	resolutionSuperseded = "superseded"
	// resolutionWithdrawn closes a question whose asked ticket closed.
	resolutionWithdrawn = "withdrawn"
)

// escalationRound is the n of meta.run.escalation "<run>/<n>".
func escalationRound(t domain.Ticket) (int, bool) {
	key := metaString(t.Meta, runMetaPack, "escalation")
	i := strings.LastIndex(key, "/")
	if i < 0 {
		return 0, false
	}
	n, err := strconv.Atoi(key[i+1:])
	return n, err == nil
}

// openEscalations lists the open escalation tickets of run runID.
func (c *Controller) openEscalations(ctx context.Context, runID string) []domain.Ticket {
	var out []domain.Ticket
	for _, st := range domain.OpenStatuses {
		ts, err := c.store.List(ctx, store.Filter{Status: st, MetaPack: runMetaPack, MetaKey: "id", MetaValue: runID, Order: store.OrderOldest})
		if err != nil {
			return out
		}
		for _, t := range ts {
			if _, ok := escalationRound(t); ok {
				out = append(out, t)
			}
		}
	}
	return out
}

// settleEscalations closes the open escalation tickets of run runID that
// no longer describe it: all of them once the run is closed, the ones of
// an earlier round or of a run that is running again otherwise. A ticket
// a session holds is closed only with the run, with a comment saying so
// (the work it was doing is moot then); while the run lives it stays open,
// so nobody loses a claim mid-work.
func (c *Controller) settleEscalations(ctx context.Context, runID string) {
	escs := c.openEscalations(ctx, runID)
	if len(escs) == 0 {
		return
	}
	run, err := c.store.Get(ctx, runID)
	if err != nil || run.Kind != domain.KindRun {
		return
	}
	if run.Status == domain.StatusClosed {
		res, why := "done", "run "+runID+" done"
		if run.Resolution == domain.RunCancelled {
			res, why = domain.RunCancelled, "run "+runID+" cancelled"
		}
		for _, e := range escs {
			c.closeDerived(ctx, e, res, why, true)
		}
		return
	}
	f, err := domain.DecodeFields[domain.RunFields](run)
	if err != nil {
		return
	}
	current := ""
	if f.State == domain.RunNeedsHuman {
		current = fmt.Sprintf("%s/%d", runID, f.Escalations)
	}
	var newer domain.Ticket
	if current != "" {
		if ts, err := c.store.List(ctx, store.Filter{MetaPack: runMetaPack, MetaKey: "escalation", MetaValue: current, Limit: 1}); err == nil && len(ts) > 0 {
			newer = ts[0]
		}
	}
	for _, e := range escs {
		if metaString(e.Meta, runMetaPack, "escalation") == current {
			continue
		}
		why := "run " + runID + " resumed"
		if n, ok := escalationRound(e); ok && current != "" && n < f.Escalations {
			why = "superseded by " + newer.ID
			if newer.ID == "" {
				why = "superseded by a later escalation of run " + runID
			}
		}
		c.closeDerived(ctx, e, resolutionSuperseded, why, false)
	}
}

// closeDerived closes derived ticket t with a comment naming why. When
// final is false and a session holds t it stays open with one comment:
// the holder may be working it. A repeat of the same comment is skipped,
// so the sweep does not pile them up.
func (c *Controller) closeDerived(ctx context.Context, t domain.Ticket, resolution, why string, final bool) {
	if t.Status == domain.StatusClosed {
		return
	}
	if t.Holder != "" && !final {
		c.commentOnce(ctx, t.ID, fmt.Sprintf("%s. Left open because %s holds it; close it when done.", why, t.Holder))
		return
	}
	text := why
	if t.Holder != "" {
		text += fmt.Sprintf(" (closed while %s held it)", t.Holder)
	}
	if _, err := store.AddComment(ctx, c.store, "controller", t.ID, text); err != nil {
		c.log.WarnContext(ctx, "derived close comment", "ticket", t.ID, "err", err)
		return
	}
	closed, err := store.Close(ctx, c.store, "controller", t.ID, nil, resolution)
	if err != nil {
		c.log.WarnContext(ctx, "derived close", "ticket", t.ID, "err", err)
		return
	}
	c.log.InfoContext(ctx, "derived ticket closed", "ticket", t.ID, "resolution", resolution, "why", why)
	c.escalationClosed(ctx, closed)
	c.wakeQuestionClosed(closed)
}

// commentOnce adds text as the controller's comment on id unless it is
// already there.
func (c *Controller) commentOnce(ctx context.Context, id, text string) {
	comments, err := store.Comments(ctx, c.store, id)
	if err != nil {
		return
	}
	for _, cm := range comments {
		if cm.Author == "controller" && cm.Body == text {
			return
		}
	}
	_, _ = store.AddComment(ctx, c.store, "controller", id, text)
}

// sweepDerived repairs what the event hooks missed (a crash, a close by
// another path, tickets filed before JOB-195): escalations of runs that
// moved on and questions about closed tickets.
func (c *Controller) sweepDerived(ctx context.Context, _ *config.Config) {
	runs := map[string]bool{}
	for _, st := range domain.OpenStatuses {
		ts, err := c.store.List(ctx, store.Filter{Label: needsHumanLabel, Status: st, Kind: domain.KindTask})
		if err != nil {
			return
		}
		for _, t := range ts {
			if _, ok := escalationRound(t); ok {
				runs[metaString(t.Meta, runMetaPack, "id")] = true
			}
		}
	}
	for id := range runs {
		c.settleEscalations(ctx, id)
	}
	c.sweepOrphanQuestions(ctx)
}

// sweepOrphanQuestions closes open questions whose asked ticket closed:
// the answer would go to nobody who needs it.
func (c *Controller) sweepOrphanQuestions(ctx context.Context) {
	qs, err := c.store.List(ctx, store.Filter{Kind: domain.KindMessage, Status: domain.StatusOpen, Label: needsHumanLabel})
	if err != nil {
		return
	}
	for _, q := range qs {
		qm, ok := questionOf(q)
		if !ok {
			continue
		}
		asked, err := c.store.Get(ctx, qm.Ticket)
		if err != nil || asked.Status != domain.StatusClosed {
			continue
		}
		res := asked.Resolution
		if res == "" {
			res = "closed"
		}
		c.closeDerived(ctx, q, resolutionWithdrawn, fmt.Sprintf("asked ticket %s closed (%s)", asked.ID, res), true)
		c.linearSvc().Unblock(ctx, q)
	}
}
