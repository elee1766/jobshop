package linear_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/linear"
)

// writeServer is a fake Linear GraphQL server that keeps created issues,
// comments and relations, so a test can count what was posted.
type writeServer struct {
	mu        sync.Mutex
	calls     []call
	mutations int
	created   map[string]bool // ids of created entities
	updatedAt string
	labels    []string
	state     string
	// dropReply makes the next mutation succeed but fail on the wire, as
	// a timeout after the write landed would.
	dropReply bool
}

const writeTeamResp = `{"data":{"team":{"id":"team-1","key":"ENG",
 "states":{"nodes":[{"id":"st-todo","name":"Todo","type":"unstarted"},{"id":"st-done","name":"Done","type":"completed"}]},
 "labels":{"nodes":[{"id":"lb-bug","name":"Bug"},{"id":"lb-ui","name":"UI"}]},
 "members":{"nodes":[{"id":"u-ann","name":"Ann Lee","displayName":"ann","email":"ann@example.com","url":"https://linear.app/acme/profiles/ann"}]}}}}`

func newWriteServer(t *testing.T) (*linear.Client, *writeServer) {
	t.Helper()
	ws := &writeServer{created: map[string]bool{}, updatedAt: "2026-11-04T01:00:00.000Z", labels: []string{"UI"}, state: "Todo"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var c call
		require.NoError(t, json.Unmarshal(b, &c))
		ws.mu.Lock()
		defer ws.mu.Unlock()
		ws.calls = append(ws.calls, c)
		q := c.Query
		if strings.HasPrefix(strings.TrimSpace(q), "mutation") {
			ws.mutations++
			if ws.dropReply {
				ws.dropReply = false
				ws.record(c)
				w.WriteHeader(http.StatusBadGateway)
				return
			}
		}
		_, _ = io.WriteString(w, ws.answer(c))
	}))
	t.Cleanup(srv.Close)
	return linear.New(srv.URL, "lin_api_x"), ws
}

func inputID(c call) string {
	in, _ := c.Variables["input"].(map[string]any)
	id, _ := in["id"].(string)
	return id
}

// record applies a mutation's effect.
func (ws *writeServer) record(c call) {
	if id := inputID(c); id != "" {
		ws.created[id] = true
	}
	if strings.Contains(c.Query, "issueUpdate") {
		in, _ := c.Variables["input"].(map[string]any)
		if s, ok := in["stateId"].(string); ok && s == "st-done" {
			ws.state = "Done"
		}
		for _, id := range anyStrings(in["addedLabelIds"]) {
			if id == "lb-bug" {
				ws.labels = append(ws.labels, "Bug")
			}
		}
		ws.updatedAt = "2026-11-04T02:00:00.000Z"
	}
}

func anyStrings(v any) []string {
	var out []string
	l, _ := v.([]any)
	for _, x := range l {
		s, _ := x.(string)
		out = append(out, s)
	}
	return out
}

func (ws *writeServer) answer(c call) string {
	q := c.Query
	switch {
	case strings.Contains(q, "issueCreate"):
		if ws.created[inputID(c)] {
			return `{"errors":[{"message":"Entity with this id already exists"}]}`
		}
		ws.record(c)
		return `{"data":{"issueCreate":{"success":true,"issue":{"id":"` + inputID(c) + `","identifier":"ENG-50","url":"https://linear.app/acme/issue/ENG-50"}}}}`
	case strings.Contains(q, "commentCreate"):
		ws.record(c)
		return `{"data":{"commentCreate":{"success":true,"comment":{"id":"` + inputID(c) + `","url":"https://linear.app/acme/issue/ENG-123#c"}}}}`
	case strings.Contains(q, "issueRelationCreate"):
		ws.record(c)
		return `{"data":{"issueRelationCreate":{"success":true,"issueRelation":{"id":"` + inputID(c) + `"}}}}`
	case strings.Contains(q, "issueUpdate"):
		ws.record(c)
		return `{"data":{"issueUpdate":{"success":true,"issue":{"id":"iss-1","identifier":"ENG-123","url":"u"}}}}`
	case strings.Contains(q, "team(id"):
		return writeTeamResp
	case strings.Contains(q, "comment(id"):
		id, _ := c.Variables["id"].(string)
		if ws.created[id] {
			return `{"data":{"comment":{"id":"` + id + `","url":"https://linear.app/acme/issue/ENG-123#c"}}}`
		}
		return notFound
	case strings.Contains(q, "updatedAt team"):
		return `{"data":{"issue":{"id":"iss-1","updatedAt":"` + ws.updatedAt + `","team":{"key":"ENG"}}}}`
	case strings.Contains(q, "archivedAt"): // context read back
		labels, _ := json.Marshal(ws.labels)
		var ls []string
		_ = json.Unmarshal(labels, &ls)
		nodes := make([]map[string]string, len(ls))
		for i, l := range ls {
			nodes[i] = map[string]string{"name": l}
		}
		nb, _ := json.Marshal(nodes)
		return `{"data":{"issue":{"id":"iss-1","identifier":"ENG-123","title":"Fix login","url":"u","state":{"id":"x","name":"` + ws.state + `","type":"started"},"team":{"id":"team-1","key":"ENG"},"labels":{"nodes":` + string(nb) + `},"children":{"nodes":[]},"relations":{"nodes":[]},"inverseRelations":{"nodes":[]},"comments":{"nodes":[]}}}}`
	case strings.Contains(q, "issue(id"):
		id, _ := c.Variables["id"].(string)
		if id == "ENG-9" {
			return `{"data":{"issue":{"id":"iss-9","identifier":"ENG-9","title":"Other","state":{"id":"s","name":"Todo"},"team":{"id":"team-1","key":"ENG"},"comments":{"nodes":[]}}}}`
		}
		if strings.HasPrefix(id, "ENG-") {
			return issueResp
		}
		if ws.created[id] {
			return `{"data":{"issue":{"id":"` + id + `","identifier":"ENG-50","url":"https://linear.app/acme/issue/ENG-50"}}}`
		}
		return notFound
	}
	return `{"errors":[{"message":"unexpected query"}]}`
}

// mapKeys is an in-memory KeyStore with the controller's semantics.
type mapKeys struct {
	mu   sync.Mutex
	rows map[string]*struct {
		op, result string
		done       bool
	}
}

func newMapKeys() *mapKeys {
	return &mapKeys{rows: map[string]*struct {
		op, result string
		done       bool
	}{}}
}

func (k *mapKeys) Claim(_ context.Context, key, op string) (string, bool, bool, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	r, ok := k.rows[key]
	if !ok {
		k.rows[key] = &struct {
			op, result string
			done       bool
		}{op: op}
		return "", false, false, nil
	}
	return r.result, r.done, !r.done, nil
}

func (k *mapKeys) Finish(_ context.Context, key, result string) (string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	r := k.rows[key]
	if !r.done {
		r.result, r.done = result, true
	}
	return r.result, nil
}

// TestWriteCreateRetryWithoutLocalKey: the local key was pruned (#415);
// the create with the same key-derived id is rejected as existing and
// Do returns the existing issue instead of failing or creating another.
func TestWriteCreateRetryWithoutLocalKey(t *testing.T) {
	c, ws := newWriteServer(t)
	ctx := t.Context()
	id := linear.KeyID("create-pruned")
	m, err := c.PlanCreate(ctx, linear.CreateParams{Team: "eng", Title: "New"}, id)
	require.NoError(t, err)
	_, err = c.Do(ctx, newMapKeys(), "create-pruned", false, m)
	require.NoError(t, err)
	res, err := c.Do(ctx, newMapKeys(), "create-pruned", false, m)
	require.NoError(t, err)
	assert.Equal(t, id, res.(linear.Result).ID)
	assert.True(t, res.(linear.Result).Replayed)
	assert.Len(t, ws.created, 1, "one entity")
}

func TestWriteRepeatedKeyPostsOnce(t *testing.T) {
	c, ws := newWriteServer(t)
	ctx := t.Context()
	keys := newMapKeys()
	id := linear.KeyID("create-1")
	m, err := c.PlanCreate(ctx, linear.CreateParams{Team: "eng", Title: "New", Labels: []string{"bug"}, Assignee: "ann", State: "todo"}, id)
	require.NoError(t, err)
	in := m.Variables["input"].(map[string]any)
	assert.Equal(t, "team-1", in["teamId"])
	assert.Equal(t, []string{"lb-bug"}, in["labelIds"])
	assert.Equal(t, "u-ann", in["assigneeId"])
	assert.Equal(t, "st-todo", in["stateId"])
	assert.Equal(t, id, in["id"])

	first, err := c.Do(ctx, keys, "create-1", false, m)
	require.NoError(t, err)
	assert.Equal(t, "ENG-50", first.(linear.Result).Identifier)
	again, err := c.Do(ctx, keys, "create-1", false, m)
	require.NoError(t, err)
	assert.Equal(t, 1, ws.mutations, "a repeated key posts once")
	assert.True(t, again.(linear.Result).Replayed)
	assert.Equal(t, first.(linear.Result).ID, again.(linear.Result).ID)
	assert.Equal(t, linear.KeyID("create-1"), id, "stable per key")
	assert.NotEqual(t, linear.KeyID("create-2"), id)
}

// A write that landed but whose reply was lost: the failed call reads the
// entity back by its key-derived id (#415), and a retry replays the
// stored result. Neither posts again.
func TestWriteUnknownOutcomeRetry(t *testing.T) {
	c, ws := newWriteServer(t)
	ctx := t.Context()
	keys := newMapKeys()
	id := linear.KeyID("ask-1")
	m, err := c.PlanAsk(ctx, "ENG-123", linear.AskParams{Body: "ready?"}, id)
	require.NoError(t, err)
	ws.dropReply = true
	first, err := c.Do(ctx, keys, "ask-1", false, m)
	require.NoError(t, err)
	assert.Equal(t, 1, ws.mutations)
	assert.True(t, first.(linear.Result).Replayed, "found by read back")

	res, err := c.Do(ctx, keys, "ask-1", false, m)
	require.NoError(t, err)
	assert.Equal(t, 1, ws.mutations, "retry found the comment instead of posting twice")
	assert.Equal(t, id, res.(linear.Result).ID)
	assert.True(t, res.(linear.Result).Replayed)
}

func TestWriteDryRunSendsNothing(t *testing.T) {
	c, ws := newWriteServer(t)
	ctx := t.Context()
	keys := newMapKeys()
	m, err := c.PlanRelate(ctx, "ENG-123", linear.RelBlocks, "ENG-9", linear.KeyID("r"))
	require.NoError(t, err)
	out, err := c.Do(ctx, keys, "r", true, m)
	require.NoError(t, err)
	assert.Equal(t, 0, ws.mutations, "dry run sends no mutation")
	assert.Empty(t, keys.rows, "dry run stores no key")
	got := out.(linear.Mutation)
	assert.Equal(t, linear.OpRelate, got.Op)
	in := got.Variables["input"].(map[string]any)
	assert.Equal(t, "iss-1", in["issueId"])
	assert.Equal(t, "iss-9", in["relatedIssueId"])
	assert.Equal(t, "blocks", in["type"])
	_, err = c.PlanRelate(ctx, "ENG-123", "duplicate", "ENG-9", "")
	require.Error(t, err)
}

func TestWriteUpdate(t *testing.T) {
	c, ws := newWriteServer(t)
	ctx := t.Context()
	stale, _ := time.Parse(time.RFC3339, "2026-11-03T00:00:00Z")
	_, err := c.PlanUpdate(ctx, "ENG-123", linear.UpdateParams{State: "Done", IfUpdatedAt: stale})
	require.ErrorIs(t, err, linear.ErrConflict, "if-updated-at conflict fails")
	assert.Equal(t, 0, ws.mutations)

	cur, _ := time.Parse(time.RFC3339, "2026-11-04T01:00:00Z")
	m, err := c.PlanUpdate(ctx, "ENG-123", linear.UpdateParams{State: "done", AddLabels: []string{"Bug"}, RemoveLabels: []string{"ui"}, IfUpdatedAt: cur})
	require.NoError(t, err)
	in := m.Variables["input"].(map[string]any)
	assert.Equal(t, []string{"lb-bug"}, in["addedLabelIds"])
	assert.Equal(t, []string{"lb-ui"}, in["removedLabelIds"])
	// The fake ignores removals, so the read back sees UI still set.
	_, err = c.Do(ctx, nil, "", false, m)
	require.ErrorContains(t, err, `label "ui" still set`)

	ws.updatedAt = "2026-11-04T01:00:00.000Z"
	m, err = c.PlanUpdate(ctx, "ENG-123", linear.UpdateParams{State: "Done", AddLabels: []string{"bug"}})
	require.NoError(t, err)
	res, err := c.Do(ctx, nil, "", false, m)
	require.NoError(t, err)
	assert.Equal(t, "ENG-123", res.(linear.Result).Identifier)
	_, err = c.PlanUpdate(ctx, "ENG-123", linear.UpdateParams{})
	require.Error(t, err)
}

func TestWriteAskMention(t *testing.T) {
	c, ws := newWriteServer(t)
	ctx := t.Context()
	m, err := c.PlanAsk(ctx, "ENG-123", linear.AskParams{Body: "Is this still needed?", Mention: []string{"@ann"}, ReplyTo: "c-2"}, "")
	require.NoError(t, err)
	in := m.Variables["input"].(map[string]any)
	assert.Equal(t, "https://linear.app/acme/profiles/ann Is this still needed?", in["body"], "mention is the profile url")
	assert.Equal(t, "c-2", in["parentId"])
	assert.Equal(t, "iss-1", in["issueId"])
	_, err = c.PlanAsk(ctx, "ENG-123", linear.AskParams{Body: "x", Mention: []string{"nobody"}}, "")
	require.ErrorContains(t, err, `no user "nobody"`)
	for _, cl := range ws.calls {
		assert.NotContains(t, cl.Query+toJSON(cl.Variables), "lin_api_x", "token never in a request body")
	}
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
