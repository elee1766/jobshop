// Package ingestqueue owns the ingest queue's coordination, split out of
// pkg/control (JOB-141): the open spool and its drain loop, the inbox
// change notifier and the `ingest.next` long poll, the admission
// pump's scheduling, external worker heartbeats, the admission hook
// runner and the external worker loop. Inbox and spool mechanics stay
// in pkg/ingest.
//
// The controller keeps every persisted decision: the authorized claim
// and admit RPCs, committing an admission (Settle's state, the hold
// ticket, the human's answer), dispatch of admitted rows and the scan
// and attention signals. It reaches this package through Host.
// See docs/notes/JOB-141-ingest-queue.md.
package ingestqueue

import (
	"context"
	"log/slog"
	"maps"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/clock"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Ingest queue (#335, #679, docs/notes/679-ingest-queue.md). Every
// source takes one path: a receiver appends to the spool (`ingest
// serve`, the controller's Linear listener, `ingest put`) or a poller
// appends to the inbox; the drain moves the spool into the inbox. Every
// inbox append advances the notifier and wakes the waits the new items
// name. The admission pump and external workers (`ingest.next`) block
// on the notifier. Timers stay as a reconcile backstop.

// DrainEvery is the drain timer. Receivers wake the drain on append: the
// controller's own listener directly, `ingest serve` over the admin
// socket (ingest.notify), so this only bounds the delay when the
// controller was unreachable at append time.
const DrainEvery = 5 * time.Second

// AdmissionIdle is how long the admission pump sleeps when nothing
// changed: the backstop for state the notifier does not see (a lease
// expiring, a hold answered).
const AdmissionIdle = time.Second

// NextMax caps one ingest.next wait below the CLI's 30s call deadline.
const NextMax = 25 * time.Second

// Host is what the queue needs from the controller.
type Host interface {
	Clock() clock.Clock
	// Store is the inbox's store.
	Store() ingest.InboxStore
	// SpoolPath is the configured spool file now.
	SpoolPath() string
	// WakeWaits wakes the waits whose conditions name a new item.
	WakeWaits(items []store.InboxItem)
	// Drained is called with the new items a drain moved, after the
	// notifier advanced: the Linear consumer reads them.
	Drained(items []store.InboxItem)
}

// Service is the queue state. The zero value is not usable; use New.
type Service struct {
	h   Host
	log *slog.Logger

	mu   sync.Mutex // one drain at a time; guards sp
	sp   *ingest.Spool
	wake chan struct{}

	inbox *ingest.Notifier

	wmu     sync.Mutex
	workers map[string]*rpc.IngestWorker
}

// New returns a Service on h.
func New(h Host, log *slog.Logger) *Service {
	return &Service{h: h, log: log, wake: make(chan struct{}, 1), inbox: ingest.NewNotifier(), workers: map[string]*rpc.IngestWorker{}}
}

// Spool opens the spool on first use and keeps it open.
func (s *Service) Spool(ctx context.Context) (*ingest.Spool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sp != nil {
		return s.sp, nil
	}
	sp, err := ingest.OpenSpool(ctx, s.h.SpoolPath())
	if err != nil {
		return nil, err
	}
	s.sp = sp
	return sp, nil
}

// spoolExists reports whether the spool is open or its file exists.
func (s *Service) spoolExists() bool {
	s.mu.Lock()
	open := s.sp != nil
	s.mu.Unlock()
	if open {
		return true
	}
	_, err := os.Stat(s.h.SpoolPath())
	return err == nil
}

// CloseSpool closes the spool if it was opened.
func (s *Service) CloseSpool() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sp != nil {
		_ = s.sp.Close()
		s.sp = nil
	}
}

// WakeDrain asks the drain loop for a pass now.
func (s *Service) WakeDrain() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

// Advance records an inbox change for the notifier's waiters.
func (s *Service) Advance() { s.inbox.Advance() }

// Changed is closed by the next inbox change. Take it before a pass, so
// a change during the pass is not missed.
func (s *Service) Changed() <-chan struct{} { return s.inbox.Changed() }

// Appended is the one hook after any inbox append, from any source
// (#679): it wakes the inbox's consumers and the waits whose conditions
// name a new item's subject. Items are the new ones only; dedupes do not
// call it.
func (s *Service) Appended(items []store.InboxItem) {
	if len(items) == 0 {
		return
	}
	s.inbox.Advance()
	s.h.WakeWaits(items)
}

// DrainResult is one drain's counts.
type DrainResult = ingest.DrainResult

// DrainOnce runs one drain until the spool is empty.
func (s *Service) DrainOnce(ctx context.Context) (DrainResult, error) {
	// No spool file means nothing was ever spooled. Opening it would
	// create and migrate it, which every test controller paid (#433).
	// Writers (the Linear listener, `ingest serve`) create it.
	if !s.spoolExists() {
		return DrainResult{}, nil
	}
	sp, err := s.Spool(ctx)
	if err != nil {
		return DrainResult{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.drain(ctx, sp)
}

// DrainFile drains the spool database at path, for `jobshop ingest
// drain --spool`.
func (s *Service) DrainFile(ctx context.Context, path string) (DrainResult, error) {
	sp, err := ingest.OpenSpool(ctx, path)
	if err != nil {
		return DrainResult{}, err
	}
	defer func() { _ = sp.Close() }()
	return s.drain(ctx, sp)
}

func (s *Service) drain(ctx context.Context, sp *ingest.Spool) (DrainResult, error) {
	res, err := ingest.Drain(ctx, sp, s.h.Store(), s.h.Clock().Now, s.log)
	if res.Moved > 0 || res.Bad > 0 {
		s.log.InfoContext(ctx, "ingest drain", "moved", res.Moved, "fresh", res.Fresh, "bad", res.Bad)
	}
	s.Appended(res.Items)
	s.h.Drained(res.Items)
	return res, err
}

// RunDrain drains on a timer and on wakeups until ctx ends. The caller
// closes the spool once every writer is done.
func (s *Service) RunDrain(ctx context.Context) {
	t := s.h.Clock().NewTicker(DrainEvery)
	defer t.Stop()
	for {
		if _, err := s.DrainOnce(ctx); err != nil && ctx.Err() == nil {
			s.log.WarnContext(ctx, "ingest drain", "err", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C():
		case <-s.wake:
		}
	}
}

// RunPass calls pass until ctx ends: once per inbox change, with
// AdmissionIdle as the backstop. pass is the controller's admit and
// dispatch pass, which does the store writes.
func (s *Service) RunPass(ctx context.Context, pass func(context.Context)) {
	for ctx.Err() == nil {
		changed := s.inbox.Changed()
		pass(ctx)
		select {
		case <-ctx.Done():
		case <-changed:
		case <-time.After(AdmissionIdle):
		}
	}
}

// Next is the inbox long poll (#679): it blocks until the inbox changed
// since after (an append or an admission), the timeout, or ctx ends.
// External workers run a pass after each return.
func (s *Service) Next(ctx context.Context, after int64, timeout time.Duration) rpc.IngestNextResult {
	if timeout <= 0 || timeout > NextMax {
		timeout = NextMax
	}
	wctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-s.h.Clock().After(timeout):
			cancel()
		case <-wctx.Done():
		}
	}()
	seq := s.inbox.Wait(after, wctx.Done())
	return rpc.IngestNextResult{Seq: seq, Changed: seq != after}
}

// PutPolled stores a poller's row. When the webhook copy came first
// under the same natural key, it keeps that one (deduped for the
// record) and adds the poller row the dispatcher delivers.
func (s *Service) PutPolled(ctx context.Context, it store.InboxItem) error {
	got, dup, err := ingest.InboxPut(ctx, s.h.Store(), it)
	if err == nil && !dup {
		s.Appended([]store.InboxItem{got})
	}
	if err != nil || !dup || got.DeliveryID == it.DeliveryID {
		return err
	}
	it.NaturalKey = ""
	got, dup, err = ingest.InboxPut(ctx, s.h.Store(), it)
	if err == nil && !dup {
		s.Appended([]store.InboxItem{got})
	}
	return err
}

// PutBatch stores activity poller rows and returns how it went; the
// fresh ones wake the consumers even when the batch failed part way.
func (s *Service) PutBatch(ctx context.Context, items []store.InboxItem) error {
	fresh, err := ingest.InboxPutBatch(ctx, s.h.Store(), items)
	s.Appended(fresh)
	return err
}

// Beat records an external worker's call. Heartbeats are kept in
// memory: after a controller restart the missing worker clock starts at
// the restart.
func (s *Service) Beat(w rpc.IngestClaimParams, claimed, admitted bool) {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	hb := s.workers[w.Worker]
	if hb == nil {
		hb = &rpc.IngestWorker{ID: w.Worker}
		s.workers[w.Worker] = hb
	}
	if w.Host != "" {
		hb.Host = w.Host
	}
	hb.Seen = s.h.Clock().Now()
	if claimed {
		hb.LastClaim = hb.Seen
	}
	if admitted {
		hb.Admitted++
	}
}

// Workers lists the heartbeats by worker id.
func (s *Service) Workers() []rpc.IngestWorker {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	out := make([]rpc.IngestWorker, 0, len(s.workers))
	for _, k := range slices.Sorted(maps.Keys(s.workers)) {
		out = append(out, *s.workers[k])
	}
	return out
}

// LastBeat is the newest heartbeat, or since when no worker has called.
func (s *Service) LastBeat(since time.Time) time.Time {
	last := since
	for _, w := range s.Workers() {
		if w.Seen.After(last) {
			last = w.Seen
		}
	}
	return last
}

// Live counts workers seen after cutoff.
func (s *Service) Live(cutoff time.Time) int64 {
	n := int64(0)
	for _, w := range s.Workers() {
		if w.Seen.After(cutoff) {
			n++
		}
	}
	return n
}
