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

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
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
	} else if !errors.Is(err, domain.ErrNotFound) {
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
		return "", fmt.Errorf("%w: reference search is too broad, use a ticket id", domain.ErrInvalid)
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
		if t.Kind == domain.KindStep && t.ParentID != "" {
			root = t.ParentID
		}
		if run := metaString(t.Meta, runMetaPack, "id"); run != "" {
			root = run
		}
		roots[root] = true
	}
	if len(roots) == 0 {
		return "", fmt.Errorf("%w: no ticket records reference %q", domain.ErrNotFound, ref)
	}
	if len(roots) != 1 {
		return "", fmt.Errorf("%w: multiple roots reference %q, use a ticket id", domain.ErrConflict, ref)
	}
	for root := range roots {
		return root, nil
	}
	return "", domain.ErrNotFound
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
	if t.Status == domain.StatusInProgress {
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
			if k.Kind == domain.KindSession || k.ID == id || slices.Contains(out, k.ID) {
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

func links(body string, comments []domain.Comment) []string {
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

// creatorIDs are the tickets t's creator held when it created t: childIDs
// read the other way, with the same rule (a hold window of the creator
// that contains the creation time, session tickets skipped). The parent
// is left out, parent_id already connects it.
func (c *Controller) creatorIDs(ctx context.Context, t domain.Ticket) ([]string, error) {
	if t.Kind == domain.KindSession || t.CreatedBy == "" {
		return nil, nil
	}
	held, err := c.store.HeldAt(ctx, t.CreatedBy, t.CreatedAt)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, id := range held {
		if id == t.ID || id == t.ParentID {
			continue
		}
		th, err := c.store.Thread(ctx, id, store.Parts{History: true})
		if err != nil {
			return nil, err
		}
		for _, w := range holdWindows(th.History, c.now()) {
			if w.holder == t.CreatedBy && !t.CreatedAt.Before(w.start) && !t.CreatedAt.After(w.end) {
				out = append(out, id)
				break
			}
		}
	}
	return out, nil
}

// creatorMemo remembers creatorIDs for one walk, so the climb to the top
// ticket and the walk itself read each ticket's creator once.
type creatorMemo map[string][]string

func (m creatorMemo) of(ctx context.Context, c *Controller, t domain.Ticket) ([]string, error) {
	if ids, ok := m[t.ID]; ok {
		return ids, nil
	}
	ids, err := c.creatorIDs(ctx, t)
	if err != nil {
		return nil, err
	}
	m[t.ID] = ids
	return ids, nil
}

// maxClimb bounds the tickets read looking for the top of a graph.
const maxClimb = 2000

// topTicket is the top of the graph id belongs to (JOB-190): the ticket
// reached by climbing parent and created links from id that has nothing
// above it. With several such tickets (a ticket created while its creator
// held two tickets, each with its own top) the oldest wins, then the
// lowest id, so the answer is the same whichever member is asked about.
// A ticket outside the caller's project is not climbed through. A cycle
// has no top: the oldest ticket in it is.
func (c *Controller) topTicket(ctx context.Context, sc scope, id string, creators creatorMemo) (string, error) {
	seen := map[string]domain.Ticket{}
	var tops []domain.Ticket
	queue := []string{id}
	for i := 0; i < len(queue) && len(seen) < maxClimb; i++ {
		t, err := c.store.Get(ctx, queue[i])
		if err != nil {
			if i > 0 && errors.Is(err, domain.ErrNotFound) {
				continue
			}
			return "", err
		}
		if !sc.admin && t.Project != sc.project {
			if i == 0 {
				return "", forbidden("ticket belongs to another project")
			}
			continue
		}
		seen[t.ID] = t
		above, err := creators.of(ctx, c, t)
		if err != nil {
			return "", err
		}
		if t.ParentID != "" {
			above = append([]string{t.ParentID}, above...)
		}
		above = slices.DeleteFunc(slices.Clone(above), func(a string) bool { return a == t.ID })
		if len(above) == 0 {
			tops = append(tops, t)
		}
		for _, a := range above {
			if _, ok := seen[a]; !ok && !slices.Contains(queue, a) {
				queue = append(queue, a)
			}
		}
	}
	if len(tops) == 0 {
		for _, t := range seen {
			tops = append(tops, t)
		}
	}
	if len(tops) == 0 {
		return id, nil
	}
	oldest := slices.MinFunc(tops, func(a, b domain.Ticket) int {
		if d := a.CreatedAt.Compare(b.CreatedAt); d != 0 {
			return d
		}
		return strings.Compare(a.ID, b.ID)
	})
	return oldest.ID, nil
}
