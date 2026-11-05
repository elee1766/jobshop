package sqlstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/ids"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqlcdb"
)

// Options configures New.
type Options struct {
	// Now overrides the clock, for tests.
	Now func() time.Time
	// Prefix maps a project to its id prefix. Default: the project name.
	Prefix func(project string) string
	// OnCommit, if set, runs after every committed write transaction.
	// The controller uses it to wake events.subscribe (#147).
	OnCommit func()
	// Concurrent opens with deferred transactions and several
	// connections, so writers are not serialized up front (#406). Tests
	// use it to check claims do not depend on a single-writer database.
	// Drivers read it in their Open; the core does not.
	Concurrent bool
}

// Driver is what a backend supplies beyond its SQL files: the opened and
// migrated database, its sqlc queries, and the engine-specific bits.
type Driver struct {
	// DB is the migrated database. Write transactions start with
	// DB.BeginTx, so the driver sets any locking mode in its DSN.
	DB *sql.DB
	// Bind returns the engine's queries over a database or transaction.
	// The method set and types are sqlcdb's (sql_package: database/sql,
	// emit_interface), so the core never converts rows.
	Bind func(sqlcdb.DBTX) sqlcdb.Querier
	// Classify maps a constraint failure to store.ErrAlreadyExist,
	// store.ErrNotFound or store.ErrInvalid; other errors return nil.
	Classify func(err error) error
	// RecallMatch turns the words of a recall query into the match
	// expressions to try in order; the first with hits wins.
	RecallMatch func(words []string) []string
	// IsBusy reports a transaction that lost a write race and may be
	// rerun (SQLITE_BUSY on a deferred upgrade). Nil means never.
	IsBusy func(err error) bool
}

// Store is the backend-neutral ticket store over a Driver. Claims and
// sweeps do not rely on a single-writer database: their writes re-check
// their conditions in the same statement (#406).
type Store struct {
	db     *sql.DB
	q      sqlcdb.Querier
	drv    Driver
	ids    *ids.Generator
	now    func() time.Time
	prefix func(string) string
	commit func()
}

var _ store.Store = (*Store)(nil)

// New returns a store over d. Close closes d.DB.
func New(d Driver, opt Options) (*Store, error) {
	gen, err := ids.New()
	if err != nil {
		return nil, err
	}
	s := &Store{db: d.DB, q: d.Bind(d.DB), drv: d, ids: gen, now: opt.Now, prefix: opt.Prefix, commit: opt.OnCommit}
	if s.now == nil {
		s.now = time.Now
	}
	if s.prefix == nil {
		s.prefix = func(p string) string { return p }
	}
	return s, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

// busyTries bounds how often tx reruns a transaction that lost a write
// race (Driver.IsBusy). Only concurrent databases hit it.
const busyTries = 16

func (s *Store) tx(ctx context.Context, fn func(q sqlcdb.Querier) error) error {
	var err error
	for range busyTries {
		if err = s.txOnce(ctx, fn); !s.isBusy(err) {
			break
		}
	}
	// A lost write race means the caller's read is stale, as with a
	// revision mismatch, and retrying is safe.
	if s.isBusy(err) {
		return fmt.Errorf("%w: %w", store.ErrConflict, err)
	}
	return err
}

func (s *Store) isBusy(err error) bool {
	return err != nil && s.drv.IsBusy != nil && s.drv.IsBusy(err)
}

func (s *Store) txOnce(ctx context.Context, fn func(q sqlcdb.Querier) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(s.drv.Bind(tx)); err != nil {
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

func ms(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixMilli()
}

func fromMS(v int64) time.Time {
	if v == 0 {
		return time.Time{}
	}
	return time.UnixMilli(v).UTC()
}

func toTicket(r sqlcdb.Ticket) store.Ticket {
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
		LeaseUntil: fromMS(r.LeaseUntil),
		Fence:      r.Fence,
		ParentID:   r.ParentID.String,
		DeferUntil: fromMS(r.DeferUntil),
		Ephemeral:  r.Ephemeral != 0,
		Fields:     json.RawMessage(r.Fields),
		Meta:       json.RawMessage(r.Meta),
		Revision:   r.Revision,
		CreatedBy:  r.CreatedBy,
		CreatedAt:  fromMS(r.CreatedAt),
		UpdatedAt:  fromMS(r.UpdatedAt),
		ClosedAt:   fromMS(r.ClosedAt),
	}
}

func notFound(err error, id string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: %s", store.ErrNotFound, id)
	}
	return err
}

func (s *Store) withLabels(ctx context.Context, q sqlcdb.Querier, ts []store.Ticket) ([]store.Ticket, error) {
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

func (s *Store) get(ctx context.Context, q sqlcdb.Querier, id string) (store.Ticket, error) {
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
	return s.get(ctx, s.q, id)
}

func history(ctx context.Context, q sqlcdb.Querier, t store.Ticket, at time.Time, actor, op string, diff any) error {
	d := []byte("{}")
	if diff != nil {
		var err error
		if d, err = json.Marshal(diff); err != nil {
			return err
		}
	}
	return q.InsertHistory(ctx, sqlcdb.InsertHistoryParams{
		TicketID: t.ID, At: ms(at), Actor: actor, Op: op, Revision: t.Revision, Diff: string(d),
	})
}

// mutate is the single write path for an existing ticket: load, apply fn,
// write with a revision guard, record history.
func (s *Store) mutate(ctx context.Context, q sqlcdb.Querier, id, actor, op string,
	fn func(t *store.Ticket) (diff any, err error),
) (store.Ticket, error) {
	return s.mutateWith(ctx, q, id, actor, op, fn, nil)
}

// writeFn writes t (already at its new revision) over prev. It returns
// the rows changed; 0 means the write's guard did not hold.
type writeFn func(ctx context.Context, q sqlcdb.Querier, t store.Ticket, prev int64, now time.Time) (int64, error)

// errGuard is a guarded write that matched no row; callers turn it into
// the error (or skip) that fits them.
var errGuard = errors.New("guard did not hold")

// mutateWith is mutate with a custom write. With w non-nil, a write that
// changes no row returns errGuard wrapped with ErrConflict.
func (s *Store) mutateWith(ctx context.Context, q sqlcdb.Querier, id, actor, op string,
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
	prev := t.Revision
	t.Revision++
	t.UpdatedAt = now.UTC().Truncate(time.Millisecond)
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
	n, err := q.WriteTicket(ctx, sqlcdb.WriteTicketParams{
		Title: t.Title, Body: t.Body, Status: string(t.Status), Resolution: t.Resolution,
		Priority: int64(t.Priority), Pool: t.Pool, Holder: t.Holder,
		LeaseUntil: ms(t.LeaseUntil), Fence: t.Fence, DeferUntil: ms(t.DeferUntil),
		Fields: string(t.Fields), Meta: string(t.Meta), Revision: t.Revision,
		FKey: cols.key, FTo: cols.to, FState: cols.state, FWorkflow: cols.workflow,
		FIndex: cols.index, FWaiting: cols.waiting,
		UpdatedAt: ms(now), ClosedAt: ms(t.ClosedAt), ID: id, ExpectedRevision: prev,
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

func (s *Store) insert(ctx context.Context, q sqlcdb.Querier, actor string, nt store.NewTicket) (store.Ticket, error) {
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
		if id, err = s.ids.Next(s.prefix(nt.Project)); err != nil {
			return store.Ticket{}, err
		}
	}
	prio := store.DefaultPriority
	if nt.Priority != nil {
		prio = *nt.Priority
	}
	now := s.now()
	var parent sql.NullString
	if nt.ParentID != "" {
		parent = sql.NullString{String: nt.ParentID, Valid: true}
	}
	cols := fieldColumns(fields)
	err = q.InsertTicket(ctx, sqlcdb.InsertTicketParams{
		ID: id, Project: nt.Project, Kind: string(nt.Kind), Title: nt.Title, Body: nt.Body,
		Status: string(store.StatusOpen), Priority: int64(prio), Pool: nt.Pool, Holder: nt.Holder,
		ParentID: parent, DeferUntil: ms(nt.DeferUntil), Ephemeral: boolInt(nt.Ephemeral),
		Fields: string(fields), Meta: string(meta), Revision: 1, CreatedBy: actor,
		CreatedAt: ms(now), UpdatedAt: ms(now),
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
	kind := s.drv.Classify(err)
	switch {
	case kind == nil:
		return err
	case errors.Is(kind, store.ErrNotFound):
		return fmt.Errorf("%w: %s: referenced ticket does not exist", store.ErrNotFound, id)
	default:
		return fmt.Errorf("%w: %s: %w", kind, id, err)
	}
}

// Create inserts one ticket.
func (s *Store) Create(ctx context.Context, actor string, nt store.NewTicket) (t store.Ticket, err error) {
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		t, err = s.insert(ctx, q, actor, nt)
		return err
	})
	return t, err
}

// CreateGraph inserts tickets then edges atomically. Tickets are inserted in
// order, so a parent must precede its children.
func (s *Store) CreateGraph(ctx context.Context, actor string, nts []store.NewTicket, edges []store.Edge) (out []store.Ticket, err error) {
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		out = make([]store.Ticket, 0, len(nts))
		for _, nt := range nts {
			t, err := s.insert(ctx, q, actor, nt)
			if err != nil {
				return err
			}
			out = append(out, t)
		}
		for _, e := range edges {
			if err := s.addEdge(ctx, q, actor, e); err != nil {
				return err
			}
		}
		return nil
	})
	return out, err
}

// ApplyGraph checks Unclaimed, inserts tickets, re-parents and adds edges
// in one transaction.
func (s *Store) ApplyGraph(ctx context.Context, actor string, g store.Graph) (out []store.Ticket, err error) {
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
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
			if err := q.SetParent(ctx, sqlcdb.SetParentParams{ParentID: sql.NullString{String: parent, Valid: parent != ""}, ID: id}); err != nil {
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

// FindOpenBatchByKey returns the open batch of project with key.
func (s *Store) FindOpenBatchByKey(ctx context.Context, project, key string) (store.Ticket, error) {
	row, err := s.q.FindOpenBatchByKey(ctx, sqlcdb.FindOpenBatchByKeyParams{Project: project, Key: nstr(key)})
	if errors.Is(err, sql.ErrNoRows) {
		return store.Ticket{}, fmt.Errorf("%w: open batch %s key %q", store.ErrNotFound, project, key)
	}
	if err != nil {
		return store.Ticket{}, err
	}
	ts, err := s.withLabels(ctx, s.q, toTickets([]sqlcdb.Ticket{row}))
	if err != nil {
		return store.Ticket{}, err
	}
	return ts[0], nil
}

// NewID mints an id with the project's prefix.
func (s *Store) NewID(project string) (string, error) {
	return s.ids.Next(s.prefix(project))
}

// ListChildren returns id's children ordered by fields.index.
func (s *Store) ListChildren(ctx context.Context, id string) ([]store.Ticket, error) {
	rows, err := s.q.ListChildren(ctx, sql.NullString{String: id, Valid: true})
	if err != nil {
		return nil, err
	}
	return s.withLabels(ctx, s.q, toTickets(rows))
}

// FindOpenRunByKey returns the open run of workflow with key.
func (s *Store) FindOpenRunByKey(ctx context.Context, workflow, key string) (store.Ticket, error) {
	row, err := s.q.FindOpenRunByKey(ctx, sqlcdb.FindOpenRunByKeyParams{Key: nstr(key), Workflow: nstr(workflow)})
	if errors.Is(err, sql.ErrNoRows) {
		return store.Ticket{}, fmt.Errorf("%w: open run %s key %q", store.ErrNotFound, workflow, key)
	}
	if err != nil {
		return store.Ticket{}, err
	}
	ts, err := s.withLabels(ctx, s.q, toTickets([]sqlcdb.Ticket{row}))
	if err != nil {
		return store.Ticket{}, err
	}
	return ts[0], nil
}

// ListRuns returns runs newest first.
func (s *Store) ListRuns(ctx context.Context, project, state string, n int) ([]store.Ticket, error) {
	rows, err := s.q.ListRuns(ctx, sqlcdb.ListRunsParams{Project: narg(project), State: narg(state), MaxRows: limit(n)})
	if err != nil {
		return nil, err
	}
	return s.withLabels(ctx, s.q, toTickets(rows))
}

func limit(n int) int64 {
	if n <= 0 {
		return store.DefaultLimit
	}
	return int64(n)
}

func nstr(s string) sql.NullString { return sql.NullString{String: s, Valid: true} }

func narg(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// List returns tickets matching f ordered by priority then age.
func (s *Store) List(ctx context.Context, f store.Filter) ([]store.Ticket, error) {
	var metaPack any
	var metaKey sql.NullString
	if f.MetaPack != "" && f.MetaKey != "" {
		metaPack = f.MetaPack
		metaKey = sql.NullString{String: f.MetaKey, Valid: true}
	}
	metaValue := sql.NullString{String: f.MetaValue, Valid: metaPack != nil}
	var rows []sqlcdb.Ticket
	var err error
	if f.Newest {
		rows, err = s.q.ListTicketsNewest(ctx, sqlcdb.ListTicketsNewestParams{
			Project: narg(f.Project), Kind: narg(string(f.Kind)), Status: narg(string(f.Status)),
			Pool: narg(f.Pool), Holder: narg(f.Holder), ParentID: narg(f.ParentID), CreatedBy: narg(f.CreatedBy),
			Label: narg(f.Label), MetaPack: metaPack, MetaKey: metaKey, MetaValue: metaValue, MaxRows: limit(f.Limit),
		})
	} else {
		rows, err = s.q.ListTickets(ctx, sqlcdb.ListTicketsParams{
			Project: narg(f.Project), Kind: narg(string(f.Kind)), Status: narg(string(f.Status)),
			Pool: narg(f.Pool), Holder: narg(f.Holder), ParentID: narg(f.ParentID), CreatedBy: narg(f.CreatedBy),
			Label: narg(f.Label), MetaPack: metaPack, MetaKey: metaKey, MetaValue: metaValue, MaxRows: limit(f.Limit),
		})
	}
	if err != nil {
		return nil, err
	}
	return s.withLabels(ctx, s.q, toTickets(rows))
}

func toTickets(rows []sqlcdb.Ticket) []store.Ticket {
	out := make([]store.Ticket, len(rows))
	for i, r := range rows {
		out[i] = toTicket(r)
	}
	return out
}

func kindsList(ks []store.Kind) []string {
	if len(ks) == 0 {
		ks = store.ClaimableKinds
	}
	out := make([]string, len(ks))
	for i, k := range ks {
		out[i] = string(k)
	}
	return out
}

// Ready returns the claimable frontier for a project, optionally one pool.
func (s *Store) Ready(ctx context.Context, project, pool string, n int) ([]store.Ticket, error) {
	rows, err := s.q.ReadyTickets(ctx, sqlcdb.ReadyTicketsParams{
		Project: project, Kinds: kindsList(nil), Column3: boolInt(pool == ""), Pool: pool,
		DeferUntil: ms(s.now()), Column6: limit(n),
	})
	if err != nil {
		return nil, err
	}
	return s.withLabels(ctx, s.q, toTickets(rows))
}

// Orphaned returns open, unclaimed tickets with no pool or an unknown one.
func (s *Store) Orphaned(ctx context.Context, project string, pools []string, n int) ([]store.Ticket, error) {
	// An empty slice binds as IN (NULL), which matches nothing under NOT
	// IN; "" is never a pool name here (pool = '' is its own case).
	known := pools
	if len(known) == 0 {
		known = []string{""}
	}
	rows, err := s.q.OrphanedTickets(ctx, sqlcdb.OrphanedTicketsParams{
		Column1: boolInt(project == ""), Project: project, Kinds: kindsList(nil), Known: known, Column5: limit(n),
	})
	if err != nil {
		return nil, err
	}
	return s.withLabels(ctx, s.q, toTickets(rows))
}

func (s *Store) claim(t *store.Ticket, holder string, lease time.Duration) (store.Claim, any) {
	resumed := t.Status == store.StatusInProgress
	t.Status = store.StatusInProgress
	t.Holder = holder
	t.LeaseUntil = s.now().Add(lease).UTC().Truncate(time.Millisecond)
	if !resumed {
		t.Fence++
	}
	return store.Claim{TicketID: t.ID, Holder: holder, Fence: t.Fence, LeaseUntil: t.LeaseUntil},
		map[string]any{"holder": holder, "fence": t.Fence, "resumed": resumed}
}

func checkLease(holder string, lease time.Duration) error {
	if holder == "" || lease <= 0 {
		return fmt.Errorf("%w: holder and a positive lease are required", store.ErrInvalid)
	}
	return nil
}

// claimWrite is the guarded write of a claim: it holds only if the
// ticket is still unblocked when written (#406).
func claimWrite(ctx context.Context, q sqlcdb.Querier, t store.Ticket, prev int64, now time.Time) (int64, error) {
	return q.ClaimWriteTicket(ctx, sqlcdb.ClaimWriteTicketParams{
		Status: string(t.Status), Holder: t.Holder, LeaseUntil: ms(t.LeaseUntil), Fence: t.Fence,
		Revision: t.Revision, UpdatedAt: ms(now), ID: t.ID, ExpectedRevision: prev,
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
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		r, err := q.NextClaimCandidate(ctx, sqlcdb.NextClaimCandidateParams{
			Project: req.Project, Kinds: kindsList(req.Kinds), Holder: req.Holder, Holder_2: req.Holder,
			Pool: req.Pool, DeferUntil: ms(s.now()),
		})
		if errors.Is(err, sql.ErrNoRows) {
			return store.ErrNoWork
		}
		if err != nil {
			return err
		}
		fresh := false
		t, err = s.mutateWith(ctx, q, r.ID, req.Holder, "claim", func(t *store.Ticket) (any, error) {
			// The candidate was read in this transaction, but without a
			// single-writer database another claimer may have taken it.
			if t.Status == store.StatusClosed || (t.Holder != "" && t.Holder != req.Holder) {
				return nil, fmt.Errorf("%w: %s was taken", store.ErrConflict, t.ID)
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
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
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
			switch {
			case t.Status == store.StatusClosed:
				return nil, fmt.Errorf("%w: %s", store.ErrClosed, id)
			case t.Status == store.StatusInProgress && t.Holder != holder:
				return nil, fmt.Errorf("%w: %s is held by %s", store.ErrNotHolder, id, t.Holder)
			case t.Status == store.StatusOpen && t.Holder != "" && t.Holder != holder:
				return nil, fmt.Errorf("%w: %s is assigned to %s", store.ErrNotHolder, id, t.Holder)
			case blocked:
				return nil, fmt.Errorf("%w: %s is blocked", store.ErrNoWork, id)
			case (full != 0 || empty != 0) && t.Status == store.StatusOpen:
				return nil, fmt.Errorf("%w: %s waits for a resource slot", store.ErrNoWork, id)
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

func (s *Store) blocked(ctx context.Context, q sqlcdb.Querier, id string) (bool, error) {
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
		// (matches ReadyTickets and NextClaimCandidate).
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
	c.LeaseUntil = now.Add(lease).UTC().Truncate(time.Millisecond)
	n, err := s.q.Heartbeat(ctx, sqlcdb.HeartbeatParams{
		LeaseUntil: ms(c.LeaseUntil), Now: ms(now), ID: c.TicketID, Holder: c.Holder, Fence: c.Fence,
	})
	if err != nil {
		return c, err
	}
	if n == 0 {
		return c, fmt.Errorf("%w: %s", store.ErrNotHolder, c.TicketID)
	}
	return c, nil
}

func holds(t *store.Ticket, c store.Claim) error {
	if t.Status != store.StatusInProgress || t.Holder != c.Holder || t.Fence != c.Fence {
		return fmt.Errorf("%w: %s", store.ErrNotHolder, t.ID)
	}
	return nil
}

func reopen(t *store.Ticket) {
	t.Status = store.StatusOpen
	t.Holder = ""
	t.LeaseUntil = time.Time{}
}

// Release returns a held ticket to open and unassigned.
func (s *Store) Release(ctx context.Context, c store.Claim, reason string, until time.Time) (t store.Ticket, err error) {
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		var note string
		t, err = s.mutate(ctx, q, c.TicketID, c.Holder, "release", func(t *store.Ticket) (any, error) {
			if err := holds(t, c); err != nil {
				return nil, err
			}
			reopen(t)
			d := map[string]any{"reason": reason}
			if !until.IsZero() {
				// A deferred release is "not yet", not a failed attempt.
				t.DeferUntil = until.UTC().Truncate(time.Millisecond)
				d["defer_until"] = t.DeferUntil
				note = fmt.Sprintf("Released until %s: %s", t.DeferUntil.Format(time.RFC3339), reason)
				return d, nil
			}
			var ferr error
			note, ferr = s.stepFailure(t, "released: "+reason, d)
			return d, ferr
		})
		if err != nil {
			return err
		}
		return s.note(ctx, q, t.ID, note)
	})
	return t, err
}

// Handoff moves a ticket to pool, releases its holder, bumps the fence
// and adds note as a comment by actor, all in one transaction (#492).
// A non-nil claim must still hold the ticket. It is not a step failure.
func (s *Store) Handoff(ctx context.Context, actor, id string, c *store.Claim, pool, note string) (t store.Ticket, err error) {
	if pool == "" {
		return store.Ticket{}, fmt.Errorf("%w: handoff needs a pool", store.ErrInvalid)
	}
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		t, err = s.mutate(ctx, q, id, actor, "handoff", func(t *store.Ticket) (any, error) {
			if t.Status == store.StatusClosed {
				return nil, fmt.Errorf("%w: %s", store.ErrClosed, id)
			}
			if c != nil {
				if err := holds(t, *c); err != nil {
					return nil, err
				}
			}
			d := map[string]any{"from_pool": t.Pool, "pool": pool, "holder": t.Holder}
			t.Pool = pool
			if t.Status == store.StatusInProgress {
				reopen(t)
				t.Fence++
			}
			if note != "" {
				d["note"] = note
			}
			return d, nil
		})
		if err != nil || strings.TrimSpace(note) == "" {
			return err
		}
		_, err = q.InsertComment(ctx, sqlcdb.InsertCommentParams{TicketID: id, Author: actor, Body: note, CreatedAt: ms(s.now())})
		return err
	})
	return t, err
}

// stepFailure counts a failure when t is a step, in the caller's
// transaction. At the retry limit the step is closed failed. It returns
// the comment to add, or "" for non-steps, and records the count in d.
func (s *Store) stepFailure(t *store.Ticket, how string, d map[string]any) (string, error) {
	f, exhausted, isStep, err := store.StepFailure(t)
	if err != nil || !isStep {
		return "", err
	}
	d["failures"] = f.Failures
	if exhausted && t.Status != store.StatusClosed {
		t.Status = store.StatusClosed
		t.Resolution = "failed"
		t.Holder = ""
		t.LeaseUntil = time.Time{}
		t.ClosedAt = s.now().UTC().Truncate(time.Millisecond)
	}
	if !exhausted && t.Status == store.StatusClosed {
		// Below the limit a failed close becomes another attempt.
		reopen(t)
		t.Resolution = ""
		t.ClosedAt = time.Time{}
	}
	return store.StepFailureComment(f, how, exhausted), nil
}

// note adds a controller comment inside a transaction; empty is a no-op.
func (s *Store) note(ctx context.Context, q sqlcdb.Querier, id, body string) error {
	if body == "" {
		return nil
	}
	_, err := q.InsertComment(ctx, sqlcdb.InsertCommentParams{TicketID: id, Author: "controller", Body: body, CreatedAt: ms(s.now())})
	return err
}

// CloseTicket closes a ticket, fenced by c when non-nil.
func (s *Store) CloseTicket(ctx context.Context, actor, id string, c *store.Claim, resolution string) (t store.Ticket, err error) {
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		var note string
		t, err = s.mutate(ctx, q, id, actor, "close", func(t *store.Ticket) (any, error) {
			if t.Status == store.StatusClosed {
				return nil, fmt.Errorf("%w: %s", store.ErrClosed, id)
			}
			if c != nil {
				if err := holds(t, *c); err != nil {
					return nil, err
				}
			}
			t.Status = store.StatusClosed
			t.Resolution = resolution
			t.LeaseUntil = time.Time{}
			t.ClosedAt = s.now().UTC().Truncate(time.Millisecond)
			d := map[string]any{"resolution": resolution}
			if resolution == "failed" {
				var ferr error
				note, ferr = s.stepFailure(t, "closed failed", d)
				return d, ferr
			}
			if resolution == store.ResolutionRebase && t.Kind == store.KindStep {
				reopen(t)
				t.Resolution = ""
				t.ClosedAt = time.Time{}
				note = "Closed for a rebase: reopened without counting a failed attempt."
			}
			return d, nil
		})
		if err != nil {
			return err
		}
		return s.note(ctx, q, t.ID, note)
	})
	return t, err
}

// Reopen returns a closed ticket to open, keeping its holder unset.
func (s *Store) Reopen(ctx context.Context, actor, id string) (t store.Ticket, err error) {
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		t, err = s.mutate(ctx, q, id, actor, "reopen", func(t *store.Ticket) (any, error) {
			if t.Status != store.StatusClosed {
				return nil, fmt.Errorf("%w: %s is not closed", store.ErrInvalid, id)
			}
			reopen(t)
			t.Resolution = ""
			t.ClosedAt = time.Time{}
			return nil, nil
		})
		return err
	})
	return t, err
}

func (s *Store) reopenAll(ctx context.Context, actor, op string, ids func(q sqlcdb.Querier) ([]string, error), diff any, w writeFn) (out []string, err error) {
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		list, err := ids(q)
		if err != nil {
			return err
		}
		for _, id := range list {
			var note string
			_, err := s.mutateWith(ctx, q, id, actor, op, func(t *store.Ticket) (any, error) {
				d := map[string]any{"holder": t.Holder, "status": t.Status}
				if m, ok := diff.(map[string]any); ok {
					for k, v := range m {
						d[k] = v
					}
				}
				// A pre-assigned step that was never claimed has not failed.
				claimed := t.Status == store.StatusInProgress
				reopen(t)
				if !claimed {
					return d, nil
				}
				how := op
				if r, ok := d["reason"].(string); ok && r != "" {
					how += ": " + r
				}
				var ferr error
				note, ferr = s.stepFailure(t, how, d)
				return d, ferr
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

// ReleaseHolder reopens everything a dead holder held or was assigned.
func (s *Store) ReleaseHolder(ctx context.Context, actor, holder, reason string) ([]string, error) {
	if holder == "" {
		return nil, fmt.Errorf("%w: holder is required", store.ErrInvalid)
	}
	return s.reopenAll(ctx, actor, "release",
		func(q sqlcdb.Querier) ([]string, error) { return q.HeldBy(ctx, holder) },
		map[string]any{"reason": reason}, nil)
}

// SweepLeases reopens tickets whose lease expired. Each write re-checks
// the expiry, so a lease renewed after the list was read is kept.
func (s *Store) SweepLeases(ctx context.Context, actor string) ([]string, error) {
	now := s.now()
	return s.reopenAll(ctx, actor, "lease_expired",
		func(q sqlcdb.Querier) ([]string, error) { return q.ExpiredLeases(ctx, ms(now)) }, nil,
		func(ctx context.Context, q sqlcdb.Querier, t store.Ticket, prev int64, at time.Time) (int64, error) {
			return q.ExpireWriteTicket(ctx, sqlcdb.ExpireWriteTicketParams{
				Status: string(t.Status), Resolution: t.Resolution, Holder: t.Holder,
				LeaseUntil: ms(t.LeaseUntil), Fields: string(t.Fields), Revision: t.Revision,
				UpdatedAt: ms(at), ClosedAt: ms(t.ClosedAt), ID: t.ID, ExpectedRevision: prev, Now: ms(now),
			})
		})
}

// Update applies u. Holder changes (pre-assignment) are allowed only while
// the ticket is open.
func (s *Store) Update(ctx context.Context, actor, id string, u store.Update) (t store.Ticket, err error) {
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		t, err = s.mutate(ctx, q, id, actor, "update", func(t *store.Ticket) (any, error) {
			if u.IfRevision != 0 && u.IfRevision != t.Revision {
				return nil, fmt.Errorf("%w: %s at revision %d, want %d", store.ErrConflict, id, t.Revision, u.IfRevision)
			}
			d := map[string]any{}
			set := func(name string, dst *string, v *string) {
				if v != nil && *v != *dst {
					*dst = *v
					d[name] = *v
				}
			}
			set("title", &t.Title, u.Title)
			set("body", &t.Body, u.Body)
			set("pool", &t.Pool, u.Pool)
			if u.Holder != nil && *u.Holder != t.Holder {
				if t.Status != store.StatusOpen {
					return nil, fmt.Errorf("%w: %s: assign only while open (status %s)", store.ErrInvalid, id, t.Status)
				}
				t.Holder = *u.Holder
				d["holder"] = *u.Holder
			}
			if u.Priority != nil && *u.Priority != t.Priority {
				t.Priority = *u.Priority
				d["priority"] = *u.Priority
			}
			if u.DeferUntil != nil {
				t.DeferUntil = u.DeferUntil.UTC().Truncate(time.Millisecond)
				d["defer_until"] = t.DeferUntil
			}
			if u.Fields != nil {
				f, err := store.NormalizeFields(t.Kind, u.Fields)
				if err != nil {
					return nil, err
				}
				t.Fields = f
				d["fields"] = f
			}
			if u.Meta != nil {
				m, err := store.NormalizeMeta(u.Meta)
				if err != nil {
					return nil, err
				}
				t.Meta = m
				d["meta"] = m
			}
			return d, nil
		})
		return err
	})
	return t, err
}

func (s *Store) addEdge(ctx context.Context, q sqlcdb.Querier, actor string, e store.Edge) error {
	if !e.Kind.Valid() {
		return fmt.Errorf("%w: edge kind %q", store.ErrInvalid, e.Kind)
	}
	if e.From == e.To {
		return fmt.Errorf("%w: self edge on %s", store.ErrInvalid, e.From)
	}
	if e.Kind.Blocking() {
		cycle, err := q.BlockingPathExists(ctx, sqlcdb.BlockingPathExistsParams{Start: e.To, Target: e.From})
		if err != nil {
			return err
		}
		if cycle {
			return fmt.Errorf("%w: %s -> %s", store.ErrCycle, e.From, e.To)
		}
	}
	now := s.now()
	if err := q.InsertEdge(ctx, sqlcdb.InsertEdgeParams{
		FromID: e.From, ToID: e.To, Kind: string(e.Kind), CreatedAt: ms(now),
	}); err != nil {
		return s.classify(err, e.From)
	}
	t, err := s.get(ctx, q, e.From)
	if err != nil {
		return err
	}
	return history(ctx, q, t, now, actor, "edge_add", map[string]any{"to": e.To, "kind": e.Kind})
}

// AddEdge records a relation. Blocking edges that would form a cycle fail.
func (s *Store) AddEdge(ctx context.Context, actor string, e store.Edge) error {
	return s.tx(ctx, func(q sqlcdb.Querier) error { return s.addEdge(ctx, q, actor, e) })
}

// RemoveEdge deletes a relation.
func (s *Store) RemoveEdge(ctx context.Context, actor string, e store.Edge) error {
	return s.tx(ctx, func(q sqlcdb.Querier) error {
		n, err := q.DeleteEdge(ctx, sqlcdb.DeleteEdgeParams{FromID: e.From, ToID: e.To, Kind: string(e.Kind)})
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
	})
}

// Edges returns edges in either direction touching id.
func (s *Store) Edges(ctx context.Context, id string) ([]store.Edge, error) {
	rows, err := s.q.ListEdges(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]store.Edge, len(rows))
	for i, r := range rows {
		out[i] = store.Edge{From: r.FromID, To: r.ToID, Kind: store.EdgeKind(r.Kind), CreatedAt: fromMS(r.CreatedAt)}
	}
	return out, nil
}

func (s *Store) addLabel(ctx context.Context, q sqlcdb.Querier, id, label string) error {
	label = strings.TrimSpace(label)
	if label == "" {
		return fmt.Errorf("%w: empty label", store.ErrInvalid)
	}
	_, err := q.InsertLabel(ctx, sqlcdb.InsertLabelParams{TicketID: id, Label: label})
	return s.classify(err, id)
}

// AddLabel adds a user label. Adding an existing label is a no-op.
func (s *Store) AddLabel(ctx context.Context, actor, id, label string) error {
	return s.tx(ctx, func(q sqlcdb.Querier) error {
		t, err := s.get(ctx, q, id)
		if err != nil {
			return err
		}
		if slices.Contains(t.Labels, label) {
			return nil
		}
		if err := s.addLabel(ctx, q, id, label); err != nil {
			return err
		}
		return history(ctx, q, t, s.now(), actor, "label_add", map[string]any{"label": label})
	})
}

// RemoveLabel removes a user label. Removing a missing label is a no-op.
func (s *Store) RemoveLabel(ctx context.Context, actor, id, label string) error {
	return s.tx(ctx, func(q sqlcdb.Querier) error {
		t, err := s.get(ctx, q, id)
		if err != nil {
			return err
		}
		n, err := q.DeleteLabel(ctx, sqlcdb.DeleteLabelParams{TicketID: id, Label: label})
		if err != nil || n == 0 {
			return err
		}
		return history(ctx, q, t, s.now(), actor, "label_remove", map[string]any{"label": label})
	})
}

// AddComment appends a comment and records a "comment" history row (no
// revision bump), so the change stream sees it.
func (s *Store) AddComment(ctx context.Context, author, id, body string) (c store.Comment, err error) {
	if strings.TrimSpace(body) == "" {
		return store.Comment{}, fmt.Errorf("%w: empty comment", store.ErrInvalid)
	}
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		now := s.now()
		r, err := q.InsertComment(ctx, sqlcdb.InsertCommentParams{
			TicketID: id, Author: author, Body: body, CreatedAt: ms(now),
		})
		if err != nil {
			return s.classify(err, id)
		}
		c = toComment(r)
		t, err := s.get(ctx, q, id)
		if err != nil {
			return err
		}
		return history(ctx, q, t, now, author, "comment", map[string]any{"comment": c.ID})
	})
	return c, err
}

func toComment(r sqlcdb.Comment) store.Comment {
	return store.Comment{ID: r.ID, TicketID: r.TicketID, Author: r.Author, Body: r.Body, CreatedAt: fromMS(r.CreatedAt)}
}

// Comments lists a ticket's comments oldest first.
func (s *Store) Comments(ctx context.Context, id string) ([]store.Comment, error) {
	rows, err := s.q.ListComments(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]store.Comment, len(rows))
	for i, r := range rows {
		out[i] = toComment(r)
	}
	return out, nil
}

// History returns a ticket's audit log oldest first.
func (s *Store) History(ctx context.Context, id string) ([]store.HistoryEntry, error) {
	rows, err := s.q.ListHistory(ctx, id)
	if err != nil {
		return nil, err
	}
	return toHistory(rows), nil
}

// Changes returns history rows of every ticket after seq, oldest first.
func (s *Store) Changes(ctx context.Context, after int64, n int) ([]store.HistoryEntry, error) {
	if n <= 0 {
		n = store.DefaultLimit
	}
	rows, err := s.q.ListChanges(ctx, sqlcdb.ListChangesParams{Seq: after, Lim: int64(n)})
	if err != nil {
		return nil, err
	}
	return toHistory(rows), nil
}

// LastSeq is the newest history seq, 0 when there is none.
func (s *Store) LastSeq(ctx context.Context) (int64, error) { return s.q.LastSeq(ctx) }

func toHistory(rows []sqlcdb.History) []store.HistoryEntry {
	out := make([]store.HistoryEntry, len(rows))
	for i, r := range rows {
		out[i] = store.HistoryEntry{
			Seq: r.Seq, TicketID: r.TicketID, At: fromMS(r.At), Actor: r.Actor,
			Op: r.Op, Revision: r.Revision, Diff: json.RawMessage(r.Diff),
		}
	}
	return out
}
