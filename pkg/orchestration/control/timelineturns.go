package control

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// sessionBusy reports whether the controller runs a turn on the session now.
func (c *Controller) sessionBusy(id string) bool {
	c.mu.Lock()
	r := c.runners[id]
	c.mu.Unlock()
	return r != nil && r.turning.Load()
}

// turnSpan is one turn found in a ticket's history, before it is placed on
// a claim. open is set when no turn_end closed it and no later turn_start
// of the session cut it.
type turnSpan struct {
	actor string
	rpc.TimelineTurn
	open bool
}

// pairTurns reads the turn_start and turn_end rows (oldest first) into
// turns per session. A turn_start while the session's last turn is
// unfinished ends that turn there, with no stop reason (the controller
// stopped mid-turn). A turn_end with no turn_start before it is dropped.
func pairTurns(rows []store.HistoryEntry) []turnSpan {
	var out []turnSpan
	pending := map[string]int{} // actor -> index in out of its unfinished turn
	for _, h := range rows {
		switch h.Op {
		case domain.OpTurnStart:
			if i, ok := pending[h.Actor]; ok {
				out[i].End, out[i].open = h.At, false
			}
			out = append(out, turnSpan{actor: h.Actor, TimelineTurn: rpc.TimelineTurn{Start: h.At}, open: true})
			pending[h.Actor] = len(out) - 1
		case domain.OpTurnEnd:
			i, ok := pending[h.Actor]
			if !ok {
				continue
			}
			var d domain.TurnDiff
			_ = json.Unmarshal(h.Diff, &d)
			out[i].End, out[i].Stop, out[i].open = h.At, d.Stop, false
			delete(pending, h.Actor)
		}
	}
	return out
}

// withTurns lists on each working or retry span the turns of its holder
// that ran inside the claim, clipped to it. A span with no turns keeps a
// nil list: its claim has no turn rows (older history, an external session,
// a step the controller runs) and the chart draws it as held, not idle.
// A turn row that falls in no claim of its session is dropped. end is where
// an open span stops (now, or the ticket's close). busy says whether a
// session is in a turn now: only then is the last turn of an open claim
// left running; otherwise a turn without a turn_end ends at the claim's end.
func withTurns(segs []rpc.TimelineSegment, rows []store.HistoryEntry, end time.Time, busy func(string) bool) []rpc.TimelineSegment {
	if len(rows) == 0 {
		return segs
	}
	spanEnd := func(s rpc.TimelineSegment) time.Time {
		if s.End.IsZero() {
			return end
		}
		return s.End
	}
	for _, t := range pairTurns(rows) {
		// The latest claim by the session that began by the time the turn
		// did: a turn belongs to the claim it started in.
		at := -1
		for i, s := range segs {
			if (s.State == rpc.TimelineWorking || s.State == rpc.TimelineRetry) && s.Actor == t.actor && !s.Start.After(t.Start) {
				at = i
			}
		}
		if at < 0 || t.Start.After(spanEnd(segs[at])) {
			continue
		}
		seg := &segs[at]
		stop := spanEnd(*seg)
		turn := t.TimelineTurn
		switch {
		case t.open && seg.End.IsZero() && busy != nil && busy(t.actor):
			turn.End = time.Time{} // running now
		case t.open || turn.End.After(stop):
			turn.End = stop
		}
		seg.Turns = append(seg.Turns, turn)
	}
	for i := range segs {
		// A running turn is the last of its claim. An earlier unfinished
		// turn was cut by the next turn_start, so it has an end already.
		slices.SortStableFunc(segs[i].Turns, func(a, b rpc.TimelineTurn) int { return a.Start.Compare(b.Start) })
	}
	return segs
}
