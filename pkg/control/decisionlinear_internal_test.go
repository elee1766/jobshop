package control

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/integrations/identity"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// propLinear is a fake Linear holding a proposal JOB-1 with sub-issues
// JOB-2 (Backlog), JOB-3 (Done) and JOB-4 (Backlog). Issue ids equal keys.
type propLinear struct {
	mu       sync.Mutex
	state    map[string]string
	labels   map[string][]string
	comments map[string][]string
}

var propStates = []map[string]any{
	{"id": "s-backlog", "name": "Backlog", "type": "backlog"},
	{"id": "s-todo", "name": "Todo", "type": "unstarted"},
	{"id": "s-done", "name": "Done", "type": "completed"},
	{"id": "s-canceled", "name": "Canceled", "type": "canceled"},
}

func stateType(name string) string {
	for _, s := range propStates {
		if s["name"] == name {
			return s["type"].(string)
		}
	}
	return ""
}

func (f *propLinear) handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var q struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		require.NoError(t, json.Unmarshal(b, &q))
		f.mu.Lock()
		defer f.mu.Unlock()
		out := func(v any) { _ = json.NewEncoder(w).Encode(map[string]any{"data": v}) }
		id, _ := q.Variables["id"].(string)
		ref := func(k string) map[string]any {
			return map[string]any{"identifier": k, "title": k, "state": map[string]any{"name": f.state[k], "type": stateType(f.state[k])}}
		}
		switch {
		case strings.Contains(q.Query, "commentCreate"):
			in, _ := q.Variables["input"].(map[string]any)
			iid, _ := in["issueId"].(string)
			body, _ := in["body"].(string)
			f.comments[iid] = append(f.comments[iid], body)
			out(map[string]any{"commentCreate": map[string]any{"success": true}})
		case strings.Contains(q.Query, "issueUpdate"):
			in, _ := q.Variables["input"].(map[string]any)
			for _, s := range propStates {
				if s["id"] == in["stateId"] {
					f.state[id] = s["name"].(string)
				}
			}
			out(map[string]any{"issueUpdate": map[string]any{"success": true}})
		case strings.Contains(q.Query, "issueAddLabel"):
			l := strings.TrimPrefix(q.Variables["labelId"].(string), "lab-")
			if !slices.Contains(f.labels[id], l) {
				f.labels[id] = append(f.labels[id], l)
			}
			out(map[string]any{"issueAddLabel": map[string]any{"success": true}})
		case strings.Contains(q.Query, "issueRemoveLabel"):
			l := strings.TrimPrefix(q.Variables["labelId"].(string), "lab-")
			f.labels[id] = slices.DeleteFunc(f.labels[id], func(s string) bool { return s == l })
			out(map[string]any{"issueRemoveLabel": map[string]any{"success": true}})
		case strings.Contains(q.Query, "issueLabels("):
			out(map[string]any{"issueLabels": map[string]any{"nodes": []any{map[string]any{"id": "lab-" + q.Variables["name"].(string), "team": nil}}}})
		case strings.Contains(q.Query, "states(first"):
			out(map[string]any{"team": map[string]any{"states": map[string]any{"nodes": propStates, "pageInfo": map[string]any{"hasNextPage": false}}}})
		case strings.Contains(q.Query, "labels { nodes { id name } }"):
			var ns []any
			for _, l := range f.labels[id] {
				ns = append(ns, map[string]any{"id": "lab-" + l, "name": l})
			}
			out(map[string]any{"issue": map[string]any{"id": id, "labels": map[string]any{"nodes": ns}}})
		case strings.Contains(q.Query, "children("):
			var ch []any
			if id == "JOB-1" {
				ch = []any{ref("JOB-2"), ref("JOB-3"), ref("JOB-4")}
			}
			out(map[string]any{"issue": map[string]any{
				"id": id, "identifier": id, "title": id, "state": map[string]any{"name": f.state[id]}, "team": map[string]any{"id": "t", "key": "JOB"},
				"labels": map[string]any{"nodes": []any{}}, "children": map[string]any{"nodes": ch},
				"relations": map[string]any{"nodes": []any{}}, "inverseRelations": map[string]any{"nodes": []any{}}, "comments": map[string]any{"nodes": []any{}},
			}})
		default:
			out(map[string]any{"issue": map[string]any{"id": id, "identifier": id, "title": id, "state": map[string]any{"name": f.state[id]}, "team": map[string]any{"id": "t", "key": "JOB"}, "comments": map[string]any{"nodes": []any{}}}})
		}
	}
}

func propController(t *testing.T) (*Controller, *propLinear) {
	t.Helper()
	c, _, _ := subController(t)
	tokFile := filepath.Join(t.TempDir(), "tok")
	require.NoError(t, os.WriteFile(tokFile, []byte("lin_api_x\n"), 0o600))
	cfg := *c.config()
	cfg.Identities = map[string]config.Identity{"lin": {Kind: config.IdentityLinear, TokenFile: config.SecretPath(tokFile), Workspace: "acme", Teams: []string{"JOB"}}}
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	c.minter = identity.NewMinter()
	m, err := telemetry.NewMetrics()
	require.NoError(t, err)
	c.metrics = m
	f := &propLinear{
		state:    map[string]string{"JOB-1": "Backlog", "JOB-2": "Backlog", "JOB-3": "Done", "JOB-4": "Backlog"},
		labels:   map[string][]string{"JOB-1": {proposedLabel}},
		comments: map[string][]string{},
	}
	srv := httptest.NewServer(f.handler(t))
	t.Cleanup(srv.Close)
	c.linearURL = srv.URL
	return c, f
}

// TestDecisionApproveLinear (JOB-17): approving a Linear proposal
// relabels it proposed -> target and moves its open sub-issues to Todo.
func TestDecisionApproveLinear(t *testing.T) {
	t.Parallel()
	c, f := propController(t)
	d := mkTicket(t, c, store.NewTicket{Title: "Proposed target: v1 (JOB-1)", Pool: "dev", Labels: []string{proposedLabel}, Meta: json.RawMessage(`{"github":{"repo":"acme/app"},"linear":{"issue":"JOB-1"}}`)})
	out, err := c.decisionResolve(t.Context(), scope{admin: true, actor: "desk:alice"}, rpc.DecisionResolveParams{ID: d.ID, Action: rpc.DecisionApprove})
	require.NoError(t, err)
	assert.Equal(t, "approved", out.(rpc.DecisionResolveResult).Ticket.Resolution)
	f.mu.Lock()
	defer f.mu.Unlock()
	assert.Equal(t, []string{targetLabel}, f.labels["JOB-1"])
	assert.Equal(t, map[string]string{"JOB-1": "Backlog", "JOB-2": "Todo", "JOB-3": "Done", "JOB-4": "Todo"}, f.state)
	assert.Equal(t, []string{"Approved as target by desk:alice."}, f.comments["JOB-1"])
}

// TestDecisionRejectLinear: rejecting cancels the proposal issue.
func TestDecisionRejectLinear(t *testing.T) {
	t.Parallel()
	c, f := propController(t)
	d := mkTicket(t, c, store.NewTicket{Title: "Proposed target: v1 (JOB-1)", Pool: "dev", Labels: []string{proposedLabel}, Meta: json.RawMessage(`{"linear":{"issue":"JOB-1"}}`)})
	_, err := c.decisionResolve(t.Context(), scope{admin: true, actor: "bob"}, rpc.DecisionResolveParams{ID: d.ID, Action: rpc.DecisionReject})
	require.NoError(t, err)
	f.mu.Lock()
	defer f.mu.Unlock()
	assert.Equal(t, "Canceled", f.state["JOB-1"])
	assert.Equal(t, "Backlog", f.state["JOB-2"])
	assert.Equal(t, []string{"Rejected by bob."}, f.comments["JOB-1"])
}

func TestContainsKey(t *testing.T) {
	t.Parallel()
	assert.True(t, containsKey("waiting on job-1 approval", "JOB-1"))
	assert.False(t, containsKey("waiting on JOB-12", "JOB-1"))
	assert.True(t, containsKey("JOB-12 and JOB-1", "JOB-1"))
}
