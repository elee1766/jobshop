package control

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// events.subscribe (#147). Every store write records a history row with a
// global seq, so history is the change log: the store calls the hub after
// each commit, the hub's pump reads the new rows once, turns them into
// events and hands each subscriber a copy through a bounded channel. A
// full channel marks that subscriber lagged instead of waiting, so a slow
// client never holds up the pump or a write. Replay reads the same rows.

// DefaultEventBuffer is the per-subscriber event buffer.
const DefaultEventBuffer = 1024

type eventSub struct {
	ch     chan rpc.Event
	lagged chan struct{} // closed when the buffer overflowed
}

type hub struct {
	wake chan struct{} // cap 1; signaled after each commit
	size int

	mu   sync.Mutex
	last int64 // last seq published
	subs map[*eventSub]struct{}
}

func newHub(size int) *hub {
	if size <= 0 {
		size = DefaultEventBuffer
	}
	return &hub{wake: make(chan struct{}, 1), size: size, subs: map[*eventSub]struct{}{}}
}

// kick is the store's OnCommit. It never blocks.
func (h *hub) kick() {
	select {
	case h.wake <- struct{}{}:
	default:
	}
}

// add registers a subscriber and returns the last published seq: events
// after it arrive on the channel, earlier ones come from replay.
func (h *hub) add() (*eventSub, int64) {
	s := &eventSub{ch: make(chan rpc.Event, h.size), lagged: make(chan struct{})}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.subs[s] = struct{}{}
	return s, h.last
}

func (h *hub) remove(s *eventSub) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.subs, s)
}

// publish fans evs out. A subscriber whose buffer is full is dropped and
// told so through lagged.
func (h *hub) publish(evs []rpc.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ev := range evs {
		h.last = max(h.last, ev.Seq)
		for s := range h.subs {
			select {
			case s.ch <- ev:
			default:
				delete(h.subs, s)
				close(s.lagged)
			}
		}
	}
}

// pumpEvents publishes new history rows until the controller stops.
func (c *Controller) pumpEvents() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.events.wake:
		}
		for c.ctx.Err() == nil {
			c.events.mu.Lock()
			after := c.events.last
			c.events.mu.Unlock()
			evs, err := c.changes(c.ctx, after)
			if err != nil {
				if c.ctx.Err() == nil {
					c.log.Warn("events: read history", "err", err)
				}
				break
			}
			if len(evs) == 0 {
				break
			}
			c.events.publish(evs)
		}
	}
}

// changes reads one batch of history after seq as events.
func (c *Controller) changes(ctx context.Context, after int64) ([]rpc.Event, error) {
	rows, err := c.store.Changes(ctx, after, 0)
	if err != nil {
		return nil, err
	}
	tickets := map[string]*store.Ticket{}
	out := make([]rpc.Event, 0, len(rows))
	for _, r := range rows {
		t, ok := tickets[r.TicketID]
		if !ok {
			if got, err := c.store.Get(ctx, r.TicketID); err == nil {
				t = &got
			}
			tickets[r.TicketID] = t
		}
		out = append(out, toEvent(r, t))
	}
	return out, nil
}

// toEvent maps a history row. t is the ticket now, nil if it is gone.
func toEvent(h store.HistoryEntry, t *store.Ticket) rpc.Event {
	ev := rpc.Event{
		Seq: h.Seq, Time: h.At, Kind: rpc.EventKindTicket, Op: h.Op,
		ID: h.TicketID, Revision: h.Revision, Actor: h.Actor,
	}
	var diff struct {
		Label      string          `json:"label"`
		Resolution string          `json:"resolution"`
		Fields     json.RawMessage `json:"fields"`
	}
	_ = json.Unmarshal(h.Diff, &diff)
	ev.Label = diff.Label
	if t != nil {
		ev.TicketKind, ev.Status, ev.Resolution, ev.Labels = t.Kind, t.Status, t.Resolution, t.Labels
	}
	switch ev.TicketKind {
	case store.KindSession:
		ev.Kind = rpc.EventKindSession
		switch h.Op {
		case "create":
			ev.Op = rpc.OpSessionStart
		case "close":
			ev.Op = rpc.OpSessionStop
			if diff.Resolution == resolutionCrashed {
				ev.Op = rpc.OpSessionCrash
			}
		}
	case store.KindRun:
		ev.Kind = rpc.EventKindRun
		if t != nil {
			if f, err := store.DecodeFields[store.RunFields](*t); err == nil {
				ev.State = f.State
			}
		}
		if h.Op == "update" && len(diff.Fields) > 0 {
			var f store.RunFields
			if json.Unmarshal(diff.Fields, &f) == nil && f.State != "" {
				ev.Op, ev.State = rpc.OpRunState, f.State
			}
		}
	}
	return ev
}

func eventMatches(p rpc.EventsSubscribeParams, ev rpc.Event) bool {
	if len(p.Kinds) > 0 && !slices.Contains(p.Kinds, ev.Kind) {
		return false
	}
	if len(p.Labels) == 0 {
		return true
	}
	for _, l := range p.Labels {
		if l == ev.Label || slices.Contains(ev.Labels, l) {
			return true
		}
	}
	return false
}

// eventsSubscribe streams change events as events.event notifications
// until the caller hangs up, the controller stops or the caller lags.
func (c *Controller) eventsSubscribe(ctx context.Context, sc scope, p rpc.EventsSubscribeParams) (any, error) {
	if err := requireAdmin(sc, "events subscribe"); err != nil {
		return nil, err
	}
	if sc.peer == nil || sc.peer.conn == nil {
		return nil, fmt.Errorf("%w: subscribe needs a socket connection", store.ErrInvalid)
	}
	sub, live := c.events.add()
	defer c.events.remove(sub)
	res := rpc.EventsSubscribeResult{Seq: p.SinceSeq}
	send := func(ev rpc.Event) error {
		// Ready, lagged and activity events are not history rows (seq 0
		// or a resume mark) and always pass.
		ctl := ev.Kind == rpc.EventKindReady || ev.Kind == rpc.EventKindLagged || ev.Kind == rpc.EventKindActivity
		if ev.Seq <= res.Seq && !ctl {
			return nil
		}
		if ev.Seq > 0 && ev.Kind != rpc.EventKindReady {
			res.Seq = max(res.Seq, ev.Seq)
		}
		if !ctl && !eventMatches(p, ev) {
			return nil
		}
		if ev.Kind == rpc.EventKindActivity && len(p.Kinds) > 0 && !slices.Contains(p.Kinds, ev.Kind) {
			return nil
		}
		return sc.peer.conn.SendNotification(ctx, rpc.NotifyEvent, ev)
	}
	// Replay rows up to what the hub had published when we registered;
	// later ones are already queued on sub.ch.
	if p.SinceSeq > 0 {
		for res.Seq < live {
			evs, err := c.changes(ctx, res.Seq)
			if err != nil {
				return nil, err
			}
			if len(evs) == 0 {
				break
			}
			for _, ev := range evs {
				if ev.Seq > live {
					break
				}
				if err := send(ev); err != nil {
					return nil, err
				}
			}
			if evs[len(evs)-1].Seq >= live {
				break
			}
		}
	}
	res.Seq = max(res.Seq, live)
	if err := send(rpc.Event{Kind: rpc.EventKindReady, Seq: res.Seq}); err != nil {
		return nil, err
	}
	for {
		select {
		case <-ctx.Done():
			return res, nil
		case ev := <-sub.ch:
			if err := send(ev); err != nil {
				return nil, err
			}
		case <-sub.lagged:
			// Events still buffered were published before the overflow.
			for len(sub.ch) > 0 {
				if err := send(<-sub.ch); err != nil {
					return nil, err
				}
			}
			res.Lagged = true
			if err := send(rpc.Event{Kind: rpc.EventKindLagged, Seq: res.Seq}); err != nil {
				return nil, err
			}
			return res, nil
		}
	}
}
