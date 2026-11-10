package linear_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/integrations/apicall"
	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
)

// complexityBudget is Linear's limit per query.
const complexityBudget = 10000

var firstArg = regexp.MustCompile(`\(first: (\$first|\d+)`)

// complexity scores q roughly as Linear does (#683): each connection
// costs its page size, and a connection nested under teams(filter:) is
// counted once per team the list may return. Measured on the real API:
// teams(filter:) with states(first: 100) scored 12550 of 10000, about
// 125 per page item; the same page under team(id:), one object, passed.
func complexity(c call) int {
	q := c.Query
	page := 0
	for _, m := range firstArg.FindAllStringSubmatch(q, -1) {
		n := 50
		if m[1] == "$first" {
			if v, ok := c.Variables["first"].(float64); ok {
				n = int(v)
			}
		} else {
			n, _ = strconv.Atoi(m[1])
		}
		page += n
	}
	if strings.Contains(q, "teams(") && page > 0 {
		return 125 * (page + 1)
	}
	return page + 1
}

// complexityServer answers create's and state's reads, and rejects any
// query over the budget, or naming more than one team connection, with
// "Query too complex". Labels come in two pages.
func complexityServer(t *testing.T) (*linear.Client, *[]call) {
	t.Helper()
	var calls []call
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var c call
		require.NoError(t, json.Unmarshal(b, &c))
		calls = append(calls, c)
		q := c.Query
		var conns []string
		for _, f := range []string{"states(", "labels(", "members(", "projects("} {
			if strings.Contains(q, f) {
				conns = append(conns, strings.TrimSuffix(f, "("))
			}
		}
		out := func(s string) { _, _ = io.WriteString(w, s) }
		switch {
		case len(conns) > 1 || complexity(c) > complexityBudget:
			out(`{"errors":[{"message":"Query too complex"}]}`)
		case strings.Contains(q, "issueUpdate"):
			out(`{"data":{"issueUpdate":{"success":true}}}`)
		case strings.Contains(q, "issue(id"):
			out(`{"data":{"issue":{"id":"iss-parent","identifier":"JOB-4","title":"Parent","state":{"id":"s","name":"Todo"},"team":{"id":"team-1","key":"JOB"},"comments":{"nodes":[]}}}}`)
		case strings.Contains(q, "team(id") && len(conns) == 0:
			out(`{"data":{"team":{"id":"team-1","key":"JOB"}}}`)
		case len(conns) == 0:
			out(`{"errors":[{"message":"unexpected query"}]}`)
		case conns[0] == "states":
			out(`{"data":{"team":{"states":{"nodes":[{"id":"st-backlog","name":"Backlog","type":"backlog"}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}}`)
		case conns[0] == "labels" && c.Variables["after"] == "c1":
			out(`{"data":{"team":{"labels":{"nodes":[{"id":"lb-repo","name":"repo:github.com/gfx-labs/jobshop"}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}}`)
		case conns[0] == "labels":
			out(`{"data":{"team":{"labels":{"nodes":[{"id":"lb-proposed","name":"proposed"}],"pageInfo":{"hasNextPage":true,"endCursor":"c1"}}}}}`)
		default:
			out(`{"errors":[{"message":"unexpected query"}]}`)
		}
	}))
	t.Cleanup(srv.Close)
	return linear.New(srv.URL, apicall.Static("lin_api_x")), &calls
}

// TestComplexityModel: the fake rejects the query shape the real API
// rejected (#683), so the tests below would catch a return to it.
func TestComplexityModel(t *testing.T) {
	old := call{
		Query:     `query($team: String!, $first: Int!) { teams(filter: {key: {eq: $team}}) { nodes { states(first: $first) { nodes { id } } } } }`,
		Variables: map[string]any{"first": float64(100)},
	}
	assert.Greater(t, complexity(old), complexityBudget)
	now := call{
		Query:     `query($team: String!, $first: Int!) { team(id: $team) { states(first: $first) { nodes { id } } } }`,
		Variables: map[string]any{"first": float64(100)},
	}
	assert.LessOrEqual(t, complexity(now), complexityBudget)
}

// TestCreateQueryComplexity (#680, #683): create with a state, two
// labels (on two pages) and a parent resolves each list in its own
// small query, so a server with Linear's budget accepts every read.
func TestCreateQueryComplexity(t *testing.T) {
	c, calls := complexityServer(t)
	m, err := c.PlanCreate(t.Context(), linear.CreateParams{
		Team: "job", Title: "Proposed target", State: "Backlog",
		Labels: []string{"proposed", "repo:github.com/gfx-labs/jobshop"}, Parent: "JOB-4",
	}, "id-1")
	require.NoError(t, err)
	in, _ := m.Variables["input"].(map[string]any)
	assert.Equal(t, "team-1", in["teamId"])
	assert.Equal(t, "st-backlog", in["stateId"])
	assert.Equal(t, []string{"lb-proposed", "lb-repo"}, in["labelIds"])
	assert.Equal(t, "iss-parent", in["parentId"])
	for _, cl := range *calls {
		assert.NotContains(t, cl.Query, "members(", "no user lookup without an assignee")
		assert.NotContains(t, cl.Query, "teams(", "a team is read by key, not listed")
	}
	assert.Len(t, *calls, 5, "team id, states, two label pages, parent")
}

// TestSetStateQueryComplexity (#683): linear state reads the team's
// states through the same paged query.
func TestSetStateQueryComplexity(t *testing.T) {
	c, _ := complexityServer(t)
	st, err := c.SetState(t.Context(), "JOB-4", "backlog")
	require.NoError(t, err)
	assert.Equal(t, "st-backlog", st.ID)
}
