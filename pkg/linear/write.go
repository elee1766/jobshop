package linear

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Write commands (#280): create, update, ask, relate. Each builds a
// Mutation from reads (team ids, states, labels, users), so --dry-run
// prints exactly what would be sent. Do runs it under an idempotency key.

// Mutation is one GraphQL write: Op names it, ID is the client-chosen id
// of the created entity (create, ask, relate), Verify checks an update
// after it is read back.
type Mutation struct {
	Op        string         `json:"op"`
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
	ID        string         `json:"id,omitempty"`
	// Issue is the key the mutation acts on (update), for read back.
	Issue string `json:"issue,omitempty"`
	check func(Context) error
}

// Result is what a write prints and the controller stores under its key.
type Result struct {
	Op         string `json:"op"`
	ID         string `json:"id"`
	Identifier string `json:"identifier,omitempty"`
	URL        string `json:"url,omitempty"`
	// Replayed is set when the result came from the key store or a read
	// back of an earlier attempt, not from this call's mutation.
	Replayed bool `json:"replayed,omitempty"`
}

// KeyStore keeps key -> result for idempotent writes. The CLI implements
// it with controller RPCs; tests with a map.
type KeyStore interface {
	// Claim returns (result, done, retry): done with the stored result,
	// or pending with retry true when an earlier attempt did not finish.
	Claim(ctx context.Context, key, op string) (result string, done, retry bool, err error)
	Finish(ctx context.Context, key, result string) (string, error)
}

// KeyID is the UUID (v4 layout) a write with key gives its entity, so a
// retry can find what an earlier attempt created.
func KeyID(key string) string {
	h := sha256.Sum256([]byte("jobshop-linear-write:" + key))
	h[6] = (h[6] & 0x0f) | 0x40
	h[8] = (h[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", h[0:4], h[4:6], h[6:8], h[8:10], h[10:16])
}

// ErrConflict is returned when --if-updated-at no longer matches.
var ErrConflict = errors.New("linear: issue changed")

// Do runs m. dry returns the mutation unsent. With a key (keys non-nil),
// a key already done returns its stored result, and a key left pending
// by an earlier attempt first looks for the entity that attempt created.
func (c *Client) Do(ctx context.Context, keys KeyStore, key string, dry bool, m Mutation) (any, error) {
	if dry {
		return m, nil
	}
	retry := false
	if keys != nil && key != "" {
		stored, done, r, err := keys.Claim(ctx, key, m.Op)
		if err != nil {
			return nil, err
		}
		if done {
			var res Result
			if err := json.Unmarshal([]byte(stored), &res); err != nil {
				return nil, fmt.Errorf("linear: stored result for key %s: %w", key, err)
			}
			res.Replayed = true
			return res, nil
		}
		retry = r
	}
	var res Result
	found := false
	if retry && m.ID != "" {
		r, ok, err := c.lookup(ctx, m.Op, m.ID)
		if err != nil {
			return nil, err
		}
		res, found = r, ok
		res.Replayed = ok
	}
	if !found {
		r, err := c.apply(ctx, m)
		if err != nil && m.ID != "" {
			// The local key may be gone (#415: a 48h cache) while the
			// entity exists: Linear rejects the reused id. The id is the
			// remote key, so finding it means the write already landed.
			if prev, ok, lerr := c.lookup(ctx, m.Op, m.ID); lerr == nil && ok {
				prev.Replayed = true
				r, err = prev, nil
			}
		}
		if err != nil {
			return nil, err
		}
		res = r
	}
	if keys != nil && key != "" {
		b, err := json.Marshal(res)
		if err != nil {
			return nil, err
		}
		stored, err := keys.Finish(ctx, key, string(b))
		if err != nil {
			return nil, err
		}
		// Another attempt finished first: its result wins.
		if stored != string(b) {
			var first Result
			if json.Unmarshal([]byte(stored), &first) == nil {
				first.Replayed = true
				return first, nil
			}
		}
	}
	return res, nil
}

func (c *Client) apply(ctx context.Context, m Mutation) (Result, error) {
	var raw map[string]struct {
		Success bool `json:"success"`
		Issue   *struct {
			ID         string `json:"id"`
			Identifier string `json:"identifier"`
			URL        string `json:"url"`
		} `json:"issue"`
		Comment *struct {
			ID  string `json:"id"`
			URL string `json:"url"`
		} `json:"comment"`
		IssueRelation *struct {
			ID string `json:"id"`
		} `json:"issueRelation"`
	}
	if err := c.do(ctx, m.Query, m.Variables, &raw); err != nil {
		return Result{}, err
	}
	res := Result{Op: m.Op, ID: m.ID}
	for _, p := range raw {
		if !p.Success {
			return Result{}, fmt.Errorf("linear: %s did not succeed", m.Op)
		}
		switch {
		case p.Issue != nil:
			res.ID, res.Identifier, res.URL = p.Issue.ID, p.Issue.Identifier, p.Issue.URL
		case p.Comment != nil:
			res.ID, res.URL = p.Comment.ID, p.Comment.URL
		case p.IssueRelation != nil:
			res.ID = p.IssueRelation.ID
		}
	}
	if m.check != nil {
		// Read back: Linear may accept an update and drop part of it.
		got, err := c.GetContext(ctx, m.Issue, false)
		if err != nil {
			return Result{}, fmt.Errorf("linear: read back %s: %w", m.Issue, err)
		}
		if err := m.check(got); err != nil {
			return Result{}, err
		}
		res.ID, res.Identifier, res.URL = got.ID, got.Identifier, got.URL
	}
	return res, nil
}

// lookup finds the entity an earlier attempt created with id.
func (c *Client) lookup(ctx context.Context, op, id string) (Result, bool, error) {
	var q string
	switch op {
	case OpCreate:
		q = `query($id: String!) { issue(id: $id) { id identifier url } }`
	case OpAsk:
		q = `query($id: String!) { comment(id: $id) { id url } }`
	case OpRelate:
		q = `query($id: String!) { issueRelation(id: $id) { id } }`
	default:
		return Result{}, false, nil
	}
	var r map[string]*struct {
		ID         string `json:"id"`
		Identifier string `json:"identifier"`
		URL        string `json:"url"`
	}
	if err := c.do(ctx, q, map[string]any{"id": id}, &r); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return Result{}, false, nil
		}
		return Result{}, false, err
	}
	for _, e := range r {
		if e != nil && e.ID != "" {
			return Result{Op: op, ID: e.ID, Identifier: e.Identifier, URL: e.URL}, true, nil
		}
	}
	return Result{}, false, nil
}

// Write ops.
const (
	OpCreate = "linear.create"
	OpUpdate = "linear.update"
	OpAsk    = "linear.ask"
	OpRelate = "linear.relate"
)

// teamInfo is what writes resolve names against.
type teamInfo struct {
	ID      string
	Key     string
	States  []Entry
	Labels  []Entry
	Members []member
}

type member struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	URL         string `json:"url"`
}

// teamIDQuery reads only the team id. States, labels and members come
// from one paged query each (teamConn): asking for all three at 250 in
// one query was rejected as "Query too complex" (#680, like #653). Both
// read the team as team(id: key), not teams(filter:), which Linear
// scores as a list and rejected even for one connection (#683).
const teamIDQuery = `query($team: String!) { team(id: $team) { id key } }`

// teamNeed says which lists team reads.
type teamNeed struct{ states, labels, members bool }

func (c *Client) team(ctx context.Context, key string, need teamNeed) (teamInfo, error) {
	var r struct {
		Team *TeamInfo `json:"team"`
	}
	if err := c.do(ctx, teamIDQuery, map[string]any{"team": key}, &r); err != nil {
		return teamInfo{}, err
	}
	if r.Team == nil {
		return teamInfo{}, fmt.Errorf("linear: team %s not found", key)
	}
	t := teamInfo{ID: r.Team.ID, Key: r.Team.Key}
	if need.states {
		if err := teamConn(ctx, c, t.Key, "states", "id name type", &t.States); err != nil {
			return teamInfo{}, err
		}
	}
	if need.labels {
		if err := teamConn(ctx, c, t.Key, "labels", "id name", &t.Labels); err != nil {
			return teamInfo{}, err
		}
	}
	if need.members {
		if err := teamConn(ctx, c, t.Key, "members", "id name displayName email url", &t.Members); err != nil {
			return teamInfo{}, err
		}
	}
	return t, nil
}

// teamConn reads one connection of a team, paged like Catalog, into out.
func teamConn[T any](ctx context.Context, c *Client, team, conn, fields string, out *[]T) error {
	q := `query($team: String!, $first: Int!, $after: String) { team(id: $team) {
  ` + conn + `(first: $first, after: $after) { nodes { ` + fields + ` } pageInfo { hasNextPage endCursor } }
} }`
	type page struct {
		Nodes    []T `json:"nodes"`
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
	}
	vars := map[string]any{"team": team, "first": catalogPage}
	for range catalogMaxPages {
		var r struct {
			Team map[string]json.RawMessage `json:"team"`
		}
		if err := c.do(ctx, q, vars, &r); err != nil {
			return err
		}
		if r.Team == nil {
			return fmt.Errorf("linear: team %s not found", team)
		}
		var p page
		if raw := r.Team[conn]; len(raw) > 0 {
			if err := json.Unmarshal(raw, &p); err != nil {
				return fmt.Errorf("linear: %s of team %s: %w", conn, team, err)
			}
		}
		*out = append(*out, p.Nodes...)
		if !p.PageInfo.HasNextPage || p.PageInfo.EndCursor == "" {
			return nil
		}
		vars["after"] = p.PageInfo.EndCursor
	}
	return fmt.Errorf("linear: %s of team %s has more than %d pages", conn, team, catalogMaxPages)
}

func pick(kind string, es []Entry, name string) (Entry, error) {
	var names []string
	for _, e := range es {
		if strings.EqualFold(e.Name, name) {
			return e, nil
		}
		names = append(names, e.Name)
	}
	return Entry{}, fmt.Errorf("linear: no %s %q (have: %s)", kind, name, strings.Join(names, ", "))
}

// user finds a team member by name, display name or email
// (case-insensitive). "me" is resolved by Client.assignee.
func (t teamInfo) user(name string) (member, error) {
	var names []string
	for _, m := range t.Members {
		if strings.EqualFold(m.Name, name) || strings.EqualFold(m.DisplayName, name) || strings.EqualFold(m.Email, name) {
			return m, nil
		}
		names = append(names, m.DisplayName)
	}
	return member{}, fmt.Errorf("linear: no user %q in team %s (have: %s)", name, t.Key, strings.Join(names, ", "))
}

func (t teamInfo) labelIDs(names []string) ([]string, error) {
	ids := []string{}
	for _, n := range names {
		e, err := pick("label", t.Labels, n)
		if err != nil {
			return nil, err
		}
		ids = append(ids, e.ID)
	}
	return ids, nil
}

// CreateParams are the fields of a new issue. Team is a team key.
type CreateParams struct {
	Team        string
	Title       string
	Description string
	State       string
	Labels      []string
	Assignee    string
	Parent      string // issue key
}

const createMutation = `mutation($input: IssueCreateInput!) { issueCreate(input: $input) { success issue { id identifier url } } }`

// PlanCreate resolves names and builds the issueCreate mutation with id.
func (c *Client) PlanCreate(ctx context.Context, p CreateParams, id string) (Mutation, error) {
	if strings.TrimSpace(p.Title) == "" {
		return Mutation{}, errors.New("linear: create needs a title")
	}
	t, err := c.team(ctx, strings.ToUpper(p.Team), teamNeed{states: p.State != "", labels: len(p.Labels) > 0, members: p.Assignee != "" && !strings.EqualFold(p.Assignee, "me")})
	if err != nil {
		return Mutation{}, err
	}
	in := map[string]any{"teamId": t.ID, "title": p.Title}
	if id != "" {
		in["id"] = id
	}
	if p.Description != "" {
		in["description"] = p.Description
	}
	if p.State != "" {
		s, err := pick("state", t.States, p.State)
		if err != nil {
			return Mutation{}, err
		}
		in["stateId"] = s.ID
	}
	if len(p.Labels) > 0 {
		ids, err := t.labelIDs(p.Labels)
		if err != nil {
			return Mutation{}, err
		}
		in["labelIds"] = ids
	}
	if p.Assignee != "" {
		field, id, _, err := c.assignee(ctx, t, p.Assignee)
		if err != nil {
			return Mutation{}, err
		}
		in[field] = id
	}
	if p.Parent != "" {
		par, err := c.GetIssue(ctx, p.Parent)
		if err != nil {
			return Mutation{}, err
		}
		in["parentId"] = par.ID
	}
	return Mutation{Op: OpCreate, Query: createMutation, Variables: map[string]any{"input": in}, ID: id}, nil
}

// UpdateParams change an issue. Empty fields are left alone.
type UpdateParams struct {
	Title        string
	State        string
	Assignee     string
	AddLabels    []string
	RemoveLabels []string
	// IfUpdatedAt fails the update when the issue's updatedAt differs.
	IfUpdatedAt time.Time
}

const updateIssueMutation = `mutation($id: String!, $input: IssueUpdateInput!) { issueUpdate(id: $id, input: $input) { success issue { id identifier url } } }`

const updatedAtQuery = `query($id: String!) { issue(id: $id) { id updatedAt team { key } } }`

// PlanUpdate reads the issue, checks IfUpdatedAt and builds issueUpdate.
// The result is verified by reading the issue back.
func (c *Client) PlanUpdate(ctx context.Context, key string, p UpdateParams) (Mutation, error) {
	var r struct {
		Issue *struct {
			ID        string    `json:"id"`
			UpdatedAt time.Time `json:"updatedAt"`
			Team      TeamInfo  `json:"team"`
		} `json:"issue"`
	}
	if err := c.do(ctx, updatedAtQuery, map[string]any{"id": key}, &r); err != nil {
		return Mutation{}, err
	}
	if r.Issue == nil {
		return Mutation{}, fmt.Errorf("linear: issue %s not found", key)
	}
	if !p.IfUpdatedAt.IsZero() && !r.Issue.UpdatedAt.Equal(p.IfUpdatedAt) {
		return Mutation{}, fmt.Errorf("%w: %s updated at %s, not %s", ErrConflict, key,
			r.Issue.UpdatedAt.UTC().Format(time.RFC3339Nano), p.IfUpdatedAt.UTC().Format(time.RFC3339Nano))
	}
	t, err := c.team(ctx, r.Issue.Team.Key, teamNeed{states: p.State != "", labels: len(p.AddLabels)+len(p.RemoveLabels) > 0, members: p.Assignee != "" && !strings.EqualFold(p.Assignee, "me")})
	if err != nil {
		return Mutation{}, err
	}
	in := map[string]any{}
	var checks []func(Context) error
	if p.Title != "" {
		in["title"] = p.Title
		checks = append(checks, func(g Context) error { return expect("title", g.Title, p.Title) })
	}
	if p.State != "" {
		s, err := pick("state", t.States, p.State)
		if err != nil {
			return Mutation{}, err
		}
		in["stateId"] = s.ID
		checks = append(checks, func(g Context) error { return expect("state", g.State.Name, s.Name) })
	}
	if p.Assignee != "" {
		field, id, name, err := c.assignee(ctx, t, p.Assignee)
		if err != nil {
			return Mutation{}, err
		}
		in[field] = id
		// An app with app:assignable that assigns itself becomes the
		// delegate instead; Linear does that, so either counts.
		checks = append(checks, func(g Context) error {
			if strings.EqualFold(g.Delegate, name) {
				return nil
			}
			return expect("assignee", g.Assignee, name)
		})
	}
	if len(p.AddLabels) > 0 {
		ids, err := t.labelIDs(p.AddLabels)
		if err != nil {
			return Mutation{}, err
		}
		in["addedLabelIds"] = ids
	}
	if len(p.RemoveLabels) > 0 {
		ids, err := t.labelIDs(p.RemoveLabels)
		if err != nil {
			return Mutation{}, err
		}
		in["removedLabelIds"] = ids
	}
	if len(in) == 0 {
		return Mutation{}, errors.New("linear: update needs --title, --state, --assignee, --add-label or --remove-label")
	}
	add, remove := p.AddLabels, p.RemoveLabels
	checks = append(checks, func(g Context) error {
		for _, l := range add {
			if !slices.ContainsFunc(g.Labels, func(x string) bool { return strings.EqualFold(x, l) }) {
				return fmt.Errorf("linear: read back: label %q missing", l)
			}
		}
		for _, l := range remove {
			if slices.ContainsFunc(g.Labels, func(x string) bool { return strings.EqualFold(x, l) }) {
				return fmt.Errorf("linear: read back: label %q still set", l)
			}
		}
		return nil
	})
	return Mutation{
		Op: OpUpdate, Query: updateIssueMutation, Variables: map[string]any{"id": r.Issue.ID, "input": in}, Issue: key,
		check: func(g Context) error {
			for _, f := range checks {
				if err := f(g); err != nil {
					return err
				}
			}
			return nil
		},
	}, nil
}

func expect(field, got, want string) error {
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("linear: read back: %s is %q, want %q", field, got, want)
	}
	return nil
}

// AskParams is a comment that asks someone. Mention is a user name,
// display name or email; ReplyTo a comment id to thread under.
type AskParams struct {
	Body    string
	Mention []string
	ReplyTo string
}

const askMutation = `mutation($input: CommentCreateInput!) { commentCreate(input: $input) { success comment { id url } } }`

// PlanAsk builds commentCreate. Each mention is resolved to the user's
// profile URL at the start of the body: Linear turns that URL into a
// mention that notifies, which a plain @name does not.
func (c *Client) PlanAsk(ctx context.Context, key string, p AskParams, id string) (Mutation, error) {
	if strings.TrimSpace(p.Body) == "" {
		return Mutation{}, errors.New("linear: ask needs a body")
	}
	is, err := c.GetIssue(ctx, key)
	if err != nil {
		return Mutation{}, err
	}
	body := p.Body
	if len(p.Mention) > 0 {
		t, err := c.team(ctx, is.Team.Key, teamNeed{members: true})
		if err != nil {
			return Mutation{}, err
		}
		var urls []string
		for _, n := range p.Mention {
			u, err := t.user(strings.TrimPrefix(n, "@"))
			if err != nil {
				return Mutation{}, err
			}
			if u.URL == "" {
				return Mutation{}, fmt.Errorf("linear: user %q has no profile url", n)
			}
			urls = append(urls, u.URL)
		}
		body = strings.Join(urls, " ") + " " + body
	}
	in := map[string]any{"issueId": is.ID, "body": body}
	if id != "" {
		in["id"] = id
	}
	if p.ReplyTo != "" {
		in["parentId"] = p.ReplyTo
	}
	return Mutation{Op: OpAsk, Query: askMutation, Variables: map[string]any{"input": in}, ID: id}, nil
}

// Relation types relate accepts.
const (
	RelBlocks  = "blocks"
	RelRelated = "related"
)

const relateMutation = `mutation($input: IssueRelationCreateInput!) { issueRelationCreate(input: $input) { success issueRelation { id } } }`

// PlanRelate builds issueRelationCreate: key blocks other, or is related.
func (c *Client) PlanRelate(ctx context.Context, key, typ, other, id string) (Mutation, error) {
	if typ != RelBlocks && typ != RelRelated {
		return Mutation{}, fmt.Errorf("linear: relation %q is not blocks or related", typ)
	}
	a, err := c.GetIssue(ctx, key)
	if err != nil {
		return Mutation{}, err
	}
	b, err := c.GetIssue(ctx, other)
	if err != nil {
		return Mutation{}, err
	}
	in := map[string]any{"issueId": a.ID, "relatedIssueId": b.ID, "type": typ}
	if id != "" {
		in["id"] = id
	}
	return Mutation{Op: OpRelate, Query: relateMutation, Variables: map[string]any{"input": in}, ID: id}, nil
}

// assignee resolves an assignee: "me" is the API key's own user (an agent
// assigning itself as it takes an issue; an app needs app:assignable),
// anything else a team member.
func (c *Client) assignee(ctx context.Context, t teamInfo, name string) (field, id, display string, err error) {
	if strings.EqualFold(name, "me") {
		v, err := c.Viewer(ctx)
		if err != nil {
			return "", "", "", err
		}
		return "assigneeId", v.ID, v.Name, nil
	}
	u, err := t.user(name)
	if err != nil {
		return "", "", "", err
	}
	return "assigneeId", u.ID, u.Name, nil
}
