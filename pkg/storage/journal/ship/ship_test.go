package ship_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/journal"
	"github.com/gfx-labs/jobshop/pkg/storage/journal/ship"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/support/redact"
	"github.com/gfx-labs/jobshop/pkg/support/telemetry"
)

// fakeHost is a ship.Host over fixed tickets and a temp dir.
type fakeHost struct {
	cfg  *config.Config
	dir  string
	now  time.Time
	open []domain.Ticket

	mu      sync.Mutex
	cursors map[string]int64
}

func (h *fakeHost) Config() *config.Config               { return h.cfg }
func (h *fakeHost) Now() time.Time                       { return h.now }
func (h *fakeHost) After(time.Duration) <-chan time.Time { return make(chan time.Time) }
func (h *fakeHost) SessionDir(id string) string          { return filepath.Join(h.dir, id) }
func (h *fakeHost) Redactor(string) *redact.Redactor     { return nil }
func (h *fakeHost) Work(ticket string) journal.Work      { return journal.Work{Run: "run-" + ticket} }

func (h *fakeHost) cursor(id string) int64  { h.mu.Lock(); defer h.mu.Unlock(); return h.cursors[id] }
func (h *fakeHost) setNow(t time.Time)      { h.now = t }
func (h *fakeHost) addOpen(t domain.Ticket) { h.open = append(h.open, t) }
func (h *fakeHost) sessionTicket(id, pool string) domain.Ticket {
	return domain.Ticket{ID: id, Project: "js", Pool: pool, Kind: domain.KindSession, Status: domain.StatusOpen}
}

func (h *fakeHost) List(_ context.Context, f store.Filter) ([]domain.Ticket, error) {
	if f.Status == domain.StatusOpen {
		return h.open, nil
	}
	return nil, nil
}

func (h *fakeHost) SaveCursor(_ context.Context, session string, seq int64) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cursors[session] = max(h.cursors[session], seq)
	return nil
}

// fakeVL is a VictoriaLogs insert endpoint that keeps rows, or fails.
type fakeVL struct {
	mu   sync.Mutex
	rows []map[string]any
	fail bool
}

func (v *fakeVL) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.fail {
		http.Error(w, "down", http.StatusServiceUnavailable)
		return
	}
	b, _ := io.ReadAll(r.Body)
	for line := range strings.SplitSeq(strings.TrimSpace(string(b)), "\n") {
		var row map[string]any
		if json.Unmarshal([]byte(line), &row) == nil {
			v.rows = append(v.rows, row)
		}
	}
}

func (v *fakeVL) count() int { v.mu.Lock(); defer v.mu.Unlock(); return len(v.rows) }

func (v *fakeVL) setFail(fail bool) { v.mu.Lock(); defer v.mu.Unlock(); v.fail = fail }

func writeFrames(t *testing.T, dir string, n int) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o700))
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, `{"seq":%d,"t":%d,"from":"agent","msg":{"jsonrpc":"2.0","method":"session/update","params":{"update":{"sessionUpdate":"agent_message_chunk"}}}}`+"\n", i, 1000+i)
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, host.JournalFile), []byte(b.String()), 0o600))
}

func setup(t *testing.T) (*fakeHost, *fakeVL, *journal.Client, *ship.Shipper) {
	t.Helper()
	vl := &fakeVL{}
	srv := httptest.NewServer(vl)
	t.Cleanup(srv.Close)
	cfg := &config.Config{
		Pools:   map[string]config.Pool{"dev": {}},
		Journal: config.Journal{VictoriaLogs: &config.VictoriaLogs{URL: srv.URL}, LagMax: "1m"},
	}
	h := &fakeHost{cfg: cfg, dir: t.TempDir(), now: time.UnixMilli(1000), cursors: map[string]int64{}}
	m, err := telemetry.NewMetrics()
	require.NoError(t, err)
	t.Cleanup(func() { _ = m.Shutdown(context.Background()) })
	cl := ship.Client(cfg)
	require.NotNil(t, cl)
	return h, vl, cl, ship.New(h, slog.New(slog.DiscardHandler), m)
}

// TestPassShipsFromCursor: a pass ships frames after the session's
// ShippedSeq and saves the new cursor; the next pass ships only appends.
func TestPassShipsFromCursor(t *testing.T) {
	t.Parallel()
	h, vl, cl, sh := setup(t)
	s := h.sessionTicket("s1", "dev")
	s.Fields = json.RawMessage(`{"shipped_seq":2}`)
	h.addOpen(s)
	writeFrames(t, h.SessionDir("s1"), 5)

	n, err := sh.Pass(t.Context(), cl, false)
	require.NoError(t, err)
	assert.Equal(t, 3, n, "frames 3..5")
	assert.Equal(t, 3, vl.count())
	// The cursor waits shipSaveEvery for an open session.
	h.setNow(h.now.Add(time.Minute))
	_, err = sh.Pass(t.Context(), cl, false)
	require.NoError(t, err)
	assert.Equal(t, int64(5), h.cursor("s1"))

	writeFrames(t, h.SessionDir("s1"), 7)
	n, err = sh.Pass(t.Context(), cl, false)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Empty(t, sh.LagAttention(h.cfg))
}

// TestLagAttention: a failed insert keeps rows pending; once the oldest
// is older than lag_max the pool shows up, with the error.
func TestLagAttention(t *testing.T) {
	t.Parallel()
	h, vl, cl, sh := setup(t)
	h.addOpen(h.sessionTicket("s1", "dev"))
	writeFrames(t, h.SessionDir("s1"), 2)
	vl.setFail(true)
	h.setNow(time.UnixMilli(1001).Add(2 * time.Minute))

	_, err := sh.Pass(t.Context(), cl, false)
	require.Error(t, err)
	items := sh.LagAttention(h.cfg)
	require.Len(t, items, 1)
	assert.Equal(t, "journal-lag:dev", items[0].ID)
	assert.Contains(t, items[0].Title, "HTTP 503")

	vl.setFail(false)
	n, err := sh.Pass(t.Context(), cl, true)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Empty(t, sh.LagAttention(h.cfg))
}
