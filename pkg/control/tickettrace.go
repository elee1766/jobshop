package control

import (
	"context"
	"encoding/json"
	"regexp"
	"slices"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// ghLink matches GitHub issue and PR URLs in ticket bodies and comments.
var ghLink = regexp.MustCompile(`https://github\.com/[\w.-]+/[\w.-]+/(?:issues|pull)/\d+`)

const defaultTraceDepth = 5

// trace builds the chain from a ticket: its trigger, holders, links, and the
// tickets it led to. A child is a ticket whose parent is this one, or one a
// holder session created while it held this ticket. The second rule is what
// connects a plan ticket to the dev tickets its planner filed.
func (c *Controller) trace(ctx context.Context, sc scope, p rpc.TraceParams) (any, error) {
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	depth := p.Depth
	if depth <= 0 {
		depth = defaultTraceDepth
	}
	seen := map[string]bool{}
	return c.traceNode(ctx, id, depth, seen)
}

func (c *Controller) traceNode(ctx context.Context, id string, depth int, seen map[string]bool) (rpc.TraceNode, error) {
	seen[id] = true
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return rpc.TraceNode{}, err
	}
	n := rpc.TraceNode{
		ID: t.ID, Title: t.Title, Pool: t.Pool, Status: t.Status,
		Resolution: t.Resolution, CreatedBy: t.CreatedBy, Trigger: triggerName(t.Meta),
	}
	th, err := c.store.Thread(ctx, id, store.AllParts)
	if err != nil {
		return n, err
	}
	n.Edges = th.Edges
	n.Links = links(t.Body, th.Comments)
	hist := th.History
	windows := holdWindows(hist, c.now())
	for _, w := range windows {
		if !slices.Contains(n.Holders, w.holder) {
			n.Holders = append(n.Holders, w.holder)
		}
	}
	if depth <= 1 {
		return n, nil
	}
	kids, err := c.childIDs(ctx, id, windows)
	if err != nil {
		return n, err
	}
	for _, k := range kids {
		if seen[k] {
			continue
		}
		child, err := c.traceNode(ctx, k, depth-1, seen)
		if err != nil {
			return n, err
		}
		n.Children = append(n.Children, child)
	}
	return n, nil
}

// holdWindow is one claim of a ticket, from claim to the next close,
// release or lease expiry.
type holdWindow struct {
	holder     string
	start, end time.Time
}

func holdWindows(hist []store.HistoryEntry, now time.Time) []holdWindow {
	var out []holdWindow
	open := -1
	for _, h := range hist {
		switch h.Op {
		case "claim":
			if open >= 0 && out[open].holder == h.Actor {
				continue // re-claim by the same holder extends the window
			}
			if open >= 0 {
				out[open].end = h.At
			}
			out = append(out, holdWindow{holder: h.Actor, start: h.At, end: now})
			open = len(out) - 1
		case "close", "release", "lease_expired":
			if open >= 0 {
				out[open].end = h.At
				open = -1
			}
		}
	}
	return out
}

func (c *Controller) childIDs(ctx context.Context, id string, windows []holdWindow) ([]string, error) {
	var out []string
	kids, err := c.store.List(ctx, store.Filter{ParentID: id})
	if err != nil {
		return nil, err
	}
	for _, k := range kids {
		out = append(out, k.ID)
	}
	for _, w := range windows {
		made, err := c.store.List(ctx, store.Filter{CreatedBy: w.holder})
		if err != nil {
			return nil, err
		}
		for _, k := range made {
			if k.Kind == store.KindSession || k.ID == id || slices.Contains(out, k.ID) {
				continue
			}
			if !k.CreatedAt.Before(w.start) && !k.CreatedAt.After(w.end) {
				out = append(out, k.ID)
			}
		}
	}
	return out, nil
}

func triggerName(meta json.RawMessage) string {
	var m struct {
		Trigger struct {
			Name string `json:"name"`
		} `json:"trigger"`
	}
	if len(meta) == 0 || json.Unmarshal(meta, &m) != nil {
		return ""
	}
	return m.Trigger.Name
}

func links(body string, comments []store.Comment) []string {
	var out []string
	add := func(s string) {
		for _, l := range ghLink.FindAllString(s, -1) {
			if !slices.Contains(out, l) {
				out = append(out, l)
			}
		}
	}
	add(body)
	for _, cm := range comments {
		add(cm.Body)
	}
	return out
}
