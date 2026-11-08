package control

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/control/linearsvc"
	"github.com/elee1766/jobshop/pkg/identity"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestLinearOutageWaits (JOB-118): while Linear hangs, a wait pass still
// clears the non-Linear waits, sends one request for the identity (the
// rest back off), and logs one WARN for the outage and no per-wait
// error. When Linear answers again after the backoff, the Linear waits
// resolve and the recovery is logged once.
func TestLinearOutageWaits(t *testing.T) {
	t.Parallel()
	var hang atomic.Bool
	var reqs atomic.Int32
	hang.Store(true)
	gql := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs.Add(1)
		// Read the body first: only then does the server see the client
		// give up and cancel r.Context().
		b, _ := io.ReadAll(r.Body)
		if hang.Load() {
			<-r.Context().Done() // until the client's request timeout
			return
		}
		var q struct {
			Variables struct {
				ID string `json:"id"`
			} `json:"variables"`
		}
		_ = json.Unmarshal(b, &q)
		_, _ = w.Write([]byte(`{"data":{"issue":{"id":"i","identifier":"` + q.Variables.ID + `","title":"t","description":"",` +
			`"state":{"name":"Done","type":"completed"},"team":{"key":"ENG"},"labels":{"nodes":[]},"children":{"nodes":[]},` +
			`"relations":{"nodes":[]},"inverseRelations":{"nodes":[]},"comments":{"nodes":[]}}}}`))
	}))
	t.Cleanup(gql.Close)

	c, clk, dir := scanControllerYAML(t, linearIntakeYAML)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "token"), []byte("lin_api_test\n"), 0o600))
	var logs lockedBuf
	c.log = slog.New(slog.NewTextHandler(&logs, nil))
	c.minter, c.linearURL = identity.NewMinter(), gql.URL
	c.linearSvc().SetRequestTimeout(100 * time.Millisecond)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	park := func(title string, cd store.WaitCond) store.Ticket {
		tk := mkTicket(t, c, store.NewTicket{Title: title, Pool: "dev"})
		_, err := c.waitAdd(ctx, admin, rpc.WaitAddParams{ID: tk.ID, Reason: title, Conds: []store.WaitCond{cd}})
		require.NoError(t, err)
		return tk
	}
	run := mkTicket(t, c, store.NewTicket{Kind: store.KindRun, Title: "dep", Fields: json.RawMessage(`{"workflow":"w","state":"done"}`)})
	st := park("state", store.WaitCond{Type: store.WaitLinearState, Thread: "ENG-1", State: "done"})
	lb := park("label", store.WaitCond{Type: store.WaitLinearLabel, Thread: "ENG-2", Label: "blocked", Absent: true})
	rn := park("run", store.WaitCond{Type: store.WaitRun, Run: run.ID, State: "done"})
	prs := &fakePRs{pr: map[string]string{}, reviews: map[string]string{}, checks: map[string]string{}, issues: map[string]string{}, runs: map[string]string{}}
	threads := &fakeThreads{}
	warns := func() int { return strings.Count(logs.String(), `msg="linear unavailable"`) }

	c.pollWaits(ctx, prs, threads)
	assert.Equal(t, []string{rn.ID}, readyIDs(t, c), "the run wait clears in the same pass")
	assert.Equal(t, int32(1), reqs.Load(), "the second Linear wait backs off instead of waiting out a timeout")
	assert.Equal(t, 1, warns())
	assert.NotContains(t, logs.String(), "waits: evaluate")

	c.pollWaits(ctx, prs, threads)
	assert.Equal(t, int32(1), reqs.Load(), "still backing off")
	assert.Equal(t, 1, warns(), "one WARN per outage, not per pass")

	hang.Store(false)
	clk.advance(linearsvc.BackoffFirst)
	c.pollWaits(ctx, prs, threads)
	assert.ElementsMatch(t, []string{rn.ID, st.ID, lb.ID}, readyIDs(t, c))
	assert.Equal(t, 1, warns())
	assert.Equal(t, 1, strings.Count(logs.String(), `msg="linear recovered"`))
}
