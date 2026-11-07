package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
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
	id, err = c.traceReference(ctx, sc, id)
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

// traceReference resolves external references without guessing between roots.
func (c *Controller) traceReference(ctx context.Context, sc scope, ref string) (string, error) {
	if _, err := c.store.Get(ctx, ref); err == nil {
		return ref, nil
	} else if !errors.Is(err, store.ErrNotFound) {
		return "", err
	}
	filter := store.Filter{Limit: 10001}
	var prReference *regexp.Regexp
	if !sc.admin {
		filter.Project = sc.project
	}
	if !strings.Contains(ref, "#") {
		filter.MetaPack, filter.MetaKey, filter.MetaValue = "linear", "issue", strings.ToUpper(ref)
	} else {
		filter.Query = ref
		repo, number, _ := strings.Cut(ref, "#")
		pattern := regexp.QuoteMeta(repo)
		if !strings.Contains(repo, "/") {
			pattern = `[\w.-]+/` + pattern
		}
		prReference = regexp.MustCompile(`(?i)(?:^|[^\w./-])` + pattern + `(?:#|/pull/|/issues/)` + regexp.QuoteMeta(number) + `(?:$|[^0-9])`)
	}
	tickets, err := c.store.List(ctx, filter)
	if err != nil {
		return "", err
	}
	if len(tickets) > 10000 {
		return "", fmt.Errorf("%w: reference search is too broad, use a ticket id", store.ErrInvalid)
	}
	roots := map[string]bool{}
	for _, t := range tickets {
		if strings.Contains(ref, "#") {
			var m struct {
				GitHub map[string]any `json:"github"`
			}
			_ = json.Unmarshal(t.Meta, &m)
			repo, number, _ := strings.Cut(ref, "#")
			actual := fmt.Sprint(m.GitHub["repo"])
			matchesRepo := strings.EqualFold(repo, actual) || strings.EqualFold(repo, strings.TrimPrefix(actual, strings.Split(actual, "/")[0]+"/"))
			matchesNumber := number == fmt.Sprint(m.GitHub["pr"]) || number == fmt.Sprint(m.GitHub["issue"])
			if (!matchesRepo || !matchesNumber) && !prReference.MatchString(t.Title+"\n"+t.Body) {
				continue
			}
		}
		root := t.ID
		if t.Kind == store.KindStep && t.ParentID != "" {
			root = t.ParentID
		}
		if run := metaString(t.Meta, runMetaPack, "id"); run != "" {
			root = run
		}
		roots[root] = true
	}
	if len(roots) == 0 {
		return "", fmt.Errorf("%w: no ticket records reference %q", store.ErrNotFound, ref)
	}
	if len(roots) != 1 {
		return "", fmt.Errorf("%w: multiple roots reference %q, use a ticket id", store.ErrConflict, ref)
	}
	for root := range roots {
		return root, nil
	}
	return "", store.ErrNotFound
}

func (c *Controller) traceNode(ctx context.Context, id string, depth int, seen map[string]bool) (rpc.TraceNode, error) {
	seen[id] = true
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return rpc.TraceNode{}, err
	}
	n := rpc.TraceNode{
		Kind: t.Kind, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt, ClosedAt: t.ClosedAt,
		ID: t.ID, Title: t.Title, Pool: t.Pool, Status: t.Status,
		Resolution: t.Resolution, CreatedBy: t.CreatedBy, Trigger: triggerName(t.Meta),
	}
	if t.Status == store.StatusInProgress {
		n.Holder = t.Holder
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
	kids, err := c.store.List(ctx, store.Filter{ParentID: id, Limit: store.NoLimit})
	if err != nil {
		return nil, err
	}
	for _, k := range kids {
		out = append(out, k.ID)
	}
	for _, w := range windows {
		made, err := c.store.List(ctx, store.Filter{CreatedBy: w.holder, Limit: store.NoLimit})
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
