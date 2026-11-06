package control

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	host "github.com/elee1766/acphost/pkg/acphost"
	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/journal"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Journal shipping (#530): every session's journal frames go to
// VictoriaLogs after a cursor, SessionFields.ShippedSeq. The shipper
// reads the journal files, not the live ACP stream, so a controller
// restart or a VictoriaLogs outage only delays rows: the next pass
// reads on from the cursor.

const (
	shipEvery       = time.Second
	shipBatchBytes  = 1 << 20
	shipPendingMax  = 4 << 20 // per session; reading pauses past it
	shipBackoffMax  = time.Minute
	shipSaveEvery   = 5 * time.Second
	shipRecentClose = 24 * time.Hour // closed sessions picked up at start
)

// shipper holds per-session shipping state. mu serializes passes: the
// pump's and journal.backfill's.
type shipper struct {
	mu       sync.Mutex
	sessions map[string]*shipSession
	started  bool
	failures int
	retryAt  time.Time
	lastErr  string
	// lagMu guards lag, so attention does not wait on a slow insert.
	lagMu sync.Mutex
	// lag is the oldest unshipped frame age per pool, from the last pass.
	lag    map[string]time.Duration
	lagErr string
}

type shipSession struct {
	t       store.Ticket
	mapper  *journal.Mapper
	off     int64 // journal bytes read
	shipped int64 // last seq VictoriaLogs accepted
	queued  int64 // last seq in pending or shipped
	saved   int64 // last seq written to the ticket
	savedAt time.Time
	pending []journal.Row
	bytes   int
	closed  bool // ticket closed: drop once caught up
	eof     bool // the last read reached the end of the file
}

func (s *shipSession) caughtUp() bool { return len(s.pending) == 0 && s.eof }

// vlClient is the configured VictoriaLogs, or nil.
func (c *Controller) vlClient(cfg *config.Config) *journal.Client {
	v := cfg.Journal.VictoriaLogs
	if v == nil {
		return nil
	}
	return &journal.Client{URL: v.URL, Headers: v.HeaderValues(os.Getenv)}
}

// pumpJournal ships journals once a second until the controller stops.
func (c *Controller) pumpJournal() {
	for {
		if cl := c.vlClient(c.config()); cl != nil {
			c.ship.mu.Lock()
			// Errors are logged and counted in sendBatches.
			_, _ = c.shipPass(c.ctx, cl, false)
			c.ship.mu.Unlock()
		}
		select {
		case <-c.ctx.Done():
			c.ship.mu.Lock()
			c.saveCursors(context.WithoutCancel(c.ctx), true)
			c.ship.mu.Unlock()
			return
		case <-c.after(shipEvery):
		}
	}
}

// shipPass reads new frames of tracked sessions, sends one or more
// batches and saves cursors. Callers hold c.ship.mu. With force it
// ignores the retry backoff (backfill, tests). It returns the frames
// shipped and the last insert error.
func (c *Controller) shipPass(ctx context.Context, cl *journal.Client, force bool) (int, error) {
	sh := &c.ship
	if sh.sessions == nil {
		sh.sessions = map[string]*shipSession{}
	}
	if err := c.trackSessions(ctx, !sh.started); err != nil {
		return 0, err
	}
	sh.started = true
	for _, s := range sh.sessions {
		c.readJournal(s)
	}
	n := 0
	var err error
	if force || !c.now().Before(sh.retryAt) {
		n, err = c.sendBatches(ctx, cl)
	}
	c.saveCursors(ctx, false)
	c.recordLag()
	for id, s := range sh.sessions {
		if s.closed && s.caughtUp() && s.saved == s.shipped {
			delete(sh.sessions, id)
		}
	}
	return n, err
}

// trackSessions adds open sessions, and at the first pass sessions closed
// in the last day, and marks closed ones.
func (c *Controller) trackSessions(ctx context.Context, first bool) error {
	sh := &c.ship
	open, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen, Limit: 1 << 20})
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, t := range open {
		seen[t.ID] = true
		if _, ok := sh.sessions[t.ID]; !ok {
			c.trackSession(t)
		}
	}
	for id, s := range sh.sessions {
		if !seen[id] && !s.closed {
			s.closed = true
		}
	}
	if first {
		closed, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusClosed, Order: store.OrderNewest, Limit: 1000})
		if err != nil {
			return err
		}
		for _, t := range closed {
			if c.since(t.ClosedAt) <= shipRecentClose {
				c.trackSession(t)
			}
		}
	}
	return nil
}

// trackSession starts shipping t from its cursor. The journal is read
// from the start anyway: the mapper needs the earlier prompts to tag
// later frames with their ticket.
func (c *Controller) trackSession(t store.Ticket) *shipSession {
	sh := &c.ship
	if s, ok := sh.sessions[t.ID]; ok {
		return s
	}
	f, _ := store.DecodeFields[store.SessionFields](t)
	s := &shipSession{
		t: t, shipped: f.ShippedSeq, queued: f.ShippedSeq, saved: f.ShippedSeq, savedAt: c.now(),
		closed: t.Status == store.StatusClosed,
		mapper: &journal.Mapper{
			Src:    journal.Source{Project: t.Project, Pool: t.Pool, Session: t.ID},
			Redact: c.redactor(journalEnv(c.root.SessionDir(t.ID))),
			Lookup: func(id string) journal.Work { return c.workOf(id) },
		},
	}
	sh.sessions[t.ID] = s
	return s
}

// workOf is a ticket's run, workflow and step, for row fields.
func (c *Controller) workOf(id string) journal.Work {
	ctx := context.WithoutCancel(c.ctx)
	t, err := c.store.Get(ctx, id)
	if err != nil || t.Kind != store.KindStep {
		return journal.Work{}
	}
	sf, err := store.DecodeFields[store.StepFields](t)
	if err != nil {
		return journal.Work{}
	}
	w := journal.Work{Run: sf.Run, Step: sf.Step}
	if run, err := c.store.Get(ctx, sf.Run); err == nil {
		if rf, err := store.DecodeFields[store.RunFields](run); err == nil {
			w.Workflow = rf.Workflow
		}
	}
	return w
}

// readJournal maps frames after the session's read offset into pending
// rows, up to shipPendingMax bytes.
func (c *Controller) readJournal(s *shipSession) {
	if s.bytes >= shipPendingMax {
		s.eof = false
		return
	}
	path := filepath.Join(c.root.SessionDir(s.t.ID), host.JournalFile)
	frames, off, err := host.ReadFrom(path, s.off)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) && c.ctx.Err() == nil {
			c.log.Warn("journal ship: read", "session", s.t.ID, "err", err)
		}
		// No journal (yet, or the dir was pruned): nothing to ship.
		s.eof = true
		return
	}
	s.eof = true
	for i, f := range frames {
		if f.Seq <= s.queued {
			s.mapper.Observe(f)
			continue
		}
		row, err := s.mapper.Map(f)
		if err != nil {
			continue
		}
		s.pending = append(s.pending, row)
		s.queued = f.Seq
		s.bytes += len(row.Line) + 1
		if s.bytes >= shipPendingMax && i < len(frames)-1 {
			// Stop here; the next read starts after this frame. ReadFrom
			// has no per-frame offsets, so re-read from the start and
			// skip by seq instead.
			s.off = 0
			s.mapper = &journal.Mapper{Src: s.mapper.Src, Redact: s.mapper.Redact, Lookup: s.mapper.Lookup}
			s.eof = false
			return
		}
	}
	s.off = off
}

// sendBatches posts pending rows in batches of about shipBatchBytes. A
// failed batch stays pending and sets the backoff.
func (c *Controller) sendBatches(ctx context.Context, cl *journal.Client) (int, error) {
	sh := &c.ship
	total := 0
	for {
		var body bytes.Buffer
		type part struct {
			s *shipSession
			n int
		}
		var parts []part
		for _, s := range sh.sessions {
			n := 0
			for _, r := range s.pending {
				if body.Len() > 0 && body.Len()+len(r.Line)+1 > shipBatchBytes {
					break
				}
				body.Write(r.Line)
				body.WriteByte('\n')
				n++
			}
			if n > 0 {
				parts = append(parts, part{s, n})
			}
			if body.Len() >= shipBatchBytes {
				break
			}
		}
		if len(parts) == 0 {
			sh.failures, sh.lastErr = 0, ""
			return total, nil
		}
		if err := cl.Insert(ctx, body.Bytes()); err != nil {
			sh.failures++
			sh.lastErr = err.Error()
			wait := min(shipEvery<<min(sh.failures, 7), shipBackoffMax)
			sh.retryAt = c.now().Add(wait)
			for _, p := range parts {
				c.metrics.JournalErrors.Add(ctx, 1, metric.WithAttributes(telemetry.AttrPool.String(p.s.t.Pool)))
			}
			if ctx.Err() == nil {
				c.log.WarnContext(ctx, "journal ship", "err", err, "retry_in", wait)
			}
			return total, err
		}
		for _, p := range parts {
			s := p.s
			sent := s.pending[:p.n]
			b := 0
			for _, r := range sent {
				b += len(r.Line) + 1
			}
			s.shipped = sent[len(sent)-1].Seq
			s.pending = s.pending[p.n:]
			s.bytes -= b
			attrs := metric.WithAttributes(telemetry.AttrPool.String(s.t.Pool))
			c.metrics.JournalFrames.Add(ctx, int64(p.n), attrs)
			c.metrics.JournalBytes.Add(ctx, int64(b), attrs)
			total += p.n
		}
	}
}

// saveCursors writes moved cursors to the session tickets, each at most
// every shipSaveEvery unless force or the session is closed.
func (c *Controller) saveCursors(ctx context.Context, force bool) {
	for _, s := range c.ship.sessions {
		if s.shipped == s.saved || (!force && !s.closed && c.since(s.savedAt) < shipSaveEvery) {
			continue
		}
		seq := s.shipped
		err := c.updateSessionFields(ctx, s.t.ID, func(f *store.SessionFields) { f.ShippedSeq = max(f.ShippedSeq, seq) })
		if err != nil {
			if ctx.Err() == nil {
				c.log.WarnContext(ctx, "journal ship: save cursor", "session", s.t.ID, "err", err)
			}
			continue
		}
		s.saved, s.savedAt = seq, c.now()
	}
}

// recordLag sets the lag gauge per pool: the age of the oldest pending
// row, 0 when every session is caught up.
func (c *Controller) recordLag() {
	lag := map[string]time.Duration{}
	for p := range c.config().Pools {
		lag[p] = 0
	}
	now := c.now()
	for _, s := range c.ship.sessions {
		if len(s.pending) == 0 {
			continue
		}
		age := max(now.Sub(time.UnixMilli(s.pending[0].T)), 0)
		lag[s.t.Pool] = max(lag[s.t.Pool], age)
	}
	for p, d := range lag {
		c.metrics.JournalLag.Record(c.ctx, d.Seconds(), metric.WithAttributes(telemetry.AttrPool.String(p)))
	}
	c.ship.lagMu.Lock()
	c.ship.lag, c.ship.lagErr = lag, c.ship.lastErr
	c.ship.lagMu.Unlock()
}

// journalLagAttention is attention journal-lag: a pool whose oldest
// unshipped frame is older than journal.lag_max.
func (c *Controller) journalLagAttention() []rpc.AttentionItem {
	cfg := c.config()
	if cfg.Journal.VictoriaLogs == nil {
		return nil
	}
	maxLag := cfg.Journal.LagMaxDuration()
	c.ship.lagMu.Lock()
	defer c.ship.lagMu.Unlock()
	var items []rpc.AttentionItem
	for p, d := range c.ship.lag {
		if d <= maxLag {
			continue
		}
		title := fmt.Sprintf("journal shipping for pool %s is %s behind (journal.lag_max %s)", p, d.Truncate(time.Second), maxLag)
		if c.ship.lagErr != "" {
			title += ": " + c.ship.lagErr
		}
		items = append(items, rpc.AttentionItem{Kind: rpc.AttentionJournalLag, ID: "journal-lag:" + p, Title: title, Since: c.now().Add(-d)})
	}
	return items
}
