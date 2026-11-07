package control

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

func (c *Controller) ticketGraph(ctx context.Context, sc scope, p rpc.GraphParams) (any, error) {
	if p.Limit < 0 || p.Limit > 2000 {
		return nil, fmt.Errorf("%w: graph limit must be between 1 and 2000", store.ErrInvalid)
	}
	if p.Limit == 0 {
		p.Limit = 500
	}
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	id, err = c.traceReference(ctx, sc, id)
	if err != nil {
		return nil, err
	}
	out := rpc.GraphResult{Root: id, Nodes: []rpc.GraphNode{}, Relations: []rpc.GraphRelation{}}
	cfg := c.config()
	queue := []string{id}
	seen := map[string]bool{id: true}
	included := map[string]bool{}
	relations := map[rpc.GraphRelation]bool{}
	connect := func(from, to, kind string) {
		if from == "" || to == "" || from == to {
			return
		}
		r := rpc.GraphRelation{From: from, To: to, Kind: kind}
		if !relations[r] {
			relations[r] = true
			out.Relations = append(out.Relations, r)
		}
		for _, next := range []string{from, to} {
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
	for i := 0; i < len(queue); i++ {
		t, err := c.store.Get(ctx, queue[i])
		if err != nil {
			return nil, err
		}
		if !sc.admin && t.Project != sc.project {
			if i == 0 {
				return nil, forbidden("ticket belongs to another project")
			}
			continue
		}
		th, err := c.store.Thread(ctx, t.ID, store.AllParts)
		if err != nil {
			return nil, err
		}
		included[t.ID] = true
		out.Nodes = append(out.Nodes, rpc.GraphNode{Ticket: t, Links: graphLinks(cfg, t, th.Comments)})
		connect(t.ParentID, t.ID, "parent")
		for _, edge := range th.Edges {
			connect(edge.From, edge.To, string(edge.Kind))
		}
		kids, err := c.childIDs(ctx, t.ID, holdWindows(th.History, c.now()))
		if err != nil {
			return nil, err
		}
		for _, kid := range kids {
			child, err := c.store.Get(ctx, kid)
			if err != nil {
				return nil, err
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
		return !included[r.From] || !included[r.To]
	})
	return out, nil
}

var trackerLink = regexp.MustCompile(`https://(?:github\.com/[\w.-]+/[\w.-]+/(?:issues|pull)/\d+|linear\.app/[\w./%+-]+)`)

func graphLinks(cfg *config.Config, t store.Ticket, comments []store.Comment) []string {
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
	if json.Unmarshal(t.Meta, &meta) == nil {
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

func commentBodies(comments []store.Comment) []string {
	out := make([]string, 0, len(comments))
	for _, comment := range comments {
		out = append(out, comment.Body)
	}
	return out
}
