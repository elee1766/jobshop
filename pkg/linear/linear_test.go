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

// Responses recorded from the Linear GraphQL API (ids shortened).
const (
	issueResp   = `{"data":{"issue":{"id":"iss-1","identifier":"ENG-123","title":"Fix login","description":"Steps...","url":"https://linear.app/acme/issue/ENG-123/fix-login","state":{"id":"st-todo","name":"Todo","type":"unstarted"},"team":{"id":"team-1","key":"ENG"},"comments":{"nodes":[{"body":"second","createdAt":"2026-11-02T10:00:00.000Z","user":{"name":"Ann"}},{"body":"first","createdAt":"2026-11-01T10:00:00.000Z","user":null}]}}}}`
	statesResp  = `{"data":{"team":{"states":{"nodes":[{"id":"st-todo","name":"Todo","type":"unstarted"},{"id":"st-review","name":"In Review","type":"started"},{"id":"st-done","name":"Done","type":"completed"}]}}}}`
	commentResp = `{"data":{"commentCreate":{"success":true}}}`
	updateResp  = `{"data":{"issueUpdate":{"success":true}}}`
	notFound    = `{"errors":[{"message":"Entity not found: Issue","extensions":{"type":"invalid input"}}],"data":null}`
)

type call struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func fake(t *testing.T) (*httptest.Server, *[]call) {
	t.Helper()
	var calls []call
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "lin_api_x", r.Header.Get("Authorization"))
		b, _ := io.ReadAll(r.Body)
		var c call
		require.NoError(t, json.Unmarshal(b, &c))
		calls = append(calls, c)
		switch {
		case strings.Contains(c.Query, "issue(id") && c.Variables["id"] == "ENG-999":
			_, _ = io.WriteString(w, notFound)
		case strings.Contains(c.Query, "issue(id"):
			_, _ = io.WriteString(w, issueResp)
		case strings.Contains(c.Query, "team(id"):
			_, _ = io.WriteString(w, statesResp)
		case strings.Contains(c.Query, "commentCreate"):
			_, _ = io.WriteString(w, commentResp)
		case strings.Contains(c.Query, "issueUpdate"):
			_, _ = io.WriteString(w, updateResp)
		default:
			t.Errorf("unexpected query %s", c.Query)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func TestIssueCommentState(t *testing.T) {
	srv, calls := fake(t)
	c := linear.New(srv.URL, "lin_api_x")
	ctx := t.Context()

	is, err := c.GetIssue(ctx, "ENG-123")
	require.NoError(t, err)
	assert.Equal(t, "Fix login", is.Title)
	assert.Equal(t, "ENG", is.Team.Key)
	require.Len(t, is.Comments, 2)
	assert.Equal(t, "first", is.Comments[0].Body, "oldest first")
	assert.Equal(t, "Ann", is.Comments[1].Author)

	require.NoError(t, c.AddComment(ctx, "ENG-123", "PR https://github.com/a/b/pull/1"))
	last := (*calls)[len(*calls)-1]
	assert.Equal(t, map[string]any{"issueId": "iss-1", "body": "PR https://github.com/a/b/pull/1"}, last.Variables["input"])

	st, err := c.SetState(ctx, "ENG-123", "done")
	require.NoError(t, err)
	assert.Equal(t, "Done", st.Name)
	last = (*calls)[len(*calls)-1]
	assert.Equal(t, "iss-1", last.Variables["id"])
	assert.Equal(t, map[string]any{"stateId": "st-done"}, last.Variables["input"])

	_, err = c.SetState(ctx, "ENG-123", "Shipped")
	require.ErrorContains(t, err, `no state "Shipped" (states: Todo, In Review, Done)`)
	_, err = c.GetIssue(ctx, "ENG-999")
	require.ErrorContains(t, err, "Entity not found")
}

func TestTeam(t *testing.T) {
	team, ok := linear.Team("ENG-123")
	assert.True(t, ok)
	assert.Equal(t, "ENG", team)
	for _, bad := range []string{"eng-1", "ENG-0", "ENG", "123", "a/b#1"} {
		_, ok := linear.Team(bad)
		assert.False(t, ok, bad)
	}
}
