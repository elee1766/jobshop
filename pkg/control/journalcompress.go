package control

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"time"

	host "github.com/elee1766/acphost/pkg/acphost"

	"github.com/elee1766/jobshop/pkg/store"
)

// Journal compression (#631, docs/notes/631-journal-compress.md). A
// closed session's journal.ndjson is compressed to journal.ndjson.zst
// journal.compress.after after it closed. Every reader goes through the
// acphost readers, which serve the .zst and its seq index the same way.

// journalCompressEvery spaces sweeps; the first runs on start.
const journalCompressEvery = time.Hour

// maybeCompressJournals starts a sweep in the background when one is due
// and none is running: a large journal takes seconds and the tick must
// not wait on it.
func (c *Controller) maybeCompressJournals() {
	cfg := c.config().Journal.Compress
	if cfg.AfterDuration() <= 0 {
		return
	}
	now := c.now()
	if !c.journalsCompressedAt.IsZero() && now.Sub(c.journalsCompressedAt) < journalCompressEvery {
		return
	}
	if !c.journalCompressing.CompareAndSwap(false, true) {
		return
	}
	c.journalsCompressedAt = now
	cutoff := now.Add(-cfg.AfterDuration())
	c.wg.Go(func() {
		defer c.journalCompressing.Store(false)
		if _, err := c.compressJournals(c.ctx, cutoff, cfg.Level); err != nil && c.ctx.Err() == nil {
			c.log.Error("journal compress", "err", err)
		}
	})
}

// compressJournals compresses, one at a time, the plain journal of every
// session closed before cutoff. It returns how many it compressed.
func (c *Controller) compressJournals(ctx context.Context, cutoff time.Time, level int) (int, error) {
	// The thread's IO priority is set below; keep this goroutine on it.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	lowIOPriority()
	closed, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusClosed})
	if err != nil {
		return 0, err
	}
	n := 0
	var errs []error
	for _, s := range closed {
		if ctx.Err() != nil {
			break
		}
		if s.ClosedAt.IsZero() || !s.ClosedAt.Before(cutoff) {
			continue
		}
		dir := c.root.SessionDir(s.ID)
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
		c.log.InfoContext(ctx, "journal compressed", "session", s.ID, "bytes_before", st.Size(), "bytes_after", after)
		n++
	}
	return n, errors.Join(errs...)
}
