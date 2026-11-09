// Package journalcompress compresses closed sessions' journals (#631),
// split out of pkg/control (JOB-89). A closed session's journal.ndjson
// becomes journal.ndjson.zst journal.compress.after after it closed.
// Every reader goes through the acphost readers, which serve the .zst and
// its seq index the same way. See docs/notes/631-journal-compress.md.
//
// The controller hands in the clock, the session list and the place to
// run the sweep through Host.
package journalcompress

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"time"

	host "github.com/elee1766/acphost/pkg/acphost"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/store"
)

// Every spaces sweeps; the first runs on start.
const Every = time.Hour

// Host is what the sweeper needs from the controller.
type Host interface {
	Now() time.Time
	// List lists tickets (closed sessions).
	List(ctx context.Context, f store.Filter) ([]domain.Ticket, error)
	// SessionDir is where a session's journal.ndjson lives.
	SessionDir(id string) string
	// Go runs f in the background, tracked so the controller waits for it
	// on shutdown.
	Go(f func())
	// Context is the controller's, not the tick's: a sweep outlives the
	// tick.
	Context() context.Context
}

// Sweeper starts compression sweeps. Its state is tick-only except
// running, which the sweep goroutine clears.
type Sweeper struct {
	h       Host
	log     *slog.Logger
	at      time.Time // the last sweep start
	running atomic.Bool
}

// New returns a Sweeper on h.
func New(h Host, log *slog.Logger) *Sweeper { return &Sweeper{h: h, log: log} }

// LastSweep is when the last sweep started; zero before the first.
func (s *Sweeper) LastSweep() time.Time { return s.at }

// Running reports whether a sweep is in progress.
func (s *Sweeper) Running() bool { return s.running.Load() }

// Maybe starts a sweep in the background when one is due and none is
// running: a large journal takes seconds and the tick must not wait on
// it. Call it from the tick only.
func (s *Sweeper) Maybe(cfg config.JournalCompress) {
	if cfg.AfterDuration() <= 0 {
		return
	}
	now := s.h.Now()
	if !s.at.IsZero() && now.Sub(s.at) < Every {
		return
	}
	if !s.running.CompareAndSwap(false, true) {
		return
	}
	s.at = now
	cutoff := now.Add(-cfg.AfterDuration())
	ctx := s.h.Context()
	s.h.Go(func() {
		defer s.running.Store(false)
		if _, err := s.Sweep(ctx, cutoff, cfg.Level); err != nil && ctx.Err() == nil {
			s.log.ErrorContext(ctx, "journal compress", "err", err)
		}
	})
}

// Sweep compresses, one at a time, the plain journal of every session
// closed before cutoff. It returns how many it compressed.
func (s *Sweeper) Sweep(ctx context.Context, cutoff time.Time, level int) (int, error) {
	// The thread's IO priority is set below; keep this goroutine on it.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	lowIOPriority()
	closed, err := s.h.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusClosed})
	if err != nil {
		return 0, err
	}
	n := 0
	var errs []error
	for _, t := range closed {
		if ctx.Err() != nil {
			break
		}
		if t.ClosedAt.IsZero() || !t.ClosedAt.Before(cutoff) {
			continue
		}
		dir := s.h.SessionDir(t.ID)
		plain := filepath.Join(dir, host.JournalFile)
		st, err := os.Stat(plain)
		if err != nil {
			continue // none, or already compressed
		}
		if err := host.CompressJournal(dir, level); err != nil {
			errs = append(errs, err)
			continue
		}
		after := int64(-1)
		if zs, err := os.Stat(filepath.Join(dir, host.CompressedJournalFile)); err == nil {
			after = zs.Size()
		}
		s.log.InfoContext(ctx, "journal compressed", "session", t.ID, "bytes_before", st.Size(), "bytes_after", after)
		n++
	}
	return n, errors.Join(errs...)
}
