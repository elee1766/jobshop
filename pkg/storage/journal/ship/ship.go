// Package ship ships session journal frames to VictoriaLogs
// (#530), split out of pkg/control (JOB-119). Every session's frames go
// after a cursor, SessionFields.ShippedSeq. The shipper reads the journal
// files, not the live ACP stream, so a controller restart or a
// VictoriaLogs outage only delays rows: the next pass reads on from the
// cursor.
//
// The controller owns tickets, config, redaction and the clock; it hands
// them in through Host. Access checks stay in the controller's RPC
// handlers.
package ship

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	host "github.com/elee1766/acphost/pkg/acphost"
	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/storage/journal"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/support/redact"
	"github.com/elee1766/jobshop/pkg/support/telemetry"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

const (
	shipEvery       = time.Second
	shipBatchBytes  = 1 << 20
	shipPendingMax  = 4 << 20 // per session; reading pauses past it
	shipBackoffMax  = time.Minute
	shipSaveEvery   = 5 * time.Second
	shipRecentClose = 24 * time.Hour // closed sessions picked up at start
)

// Host is what the shipper needs from the controller.
type Host interface {
	// Config is the active config: journal.victorialogs and the pools.
	Config() *config.Config
	Now() time.Time
	After(d time.Duration) <-chan time.Time
	// List lists tickets (sessions, open and closed).
	List(ctx context.Context, f store.Filter) ([]domain.Ticket, error)
	// SessionDir is where a session's journal.ndjson lives.
	SessionDir(id string) string
	// Redactor redacts a session's rows with its env's secret values.
	Redactor(session string) *redact.Redactor
	// Work is a ticket's run, workflow and step, for row fields.
	Work(ticket string) journal.Work
	// SaveCursor raises the session's ShippedSeq to seq.
	SaveCursor(ctx context.Context, session string, seq int64) error
}

// Shipper holds per-session shipping state. mu serializes passes: the
// pump's and journal.backfill's.
type Shipper struct {
	host    Host
	log     *slog.Logger
	metrics *telemetry.Metrics

	mu       sync.Mutex
	sessions map[string]*session
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

// New returns a shipper with nothing tracked; the first pass picks up
// open sessions and those closed in the last day.
func New(h Host, log *slog.Logger, m *telemetry.Metrics) *Shipper {
	return &Shipper{host: h, log: log, metrics: m, sessions: map[string]*session{}}
}

type session struct {
	t       domain.Ticket
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

func (s *session) caughtUp() bool { return len(s.pending) == 0 && s.eof }

// Client is the configured VictoriaLogs, or nil.
func Client(cfg *config.Config) *journal.Client {
	v := cfg.Journal.VictoriaLogs
	if v == nil {
		return nil
	}
	return &journal.Client{URL: v.URL, Headers: v.HeaderValues(os.Getenv)}
}

func (sh *Shipper) since(t time.Time) time.Duration { return sh.host.Now().Sub(t) }

// Run ships journals once a second until ctx ends, then saves cursors.
func (sh *Shipper) Run(ctx context.Context) {
	for {
		if cl := Client(sh.host.Config()); cl != nil {
			// Errors are logged and counted in sendBatches.
			_, _ = sh.Pass(ctx, cl, false)
		}
		select {
		case <-ctx.Done():
			sh.mu.Lock()
			sh.saveCursors(context.WithoutCancel(ctx), true)
			sh.mu.Unlock()
			return
		case <-sh.host.After(shipEvery):
		}
	}
}

// Pass reads new frames of tracked sessions, sends one or more batches
// and saves cursors. With force it ignores the retry backoff (backfill,
// tests). It returns the frames shipped and the last insert error.
func (sh *Shipper) Pass(ctx context.Context, cl *journal.Client, force bool) (int, error) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	return sh.pass(ctx, cl, force)
}

// pass is Pass under sh.mu.
func (sh *Shipper) pass(ctx context.Context, cl *journal.Client, force bool) (int, error) {
	if err := sh.trackSessions(ctx, !sh.started); err != nil {
		return 0, err
	}
	sh.started = true
	for _, s := range sh.sessions {
		sh.readJournal(ctx, s)
	}
	n := 0
	var err error
	if force || !sh.host.Now().Before(sh.retryAt) {
		n, err = sh.sendBatches(ctx, cl)
	}
	sh.saveCursors(ctx, false)
	sh.recordLag(ctx)
	sh.dropDone()
	return n, err
}

// dropDone forgets closed sessions that are shipped and saved.
func (sh *Shipper) dropDone() {
	for id, s := range sh.sessions {
		if s.closed && s.caughtUp() && s.saved == s.shipped {
			delete(sh.sessions, id)
		}
	}
}

// Backfill ships every tracked journal from its cursor now, adding
// sessions closed within since (all closed sessions when zero), and
// flushes VictoriaLogs so the rows are searchable when it returns.
func (sh *Shipper) Backfill(ctx context.Context, cl *journal.Client, since time.Duration) (rpc.JournalBackfillResult, error) {
	res := rpc.JournalBackfillResult{}
	closed, err := sh.host.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusClosed, Order: store.OrderNewest, Limit: 1 << 20})
	if err != nil {
		return res, err
	}
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if err := sh.trackSessions(ctx, !sh.started); err != nil {
		return res, err
	}
	sh.started = true
	for _, t := range closed {
		if since > 0 && sh.since(t.ClosedAt) > since {
			continue
		}
		sh.trackSession(t)
	}
	before := map[*session]int64{}
	for _, s := range sh.sessions {
		before[s] = s.shipped
	}
	for {
		n, err := sh.pass(ctx, cl, true)
		res.Frames += n
		if err != nil {
			return rpc.JournalBackfillResult{}, err
		}
		// A session reading in chunks has more after a pass that shipped.
		more := false
		for _, s := range sh.sessions {
			if !s.caughtUp() {
				more = true
			}
		}
		if !more || n == 0 {
			break
		}
	}
	sh.saveCursors(ctx, true)
	sh.dropDone()
	for s, seq := range before {
		if s.shipped > seq {
			res.Sessions++
		}
	}
	if err := cl.Flush(ctx); err != nil {
		return rpc.JournalBackfillResult{}, err
	}
	return res, nil
}

// trackSessions adds open sessions, and at the first pass sessions closed
// in the last day, and marks closed ones.
func (sh *Shipper) trackSessions(ctx context.Context, first bool) error {
	open, err := sh.host.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusOpen, Limit: 1 << 20})
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, t := range open {
		seen[t.ID] = true
		if _, ok := sh.sessions[t.ID]; !ok {
			sh.trackSession(t)
		}
	}
	for id, s := range sh.sessions {
		if !seen[id] && !s.closed {
			s.closed = true
		}
	}
	if first {
		closed, err := sh.host.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusClosed, Order: store.OrderNewest, Limit: 1000})
		if err != nil {
			return err
		}
		for _, t := range closed {
			if sh.since(t.ClosedAt) <= shipRecentClose {
				sh.trackSession(t)
			}
		}
	}
	return nil
}

// trackSession starts shipping t from its cursor. The journal is read
// from the start anyway: the mapper needs the earlier prompts to tag
// later frames with their ticket.
func (sh *Shipper) trackSession(t domain.Ticket) {
	if _, ok := sh.sessions[t.ID]; ok {
		return
	}
	f, _ := domain.DecodeFields[domain.SessionFields](t)
	sh.sessions[t.ID] = &session{
		t: t, shipped: f.ShippedSeq, queued: f.ShippedSeq, saved: f.ShippedSeq, savedAt: sh.host.Now(),
		closed: t.Status == domain.StatusClosed,
		mapper: &journal.Mapper{
			Src:    journal.Source{Project: t.Project, Pool: t.Pool, Session: t.ID},
			Redact: sh.host.Redactor(t.ID),
			Lookup: sh.host.Work,
		},
	}
}

// readJournal maps frames after the session's read offset into pending
// rows, up to shipPendingMax bytes.
func (sh *Shipper) readJournal(ctx context.Context, s *session) {
	if s.bytes >= shipPendingMax {
		s.eof = false
		return
	}
	path := filepath.Join(sh.host.SessionDir(s.t.ID), host.JournalFile)
	frames, off, err := host.ReadFrom(path, s.off)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) && ctx.Err() == nil {
			sh.log.WarnContext(ctx, "journal ship: read", "session", s.t.ID, "err", err)
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
func (sh *Shipper) sendBatches(ctx context.Context, cl *journal.Client) (int, error) {
	total := 0
	for {
		var body bytes.Buffer
		type part struct {
			s *session
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
			sh.retryAt = sh.host.Now().Add(wait)
			for _, p := range parts {
				sh.metrics.JournalErrors.Add(ctx, 1, metric.WithAttributes(telemetry.AttrPool.String(p.s.t.Pool)))
			}
			if ctx.Err() == nil {
				sh.log.WarnContext(ctx, "journal ship", "err", err, "retry_in", wait)
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
			sh.metrics.JournalFrames.Add(ctx, int64(p.n), attrs)
			sh.metrics.JournalBytes.Add(ctx, int64(b), attrs)
			total += p.n
		}
	}
}

// saveCursors writes moved cursors to the session tickets, each at most
// every shipSaveEvery unless force or the session is closed.
func (sh *Shipper) saveCursors(ctx context.Context, force bool) {
	for _, s := range sh.sessions {
		if s.shipped == s.saved || (!force && !s.closed && sh.since(s.savedAt) < shipSaveEvery) {
			continue
		}
		seq := s.shipped
		if err := sh.host.SaveCursor(ctx, s.t.ID, seq); err != nil {
			if ctx.Err() == nil {
				sh.log.WarnContext(ctx, "journal ship: save cursor", "session", s.t.ID, "err", err)
			}
			continue
		}
		s.saved, s.savedAt = seq, sh.host.Now()
	}
}

// recordLag sets the lag gauge per pool: the age of the oldest pending
// row, 0 when every session is caught up.
func (sh *Shipper) recordLag(ctx context.Context) {
	lag := map[string]time.Duration{}
	for p := range sh.host.Config().Pools {
		lag[p] = 0
	}
	now := sh.host.Now()
	for _, s := range sh.sessions {
		if len(s.pending) == 0 {
			continue
		}
		age := max(now.Sub(time.UnixMilli(s.pending[0].T)), 0)
		lag[s.t.Pool] = max(lag[s.t.Pool], age)
	}
	for p, d := range lag {
		sh.metrics.JournalLag.Record(ctx, d.Seconds(), metric.WithAttributes(telemetry.AttrPool.String(p)))
	}
	sh.lagMu.Lock()
	sh.lag, sh.lagErr = lag, sh.lastErr
	sh.lagMu.Unlock()
}

// LagAttention is attention journal-lag: a pool whose oldest unshipped
// frame is older than journal.lag_max.
func (sh *Shipper) LagAttention(cfg *config.Config) []rpc.AttentionItem {
	if cfg.Journal.VictoriaLogs == nil {
		return nil
	}
	maxLag := cfg.Journal.LagMaxDuration()
	sh.lagMu.Lock()
	defer sh.lagMu.Unlock()
	var items []rpc.AttentionItem
	for p, d := range sh.lag {
		if d <= maxLag {
			continue
		}
		title := fmt.Sprintf("journal shipping for pool %s is %s behind (journal.lag_max %s)", p, d.Truncate(time.Second), maxLag)
		if sh.lagErr != "" {
			title += ": " + sh.lagErr
		}
		items = append(items, rpc.AttentionItem{Kind: rpc.AttentionJournalLag, ID: "journal-lag:" + p, Title: title, Since: sh.host.Now().Add(-d)})
	}
	return items
}
