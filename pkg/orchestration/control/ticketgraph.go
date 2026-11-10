package control

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

func (c *Controller) ticketGraph(ctx context.Context, sc scope, p rpc.GraphParams) (any, error) {
	g, _, err := c.graphWalk(ctx, sc, p)
	if err != nil {
		return nil, err
	}
	return g, nil
}

// graphWalk builds the graph and returns the thread of every node it read,
// so the timeline (JOB-1103) reads history once.
func (c *Controller) graphWalk(ctx context.Context, sc scope, p rpc.GraphParams) (rpc.GraphResult, map[string]store.Thread, error) {
	if p.Limit < 0 || p.Limit > 2000 {
		return rpc.GraphResult{}, nil, fmt.Errorf("%w: graph limit must be between 1 and 2000", domain.ErrInvalid)
	}
	if p.Limit == 0 {
		p.Limit = 500
	}
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return rpc.GraphResult{}, nil, err
	}
	id, err = c.traceReference(ctx, sc, id)
	if err != nil {
		return rpc.GraphResult{}, nil, err
	}
	threads := map[string]store.Thread{}
	out := rpc.GraphResult{Root: id, Nodes: []rpc.GraphNode{}, Relations: []rpc.GraphRelation{}, External: []rpc.GraphExternal{}}
	cfg := c.config()
	queue := []string{id}
	seen := map[string]bool{id: true}
	included := map[string]bool{}
	relations := map[rpc.GraphRelation]bool{}
	enqueue := func(ids ...string) {
		for _, next := range ids {
			if !seen[next] {
				seen[next] = true
				if len(queue) < p.Limit {
					queue = append(queue, next)
				} else {
					out.Truncated = true
				}
			}
		}
	}
	connect := func(from, to, kind string) {
		if from == "" || to == "" || from == to {
			return
		}
		r := rpc.GraphRelation{From: from, To: to, Kind: kind}
		if !relations[r] {
			relations[r] = true
			out.Relations = append(out.Relations, r)
		}
		enqueue(from, to)
	}
	// Wait relations (#1091): one per waiting ticket and target, the
	// newest wait winning. An external target is a node of its own.
	waitAt := map[[2]string]int{}
	externals := map[string]int{}
	connectWait := func(w graphWait) {
		r := w.rel
		if w.ext == nil {
			if r.To == r.From {
				return
			}
			if _, err := c.store.Get(ctx, r.To); err != nil {
				return // a wait on a ticket that does not exist has nothing to show
			}
			enqueue(r.To)
		} else if i, ok := externals[w.ext.ID]; ok {
			if r.Since.Before(out.External[i].At) {
				out.External[i].At = r.Since
			}
		} else {
			externals[w.ext.ID] = len(out.External)
			e := *w.ext
			e.At = r.Since
			out.External = append(out.External, e)
		}
		key := [2]string{r.From, r.To}
		if i, ok := waitAt[key]; ok {
			out.Relations[i] = r
			return
		}
		waitAt[key] = len(out.Relations)
		out.Relations = append(out.Relations, r)
	}
	for i := 0; i < len(queue); i++ {
		t, err := c.store.Get(ctx, queue[i])
		if err != nil {
			return rpc.GraphResult{}, nil, err
		}
		if !sc.admin && t.Project != sc.project {
			if i == 0 {
				return rpc.GraphResult{}, nil, forbidden("ticket belongs to another project")
			}
			continue
		}
		th, err := c.store.Thread(ctx, t.ID, store.AllParts)
		if err != nil {
			return rpc.GraphResult{}, nil, err
		}
		included[t.ID] = true
		threads[t.ID] = th
		out.Nodes = append(out.Nodes, rpc.GraphNode{Ticket: t, Links: graphLinks(cfg, t, th.Comments)})
		connect(t.ParentID, t.ID, "parent")
		for _, edge := range th.Edges {
			connect(edge.From, edge.To, string(edge.Kind))
		}
		for _, w := range graphWaits(cfg, t, th) {
			connectWait(w)
		}
		kids, err := c.childIDs(ctx, t.ID, holdWindows(th.History, c.now()))
		if err != nil {
			return rpc.GraphResult{}, nil, err
		}
		for _, kid := range kids {
			child, err := c.store.Get(ctx, kid)
			if err != nil {
				return rpc.GraphResult{}, nil, err
			}
			kind := "created"
			if child.ParentID == t.ID {
				kind = "parent"
			}
			connect(t.ID, kid, kind)
		}
	}
	// A bounded graph never contains dangling or out-of-scope endpoints.
	out.Relations = slices.DeleteFunc(out.Relations, func(r rpc.GraphRelation) bool {
		_, external := externals[r.To]
		return !included[r.From] || !included[r.To] && !external
	})
	return out, threads, nil
}

var trackerLink = regexp.MustCompile(`https://(?:github\.com/[\w.-]+/[\w.-]+/(?:issues|pull)/\d+|linear\.app/[\w./%+-]+)`)

func graphLinks(cfg *config.Config, t domain.Ticket, comments []domain.Comment) []string {
	out := []string{}
	add := func(raw string) {
		u, err := url.Parse(raw)
		if err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != "" && !slices.Contains(out, raw) {
			out = append(out, raw)
		}
	}
	var meta struct {
		GitHub struct {
			Repo  string      `json:"repo"`
			Issue json.Number `json:"issue"`
			PR    json.Number `json:"pr"`
		} `json:"github"`
		Linear struct {
			Issue      string `json:"issue"`
			Identifier string `json:"identifier"`
			Workspace  string `json:"workspace"`
			URL        string `json:"url"`
		} `json:"linear"`
	}
	if len(t.Meta) == 0 || json.Unmarshal(t.Meta, &meta) == nil {
		if f, err := domain.DecodeFields[domain.RunFields](t); err == nil && t.Kind == domain.KindRun {
			if meta.GitHub.Repo == "" {
				meta.GitHub.Repo = f.Vars["repo"]
			}
			if meta.GitHub.PR == "" {
				meta.GitHub.PR = json.Number(f.Vars["pr"])
			}
			if meta.Linear.Issue == "" {
				meta.Linear.Issue = f.Vars["issue"]
			}
		}
		gh, ln := meta.GitHub, meta.Linear
		if gh.Repo != "" {
			if _, err := gh.Issue.Int64(); err == nil {
				add("https://github.com/" + gh.Repo + "/issues/" + gh.Issue.String())
			}
			if _, err := gh.PR.Int64(); err == nil {
				add("https://github.com/" + gh.Repo + "/pull/" + gh.PR.String())
			}
		}
		key := ln.Issue
		if key == "" {
			key = ln.Identifier
		}
		if ln.URL != "" {
			add(ln.URL)
		} else if key != "" {
			workspace := ln.Workspace
			if workspace == "" {
				// Old tickets may omit workspace. Only use a unique configured match.
				team, _, _ := strings.Cut(key, "-")
				matches := map[string]bool{}
				for _, name := range cfg.Projects[t.Project].Identities {
					identity := cfg.Identities[name]
					if config.IsLinear(identity.Kind) && identity.CoversTeam(team) {
						matches[identity.Workspace] = true
					}
				}
				if len(matches) == 1 {
					for match := range matches {
						workspace = match
					}
				}
			}
			if workspace != "" {
				add("https://linear.app/" + url.PathEscape(workspace) + "/issue/" + url.PathEscape(key))
			}
		}
	}
	for _, body := range append([]string{t.Body}, commentBodies(comments)...) {
		for _, link := range trackerLink.FindAllString(body, -1) {
			add(link)
		}
	}
	return out
}

func commentBodies(comments []domain.Comment) []string {
	out := make([]string, 0, len(comments))
	for _, comment := range comments {
		out = append(out, comment.Body)
	}
	return out
}

// stepPRComment is the handoff line a step records for its PR: "PR <url>
// head <sha>" at the start of a line (the pr-work prompts, JOB-131). A PR
// merely mentioned in a comment does not match.
var stepPRComment = regexp.MustCompile(`(?m)^PR (https://github\.com/[\w.-]+/[\w.-]+/pull/\d+)(?:\s|$)`)

// stepPRLinks are the PRs a step records itself: those its wait conditions
// watch (structured, set right after the PR is opened) and those it
// announced with a "PR <url>" comment. It never reads an issue number or
// a casual mention.
func stepPRLinks(t domain.Ticket, comments []domain.Comment) []string {
	var out []string
	add := func(link string) {
		if !slices.Contains(out, link) {
			out = append(out, link)
		}
	}
	if w := waitOf(t); w != nil {
		for _, cd := range w.Conds {
			if cd.Repo != "" && cd.PR > 0 {
				add(fmt.Sprintf("https://github.com/%s/pull/%d", cd.Repo, cd.PR))
			}
		}
	}
	for _, body := range commentBodies(comments) {
		for _, m := range stepPRComment.FindAllStringSubmatch(body, -1) {
			add(m[1])
		}
	}
	return out
}

// runLinks are a run's links: what its vars and meta name, plus the PRs
// its steps recorded after it started (JOB-131), in step order.
func (c *Controller) runLinks(ctx context.Context, run domain.Ticket, steps []domain.Ticket) ([]string, error) {
	links := graphLinks(c.config(), run, nil)
	for _, s := range steps {
		th, err := c.store.Thread(ctx, s.ID, store.Parts{Comments: true})
		if err != nil {
			return nil, err
		}
		for _, link := range stepPRLinks(s, th.Comments) {
			if !slices.Contains(links, link) {
				links = append(links, link)
			}
		}
	}
	return links, nil
}
