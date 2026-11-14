package control

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/integrations/identity"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// linearStartYAML is twoTrackersYAML with start on the signal and a skip
// label: the scan starts pr-work-linear itself (JOB-211).
var linearStartYAML = strings.Replace(twoTrackersYAML,
	"{signal: linear_issue_no_run, team: JOB, states: Todo, poll: 1m}",
	"{signal: linear_issue_no_run, team: JOB, states: Todo, skip_labels: Blocked, poll: 1m, start: pr-work-linear}", 1)

// TestLinearScanStart (JOB-211): an eligible Todo issue gets a
// pr-work-linear run keyed linear-<KEY> from the scan, with no triage run.
// Skip-labeled and already-running issues get none, and a second pass
// starts nothing new.
func TestLinearScanStart(t *testing.T) {
	t.Parallel()
	c, clk, dir := scanControllerYAML(t, linearStartYAML)
	ctx := t.Context()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "token"), []byte("lin_api_x\n"), 0o600))
	issue := func(key string, labels ...any) map[string]any {
		return map[string]any{"identifier": key, "title": "Issue " + key, "url": "u", "updatedAt": "2026-11-01T00:00:00Z", "state": map[string]any{"id": "s", "name": "Todo", "type": "unstarted"}, "assignee": nil, "labels": map[string]any{"nodes": labels}}
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var q struct {
			Query string `json:"query"`
		}
		require.NoError(t, json.Unmarshal(b, &q))
		if strings.Contains(q.Query, "issues(") {
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"issues": map[string]any{
				"nodes": []any{
					issue("JOB-1"),
					issue("JOB-2", map[string]any{"name": "Blocked"}),
					issue("JOB-3"),
				},
				"pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
			}}})
			return
		}
		_, _ = io.WriteString(w, `{"data":{}}`)
	}))
	t.Cleanup(srv.Close)
	c.minter, c.linearURL = identity.NewMinter(), srv.URL

	// The planner already started JOB-3 under the shared key.
	admin := scope{admin: true, actor: "admin"}
	pre, err := c.runStart(ctx, admin, rpc.RunStartParams{Workflow: "pr-work-linear", Vars: map[string]string{"issue": "JOB-3"}, Key: "linear-JOB-3", Force: true})
	require.NoError(t, err)

	runs := func() []domain.Ticket {
		rs, err := c.store.List(ctx, store.Filter{Kind: domain.KindRun})
		require.NoError(t, err)
		return rs
	}
	c.runTriggers(ctx, c.config())
	rs := runs()
	require.Len(t, rs, 2, "JOB-3's planner run and JOB-1's; JOB-2 is skipped")
	var got domain.Ticket
	for _, r := range rs {
		f, err := domain.DecodeFields[domain.RunFields](r)
		require.NoError(t, err)
		assert.Equal(t, "pr-work-linear", f.Workflow, "no triage run")
		if r.ID != pre.(rpc.RunResult).Run.ID {
			got = r
		}
	}
	f, err := domain.DecodeFields[domain.RunFields](got)
	require.NoError(t, err)
	assert.Equal(t, "linear-JOB-1", f.Key)
	assert.Equal(t, "JOB-1", metaString(got.Meta, "linear", "issue"))
	assert.Equal(t, "linear_issue_no_run:JOB-1", metaString(got.Meta, "trigger", "finding"))

	// Idempotent: the open runs hide both issues, so nothing starts again.
	c.runTriggers(ctx, c.config())
	clk.advance(2 * time.Minute)
	c.runTriggers(ctx, c.config())
	assert.Len(t, runs(), 2)

	// The same key started again returns the same run.
	again, err := c.runStart(ctx, admin, rpc.RunStartParams{Workflow: "pr-work-linear", Vars: map[string]string{"issue": "JOB-1"}, Key: "linear-JOB-1", Force: true})
	require.NoError(t, err)
	assert.Equal(t, got.ID, again.(rpc.RunResult).Run.ID)
}
