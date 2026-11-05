// Package linear is a small Linear GraphQL client for the tracker
// commands (#201): read an issue, comment on it, move its state. The API
// key comes from the controller's identity for the issue's team (#200).
package linear

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

// DefaultAPIURL is Linear's GraphQL endpoint.
const DefaultAPIURL = "https://api.linear.app/graphql"

// KeyPattern is an issue key such as ENG-123.
var KeyPattern = regexp.MustCompile(`^([A-Z][A-Z0-9]*)-[1-9][0-9]*$`)

// Team returns the team key of an issue key ("ENG" for "ENG-123").
func Team(key string) (string, bool) {
	m := KeyPattern.FindStringSubmatch(key)
	if m == nil {
		return "", false
	}
	return m[1], true
}

// Client calls the Linear API with one API key.
type Client struct {
	URL   string
	Token string
	HTTP  *http.Client
}

// New returns a client for token. url "" is DefaultAPIURL.
func New(url, token string) *Client {
	if url == "" {
		url = DefaultAPIURL
	}
	return &Client{URL: url, Token: token, HTTP: &http.Client{Timeout: 30 * time.Second}}
}

// Issue is the part of a Linear issue the prompts need.
type Issue struct {
	ID          string    `json:"id"`
	Identifier  string    `json:"identifier"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	State       State     `json:"state"`
	Team        TeamInfo  `json:"team"`
	Comments    []Comment `json:"comments"`
}

// State is a workflow state.
type State struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

// TeamInfo names the issue's team.
type TeamInfo struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

// Comment is one issue comment. Self marks comments by the API key's own
// user (jobshop's sessions); Bot marks comments with no user
// (integrations).
type Comment struct {
	ID        string    `json:"id,omitempty"`
	Body      string    `json:"body"`
	Author    string    `json:"author"`
	URL       string    `json:"url,omitempty"`
	Self      bool      `json:"self,omitempty"`
	Bot       bool      `json:"bot,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	// ParentID is the comment this one replies to (GetIssue only).
	ParentID string `json:"parent_id,omitempty"`
	// ParentURL is the replied-to comment's URL (GetIssue only).
	ParentURL string `json:"parent_url,omitempty"`
	// AuthorID is the user's (or bot actor's) id, Handle the user's
	// displayName and Email the user's email (#595). Author stays the
	// full name, which is neither unique nor stable.
	AuthorID string `json:"author_id,omitempty"`
	Handle   string `json:"handle,omitempty"`
	Email    string `json:"email,omitempty"`
	// EditedAt is when the body was last edited, nil if never (GetIssue).
	EditedAt *time.Time `json:"edited_at,omitempty"`
}

// Display is "displayName (name)", or the name alone when they match
// or there is no displayName (bots).
func (cm Comment) Display() string {
	if cm.Handle == "" || cm.Handle == cm.Author {
		return cm.Author
	}
	if cm.Author == "" {
		return cm.Handle
	}
	return cm.Handle + " (" + cm.Author + ")"
}

// commentUser is a comment's user as the comment queries read it.
type commentUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	IsMe        bool   `json:"isMe"`
}

// commentBot is a comment's bot actor.
type commentBot struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// setAuthor fills the author fields from the user or bot actor.
func (cm *Comment) setAuthor(u *commentUser, b *commentBot) {
	switch {
	case u != nil:
		cm.Author, cm.AuthorID, cm.Handle, cm.Email, cm.Self = u.Name, u.ID, u.DisplayName, u.Email, u.IsMe
	case b != nil:
		cm.Author, cm.AuthorID, cm.Bot = b.Name, b.ID, true
	default:
		cm.Bot = true
	}
}

// commentAuthorFields is the selection setAuthor reads.
const commentAuthorFields = `user { id name displayName email isMe } botActor { id name }`

// Viewer is the user (or app actor) a token acts as.
type Viewer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Viewer reads the token's own user, so a bad token fails here.
func (c *Client) Viewer(ctx context.Context) (Viewer, error) {
	var r struct {
		Viewer Viewer `json:"viewer"`
	}
	if err := c.do(ctx, `query { viewer { id name } }`, nil, &r); err != nil {
		return Viewer{}, err
	}
	if r.Viewer.ID == "" {
		return Viewer{}, fmt.Errorf("linear: viewer: empty response")
	}
	return r.Viewer, nil
}

// Me is the token's own actor id and its organization id, for checking
// who sent a webhook (#300).
func (c *Client) Me(ctx context.Context) (user, org string, err error) {
	var r struct {
		Viewer struct {
			ID           string `json:"id"`
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
		} `json:"viewer"`
	}
	if err := c.do(ctx, `query { viewer { id organization { id } } }`, nil, &r); err != nil {
		return "", "", err
	}
	if r.Viewer.ID == "" || r.Viewer.Organization.ID == "" {
		return "", "", fmt.Errorf("linear: viewer organization: empty response")
	}
	return r.Viewer.ID, r.Viewer.Organization.ID, nil
}

func (c *Client) do(ctx context.Context, query string, vars map[string]any, out any) error {
	b, err := json.Marshal(map[string]any{"query": query, "variables": vars})
	if err != nil {
		return err
	}
	resp, body, err := c.post(ctx, b)
	if err != nil {
		return err
	}
	var r struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return fmt.Errorf("linear: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	if len(r.Errors) > 0 {
		msgs := make([]string, len(r.Errors))
		for i, e := range r.Errors {
			msgs[i] = e.Message
		}
		return fmt.Errorf("linear: %s", strings.Join(msgs, "; "))
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("linear: %s", resp.Status)
	}
	return json.Unmarshal(r.Data, out)
}

// post sends one GraphQL request body and reads the response.
func (c *Client) post(ctx context.Context, b []byte) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(b))
	if err != nil {
		return nil, nil, err
	}
	// Personal API keys (lin_api_...) go in Authorization as they are;
	// OAuth tokens, such as a linear-app identity's (#278), need Bearer.
	auth := c.Token
	if !strings.HasPrefix(auth, "lin_api_") && !strings.HasPrefix(auth, "Bearer ") {
		auth = "Bearer " + auth
	}
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	return resp, body, err
}

// Raw posts body (a JSON GraphQL request) and returns the response body
// as is (#546). A GraphQL errors response is not an error here, so the
// caller can print it; a non-200 status is, with the body still returned.
func (c *Client) Raw(ctx context.Context, body []byte) ([]byte, error) {
	resp, out, err := c.post(ctx, body)
	if err != nil {
		return out, err
	}
	if resp.StatusCode != http.StatusOK {
		return out, fmt.Errorf("linear: %s", resp.Status)
	}
	return out, nil
}

const issueQuery = `query($id: String!) { issue(id: $id) {
  id identifier title description url
  state { id name type }
  team { id key }
  comments(first: 100) { nodes { id body url createdAt updatedAt editedAt parent { id url } ` + commentAuthorFields + ` } }
} }`

// GetIssue reads an issue by key (ENG-123) with its comments, oldest first.
func (c *Client) GetIssue(ctx context.Context, key string) (Issue, error) {
	var r struct {
		Issue *struct {
			Issue
			Comments struct {
				Nodes []struct {
					ID        string     `json:"id"`
					Body      string     `json:"body"`
					URL       string     `json:"url"`
					CreatedAt time.Time  `json:"createdAt"`
					EditedAt  *time.Time `json:"editedAt"`
					Parent    *struct {
						ID  string `json:"id"`
						URL string `json:"url"`
					} `json:"parent"`
					User     *commentUser `json:"user"`
					BotActor *commentBot  `json:"botActor"`
				} `json:"nodes"`
			} `json:"comments"`
		} `json:"issue"`
	}
	if err := c.do(ctx, issueQuery, map[string]any{"id": key}, &r); err != nil {
		return Issue{}, err
	}
	if r.Issue == nil {
		return Issue{}, fmt.Errorf("linear: issue %s not found", key)
	}
	is := r.Issue.Issue
	is.Comments = nil
	for _, n := range r.Issue.Comments.Nodes {
		cm := Comment{ID: n.ID, Body: n.Body, URL: n.URL, CreatedAt: n.CreatedAt, EditedAt: n.EditedAt}
		if n.Parent != nil {
			cm.ParentID, cm.ParentURL = n.Parent.ID, n.Parent.URL
		}
		cm.setAuthor(n.User, n.BotActor)
		is.Comments = append(is.Comments, cm)
	}
	// Linear returns newest first.
	for i, j := 0, len(is.Comments)-1; i < j; i, j = i+1, j-1 {
		is.Comments[i], is.Comments[j] = is.Comments[j], is.Comments[i]
	}
	return is, nil
}

const commentMutation = `mutation($input: CommentCreateInput!) { commentCreate(input: $input) { success } }`

// AddComment posts body on the issue.
func (c *Client) AddComment(ctx context.Context, key, body string) error {
	is, err := c.GetIssue(ctx, key)
	if err != nil {
		return err
	}
	var r struct {
		CommentCreate struct {
			Success bool `json:"success"`
		} `json:"commentCreate"`
	}
	if err := c.do(ctx, commentMutation, map[string]any{"input": map[string]any{"issueId": is.ID, "body": body}}, &r); err != nil {
		return err
	}
	if !r.CommentCreate.Success {
		return errors.New("linear: comment not created")
	}
	return nil
}

const updateMutation = `mutation($id: String!, $input: IssueUpdateInput!) { issueUpdate(id: $id, input: $input) { success } }`

// SetState moves the issue to the team state named name (case-insensitive).
// The states come from the paged team reader writes use (#683).
func (c *Client) SetState(ctx context.Context, key, name string) (State, error) {
	is, err := c.GetIssue(ctx, key)
	if err != nil {
		return State{}, err
	}
	var states []State
	if err := teamConn(ctx, c, is.Team.ID, "states", "id name type", &states); err != nil {
		return State{}, err
	}
	var st *State
	var names []string
	for i, s := range states {
		names = append(names, s.Name)
		if strings.EqualFold(s.Name, name) {
			st = &states[i]
		}
	}
	if st == nil {
		return State{}, fmt.Errorf("linear: team %s has no state %q (states: %s)", is.Team.Key, name, strings.Join(names, ", "))
	}
	var ur struct {
		IssueUpdate struct {
			Success bool `json:"success"`
		} `json:"issueUpdate"`
	}
	if err := c.do(ctx, updateMutation, map[string]any{"id": is.ID, "input": map[string]any{"stateId": st.ID}}, &ur); err != nil {
		return State{}, err
	}
	if !ur.IssueUpdate.Success {
		return State{}, errors.New("linear: state not updated")
	}
	return *st, nil
}

// Change is one thing that happened on an issue (#255): created, a
// comment, a state change or an assignee change.
type Change struct {
	Issue string `json:"issue"` // key, ENG-123
	Title string `json:"title"` // issue title
	Kind  string `json:"kind"`  // issue, comment, state, assignee
	Actor string `json:"actor"`
	// ActorID is the comment author's user or bot id (comments only).
	ActorID string    `json:"actor_id,omitempty"`
	Self    bool      `json:"self,omitempty"` // by the API key's own user
	Bot     bool      `json:"bot,omitempty"`  // no user (integration)
	Text    string    `json:"text"`           // comment body, or "Todo -> Done"
	URL     string    `json:"url"`
	At      time.Time `json:"at"`
	ItemID  string    `json:"item_id"` // stable id of the comment or history entry
}

const activityQuery = `query($team: String!, $since: DateTimeOrDuration!) {
  issues(first: 50, orderBy: updatedAt, filter: {team: {key: {eq: $team}}, updatedAt: {gt: $since}}) { nodes {
    id identifier title url createdAt
    creator { name isMe }
    comments(first: 50, filter: {createdAt: {gt: $since}}) { nodes { id body url createdAt ` + commentAuthorFields + ` } }
    history(first: 50) { nodes { id createdAt actor { name isMe } fromState { name } toState { name } fromAssignee { name } toAssignee { name } } }
  } }
}`

type actor struct {
	Name string `json:"name"`
	IsMe bool   `json:"isMe"`
}

type named struct {
	Name string `json:"name"`
}

// Changes lists what happened in the team's issues after since, oldest
// first. It reads the 50 most recently updated issues and up to 50
// comments and history entries each.
func (c *Client) Changes(ctx context.Context, team string, since time.Time) ([]Change, error) {
	var r struct {
		Issues struct {
			Nodes []struct {
				Identifier string    `json:"identifier"`
				Title      string    `json:"title"`
				URL        string    `json:"url"`
				CreatedAt  time.Time `json:"createdAt"`
				Creator    *actor    `json:"creator"`
				Comments   struct {
					Nodes []struct {
						ID        string       `json:"id"`
						Body      string       `json:"body"`
						URL       string       `json:"url"`
						CreatedAt time.Time    `json:"createdAt"`
						User      *commentUser `json:"user"`
						BotActor  *commentBot  `json:"botActor"`
					} `json:"nodes"`
				} `json:"comments"`
				History struct {
					Nodes []struct {
						ID           string    `json:"id"`
						CreatedAt    time.Time `json:"createdAt"`
						Actor        *actor    `json:"actor"`
						FromState    *named    `json:"fromState"`
						ToState      *named    `json:"toState"`
						FromAssignee *named    `json:"fromAssignee"`
						ToAssignee   *named    `json:"toAssignee"`
					} `json:"nodes"`
				} `json:"history"`
			} `json:"nodes"`
		} `json:"issues"`
	}
	vars := map[string]any{"team": team, "since": since.UTC().Format(time.RFC3339Nano)}
	if err := c.do(ctx, activityQuery, vars, &r); err != nil {
		return nil, err
	}
	who := func(a *actor) (string, bool, bool) {
		if a == nil {
			return "", false, true
		}
		return a.Name, a.IsMe, false
	}
	name := func(n *named) string {
		if n == nil {
			return "none"
		}
		return n.Name
	}
	var out []Change
	for _, is := range r.Issues.Nodes {
		base := Change{Issue: is.Identifier, Title: is.Title, URL: is.URL}
		if is.CreatedAt.After(since) {
			ch := base
			ch.Kind, ch.At, ch.ItemID, ch.Text = "issue", is.CreatedAt, "issue-"+is.Identifier, is.Title
			ch.Actor, ch.Self, ch.Bot = who(is.Creator)
			out = append(out, ch)
		}
		for _, cm := range is.Comments.Nodes {
			ch := base
			ch.Kind, ch.At, ch.ItemID, ch.Text = "comment", cm.CreatedAt, cm.ID, cm.Body
			if cm.URL != "" {
				ch.URL = cm.URL
			}
			var a Comment
			a.setAuthor(cm.User, cm.BotActor)
			ch.Actor, ch.ActorID, ch.Self, ch.Bot = a.Display(), a.AuthorID, a.Self, a.Bot
			out = append(out, ch)
		}
		for _, h := range is.History.Nodes {
			if !h.CreatedAt.After(since) {
				continue
			}
			ch := base
			ch.At, ch.ItemID = h.CreatedAt, h.ID
			ch.Actor, ch.Self, ch.Bot = who(h.Actor)
			switch {
			case h.ToState != nil || h.FromState != nil:
				ch.Kind, ch.Text = "state", name(h.FromState)+" -> "+name(h.ToState)
			case h.ToAssignee != nil || h.FromAssignee != nil:
				ch.Kind, ch.Text = "assignee", name(h.FromAssignee)+" -> "+name(h.ToAssignee)
			default:
				continue
			}
			out = append(out, ch)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out, nil
}

const labelsQuery = `query($name: String!) { issueLabels(filter: {name: {eqIgnoreCase: $name}}) { nodes { id team { id } } } }`

const labelCreateMutation = `mutation($input: IssueLabelCreateInput!) { issueLabelCreate(input: $input) { success issueLabel { id } } }`

const addLabelMutation = `mutation($id: String!, $labelId: String!) { issueAddLabel(id: $id, labelId: $labelId) { success } }`

// AddLabel puts the label named name on the issue, creating it in the
// issue's team when neither the team nor the workspace has it.
func (c *Client) AddLabel(ctx context.Context, key, name string) error {
	is, err := c.GetIssue(ctx, key)
	if err != nil {
		return err
	}
	var lr struct {
		IssueLabels struct {
			Nodes []struct {
				ID   string  `json:"id"`
				Team *TeamID `json:"team"`
			} `json:"nodes"`
		} `json:"issueLabels"`
	}
	if err := c.do(ctx, labelsQuery, map[string]any{"name": name}, &lr); err != nil {
		return err
	}
	id := ""
	for _, n := range lr.IssueLabels.Nodes {
		if n.Team == nil || n.Team.ID == is.Team.ID {
			id = n.ID
			break
		}
	}
	if id == "" {
		var cr struct {
			IssueLabelCreate struct {
				Success    bool   `json:"success"`
				IssueLabel TeamID `json:"issueLabel"`
			} `json:"issueLabelCreate"`
		}
		if err := c.do(ctx, labelCreateMutation, map[string]any{"input": map[string]any{"name": name, "teamId": is.Team.ID}}, &cr); err != nil {
			return err
		}
		if !cr.IssueLabelCreate.Success || cr.IssueLabelCreate.IssueLabel.ID == "" {
			return errors.New("linear: label not created")
		}
		id = cr.IssueLabelCreate.IssueLabel.ID
	}
	var ar struct {
		IssueAddLabel struct {
			Success bool `json:"success"`
		} `json:"issueAddLabel"`
	}
	if err := c.do(ctx, addLabelMutation, map[string]any{"id": is.ID, "labelId": id}, &ar); err != nil {
		return err
	}
	if !ar.IssueAddLabel.Success {
		return errors.New("linear: label not added")
	}
	return nil
}

const removeLabelMutation = `mutation($id: String!, $labelId: String!) { issueRemoveLabel(id: $id, labelId: $labelId) { success } }`

const issueLabelsQuery = `query($id: String!) { issue(id: $id) { id labels { nodes { id name } } } }`

// RemoveLabel takes the label named name (case-insensitive) off the
// issue. A label the issue does not have is not an error.
func (c *Client) RemoveLabel(ctx context.Context, key, name string) error {
	var r struct {
		Issue *struct {
			ID     string `json:"id"`
			Labels struct {
				Nodes []Entry `json:"nodes"`
			} `json:"labels"`
		} `json:"issue"`
	}
	if err := c.do(ctx, issueLabelsQuery, map[string]any{"id": key}, &r); err != nil {
		return err
	}
	if r.Issue == nil {
		return fmt.Errorf("linear: issue %s not found", key)
	}
	for _, l := range r.Issue.Labels.Nodes {
		if !strings.EqualFold(l.Name, name) {
			continue
		}
		var rr struct {
			IssueRemoveLabel struct {
				Success bool `json:"success"`
			} `json:"issueRemoveLabel"`
		}
		if err := c.do(ctx, removeLabelMutation, map[string]any{"id": r.Issue.ID, "labelId": l.ID}, &rr); err != nil {
			return err
		}
		if !rr.IssueRemoveLabel.Success {
			return errors.New("linear: label not removed")
		}
	}
	return nil
}

// TeamID is an object with only an id.
type TeamID struct {
	ID string `json:"id"`
}
