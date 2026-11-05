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

// upsertMetaPack holds the key of a ticket made by ticket.upsert (#577):
// meta.upsert.key.
const upsertMetaPack = "upsert"

// upsertMu serializes ticket.upsert, so two reports of one problem at
// once make one ticket.
var upsertMu sync.Mutex

// upsert creates the open task with p.Key, updates its title, body and
// labels, or closes it. Outside tools (a deploy wrapper) report through
// it with no tool-specific code in the controller.
func (c *Controller) upsert(ctx context.Context, sc scope, p rpc.UpsertParams) (any, error) {
	if err := requireAdmin(sc, "ticket upsert"); err != nil {
		return nil, err
	}
	if p.Key == "" {
		return nil, fmt.Errorf("%w: upsert needs a key", store.ErrInvalid)
	}
	proj, err := c.project(sc, p.Project)
	if err != nil {
		return nil, err
	}
	upsertMu.Lock()
	defer upsertMu.Unlock()
	// Upsert never makes a second open ticket for a key, so the newest
	// one is the open one if any is open.
	found, err := c.store.List(ctx, store.Filter{Project: proj, MetaPack: upsertMetaPack, MetaKey: "key", MetaValue: p.Key, Order: store.OrderNewest, Limit: 1})
	if err != nil {
		return nil, err
	}
	var cur *store.Ticket
	if len(found) > 0 && found[0].Status != store.StatusClosed {
		cur = &found[0]
	}
	if p.Close {
		if cur == nil {
			return rpc.UpsertResult{Op: "none"}, nil
		}
		if note := joinNonEmpty(p.Title, p.Body); note != "" {
			if _, err := store.AddComment(ctx, c.store, sc.actor, cur.ID, note); err != nil {
				return nil, err
			}
		}
		res, err := c.closeTicket(ctx, sc, rpc.CloseParams{ID: cur.ID, Resolution: p.Resolution})
		if err != nil {
			return nil, err
		}
		t := res.(store.Ticket)
		return rpc.UpsertResult{Op: "closed", Ticket: &t}, nil
	}
	if p.Title == "" {
		return nil, fmt.Errorf("%w: upsert needs a title", store.ErrInvalid)
	}
	if cur == nil {
		meta, err := json.Marshal(map[string]map[string]string{upsertMetaPack: {"key": p.Key}})
		if err != nil {
			return nil, err
		}
		t, err := store.Create(ctx, c.store, sc.actor, store.NewTicket{
			Project: proj, Kind: store.KindTask, Title: p.Title, Body: p.Body,
			Pool: p.Pool, Labels: p.Labels, Meta: meta,
		})
		if err != nil {
			return nil, err
		}
		return rpc.UpsertResult{Op: "created", Ticket: &t}, nil
	}
	op := "unchanged"
	var u store.Update
	if cur.Title != p.Title {
		u.Title = &p.Title
	}
	if cur.Body != p.Body {
		u.Body = &p.Body
	}
	if u.Title != nil || u.Body != nil {
		t, err := c.store.Update(ctx, sc.actor, cur.ID, u)
		if err != nil {
			return nil, err
		}
		cur, op = &t, "updated"
	}
	for _, l := range p.Labels {
		if slices.Contains(cur.Labels, l) {
			continue
		}
		if err := store.AddLabel(ctx, c.store, sc.actor, cur.ID, l); err != nil {
			return nil, err
		}
		cur.Labels = append(cur.Labels, l)
		op = "updated"
	}
	return rpc.UpsertResult{Op: op, Ticket: cur}, nil
}

func joinNonEmpty(a, b string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	}
	return a + "\n\n" + b
}
