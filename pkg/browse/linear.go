package browse

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/integrations/linear"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

var linearID = regexp.MustCompile(`^[0-9a-fA-F-]{8,64}$`)

// linearViews maps a view to the state types it keeps (in) or drops (out).
var linearViews = map[string]struct{ in, out []string }{
	"open":    {out: []string{"completed", "canceled"}},
	"active":  {in: []string{"started"}},
	"backlog": {in: []string{"backlog", "unstarted", "triage"}},
	"closed":  {in: []string{"completed", "canceled"}},
	"all":     {},
}

const linearListFields = `identifier title url updatedAt priority priorityLabel
    state { name type } assignee { name displayName } project { name }
    labels(first: 10) { nodes { name } }`

const linearListQuery = `query($filter: IssueFilter, $first: Int!, $after: String) {
  issues(filter: $filter, first: $first, after: $after, orderBy: updatedAt) {
    nodes { ` + linearListFields + ` }
    pageInfo { hasNextPage endCursor }
  }
}`

const linearSearchQuery = `query($term: String!, $filter: IssueFilter, $first: Int!, $after: String) {
  searchIssues(term: $term, filter: $filter, first: $first, after: $after) {
    nodes { ` + linearListFields + ` }
    pageInfo { hasNextPage endCursor }
  }
}`

type linearNode struct {
	Identifier    string    `json:"identifier"`
	Title         string    `json:"title"`
	URL           string    `json:"url"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Priority      int       `json:"priority"`
	PriorityLabel string    `json:"priorityLabel"`
	State         struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"state"`
	Assignee *linearUser `json:"assignee"`
	Project  *struct {
		Name string `json:"name"`
	} `json:"project"`
	Labels struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	} `json:"labels"`
}

type linearUser struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

func (u *linearUser) label() string {
	switch {
	case u == nil:
		return ""
	case u.DisplayName != "":
		return u.DisplayName
	}
	return u.Name
}

type linearPageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

func (n linearNode) item(team string) rpc.BrowseItem {
	it := rpc.BrowseItem{
		Provider: rpc.BrowseLinear, Scope: team, Ref: n.Identifier, ID: n.Identifier, Kind: "issue",
		Title: n.Title, State: n.State.Name, StateType: n.State.Type, Owner: n.Assignee.label(),
		Priority: priorityName(n.Priority, n.PriorityLabel), Labels: []string{}, UpdatedAt: n.UpdatedAt, URL: n.URL,
	}
	if n.Project != nil {
		it.Project = n.Project.Name
	}
	for _, l := range n.Labels.Nodes {
		it.Labels = append(it.Labels, l.Name)
	}
	return it
}

// priorityName is empty for no priority, so the row stays quiet.
func priorityName(v int, label string) string {
	if v == 0 {
		return ""
	}
	return strings.ToLower(label)
}

// searchTerm keeps what a person typed as plain words: control characters
// go, the length is bounded.
func searchTerm(q string) string {
	q = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, q)
	q = strings.Join(strings.Fields(q), " ")
	if len(q) > maxQuery {
		q = strings.ToValidUTF8(q[:maxQuery], "")
	}
	return q
}

func (s *Service) listLinear(ctx context.Context, p rpc.BrowseListParams) (rpc.BrowseListResult, error) {
	if err := s.linearTeam(p.Scope); err != nil {
		return rpc.BrowseListResult{}, err
	}
	view := p.View
	if view == "" {
		view = "open"
	}
	v, ok := linearViews[view]
	if !ok {
		return rpc.BrowseListResult{}, fmt.Errorf("%w: view %q is not open, active, backlog, closed or all", domain.ErrInvalid, view)
	}
	filter := map[string]any{"team": map[string]any{"key": map[string]any{"eq": p.Scope}}}
	switch {
	case len(v.in) > 0:
		filter["state"] = map[string]any{"type": map[string]any{"in": v.in}}
	case len(v.out) > 0:
		filter["state"] = map[string]any{"type": map[string]any{"nin": v.out}}
	}
	switch {
	case p.Project == "none":
		filter["project"] = map[string]any{"null": true}
	case p.Project != "":
		if !linearID.MatchString(p.Project) {
			return rpc.BrowseListResult{}, fmt.Errorf("%w: project must be an id from the catalog", domain.ErrInvalid)
		}
		filter["project"] = map[string]any{"id": map[string]any{"eq": p.Project}}
	}
	switch {
	case p.Owner == "none":
		filter["assignee"] = map[string]any{"null": true}
	case p.Owner != "":
		if !linearID.MatchString(p.Owner) {
			return rpc.BrowseListResult{}, fmt.Errorf("%w: owner must be an id from the catalog", domain.ErrInvalid)
		}
		filter["assignee"] = map[string]any{"id": map[string]any{"eq": p.Owner}}
	}
	vars := map[string]any{"filter": filter, "first": bound(p.Limit)}
	if p.Cursor != "" {
		vars["after"] = p.Cursor
	}
	term := searchTerm(p.Query)
	query, field := linearListQuery, "issues"
	// An exact key of this team is a lookup by number, which finds it at
	// any age without a search.
	if team, ok := linear.Team(strings.ToUpper(term)); ok && team == p.Scope {
		n, _ := strconv.Atoi(strings.ToUpper(term)[len(team)+1:])
		filter["number"] = map[string]any{"eq": n}
		delete(filter, "state")
		term = ""
	}
	if term != "" {
		query, field = linearSearchQuery, "searchIssues"
		vars["term"] = term
	}
	key := fmt.Sprintf("linear:list:%s|%s|%s|%s|%s|%s|%d|%s", p.Scope, view, p.Project, p.Owner, term, p.Cursor, vars["first"], fmt.Sprint(filter["number"]))
	res, st := fetch(ctx, s, key, "linear:"+p.Scope, listTTL, p.Refresh, func(ctx context.Context) (rpc.BrowseListResult, error) {
		cl, err := s.h.Linear(p.Scope)
		if err != nil {
			return rpc.BrowseListResult{}, err
		}
		var r map[string]struct {
			Nodes    []linearNode   `json:"nodes"`
			PageInfo linearPageInfo `json:"pageInfo"`
		}
		if err := cl.GraphQL(ctx, query, vars, &r); err != nil {
			return rpc.BrowseListResult{}, err
		}
		page := r[field]
		out := rpc.BrowseListResult{Items: make([]rpc.BrowseItem, 0, len(page.Nodes)), HasMore: page.PageInfo.HasNextPage}
		for _, n := range page.Nodes {
			out.Items = append(out.Items, n.item(p.Scope))
		}
		if out.HasMore {
			out.Next = page.PageInfo.EndCursor
		}
		return out, nil
	})
	if res.Items == nil {
		res.Items = []rpc.BrowseItem{}
	}
	res.Status = st
	return res, nil
}

const linearRef = `identifier title url state { name type }`

const linearCommentFields = `id body url createdAt editedAt
      user { name displayName } botActor { name }
      parent { id body url createdAt user { name displayName } botActor { name } }`

const linearHead = `identifier title description url createdAt updatedAt priority priorityLabel
  state { name type } team { key }
  assignee { name displayName } project { name }
  labels(first: 30) { nodes { name } }
  parent { ` + linearRef + ` }
  children(first: 25) { nodes { ` + linearRef + ` } }
  relations(first: 25) { nodes { type relatedIssue { ` + linearRef + ` } } }
  inverseRelations(first: 25) { nodes { type issue { ` + linearRef + ` } } }
  attachments(first: 20) { nodes { url title sourceType metadata } }`

// Comments newest first: a reply names its parent, so a thread is rebuilt
// from what is read, and an older page extends it.
const linearDetailQuery = `query($id: String!, $first: Int!, $after: String) { issue(id: $id) {
  ` + linearHead + `
  comments(first: $first, after: $after, orderBy: createdAt) { nodes { ` + linearCommentFields + ` } pageInfo { hasNextPage endCursor } }
} }`

const linearBriefQuery = `query($id: String!) { issue(id: $id) { ` + linearHead + ` } }`

const linearMoreQuery = `query($id: String!, $first: Int!, $after: String) { issue(id: $id) {
  comments(first: $first, after: $after, orderBy: createdAt) { nodes { ` + linearCommentFields + ` } pageInfo { hasNextPage endCursor } }
} }`

type linearRefNode struct {
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	State      struct {
		Name string `json:"name"`
	} `json:"state"`
}

type linearComment struct {
	ID        string                 `json:"id"`
	Body      string                 `json:"body"`
	URL       string                 `json:"url"`
	CreatedAt time.Time              `json:"createdAt"`
	EditedAt  *time.Time             `json:"editedAt"`
	User      *linearUser            `json:"user"`
	BotActor  *struct{ Name string } `json:"botActor"`
	Parent    *linearComment         `json:"parent"`
}

func (c *linearComment) author() (string, bool) {
	switch {
	case c.User != nil:
		return c.User.label(), false
	case c.BotActor != nil:
		return c.BotActor.Name, true
	}
	return "", true
}

type linearComments struct {
	Nodes    []linearComment `json:"nodes"`
	PageInfo linearPageInfo  `json:"pageInfo"`
}

type linearIssue struct {
	linearNode
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	Team        struct {
		Key string `json:"key"`
	} `json:"team"`
	Parent   *linearRefNode `json:"parent"`
	Children struct {
		Nodes []linearRefNode `json:"nodes"`
	} `json:"children"`
	Relations struct {
		Nodes []struct {
			Type         string         `json:"type"`
			RelatedIssue *linearRefNode `json:"relatedIssue"`
		} `json:"nodes"`
	} `json:"relations"`
	InverseRelations struct {
		Nodes []struct {
			Type  string         `json:"type"`
			Issue *linearRefNode `json:"issue"`
		} `json:"nodes"`
	} `json:"inverseRelations"`
	Attachments struct {
		Nodes []struct {
			URL        string         `json:"url"`
			Title      string         `json:"title"`
			SourceType string         `json:"sourceType"`
			Metadata   map[string]any `json:"metadata"`
		} `json:"nodes"`
	} `json:"attachments"`
	Comments linearComments `json:"comments"`
}

var pullURL = regexp.MustCompile(`^https://github\.com/([A-Za-z0-9][A-Za-z0-9-]*)/([A-Za-z0-9_.-]+)/pull/([1-9][0-9]*)$`)

func (i linearIssue) linked() []rpc.BrowseLinked {
	out := []rpc.BrowseLinked{}
	ref := func(r *linearRefNode, rel string) {
		if r != nil {
			out = append(out, rpc.BrowseLinked{Kind: "issue", Ref: r.Identifier, Title: r.Title, State: r.State.Name, URL: r.URL, Relation: rel})
		}
	}
	ref(i.Parent, "parent")
	for k := range i.Children.Nodes {
		ref(&i.Children.Nodes[k], "child")
	}
	for _, r := range i.Relations.Nodes {
		ref(r.RelatedIssue, strings.ReplaceAll(r.Type, "_", " "))
	}
	for _, r := range i.InverseRelations.Nodes {
		rel := strings.ReplaceAll(r.Type, "_", " ")
		if r.Type == "blocks" {
			rel = "blocked by"
		}
		ref(r.Issue, rel)
	}
	for _, a := range i.Attachments.Nodes {
		m := pullURL.FindStringSubmatch(a.URL)
		if a.SourceType != "github" || m == nil {
			continue
		}
		state, _ := a.Metadata["status"].(string)
		if state == "" {
			state = "open"
		}
		out = append(out, rpc.BrowseLinked{Kind: "pr", Ref: m[1] + "/" + m[2] + "#" + m[3], Title: a.Title, State: state, URL: a.URL, Relation: "attached"})
	}
	return out
}

// linearThreads groups newest-first comments into threads, newest activity
// first. A reply whose parent is outside this page brings the parent along,
// so the thread still opens with what was answered. Nesting is the provider's
// parent link only: a reply to a reply sits under its parent.
func linearThreads(nodes []linearComment) []rpc.BrowseThread {
	byID := map[string]*linearComment{}
	for i := range nodes {
		byID[nodes[i].ID] = &nodes[i]
	}
	rootOf := func(c *linearComment) string {
		for range 20 {
			if c.Parent == nil {
				break
			}
			up, ok := byID[c.Parent.ID]
			if !ok {
				return c.Parent.ID
			}
			c = up
		}
		return c.ID
	}
	threads := map[string]*rpc.BrowseThread{}
	var order []string
	add := func(root string, c *linearComment, replyTo *linearComment) {
		th := threads[root]
		if th == nil {
			th = &rpc.BrowseThread{ID: root, Kind: "comment", Comments: []rpc.BrowseComment{}}
			threads[root] = th
			order = append(order, root)
		}
		if slices.ContainsFunc(th.Comments, func(x rpc.BrowseComment) bool { return x.ID == c.ID }) {
			return
		}
		author, bot := c.author()
		body, cut := clip(c.Body, maxComment)
		bc := rpc.BrowseComment{ID: c.ID, Author: author, Bot: bot, At: c.CreatedAt, URL: c.URL, Body: body, Truncated: cut}
		if c.EditedAt != nil {
			bc.EditedAt = *c.EditedAt
		}
		if replyTo != nil {
			bc.ReplyTo = replyTo.ID
			bc.ReplyToAuthor, _ = replyTo.author()
		}
		th.Comments = append(th.Comments, bc)
		if th.URL == "" || c.ID == root {
			th.URL = c.URL
		}
	}
	for i := len(nodes) - 1; i >= 0; i-- { // oldest first
		c := &nodes[i]
		root := rootOf(c)
		if c.Parent != nil && byID[c.Parent.ID] == nil {
			add(root, c.Parent, nil)
		}
		var to *linearComment
		if c.Parent != nil {
			to = byID[c.Parent.ID]
			if to == nil {
				to = c.Parent
			}
		}
		add(root, c, to)
	}
	out := make([]rpc.BrowseThread, 0, len(order))
	for _, id := range order {
		th := threads[id]
		slices.SortStableFunc(th.Comments, func(a, b rpc.BrowseComment) int { return a.At.Compare(b.At) })
		out = append(out, *th)
	}
	slices.SortStableFunc(out, func(a, b rpc.BrowseThread) int {
		return out2last(b).Compare(out2last(a))
	})
	return out
}

func out2last(t rpc.BrowseThread) time.Time {
	if len(t.Comments) == 0 {
		return time.Time{}
	}
	return t.Comments[len(t.Comments)-1].At
}

func (s *Service) showLinear(ctx context.Context, p rpc.BrowseShowParams) (rpc.BrowseDetail, error) {
	key := strings.ToUpper(strings.TrimSpace(p.Ref))
	team, ok := linear.Team(key)
	if !ok {
		return rpc.BrowseDetail{}, fmt.Errorf("%w: %q is not a Linear issue key", domain.ErrInvalid, p.Ref)
	}
	if err := s.linearTeam(team); err != nil {
		return rpc.BrowseDetail{}, err
	}
	ck := fmt.Sprintf("linear:show:%s|%s|%t", key, p.Cursor, p.Brief)
	d, st := fetch(ctx, s, ck, "linear:"+team, detailTTL, p.Refresh, func(ctx context.Context) (rpc.BrowseDetail, error) {
		cl, err := s.h.Linear(team)
		if err != nil {
			return rpc.BrowseDetail{}, err
		}
		vars := map[string]any{"id": key}
		query := linearDetailQuery
		switch {
		case p.Brief:
			query = linearBriefQuery
		default:
			vars["first"] = threadPage
			if p.Cursor != "" {
				vars["after"] = p.Cursor
				query = linearMoreQuery
			}
		}
		var r struct {
			Issue *linearIssue `json:"issue"`
		}
		if err := cl.GraphQL(ctx, query, vars, &r); err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Entity not found") {
				return rpc.BrowseDetail{}, errNotFound
			}
			return rpc.BrowseDetail{}, err
		}
		if r.Issue == nil {
			return rpc.BrowseDetail{}, errNotFound
		}
		return linearDetail(team, key, r.Issue, p), nil
	})
	if st.State == rpc.BrowseError && d.Item.Ref == "" {
		d.Item = rpc.BrowseItem{Provider: rpc.BrowseLinear, Scope: team, Ref: key, ID: key, Kind: "issue", Labels: []string{}}
	}
	d = fillDetail(d)
	d.Status = st
	return d, nil
}

func linearDetail(team, key string, i *linearIssue, p rpc.BrowseShowParams) rpc.BrowseDetail {
	d := rpc.BrowseDetail{Item: i.item(team), Threads: []rpc.BrowseThread{}}
	if p.Cursor != "" {
		// An older page carries only discussion.
		d.Item = rpc.BrowseItem{Provider: rpc.BrowseLinear, Scope: team, Ref: key, ID: key, Kind: "issue", Labels: []string{}}
		d.Threads = linearThreads(i.Comments.Nodes)
		if i.Comments.PageInfo.HasNextPage {
			d.Older = i.Comments.PageInfo.EndCursor
		}
		return d
	}
	d.Body, d.BodyTruncated = clip(i.Description, maxBody)
	d.Props = rpc.BrowseProps{
		Assignee: d.Item.Owner, Priority: priorityName(i.Priority, i.PriorityLabel), PriorityValue: i.Priority,
		Project: d.Item.Project, Labels: d.Item.Labels, CreatedAt: i.CreatedAt,
	}
	d.Linked = i.linked()
	if !p.Brief {
		d.Threads = linearThreads(i.Comments.Nodes)
		if i.Comments.PageInfo.HasNextPage {
			d.Older = i.Comments.PageInfo.EndCursor
		}
	}
	return d
}

// fillDetail makes the slices the JSON shape promises non-nil.
func fillDetail(d rpc.BrowseDetail) rpc.BrowseDetail {
	if d.Item.Labels == nil {
		d.Item.Labels = []string{}
	}
	if d.Props.Labels == nil {
		d.Props.Labels = []string{}
	}
	if d.Linked == nil {
		d.Linked = []rpc.BrowseLinked{}
	}
	if d.Threads == nil {
		d.Threads = []rpc.BrowseThread{}
	}
	if d.Files == nil {
		d.Files = []rpc.BrowseFile{}
	}
	if d.Checks == nil {
		d.Checks = []rpc.BrowseCheck{}
	}
	return d
}
