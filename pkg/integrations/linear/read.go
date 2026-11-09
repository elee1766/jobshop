package linear

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

// Read commands (#279): issue context, list and search with paging, team
// catalogs. All return plain structs printed as JSON by the CLI.

// Ref is a short reference to another issue.
type Ref struct {
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
	State      State  `json:"state"`
}

// Open reports whether the referenced issue is not done or canceled.
func (r Ref) Open() bool { return OpenStateType(r.State.Type) }

// OpenStateType is false for Linear's completed and canceled state types.
func OpenStateType(t string) bool { return t != "completed" && t != "canceled" }

// Relation is a link from the issue to another: blocks, blocked_by,
// related, duplicate or similar.
type Relation struct {
	Type  string `json:"type"`
	Issue Ref    `json:"issue"`
}

// Context is an issue with what an agent needs to decide on it.
type Context struct {
	Issue
	Labels   []string `json:"labels"`
	Assignee string   `json:"assignee,omitempty"`
	// Delegate is the agent (app) working the issue for the assignee.
	Delegate   string     `json:"delegate,omitempty"`
	Parent     *Ref       `json:"parent,omitempty"`
	Children   []Ref      `json:"children"`
	Relations  []Relation `json:"relations"`
	Blockers   []Ref      `json:"blockers"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	// OpenDeps are the open GitHub issues the description says it
	// depends on (JOB-68), as owner/repo#n. The caller fills it after
	// reading GitHub: the Linear read does not.
	OpenDeps []string `json:"open_deps,omitempty"`
}

// GitHubDep is one "Depends on owner/repo#n" line of a description.
type GitHubDep struct {
	Repo   string // owner/name
	Number int
}

func (d GitHubDep) String() string { return fmt.Sprintf("%s#%d", d.Repo, d.Number) }

var dependsOn = regexp.MustCompile(`(?im)^\W*depends on:?\s+(?:https://github\.com/)?([A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)?)(?:#|/issues/)([1-9][0-9]*)`)

// Dependencies parses "Depends on <owner/repo>#<n>" lines (also
// "repo#n", which takes owner, and GitHub issue URLs) from a description.
// A bare repo with no owner is skipped.
func Dependencies(desc, owner string) []GitHubDep {
	var out []GitHubDep
	for _, m := range dependsOn.FindAllStringSubmatch(desc, -1) {
		repo := m[1]
		if !strings.Contains(repo, "/") {
			if owner == "" {
				continue
			}
			repo = owner + "/" + repo
		}
		var n int
		_, _ = fmt.Sscan(m[2], &n)
		d := GitHubDep{Repo: repo, Number: n}
		if !slices.Contains(out, d) {
			out = append(out, d)
		}
	}
	return out
}

// CheckDeps sets c.OpenDeps to the description's dependencies that open
// reports open. owner is the GitHub owner for a bare "repo#n".
func (c *Context) CheckDeps(ctx context.Context, owner string, open func(context.Context, GitHubDep) (bool, error)) error {
	c.OpenDeps = nil
	for _, d := range Dependencies(c.Description, owner) {
		ok, err := open(ctx, d)
		if err != nil {
			return fmt.Errorf("dependency %s: %w", d, err)
		}
		if ok {
			c.OpenDeps = append(c.OpenDeps, d.String())
		}
	}
	return nil
}

// OpenBlockers are the blockers not done or canceled.
func (c Context) OpenBlockers() []Ref {
	var out []Ref
	for _, b := range c.Blockers {
		if b.Open() {
			out = append(out, b)
		}
	}
	return out
}

const refFields = `identifier title state { name type }`

const contextQuery = `query($id: String!, $comments: Int!) { issue(id: $id) {
  id identifier title description url archivedAt
  state { id name type }
  team { id key }
  assignee { name }
  delegate { name }
  labels(first: 50) { nodes { name } }
  parent { ` + refFields + ` }
  children(first: 100) { nodes { ` + refFields + ` } }
  relations(first: 100) { nodes { type relatedIssue { ` + refFields + ` } } }
  inverseRelations(first: 100) { nodes { type issue { ` + refFields + ` } } }
  comments(first: $comments) { nodes { id body url createdAt ` + commentAuthorFields + ` } }
} }`

// RecentComments is how many comments Context reads without all.
const RecentComments = 10

// GetContext reads an issue with labels, assignee, parent, children,
// relations and blockers (issues that block it), and its newest
// comments (RecentComments, or up to 250 with all), oldest first.
func (c *Client) GetContext(ctx context.Context, key string, all bool) (Context, error) {
	n := RecentComments
	if all {
		n = 250
	}
	type rel struct {
		Type         string `json:"type"`
		RelatedIssue *Ref   `json:"relatedIssue"`
		Issue        *Ref   `json:"issue"`
	}
	var r struct {
		Issue *struct {
			Issue
			ArchivedAt *time.Time `json:"archivedAt"`
			Assignee   *named     `json:"assignee"`
			Delegate   *named     `json:"delegate"`
			Labels     struct {
				Nodes []named `json:"nodes"`
			} `json:"labels"`
			Parent   *Ref `json:"parent"`
			Children struct {
				Nodes []Ref `json:"nodes"`
			} `json:"children"`
			Relations struct {
				Nodes []rel `json:"nodes"`
			} `json:"relations"`
			InverseRelations struct {
				Nodes []rel `json:"nodes"`
			} `json:"inverseRelations"`
			Comments commentNodes `json:"comments"`
		} `json:"issue"`
	}
	if err := c.GraphQL(ctx, contextQuery, map[string]any{"id": key, "comments": n}, &r); err != nil {
		return Context{}, err
	}
	if r.Issue == nil {
		return Context{}, fmt.Errorf("linear: issue %s not found", key)
	}
	is := r.Issue
	out := Context{Issue: is.Issue, ArchivedAt: is.ArchivedAt, Parent: is.Parent, Labels: []string{}, Children: []Ref{}, Relations: []Relation{}, Blockers: []Ref{}}
	out.Comments = is.Comments.list()
	if is.Delegate != nil {
		out.Delegate = is.Delegate.Name
	}
	if is.Assignee != nil {
		out.Assignee = is.Assignee.Name
	}
	for _, l := range is.Labels.Nodes {
		out.Labels = append(out.Labels, l.Name)
	}
	out.Children = append(out.Children, is.Children.Nodes...)
	for _, x := range is.Relations.Nodes {
		if x.RelatedIssue != nil {
			out.Relations = append(out.Relations, Relation{Type: x.Type, Issue: *x.RelatedIssue})
		}
	}
	// An inverse "blocks" relation is an issue that blocks this one.
	for _, x := range is.InverseRelations.Nodes {
		if x.Issue == nil {
			continue
		}
		t := x.Type
		if t == "blocks" {
			t = "blocked_by"
			out.Blockers = append(out.Blockers, *x.Issue)
		}
		out.Relations = append(out.Relations, Relation{Type: t, Issue: *x.Issue})
	}
	return out, nil
}

type commentNodes struct {
	Nodes []struct {
		ID        string       `json:"id"`
		Body      string       `json:"body"`
		URL       string       `json:"url"`
		CreatedAt time.Time    `json:"createdAt"`
		User      *commentUser `json:"user"`
		BotActor  *commentBot  `json:"botActor"`
	} `json:"nodes"`
}

// list converts newest-first nodes to comments, oldest first.
func (cn commentNodes) list() []Comment {
	out := []Comment{}
	for _, n := range cn.Nodes {
		cm := Comment{ID: n.ID, Body: n.Body, URL: n.URL, CreatedAt: n.CreatedAt}
		cm.setAuthor(n.User, n.BotActor)
		out = append(out, cm)
	}
	slices.Reverse(out)
	return out
}

// Filter narrows List and Search. Empty fields match everything.
// Assignee "me" is the API key's own user.
type Filter struct {
	Team     string
	State    string
	Label    string
	Assignee string
	Limit    int    // page size, default 50, max 250
	After    string // cursor from a previous Page.EndCursor
}

// Summary is one issue in a list.
type Summary struct {
	Identifier string    `json:"identifier"`
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	State      State     `json:"state"`
	Assignee   string    `json:"assignee,omitempty"`
	Labels     []string  `json:"labels"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Page is one page of issues. EndCursor goes in Filter.After for the
// next page while HasNext.
type Page struct {
	Issues    []Summary `json:"issues"`
	HasNext   bool      `json:"has_next"`
	EndCursor string    `json:"end_cursor,omitempty"`
}

func (f Filter) gql() map[string]any {
	m := map[string]any{}
	if f.Team != "" {
		m["team"] = map[string]any{"key": map[string]any{"eq": f.Team}}
	}
	if f.State != "" {
		m["state"] = map[string]any{"name": map[string]any{"eqIgnoreCase": f.State}}
	}
	if f.Label != "" {
		m["labels"] = map[string]any{"name": map[string]any{"eqIgnoreCase": f.Label}}
	}
	switch {
	case strings.EqualFold(f.Assignee, "me"):
		m["assignee"] = map[string]any{"isMe": map[string]any{"eq": true}}
	case f.Assignee != "":
		m["assignee"] = map[string]any{"name": map[string]any{"eqIgnoreCase": f.Assignee}}
	}
	return m
}

func (f Filter) vars() map[string]any {
	n := f.Limit
	if n <= 0 {
		n = 50
	}
	n = min(n, 250)
	v := map[string]any{"filter": f.gql(), "first": n}
	if f.After != "" {
		v["after"] = f.After
	}
	return v
}

const summaryFields = `nodes { identifier title url updatedAt state { id name type } assignee { name } labels(first: 20) { nodes { name } } }
  pageInfo { hasNextPage endCursor }`

const listQuery = `query($filter: IssueFilter, $first: Int!, $after: String) {
  issues(filter: $filter, first: $first, after: $after, orderBy: updatedAt) { ` + summaryFields + ` }
}`

const searchQuery = `query($term: String!, $filter: IssueFilter, $first: Int!, $after: String) {
  searchIssues(term: $term, filter: $filter, first: $first, after: $after) { ` + summaryFields + ` }
}`

type pageNodes struct {
	Nodes []struct {
		Identifier string    `json:"identifier"`
		Title      string    `json:"title"`
		URL        string    `json:"url"`
		UpdatedAt  time.Time `json:"updatedAt"`
		State      State     `json:"state"`
		Assignee   *named    `json:"assignee"`
		Labels     struct {
			Nodes []named `json:"nodes"`
		} `json:"labels"`
	} `json:"nodes"`
	PageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	} `json:"pageInfo"`
}

func (p pageNodes) page() Page {
	out := Page{Issues: []Summary{}, HasNext: p.PageInfo.HasNextPage}
	if out.HasNext {
		out.EndCursor = p.PageInfo.EndCursor
	}
	for _, n := range p.Nodes {
		s := Summary{Identifier: n.Identifier, Title: n.Title, URL: n.URL, State: n.State, UpdatedAt: n.UpdatedAt, Labels: []string{}}
		if n.Assignee != nil {
			s.Assignee = n.Assignee.Name
		}
		for _, l := range n.Labels.Nodes {
			s.Labels = append(s.Labels, l.Name)
		}
		out.Issues = append(out.Issues, s)
	}
	return out
}

// List returns one page of issues matching f, most recently updated first.
func (c *Client) List(ctx context.Context, f Filter) (Page, error) {
	var r struct {
		Issues pageNodes `json:"issues"`
	}
	if err := c.GraphQL(ctx, listQuery, f.vars(), &r); err != nil {
		return Page{}, err
	}
	return r.Issues.page(), nil
}

// Search returns one page of issues matching the full-text term and f.
func (c *Client) Search(ctx context.Context, term string, f Filter) (Page, error) {
	if strings.TrimSpace(term) == "" {
		return Page{}, fmt.Errorf("linear: empty search")
	}
	v := f.vars()
	v["term"] = term
	var r struct {
		SearchIssues pageNodes `json:"searchIssues"`
	}
	if err := c.GraphQL(ctx, searchQuery, v, &r); err != nil {
		return Page{}, err
	}
	return r.SearchIssues.page(), nil
}

// CatalogKinds are the lists Catalog reads.
var CatalogKinds = []string{"projects", "labels", "states", "users"}

// Entry is one catalog item. Type is the state type, project state or
// user email.
type Entry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

// catalogFields maps a kind to the team connection and node fields it
// reads. One kind per query: all four at 250 each was rejected as "Query
// too complex" on a real workspace (#653). The team is read with
// team(id: key), one object: under teams(filter:) Linear counts the
// nested page once per possible team, and 100 states scored 12550 of a
// 10000 budget (#683).
var catalogFields = map[string][2]string{
	"projects": {"projects", "id name state"},
	"labels":   {"labels", "id name"},
	"states":   {"states", "id name type"},
	"users":    {"members", "id name email"},
}

// catalogPage is the page size; Catalog follows the cursor.
const catalogPage = 100

// catalogMaxPages bounds the paging so a bad cursor cannot loop.
const catalogMaxPages = 50

func catalogQuery(kind string) string {
	f := catalogFields[kind]
	return `query($team: String!, $first: Int!, $after: String) { team(id: $team) {
  ` + f[0] + `(first: $first, after: $after) { nodes { ` + f[1] + ` } pageInfo { hasNextPage endCursor } }
} }`
}

// Catalog lists a team's projects, labels, states or users.
func (c *Client) Catalog(ctx context.Context, team, kind string) ([]Entry, error) {
	if !slices.Contains(CatalogKinds, kind) {
		return nil, fmt.Errorf("linear: catalog %q is not one of %s", kind, strings.Join(CatalogKinds, ", "))
	}
	type node struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		State string `json:"state"`
		Type  string `json:"type"`
		Email string `json:"email"`
	}
	type conn struct {
		Nodes    []node `json:"nodes"`
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
	}
	field := catalogFields[kind][0]
	q := catalogQuery(kind)
	out := []Entry{}
	vars := map[string]any{"team": team, "first": catalogPage}
	for range catalogMaxPages {
		var r struct {
			Team map[string]conn `json:"team"`
		}
		if err := c.GraphQL(ctx, q, vars, &r); err != nil {
			return nil, err
		}
		if r.Team == nil {
			return nil, fmt.Errorf("linear: team %s not found", team)
		}
		p := r.Team[field]
		for _, n := range p.Nodes {
			out = append(out, Entry{ID: n.ID, Name: n.Name, Type: n.State + n.Type + n.Email})
		}
		if !p.PageInfo.HasNextPage || p.PageInfo.EndCursor == "" {
			return out, nil
		}
		vars["after"] = p.PageInfo.EndCursor
	}
	return nil, fmt.Errorf("linear: catalog %s for %s has more than %d pages", kind, team, catalogMaxPages)
}

// Eligibility is the intake gate's answer for one issue.
type Eligibility struct {
	Issue    string `json:"issue"`
	Eligible bool   `json:"eligible"`
	Reason   string `json:"reason,omitempty"`
}

// Eligible decides whether an issue may start a run: not done, canceled
// or archived, no open blockers, no label in skip (case-insensitive) and
// no open run (openRun names it, "" if none).
func Eligible(c Context, skip []string, openRun string) Eligibility {
	e := Eligibility{Issue: c.Identifier}
	switch {
	case c.ArchivedAt != nil:
		e.Reason = "archived"
	case c.State.Type == "completed":
		e.Reason = "state " + c.State.Name + " is done"
	case c.State.Type == "canceled":
		e.Reason = "state " + c.State.Name + " is canceled"
	}
	if e.Reason == "" {
		if bs := c.OpenBlockers(); len(bs) > 0 {
			ids := make([]string, len(bs))
			for i, b := range bs {
				ids[i] = b.Identifier
			}
			e.Reason = "blocked by " + strings.Join(ids, ", ")
		}
	}
	if e.Reason == "" && len(c.OpenDeps) > 0 {
		e.Reason = "depends on open " + strings.Join(c.OpenDeps, ", ")
	}
	if e.Reason == "" && openRun != "" {
		e.Reason = "open run " + openRun
	}
	if e.Reason == "" {
		for _, l := range c.Labels {
			for _, s := range skip {
				if s = strings.TrimSpace(s); s != "" && strings.EqualFold(l, s) {
					e.Reason = "label " + l + " is skipped"
				}
			}
		}
	}
	e.Eligible = e.Reason == ""
	return e
}
