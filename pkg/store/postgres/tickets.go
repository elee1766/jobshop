package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math/rand/v2"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/elee1766/jobshop/pkg/ids"
	"github.com/elee1766/jobshop/pkg/store"
	sqldb "github.com/elee1766/jobshop/pkg/store/postgres/pgdb"
	"github.com/elee1766/jobshop/pkg/store/storerules"
)

// Store is the Postgres ticket store (JOB-41). It owns its
// transactions: writes run serializable, set per transaction, and are
// rerun on a serialization failure or deadlock. Reads outside a
// transaction run on a separate read-only pool at the server default
// (read committed), so Desk reads neither fail with 40001 nor compete
// with claims for connections. Claims lock their candidate with
// FOR UPDATE SKIP LOCKED, so concurrent claimers take different rows
// instead of colliding on one and retrying.
type Store struct {
	db *sql.DB
	q  sqldb.Querier
	// rdb and rq serve reads outside a transaction.
	rdb     *sql.DB
	rq      sqldb.Querier
	reg     *registry
	ids     *ids.Set
	idSpec  func(string) ids.Spec
	now     func() time.Time
	created storerules.Clock
	base    func() time.Time // the clock before Floor
	prefix  func(string) string
	commit  func()
	// retries counts transactions rerun after a lost race, for the
	// claim benchmark.
	retries atomic.Int64
}

var _ store.Store = (*Store)(nil)

// newStore returns a store over the write pool db and the read pool rdb.
// Close closes both.
func newStore(db, rdb *sql.DB, opt Options) (*Store, error) {
	s := &Store{
		db: db, q: sqldb.New(db), rdb: rdb, rq: sqldb.New(rdb), reg: &registry{},
		ids: &ids.Set{}, idSpec: opt.IDs, now: opt.Now, prefix: opt.Prefix, commit: opt.OnCommit,
	}
	if s.now == nil {
		s.now = time.Now
	}
	// Creation times are strictly increasing (JOB-58); other writes never
	// go below the last one.
	clock := s.now
	s.now = func() time.Time { return s.created.Floor(clock()) }
	s.base = clock
	if s.prefix == nil {
		s.prefix = func(p string) string { return p }
	}
	var last int64
	if err := db.QueryRow("SELECT COALESCE(MAX(created_at), 0) FROM tickets").Scan(&last); err != nil {
		return nil, fmt.Errorf("seed creation clock: %w", err)
	}
	s.created.Seed(last)
	return s, nil
}

// Close closes the database.
func (s *Store) Close() error {
	_ = s.rdb.Close()
	return s.db.Close()
}

// Retries is how many write transactions were rerun after losing a
// race since Open.
func (s *Store) Retries() int64 { return s.retries.Load() }

// busyTries bounds how often tx reruns a transaction that lost a write
// race (a serialization failure or deadlock).
const busyTries = 16

// serializable is the isolation of every write transaction. Serializable:
// the claim and sweep guards (#406) read other rows in the same
// statement as the write; under read committed a concurrent writer could
// pass them on a stale snapshot.
var serializable = &sql.TxOptions{Isolation: sql.LevelSerializable}

func (s *Store) tx(ctx context.Context, fn func(q sqldb.Querier) error) error {
	return s.sqlTx(ctx, func(tx *sql.Tx) error { return fn(sqldb.New(tx)) })
}

// sqlTx runs fn in a write transaction, rerunning it when it loses a
// write race. Collections (#579) use it with plain SQL.
func (s *Store) sqlTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	var err error
	for try := range busyTries {
		if err = s.txOnce(ctx, fn); !isBusy(err) {
			break
		}
		s.retries.Add(1)
		// Rerunning at once lets the same writers collide again in
		// lockstep until the tries run out (#527); a random, growing
		// wait spreads them.
		if try < busyTries-1 && !busyWait(ctx, try) {
			break
		}
	}
	// A lost write race means the caller's read is stale, as with a
	// revision mismatch, and retrying is safe.
	if isBusy(err) {
		return fmt.Errorf("%w: %w", store.ErrConflict, err)
	}
	return err
}

// busyWait sleeps a random time up to 1ms<<try (capped at 64ms). It
// reports false if ctx ends first.
func busyWait(ctx context.Context, try int) bool {
	d := time.Duration(rand.Int64N(int64(time.Millisecond << min(try, 6))))
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func (s *Store) txOnce(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, serializable)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if s.commit != nil {
		s.commit()
	}
	return nil
}

func toTicket(r sqldb.Ticket) store.Ticket {
	return store.Ticket{
		ID:         r.ID,
		Project:    r.Project,
		Kind:       store.Kind(r.Kind),
		Title:      r.Title,
		Body:       r.Body,
		Status:     store.Status(r.Status),
		Resolution: r.Resolution,
		Priority:   int(r.Priority),
		Pool:       r.Pool,
		Holder:     r.Holder,
		LeaseUntil: fromNS(r.LeaseUntil),
		Fence:      r.Fence,
		ParentID:   r.ParentID.String,
		DeferUntil: fromNS(r.DeferUntil),
		Ephemeral:  r.Ephemeral != 0,
		Fields:     json.RawMessage(r.Fields),
		Meta:       json.RawMessage(r.Meta),
		Revision:   r.Revision,
		CreatedBy:  r.CreatedBy,
		CreatedAt:  fromNS(r.CreatedAt),
		UpdatedAt:  fromNS(r.UpdatedAt),
		ClosedAt:   fromNS(r.ClosedAt),
	}
}

func notFound(err error, id string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: %s", store.ErrNotFound, id)
	}
	return err
}

func (s *Store) withLabels(ctx context.Context, q sqldb.Querier, ts []store.Ticket) ([]store.Ticket, error) {
	if len(ts) == 0 {
		return ts, nil
	}
	idx := make(map[string]int, len(ts))
	keys := make([]string, len(ts))
	for i, t := range ts {
		idx[t.ID] = i
		keys[i] = t.ID
	}
	rows, err := q.LabelsFor(ctx, keys)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		i := idx[r.TicketID]
		ts[i].Labels = append(ts[i].Labels, r.Label)
	}
	return ts, nil
}

func (s *Store) get(ctx context.Context, q sqldb.Querier, id string) (store.Ticket, error) {
	r, err := q.GetTicket(ctx, id)
	if err != nil {
		return store.Ticket{}, notFound(err, id)
	}
	ts, err := s.withLabels(ctx, q, []store.Ticket{toTicket(r)})
	if err != nil {
		return store.Ticket{}, err
	}
	return ts[0], nil
}

// Get returns one ticket with its labels.
func (s *Store) Get(ctx context.Context, id string) (store.Ticket, error) {
	return s.get(ctx, s.rq, id)
}

func history(ctx context.Context, q sqldb.Querier, t store.Ticket, at time.Time, actor, op string, diff any) error {
	d := []byte("{}")
	if diff != nil {
		var err error
		if d, err = json.Marshal(diff); err != nil {
			return err
		}
	}
	return q.InsertHistory(ctx, sqldb.InsertHistoryParams{
		TicketID: t.ID, At: ns(at), Actor: actor, Op: op, Revision: t.Revision, Diff: string(d),
	})
}

// mutate is the single write path for an existing ticket: load, apply fn,
// write with a revision guard, record history.
func (s *Store) mutate(ctx context.Context, q sqldb.Querier, id, actor, op string,
	fn func(t *store.Ticket) (diff any, err error),
) (store.Ticket, error) {
	return s.mutateWith(ctx, q, id, actor, op, fn, nil)
}

// writeFn writes t (already at its new revision) over prev. It returns
// the rows changed; 0 means the write's guard did not hold.
type writeFn func(ctx context.Context, q sqldb.Querier, t store.Ticket, prev int64, now time.Time) (int64, error)

// errGuard is a guarded write that matched no row; callers turn it into
// the error (or skip) that fits them.
var errGuard = errors.New("guard did not hold")

// mutateWith is mutate with a custom write. With w non-nil, a write that
// changes no row returns errGuard wrapped with ErrConflict.
func (s *Store) mutateWith(ctx context.Context, q sqldb.Querier, id, actor, op string,
	fn func(t *store.Ticket) (diff any, err error), w writeFn,
) (store.Ticket, error) {
	t, err := s.get(ctx, q, id)
	if err != nil {
		return t, err
	}
	diff, err := fn(&t)
	if err != nil {
		return store.Ticket{}, err
	}
	now := s.now()
	prev := storerules.Bump(&t, now)
	if w != nil {
		n, err := w(ctx, q, t, prev, now)
		if err != nil {
			return store.Ticket{}, err
		}
		if n == 0 {
			return store.Ticket{}, fmt.Errorf("%w: %s: %w", store.ErrConflict, id, errGuard)
		}
		if err := history(ctx, q, t, now, actor, op, diff); err != nil {
			return store.Ticket{}, err
		}
		return t, nil
	}
	cols := fieldColumns(t.Fields)
	n, err := q.WriteTicket(ctx, sqldb.WriteTicketParams{
		Title: t.Title, Body: t.Body, Status: string(t.Status), Resolution: t.Resolution,
		Priority: int64(t.Priority), Pool: t.Pool, Holder: t.Holder,
		LeaseUntil: ns(t.LeaseUntil), Fence: t.Fence, DeferUntil: ns(t.DeferUntil),
		Fields: string(t.Fields), Meta: string(t.Meta), Revision: t.Revision,
		FKey: cols.key, FTo: cols.to, FState: cols.state, FWorkflow: cols.workflow,
		FIndex: cols.index, FWaiting: cols.waiting,
		UpdatedAt: ns(now), ClosedAt: ns(t.ClosedAt), ID: id, ExpectedRevision: prev,
	})
	if err != nil {
		return store.Ticket{}, err
	}
	if n == 0 {
		return store.Ticket{}, fmt.Errorf("%w: %s", store.ErrConflict, id)
	}
	if err := syncMeta(ctx, q, id, t.Meta); err != nil {
		return store.Ticket{}, err
	}
	if err := syncResources(ctx, q, id, t.Fields); err != nil {
		return store.Ticket{}, err
	}
	if err := history(ctx, q, t, now, actor, op, diff); err != nil {
		return store.Ticket{}, err
	}
	return t, nil
}

func (s *Store) insert(ctx context.Context, q sqldb.Querier, actor string, nt store.NewTicket) (store.Ticket, error) {
	if nt.Project == "" {
		return store.Ticket{}, fmt.Errorf("%w: project is required", store.ErrInvalid)
	}
	fields, err := store.NormalizeFields(nt.Kind, nt.Fields)
	if err != nil {
		return store.Ticket{}, err
	}
	meta, err := store.NormalizeMeta(nt.Meta)
	if err != nil {
		return store.Ticket{}, err
	}
	id := nt.ID
	if id == "" {
		if id, err = s.mint(ctx, q, nt.Project, idTag(nt.Kind)); err != nil {
			return store.Ticket{}, err
		}
	}
	prio := store.DefaultPriority
	if nt.Priority != nil {
		prio = *nt.Priority
	}
	// Strictly increasing per store (JOB-58), so tickets created in one
	// batch get distinct created_at in creation order.
	now := s.created.Next(s.base())
	var parent sql.NullString
	if nt.ParentID != "" {
		parent = sql.NullString{String: nt.ParentID, Valid: true}
	}
	cols := fieldColumns(fields)
	err = q.InsertTicket(ctx, sqldb.InsertTicketParams{
		ID: id, Project: nt.Project, Kind: string(nt.Kind), Title: nt.Title, Body: nt.Body,
		Status: string(store.StatusOpen), Priority: int64(prio), Pool: nt.Pool, Holder: nt.Holder,
		ParentID: parent, DeferUntil: ns(nt.DeferUntil), Ephemeral: boolInt(nt.Ephemeral),
		Fields: string(fields), Meta: string(meta), Revision: 1, CreatedBy: actor,
		CreatedAt: ns(now), UpdatedAt: ns(now),
		FKey: cols.key, FTo: cols.to, FState: cols.state, FWorkflow: cols.workflow,
		FIndex: cols.index, FWaiting: cols.waiting,
	})
	if err != nil {
		return store.Ticket{}, s.classify(err, id)
	}
	if err := syncMeta(ctx, q, id, meta); err != nil {
		return store.Ticket{}, err
	}
	if err := syncResources(ctx, q, id, fields); err != nil {
		return store.Ticket{}, err
	}
	for _, l := range nt.Labels {
		if err := s.addLabel(ctx, q, id, l); err != nil {
			return store.Ticket{}, err
		}
	}
	t, err := s.get(ctx, q, id)
	if err != nil {
		return t, err
	}
	return t, history(ctx, q, t, now, actor, "create", nil)
}

func boolInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// classify maps the driver's constraint failures to store errors.
func (s *Store) classify(err error, id string) error {
	if err == nil {
		return nil
	}
	kind := classify(err)
	switch {
	case kind == nil:
		return err
	case errors.Is(kind, store.ErrNotFound):
		return fmt.Errorf("%w: %s: referenced ticket does not exist", store.ErrNotFound, id)
	default:
		return fmt.Errorf("%w: %s: %w", kind, id, err)
	}
}

// Apply checks Unclaimed, inserts tickets, re-parents, removes and adds
// edges in one transaction.
func (s *Store) Apply(ctx context.Context, actor string, g store.Graph) (out []store.Ticket, err error) {
	err = s.tx(ctx, func(q sqldb.Querier) error {
		for _, id := range g.Unclaimed {
			t, err := s.get(ctx, q, id)
			if err != nil {
				return err
			}
			if t.Status != store.StatusOpen || t.Holder != "" {
				return fmt.Errorf("%w: %s is %s (holder %q)", store.ErrNotHolder, id, t.Status, t.Holder)
			}
		}
		out = make([]store.Ticket, 0, len(g.Tickets))
		for _, nt := range g.Tickets {
			t, err := s.insert(ctx, q, actor, nt)
			if err != nil {
				return err
			}
			out = append(out, t)
		}
		ids := slices.Sorted(maps.Keys(g.Reparent))
		for _, id := range ids {
			parent := g.Reparent[id]
			if _, err := s.mutate(ctx, q, id, actor, "update", func(t *store.Ticket) (any, error) {
				t.ParentID = parent
				return map[string]any{"parent_id": parent}, nil
			}); err != nil {
				return err
			}
			if err := q.SetParent(ctx, sqldb.SetParentParams{ParentID: sql.NullString{String: parent, Valid: parent != ""}, ID: id}); err != nil {
				return err
			}
		}
		for _, e := range g.RemoveEdges {
			if err := s.removeEdge(ctx, q, actor, e); err != nil {
				return err
			}
		}
		for _, e := range g.Edges {
			if err := s.addEdge(ctx, q, actor, e); err != nil {
				return err
			}
		}
		return nil
	})
	return out, err
}

// NewID mints an id with the project's prefix, in its own transaction
// (seq bumps a counter).
func (s *Store) NewID(project string) (string, error) {
	var id string
	err := s.tx(context.Background(), func(q sqldb.Querier) error {
		var err error
		id, err = s.mint(context.Background(), q, project, "")
		return err
	})
	return id, err
}

// mint returns a new id for project inside the write transaction q.
func (s *Store) mint(ctx context.Context, q sqldb.Querier, project, tag string) (string, error) {
	var spec ids.Spec
	if s.idSpec != nil {
		spec = s.idSpec(project)
	}
	g, err := s.ids.Get(spec)
	if err != nil {
		return "", err
	}
	return ids.Mint(g, s.prefix(project), tag, func(prefix string) (int64, error) {
		return nextSeq(ctx, q, prefix)
	}, func(id string) (bool, error) { return q.TicketExists(ctx, id) })
}

// nextSeq bumps prefix's seq counter. The first use starts it past every
// seq-shaped id already under the prefix. Serializable isolation makes
// two racing first uses conflict and retry.
func nextSeq(ctx context.Context, q sqldb.Querier, prefix string) (int64, error) {
	n, err := q.BumpSeq(ctx, prefix)
	if !errors.Is(err, sql.ErrNoRows) {
		return n, err
	}
	existing, err := q.TicketIDsWithPrefix(ctx, prefix+"-")
	if err != nil {
		return 0, err
	}
	n = ids.SeqStart(prefix, existing)
	return n, q.StartSeq(ctx, prefix, n)
}

// idTag is the kind tag in a new id: sessions get one so their ids are
// distinguishable from ticket ids (#615).
func idTag(k store.Kind) string {
	if k == store.KindSession {
		return ids.SessionTag
	}
	return ""
}

func limit(n int) int64 {
	if n <= 0 {
		return store.DefaultLimit
	}
	return int64(n)
}

func toTickets(rows []sqldb.Ticket) []store.Ticket {
	out := make([]store.Ticket, len(rows))
	for i, r := range rows {
		out[i] = toTicket(r)
	}
	return out
}

func (s *Store) claim(t *store.Ticket, holder string, lease time.Duration) (store.Claim, any) {
	c, d, _ := storerules.Claim(t, holder, lease, s.now())
	return c, d
}

func checkLease(holder string, lease time.Duration) error {
	return storerules.CheckLease(holder, lease)
}

// claimWrite is the guarded write of a claim: it holds only if the
// ticket is still unblocked when written (#406).
func claimWrite(ctx context.Context, q sqldb.Querier, t store.Ticket, prev int64, now time.Time) (int64, error) {
	return q.ClaimWriteTicket(ctx, sqldb.ClaimWriteTicketParams{
		Status: string(t.Status), Holder: t.Holder, LeaseUntil: ns(t.LeaseUntil), Fence: t.Fence,
		Revision: t.Revision, UpdatedAt: ns(now), ID: t.ID, ExpectedRevision: prev,
	})
}

// claimNextTries bounds how many candidates ClaimNext tries when other
// claimers win the race for them.
const claimNextTries = 64

// ClaimNext claims the holder's next ticket. A ticket the holder already has
// in progress is returned again with the same fence and a renewed lease.
// A candidate lost to a concurrent claimer (or blocked meanwhile) is
// skipped for the next one, up to claimNextTries.
func (s *Store) ClaimNext(ctx context.Context, req store.ClaimNext) (t store.Ticket, c store.Claim, err error) {
	if err := checkLease(req.Holder, req.Lease); err != nil {
		return t, c, err
	}
	for range claimNextTries {
		t, c, err = s.claimNextOnce(ctx, req)
		if !errors.Is(err, store.ErrConflict) {
			return t, c, err
		}
	}
	return t, c, err
}

func (s *Store) claimNextOnce(ctx context.Context, req store.ClaimNext) (t store.Ticket, c store.Claim, err error) {
	err = s.sqlTx(ctx, func(tx *sql.Tx) error {
		q := sqldb.New(tx)
		id, err := s.claimCandidate(ctx, tx, req)
		if errors.Is(err, sql.ErrNoRows) {
			return store.ErrNoWork
		}
		if err != nil {
			return err
		}
		fresh := false
		t, err = s.mutateWith(ctx, q, id, req.Holder, "claim", func(t *store.Ticket) (any, error) {
			if err := storerules.StillCandidate(*t, req.Holder); err != nil {
				return nil, err
			}
			fresh = t.Status == store.StatusOpen
			var d any
			c, d = s.claim(t, req.Holder, req.Lease)
			return d, nil
		}, claimWrite)
		if err != nil || !fresh {
			return err
		}
		// A fresh claim takes its provider slots (#407).
		return s.takeSlots(ctx, q, t)
	})
	return t, c, err
}

// Claim claims one specific ticket. It must be open and either unassigned or
// pre-assigned to holder, and unblocked.
func (s *Store) Claim(ctx context.Context, id, holder string, lease time.Duration) (t store.Ticket, c store.Claim, err error) {
	if err := checkLease(holder, lease); err != nil {
		return t, c, err
	}
	err = s.tx(ctx, func(q sqldb.Querier) error {
		blocked, err := s.blocked(ctx, q, id)
		if err != nil {
			return err
		}
		full, err := q.ResourceFull(ctx, id)
		if err != nil {
			return err
		}
		empty, err := q.ProviderEmpty(ctx, id)
		if err != nil {
			return err
		}
		fresh := false
		t, err = s.mutateWith(ctx, q, id, holder, "claim", func(t *store.Ticket) (any, error) {
			if err := storerules.CanClaim(*t, holder, blocked, full != 0 || empty != 0); err != nil {
				return nil, err
			}
			fresh = t.Status == store.StatusOpen
			var d any
			c, d = s.claim(t, holder, lease)
			return d, nil
		}, claimWrite)
		if err != nil || !fresh {
			return err
		}
		return s.takeSlots(ctx, q, t)
	})
	return t, c, err
}

func (s *Store) blocked(ctx context.Context, q sqldb.Querier, id string) (bool, error) {
	edges, err := q.ListEdges(ctx, id)
	if err != nil {
		return false, err
	}
	for _, e := range edges {
		if e.FromID != id || !store.EdgeKind(e.Kind).Blocking() {
			continue
		}
		b, err := q.GetTicket(ctx, e.ToID)
		if err != nil {
			return false, err
		}
		if b.Status != string(store.StatusClosed) {
			return true, nil
		}
		// A step unblocks the next only when it closed done or skipped
		// (matches notBlocked in list.go).
		if b.Kind == string(store.KindStep) && b.Resolution != "done" && b.Resolution != store.StepSkipped {
			return true, nil
		}
	}
	return false, nil
}

// Heartbeat extends a held lease without a revision bump.
func (s *Store) Heartbeat(ctx context.Context, c store.Claim, lease time.Duration) (store.Claim, error) {
	if err := checkLease(c.Holder, lease); err != nil {
		return c, err
	}
	now := s.now()
	c.LeaseUntil = storerules.Stamp(now.Add(lease))
	// Through tx so a lost race (Postgres 40001) is rerun and, at the
	// limit, reported as ErrConflict, not a raw driver error (JOB-38).
	var n int64
	err := s.tx(ctx, func(q sqldb.Querier) error {
		var err error
		n, err = q.Heartbeat(ctx, sqldb.HeartbeatParams{
			LeaseUntil: ns(c.LeaseUntil), Now: ns(now), ID: c.TicketID, Holder: c.Holder, Fence: c.Fence,
		})
		return err
	})
	if err != nil {
		return c, err
	}
	if n == 0 {
		return c, fmt.Errorf("%w: %s", store.ErrNotHolder, c.TicketID)
	}
	return c, nil
}

func reopen(t *store.Ticket) { storerules.Reopen(t) }

// Release lets go of a ticket as o says (#579): close, hand off, or
// return to open.
func (s *Store) Release(ctx context.Context, actor, id string, c *store.Claim, o store.Outcome) (store.Ticket, error) {
	if id == "" && c != nil {
		id = c.TicketID
	}
	if c != nil && c.TicketID != "" && c.TicketID != id {
		return store.Ticket{}, fmt.Errorf("%w: claim is for %s, not %s", store.ErrInvalid, c.TicketID, id)
	}
	if actor == "" && c != nil {
		actor = c.Holder
	}
	switch {
	case o.Close != "" && o.Pool != "":
		return store.Ticket{}, fmt.Errorf("%w: release: close and pool are exclusive", store.ErrInvalid)
	case o.Close != "":
		return s.closeTicket(ctx, actor, id, c, o.Close)
	case o.Pool != "":
		return s.handoff(ctx, actor, id, c, o.Pool, o.Note)
	case c == nil:
		return store.Ticket{}, fmt.Errorf("%w: release to open needs a claim", store.ErrInvalid)
	}
	return s.release(ctx, *c, o.Reason, o.Until)
}

// Expire reopens what holder held or was assigned, or with holder ""
// every ticket whose lease expired (#579).
func (s *Store) Expire(ctx context.Context, actor, holder, reason string) ([]string, error) {
	if holder == "" {
		return s.sweepLeases(ctx, actor)
	}
	return s.releaseHolder(ctx, actor, holder, reason)
}

// release returns a held ticket to open and unassigned.
func (s *Store) release(ctx context.Context, c store.Claim, reason string, until time.Time) (t store.Ticket, err error) {
	err = s.tx(ctx, func(q sqldb.Querier) error {
		var note string
		t, err = s.mutate(ctx, q, c.TicketID, c.Holder, "release", func(t *store.Ticket) (any, error) {
			d, n, err := storerules.Release(t, c, reason, until, s.now())
			note = n
			return d, err
		})
		if err != nil {
			return err
		}
		return s.note(ctx, q, t.ID, note)
	})
	return t, err
}

// handoff moves a ticket to pool, releases its holder, bumps the fence
// and adds note as a comment by actor, all in one transaction (#492).
// A non-nil claim must still hold the ticket. It is not a step failure.
func (s *Store) handoff(ctx context.Context, actor, id string, c *store.Claim, pool, note string) (t store.Ticket, err error) {
	if pool == "" {
		return store.Ticket{}, fmt.Errorf("%w: handoff needs a pool", store.ErrInvalid)
	}
	err = s.tx(ctx, func(q sqldb.Querier) error {
		t, err = s.mutate(ctx, q, id, actor, "handoff", func(t *store.Ticket) (any, error) {
			return storerules.Handoff(t, c, pool, note)
		})
		if err != nil || strings.TrimSpace(note) == "" {
			return err
		}
		_, err = q.InsertComment(ctx, sqldb.InsertCommentParams{TicketID: id, Author: actor, Body: note, CreatedAt: ns(s.now())})
		return err
	})
	return t, err
}

// note adds a controller comment inside a transaction; empty is a no-op.
func (s *Store) note(ctx context.Context, q sqldb.Querier, id, body string) error {
	if body == "" {
		return nil
	}
	_, err := q.InsertComment(ctx, sqldb.InsertCommentParams{TicketID: id, Author: "controller", Body: body, CreatedAt: ns(s.now())})
	return err
}

// closeTicket closes a ticket, fenced by c when non-nil.
func (s *Store) closeTicket(ctx context.Context, actor, id string, c *store.Claim, resolution string) (t store.Ticket, err error) {
	err = s.tx(ctx, func(q sqldb.Querier) error {
		var note string
		t, err = s.mutate(ctx, q, id, actor, "close", func(t *store.Ticket) (any, error) {
			d, n, err := storerules.Close(t, c, resolution, s.now())
			note = n
			return d, err
		})
		if err != nil {
			return err
		}
		return s.note(ctx, q, t.ID, note)
	})
	return t, err
}

// reopen returns a closed ticket to open, keeping its holder unset.
func (s *Store) reopen(ctx context.Context, q sqldb.Querier, actor, id string, ifRev int64) (store.Ticket, error) {
	return s.mutate(ctx, q, id, actor, "reopen", func(t *store.Ticket) (any, error) {
		return nil, storerules.ReopenClosed(t, ifRev)
	})
}

func (s *Store) reopenAll(ctx context.Context, actor, op string, ids func(q sqldb.Querier) ([]string, error), diff any, w writeFn) (out []string, err error) {
	err = s.tx(ctx, func(q sqldb.Querier) error {
		// tx reruns this after a lost write race: ids a rolled-back
		// attempt reopened must not be reported (JOB-38).
		out = nil
		list, err := ids(q)
		if err != nil {
			return err
		}
		for _, id := range list {
			var note string
			_, err := s.mutateWith(ctx, q, id, actor, op, func(t *store.Ticket) (any, error) {
				extra, _ := diff.(map[string]any)
				d, n, err := storerules.Expire(t, op, extra, s.now())
				note = n
				return d, err
			}, w)
			if errors.Is(err, store.ErrConflict) {
				// Changed since the list was read (renewed, released or
				// closed by its holder): not this sweep's to reopen.
				continue
			}
			if err != nil {
				return err
			}
			out = append(out, id)
			if err := s.note(ctx, q, id, note); err != nil {
				return err
			}
		}
		return nil
	})
	return out, err
}

// releaseHolder reopens everything a dead holder held or was assigned.
func (s *Store) releaseHolder(ctx context.Context, actor, holder, reason string) ([]string, error) {
	if holder == "" {
		return nil, fmt.Errorf("%w: holder is required", store.ErrInvalid)
	}
	return s.reopenAll(ctx, actor, "release",
		func(q sqldb.Querier) ([]string, error) { return q.HeldBy(ctx, holder) },
		map[string]any{"reason": reason}, nil)
}

// sweepLeases reopens tickets whose lease expired. Each write re-checks
// the expiry, so a lease renewed after the list was read is kept.
func (s *Store) sweepLeases(ctx context.Context, actor string) ([]string, error) {
	now := s.now()
	return s.reopenAll(ctx, actor, "lease_expired",
		func(q sqldb.Querier) ([]string, error) { return q.ExpiredLeases(ctx, ns(now)) }, nil,
		func(ctx context.Context, q sqldb.Querier, t store.Ticket, prev int64, at time.Time) (int64, error) {
			return q.ExpireWriteTicket(ctx, sqldb.ExpireWriteTicketParams{
				Status: string(t.Status), Resolution: t.Resolution, Holder: t.Holder,
				LeaseUntil: ns(t.LeaseUntil), Fields: string(t.Fields), Revision: t.Revision,
				UpdatedAt: ns(at), ClosedAt: ns(t.ClosedAt), ID: t.ID, ExpectedRevision: prev, Now: ns(now),
			})
		})
}

// Update applies u in one transaction: reopen, fields, labels, comment.
// Holder changes (pre-assignment) are allowed only while the ticket is
// open.
func (s *Store) Update(ctx context.Context, actor, id string, u store.Update) (t store.Ticket, err error) {
	if u.Comment != "" && strings.TrimSpace(u.Comment) == "" {
		return store.Ticket{}, fmt.Errorf("%w: empty comment", store.ErrInvalid)
	}
	extras := u.Reopen || len(u.AddLabels) > 0 || len(u.RemoveLabels) > 0 || u.Comment != ""
	// An update with nothing but labels or a comment leaves the revision
	// alone, as those writes always have.
	fields := !extras || u.IfRevision != 0 || u.Title != nil || u.Body != nil || u.Priority != nil ||
		u.Pool != nil || u.Holder != nil || u.DeferUntil != nil || u.Fields != nil || u.Meta != nil
	err = s.tx(ctx, func(q sqldb.Querier) error {
		if u.Reopen {
			if t, err = s.reopen(ctx, q, actor, id, u.IfRevision); err != nil {
				return err
			}
			u.IfRevision = 0
		}
		if fields {
			if t, err = s.updateFields(ctx, q, actor, id, u); err != nil {
				return err
			}
		}
		if !extras {
			return nil
		}
		if t, err = s.get(ctx, q, id); err != nil {
			return err
		}
		now := s.now()
		for _, l := range u.AddLabels {
			if slices.Contains(t.Labels, l) {
				continue
			}
			if err := s.addLabel(ctx, q, id, l); err != nil {
				return err
			}
			if err := history(ctx, q, t, now, actor, "label_add", map[string]any{"label": l}); err != nil {
				return err
			}
			t.Labels = append(t.Labels, l)
		}
		for _, l := range u.RemoveLabels {
			n, err := q.DeleteLabel(ctx, sqldb.DeleteLabelParams{TicketID: id, Label: l})
			if err != nil {
				return err
			}
			if n == 0 {
				continue
			}
			if err := history(ctx, q, t, now, actor, "label_remove", map[string]any{"label": l}); err != nil {
				return err
			}
			t.Labels = slices.DeleteFunc(t.Labels, func(x string) bool { return x == l })
		}
		if u.Comment != "" {
			r, err := q.InsertComment(ctx, sqldb.InsertCommentParams{
				TicketID: id, Author: actor, Body: u.Comment, CreatedAt: ns(now),
			})
			if err != nil {
				return s.classify(err, id)
			}
			if err := history(ctx, q, t, now, actor, "comment", map[string]any{"comment": r.ID}); err != nil {
				return err
			}
		}
		if len(u.AddLabels) > 0 || len(u.RemoveLabels) > 0 {
			// Read back for the stored label order.
			t, err = s.get(ctx, q, id)
		}
		return err
	})
	return t, err
}

func (s *Store) updateFields(ctx context.Context, q sqldb.Querier, actor, id string, u store.Update) (store.Ticket, error) {
	return s.mutate(ctx, q, id, actor, "update", func(t *store.Ticket) (any, error) {
		return storerules.ApplyUpdate(t, u)
	})
}

func (s *Store) addEdge(ctx context.Context, q sqldb.Querier, actor string, e store.Edge) error {
	if !e.Kind.Valid() {
		return fmt.Errorf("%w: edge kind %q", store.ErrInvalid, e.Kind)
	}
	if e.From == e.To {
		return fmt.Errorf("%w: self edge on %s", store.ErrInvalid, e.From)
	}
	if e.Kind.Blocking() {
		cycle, err := q.BlockingPathExists(ctx, sqldb.BlockingPathExistsParams{Start: e.To, Target: e.From})
		if err != nil {
			return err
		}
		if cycle {
			return fmt.Errorf("%w: %s -> %s", store.ErrCycle, e.From, e.To)
		}
	}
	now := s.now()
	if err := q.InsertEdge(ctx, sqldb.InsertEdgeParams{
		FromID: e.From, ToID: e.To, Kind: string(e.Kind), CreatedAt: ns(now),
	}); err != nil {
		return s.classify(err, e.From)
	}
	t, err := s.get(ctx, q, e.From)
	if err != nil {
		return err
	}
	return history(ctx, q, t, now, actor, "edge_add", map[string]any{"to": e.To, "kind": e.Kind})
}

func (s *Store) removeEdge(ctx context.Context, q sqldb.Querier, actor string, e store.Edge) error {
	n, err := q.DeleteEdge(ctx, sqldb.DeleteEdgeParams{FromID: e.From, ToID: e.To, Kind: string(e.Kind)})
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("%w: edge %s -> %s (%s)", store.ErrNotFound, e.From, e.To, e.Kind)
	}
	t, err := s.get(ctx, q, e.From)
	if err != nil {
		return err
	}
	return history(ctx, q, t, s.now(), actor, "edge_remove", map[string]any{"to": e.To, "kind": e.Kind})
}

func (s *Store) addLabel(ctx context.Context, q sqldb.Querier, id, label string) error {
	label = strings.TrimSpace(label)
	if label == "" {
		return fmt.Errorf("%w: empty label", store.ErrInvalid)
	}
	_, err := q.InsertLabel(ctx, sqldb.InsertLabelParams{TicketID: id, Label: label})
	return s.classify(err, id)
}

// Record adds a history row without changing the ticket.
func (s *Store) Record(ctx context.Context, actor, id, op string, diff any) error {
	return s.tx(ctx, func(q sqldb.Querier) error {
		t, err := s.get(ctx, q, id)
		if err != nil {
			return err
		}
		return history(ctx, q, t, s.now(), actor, op, diff)
	})
}
