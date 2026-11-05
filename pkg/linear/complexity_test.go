package linear_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/linear"
)

// complexityServer answers like Linear for create's reads, but rejects
// a team query that names more than one connection with "Query too
// complex", as the real workspace did (#680). Labels come in two pages.
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
		case len(conns) > 1:
			out(`{"errors":[{"message":"Query too complex"}]}`)
		case strings.Contains(q, "issue(id"):
			out(`{"data":{"issue":{"id":"iss-parent","identifier":"JOB-4","title":"Parent","state":{"id":"s","name":"Todo"},"team":{"id":"team-1","key":"JOB"},"comments":{"nodes":[]}}}}`)
		case strings.Contains(q, "teams(filter") && len(conns) == 0:
			out(`{"data":{"teams":{"nodes":[{"id":"team-1","key":"JOB"}]}}}`)
		case conns[0] == "states":
			out(`{"data":{"teams":{"nodes":[{"states":{"nodes":[{"id":"st-backlog","name":"Backlog","type":"backlog"}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}]}}}`)
		case conns[0] == "labels" && c.Variables["after"] == "c1":
			out(`{"data":{"teams":{"nodes":[{"labels":{"nodes":[{"id":"lb-repo","name":"repo:github.com/elee1766/jobshop"}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}]}}}`)
		case conns[0] == "labels":
			out(`{"data":{"teams":{"nodes":[{"labels":{"nodes":[{"id":"lb-proposed","name":"proposed"}],"pageInfo":{"hasNextPage":true,"endCursor":"c1"}}}]}}}`)
		default:
			out(`{"errors":[{"message":"unexpected query"}]}`)
		}
	}))
	t.Cleanup(srv.Close)
	return linear.New(srv.URL, "lin_api_x"), &calls
}

// TestCreateQueryComplexity (#680): create with a state, two labels (on
// two pages) and a parent resolves each list in its own query, so a
// server that rejects combined team queries accepts every read.
func TestCreateQueryComplexity(t *testing.T) {
	c, calls := complexityServer(t)
	m, err := c.PlanCreate(t.Context(), linear.CreateParams{
		Team: "job", Title: "Proposed target", State: "Backlog",
		Labels: []string{"proposed", "repo:github.com/elee1766/jobshop"}, Parent: "JOB-4",
	}, "id-1")
	require.NoError(t, err)
	in, _ := m.Variables["input"].(map[string]any)
	assert.Equal(t, "team-1", in["teamId"])
	assert.Equal(t, "st-backlog", in["stateId"])
	assert.Equal(t, []string{"lb-proposed", "lb-repo"}, in["labelIds"])
	assert.Equal(t, "iss-parent", in["parentId"])
	for _, cl := range *calls {
		assert.NotContains(t, cl.Query, "members(", "no user lookup without an assignee")
	}
	assert.Len(t, *calls, 5, "team id, states, two label pages, parent")
}
