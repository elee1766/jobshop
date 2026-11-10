package compress_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	host "github.com/gfx-labs/acphost/pkg/acphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/journal/compress"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// fakeHost serves a fixed list of sessions from a temp dir and runs
// background work on a WaitGroup the test waits on.
type fakeHost struct {
	t        *testing.T
	root     string
	now      time.Time
	sessions []domain.Ticket
	wg       sync.WaitGroup
}

func (h *fakeHost) Now() time.Time              { return h.now }
func (h *fakeHost) SessionDir(id string) string { return filepath.Join(h.root, id) }
func (h *fakeHost) Go(f func())                 { h.wg.Go(f) }
func (h *fakeHost) Context() context.Context    { return h.t.Context() }
func (h *fakeHost) List(_ context.Context, f store.Filter) ([]domain.Ticket, error) {
	var out []domain.Ticket
	for _, s := range h.sessions {
		if s.Kind == f.Kind && s.Status == f.Status {
			out = append(out, s)
		}
	}
	return out, nil
}

func (h *fakeHost) add(id string, status domain.Status, closedAgo time.Duration) {
	h.t.Helper()
	s := domain.Ticket{ID: id, Kind: domain.KindSession, Status: status}
	if status == domain.StatusClosed {
		s.ClosedAt = h.now.Add(-closedAgo)
	}
	h.sessions = append(h.sessions, s)
	dir := h.SessionDir(id)
	require.NoError(h.t, os.MkdirAll(dir, 0o700))
	var b strings.Builder
	for i := 1; i <= 5; i++ {
		fmt.Fprintf(&b, `{"seq":%d,"t":%d,"from":"agent","msg":{"jsonrpc":"2.0","method":"session/update","params":{}}}`+"\n", i, 1000+i)
	}
	require.NoError(h.t, os.WriteFile(filepath.Join(dir, host.JournalFile), []byte(b.String()), 0o600))
}

func (h *fakeHost) plain(id string) string { return filepath.Join(h.SessionDir(id), host.JournalFile) }

// TestMaybe: a due sweep compresses only sessions closed before the
// cutoff, a second call inside the hour does nothing, and an empty After
// is off.
func TestMaybe(t *testing.T) {
	t.Parallel()
	h := &fakeHost{t: t, root: t.TempDir(), now: time.Date(2026, 11, 8, 8, 0, 0, 0, time.UTC)}
	h.add("old", domain.StatusClosed, 48*time.Hour)
	h.add("newer", domain.StatusClosed, time.Hour)
	h.add("open", domain.StatusOpen, 0)
	sw := compress.New(h, slog.New(slog.DiscardHandler))

	sw.Maybe(config.JournalCompress{})
	h.wg.Wait()
	assert.True(t, sw.LastSweep().IsZero(), "off")
	assert.FileExists(t, h.plain("old"))

	cfg := config.JournalCompress{After: "24h", Level: 3}
	sw.Maybe(cfg)
	h.wg.Wait()
	assert.False(t, sw.Running())
	assert.Equal(t, h.now, sw.LastSweep())
	assert.NoFileExists(t, h.plain("old"))
	assert.FileExists(t, filepath.Join(h.SessionDir("old"), host.CompressedJournalFile))
	assert.FileExists(t, h.plain("newer"), "closed less than after ago")
	assert.FileExists(t, h.plain("open"), "still open")

	// Inside the hour: no second sweep, even though newer is now due.
	h.now = h.now.Add(30 * time.Minute)
	h.sessions[1].ClosedAt = h.now.Add(-48 * time.Hour)
	sw.Maybe(cfg)
	h.wg.Wait()
	assert.FileExists(t, h.plain("newer"))

	h.now = h.now.Add(compress.Every)
	sw.Maybe(cfg)
	h.wg.Wait()
	assert.NoFileExists(t, h.plain("newer"))
}
