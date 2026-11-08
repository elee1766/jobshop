package control

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/identity"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// lockedBuf is a log sink written from off-tick goroutines.
type lockedBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuf) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuf) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

// offTickController is a finding controller that runs calls outside the
// tick as Run sets it up (JOB-103), logging to log, with metrics.
func offTickController(t *testing.T, yaml string, log io.Writer) (*Controller, *scanClock, string) {
	t.Helper()
	c, clk, dir := scanControllerYAML(t, yaml)
	c.root.Dir = dir
	c.log = slog.New(slog.NewTextHandler(log, nil))
	m, err := telemetry.NewMetrics()
	require.NoError(t, err)
	c.metrics = m
	c.backgroundChecks = true
	c.offWake = make(chan struct{}, 1)
	t.Cleanup(c.wg.Wait)
	return c, clk, dir
}

// testTriggerPass runs the trigger step of a tick under the tick lock, as
// tick does.
func (c *Controller) testTriggerPass(t *testing.T) {
	t.Helper()
	unlock := c.lockTick(t.Context(), "tick")
	defer unlock()
	c.runTriggers(t.Context(), c.config())
}

// TestLinearIntakeOffTick (JOB-103): the linear_issue_no_run read runs
// outside the tick. With Linear blocked, the trigger pass returns and a
// reload can take place; once Linear answers, the replay fires the run
// once.
func TestLinearIntakeOffTick(t *testing.T) {
	t.Parallel()
	release := make(chan struct{})
	var reads atomic.Int32
	gql := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if strings.Contains(string(b), "issue(id:") {
			_, _ = w.Write([]byte(linearContextReply(b)))
			return
		}
		reads.Add(1)
		<-release
		_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[{"identifier":"ENG-1","title":"t","url":"u","updatedAt":"2026-11-01T00:00:00Z","state":{"id":"s","name":"Todo","type":"unstarted"},"assignee":null,"labels":{"nodes":[{"name":"agent"}]}}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	}))
	t.Cleanup(gql.Close)
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})
	c, _, dir := offTickController(t, linearIntakeYAML, io.Discard)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "token"), []byte("lin_api_test\n"), 0o600))
	c.minter, c.linearURL = identity.NewMinter(), gql.URL
	runs := func() int {
		rs, err := c.store.List(t.Context(), store.Filter{Kind: store.KindRun})
		require.NoError(t, err)
		return len(rs)
	}

	// Before JOB-103 this pass waited on Linear.
	c.testTriggerPass(t)
	assert.Equal(t, 0, runs())
	c.testTriggerPass(t)
	require.NoError(t, c.reload(), "reload while the read runs")

	close(release)
	c.wg.Wait() // the check job: read, replay, fire
	assert.Equal(t, 1, runs(), "the replay fires once the read is back")
	assert.Equal(t, int32(1), reads.Load(), "one read in flight, then the cache")
	c.testTriggerPass(t)
	assert.Equal(t, 1, runs())
	assert.Equal(t, int32(1), reads.Load())
}

const offTickCommandYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  dev: {agent: fake, dir: %[1]q}
triggers:
  slow-check:
    when: {check: "exec sleep 30", check_timeout: 100ms}
    ticket: {title: never, pool: dev}
  cmd:
    when: {open_below: 100}
    command: {run: "while [ ! -f go ]; do sleep 0.01; done; echo ran"}
`

// TestTriggerCommandOffTick (JOB-103): a trigger command runs outside the
// tick. The pass files its record and returns; the record holds the
// trigger while the command runs and is closed when it ends.
func TestTriggerCommandOffTick(t *testing.T) {
	t.Parallel()
	c, _, dir := offTickController(t, offTickCommandYAML, io.Discard)
	records := func() []store.Ticket { return triggerTickets(t, c, "cmd") }
	c.testTriggerPass(t)
	rs := records()
	require.Len(t, rs, 1, "the pass files the record without waiting")
	assert.Equal(t, store.StatusOpen, rs[0].Status)
	c.testTriggerPass(t)
	assert.Len(t, records(), 1, "the open record holds the trigger")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "go"), nil, 0o600))
	c.wg.Wait()
	rs = records()
	require.Len(t, rs, 1)
	assert.Equal(t, store.StatusClosed, rs[0].Status)
}

// TestOffTickTimeout (JOB-103): a check that runs out its timeout is
// logged and counted in jobshop.poller.timeouts by poller.
func TestOffTickTimeout(t *testing.T) {
	t.Parallel()
	var log lockedBuf
	c, _, _ := offTickController(t, offTickCommandYAML, &log)
	tr := c.config().Triggers["slow-check"]
	unlock := c.lockTick(t.Context(), "tick")
	c.triggerEngine().RunChecked(t.Context(), c.config(), "slow-check", false, func(ctx context.Context, cfg *config.Config) {
		c.triggerEngine().RunTrigger(ctx, cfg, "slow-check", tr)
	})
	unlock()
	c.wg.Wait()
	assert.Empty(t, triggerTickets(t, c, "slow-check"))
	assert.Contains(t, log.String(), "off-pass call exceeded its timeout")
	assert.Contains(t, log.String(), "poller=check")

	rec := httptest.NewRecorder()
	c.metrics.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	assert.Regexp(t, `jobshop_poller_timeouts_total\{[^}]*jobshop_poller="check"[^}]*\} 1`, rec.Body.String())
}
