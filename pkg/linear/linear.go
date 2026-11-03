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
}

func (c *Client) do(ctx context.Context, query string, vars map[string]any, out any) error {
	b, err := json.Marshal(map[string]any{"query": query, "variables": vars})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(b))
	if err != nil {
		return err
	}
	// Personal API keys go in Authorization without "Bearer".
	req.Header.Set("Authorization", c.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
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

const issueQuery = `query($id: String!) { issue(id: $id) {
  id identifier title description url
  state { id name type }
  team { id key }
  comments(first: 100) { nodes { id body url createdAt user { name isMe } botActor { name } } }
} }`

// GetIssue reads an issue by key (ENG-123) with its comments, oldest first.
func (c *Client) GetIssue(ctx context.Context, key string) (Issue, error) {
	var r struct {
		Issue *struct {
			Issue
			Comments struct {
				Nodes []struct {
					ID        string    `json:"id"`
					Body      string    `json:"body"`
					URL       string    `json:"url"`
					CreatedAt time.Time `json:"createdAt"`
					User      *struct {
						Name string `json:"name"`
						IsMe bool   `json:"isMe"`
					} `json:"user"`
					BotActor *struct {
						Name string `json:"name"`
					} `json:"botActor"`
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
		cm := Comment{ID: n.ID, Body: n.Body, URL: n.URL, CreatedAt: n.CreatedAt}
		switch {
		case n.User != nil:
			cm.Author, cm.Self = n.User.Name, n.User.IsMe
		case n.BotActor != nil:
			cm.Author, cm.Bot = n.BotActor.Name, true
		default:
			cm.Bot = true
		}
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

const statesQuery = `query($team: String!) { team(id: $team) { states { nodes { id name type } } } }`

const updateMutation = `mutation($id: String!, $input: IssueUpdateInput!) { issueUpdate(id: $id, input: $input) { success } }`

// SetState moves the issue to the team state named name (case-insensitive).
func (c *Client) SetState(ctx context.Context, key, name string) (State, error) {
	is, err := c.GetIssue(ctx, key)
	if err != nil {
		return State{}, err
	}
	var sr struct {
		Team struct {
			States struct {
				Nodes []State `json:"nodes"`
			} `json:"states"`
		} `json:"team"`
	}
	if err := c.do(ctx, statesQuery, map[string]any{"team": is.Team.ID}, &sr); err != nil {
		return State{}, err
	}
	var st *State
	var names []string
	for i, s := range sr.Team.States.Nodes {
		names = append(names, s.Name)
		if strings.EqualFold(s.Name, name) {
			st = &sr.Team.States.Nodes[i]
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
