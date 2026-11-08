package linear

import "context"

// AgentSession is a Linear agent session on an issue (delegation or a
// mention of an app). Jobshop does not hold Linear conversations: it
// answers each one once, which ends it.
type AgentSession struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Issue  string `json:"issue"`
	Team   string `json:"team"`
}

// AgentSessionReply is the response jobshop ends a session with.
const AgentSessionReply = "Jobshop tracks this work through its runs, issue comments and pull requests, not this chat."

// OpenAgentSessions lists the viewer app's sessions that are not complete
// or errored, newest first (the API takes no status filter).
func (c *Client) OpenAgentSessions(ctx context.Context) ([]AgentSession, error) {
	var r struct {
		Viewer        struct{ ID string } `json:"viewer"`
		AgentSessions struct {
			Nodes []struct {
				ID      string `json:"id"`
				Status  string `json:"status"`
				AppUser *struct {
					ID string `json:"id"`
				} `json:"appUser"`
				Issue *struct {
					Identifier string   `json:"identifier"`
					Team       TeamInfo `json:"team"`
				} `json:"issue"`
			} `json:"nodes"`
		} `json:"agentSessions"`
	}
	const q = `query { viewer { id } agentSessions(first: 100, orderBy: updatedAt) { nodes { id status appUser { id } issue { identifier team { id key } } } } }`
	if err := c.GraphQL(ctx, q, nil, &r); err != nil {
		return nil, err
	}
	var out []AgentSession
	for _, n := range r.AgentSessions.Nodes {
		if n.Status == "complete" || n.Status == "error" || n.AppUser == nil || n.AppUser.ID != r.Viewer.ID {
			continue
		}
		s := AgentSession{ID: n.ID, Status: n.Status}
		if n.Issue != nil {
			s.Issue, s.Team = n.Issue.Identifier, n.Issue.Team.Key
		}
		out = append(out, s)
	}
	return out, nil
}

// EndAgentSession posts body as the session's response, which completes it.
func (c *Client) EndAgentSession(ctx context.Context, id, body string) error {
	var r struct {
		AgentActivityCreate struct {
			Success bool `json:"success"`
		} `json:"agentActivityCreate"`
	}
	const m = `mutation($s: String!, $content: JSONObject!) { agentActivityCreate(input: {agentSessionId: $s, content: $content}) { success } }`
	return c.GraphQL(ctx, m, map[string]any{"s": id, "content": map[string]any{"type": "response", "body": body}}, &r)
}
