package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // database/sql driver "sqlite"

	"github.com/elee1766/jobshop/pkg/ids"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite/sqlcdb"
)

//go:generate go tool sqlc generate

//go:embed migrations/*.sql
var migrations embed.FS

// Options configures Open.
type Options struct {
	// Now overrides the clock, for tests.
	Now func() time.Time
	// Prefix maps a project to its id prefix. Default: the project name.
	Prefix func(project string) string
}

// Store is the SQLite backend. All access goes through one connection: the
// controller is the only writer and SQLite serializes writes anyway.
type Store struct {
	db     *sql.DB
	q      *sqlcdb.Queries
	ids    *ids.Generator
	now    func() time.Time
	prefix func(string) string
}

var _ store.Store = (*Store)(nil)

// Open opens (creating if needed) the database at path and migrates it.
// Use ":memory:" for a private in-memory database.
func Open(ctx context.Context, path string, opt Options) (*Store, error) {
	dsn := "file:" + path + "?_txlock=immediate" +
		"&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)" +
		"&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetConnMaxIdleTime(0)
	db.SetConnMaxLifetime(0)
	if err := migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	gen, err := ids.New()
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	s := &Store{db: db, q: sqlcdb.New(db), ids: gen, now: opt.Now, prefix: opt.Prefix}
	if s.now == nil {
		s.now = time.Now
	}
	if s.prefix == nil {
		s.prefix = func(p string) string { return p }
	}
	return s, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	p, err := goose.NewProvider(goose.DialectSQLite3, db, mustSub(migrations, "migrations"))
	if err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	if _, err := p.Up(ctx); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) tx(ctx context.Context, fn func(q *sqlcdb.Queries) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(s.q.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit()
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

func (s *Store) withLabels(ctx context.Context, q *sqlcdb.Queries, ts []store.Ticket) ([]store.Ticket, error) {
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

func (s *Store) get(ctx context.Context, q *sqlcdb.Queries, id string) (store.Ticket, error) {
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

func history(ctx context.Context, q *sqlcdb.Queries, t store.Ticket, at time.Time, actor, op string, diff any) error {
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
func (s *Store) mutate(ctx context.Context, q *sqlcdb.Queries, id, actor, op string,
	fn func(t *store.Ticket) (diff any, err error),
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
	n, err := q.WriteTicket(ctx, sqlcdb.WriteTicketParams{
		Title: t.Title, Body: t.Body, Status: string(t.Status), Resolution: t.Resolution,
		Priority: int64(t.Priority), Pool: t.Pool, Holder: t.Holder,
		LeaseUntil: ms(t.LeaseUntil), Fence: t.Fence, DeferUntil: ms(t.DeferUntil),
		Fields: string(t.Fields), Meta: string(t.Meta), Revision: t.Revision,
		UpdatedAt: ms(now), ClosedAt: ms(t.ClosedAt), ID: id, ExpectedRevision: prev,
	})
	if err != nil {
		return store.Ticket{}, err
	}
	if n == 0 {
		return store.Ticket{}, fmt.Errorf("%w: %s", store.ErrConflict, id)
	}
	if err := history(ctx, q, t, now, actor, op, diff); err != nil {
		return store.Ticket{}, err
	}
	return t, nil
}

func (s *Store) insert(ctx context.Context, q *sqlcdb.Queries, actor string, nt store.NewTicket) (store.Ticket, error) {
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
	err = q.InsertTicket(ctx, sqlcdb.InsertTicketParams{
		ID: id, Project: nt.Project, Kind: string(nt.Kind), Title: nt.Title, Body: nt.Body,
		Status: string(store.StatusOpen), Priority: int64(prio), Pool: nt.Pool, Holder: nt.Holder,
		ParentID: parent, DeferUntil: ms(nt.DeferUntil), Ephemeral: boolInt(nt.Ephemeral),
		Fields: string(fields), Meta: string(meta), Revision: 1, CreatedBy: actor,
		CreatedAt: ms(now), UpdatedAt: ms(now),
	})
	if err != nil {
		return store.Ticket{}, classify(err, id)
	}
	for _, l := range nt.Labels {
		if err := addLabel(ctx, q, id, l); err != nil {
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

// classify maps SQLite constraint failures to store errors.
func classify(err error, id string) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "UNIQUE constraint failed"), strings.Contains(msg, "PRIMARY KEY"):
		return fmt.Errorf("%w: %s: %w", store.ErrAlreadyExist, id, err)
	case strings.Contains(msg, "FOREIGN KEY constraint failed"):
		return fmt.Errorf("%w: %s: referenced ticket does not exist", store.ErrNotFound, id)
	case strings.Contains(msg, "CHECK constraint failed"):
		return fmt.Errorf("%w: %s: %w", store.ErrInvalid, id, err)
	}
	return err
}

// Create inserts one ticket.
func (s *Store) Create(ctx context.Context, actor string, nt store.NewTicket) (t store.Ticket, err error) {
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
		t, err = s.insert(ctx, q, actor, nt)
		return err
	})
	return t, err
}

// CreateGraph inserts tickets then edges atomically. Tickets are inserted in
// order, so a parent must precede its children.
func (s *Store) CreateGraph(ctx context.Context, actor string, nts []store.NewTicket, edges []store.Edge) (out []store.Ticket, err error) {
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
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

func limit(n int) int64 {
	if n <= 0 {
		return store.DefaultLimit
	}
	return int64(n)
}

func narg(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// List returns tickets matching f ordered by priority then age.
func (s *Store) List(ctx context.Context, f store.Filter) ([]store.Ticket, error) {
	rows, err := s.q.ListTickets(ctx, sqlcdb.ListTicketsParams{
		Project: narg(f.Project), Kind: narg(string(f.Kind)), Status: narg(string(f.Status)),
		Pool: narg(f.Pool), Holder: narg(f.Holder), ParentID: narg(f.ParentID),
		Label: narg(f.Label), MaxRows: limit(f.Limit),
	})
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

func kindsJSON(ks []store.Kind) string {
	if len(ks) == 0 {
		ks = store.ClaimableKinds
	}
	b, _ := json.Marshal(ks) // []string-like, cannot fail
	return string(b)
}

// Ready returns the claimable frontier for a project, optionally one pool.
func (s *Store) Ready(ctx context.Context, project, pool string, n int) ([]store.Ticket, error) {
	rows, err := s.q.ReadyTickets(ctx, sqlcdb.ReadyTicketsParams{
		Project: project, KindsJson: kindsJSON(nil), Pool: narg(pool),
		Now: ms(s.now()), MaxRows: limit(n),
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

// ClaimNext claims the holder's next ticket. A ticket the holder already has
// in progress is returned again with the same fence and a renewed lease.
func (s *Store) ClaimNext(ctx context.Context, req store.ClaimNext) (t store.Ticket, c store.Claim, err error) {
	if err := checkLease(req.Holder, req.Lease); err != nil {
		return t, c, err
	}
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
		r, err := q.NextClaimCandidate(ctx, sqlcdb.NextClaimCandidateParams{
			Project: req.Project, KindsJson: kindsJSON(req.Kinds), Holder: req.Holder,
			Pool: req.Pool, Now: ms(s.now()),
		})
		if errors.Is(err, sql.ErrNoRows) {
			return store.ErrNoWork
		}
		if err != nil {
			return err
		}
		t, err = s.mutate(ctx, q, r.ID, req.Holder, "claim", func(t *store.Ticket) (any, error) {
			var d any
			c, d = s.claim(t, req.Holder, req.Lease)
			return d, nil
		})
		return err
	})
	return t, c, err
}

// Claim claims one specific ticket. It must be open and either unassigned or
// pre-assigned to holder, and unblocked.
func (s *Store) Claim(ctx context.Context, id, holder string, lease time.Duration) (t store.Ticket, c store.Claim, err error) {
	if err := checkLease(holder, lease); err != nil {
		return t, c, err
	}
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
		blocked, err := s.blocked(ctx, q, id)
		if err != nil {
			return err
		}
		t, err = s.mutate(ctx, q, id, holder, "claim", func(t *store.Ticket) (any, error) {
			switch {
			case t.Status == store.StatusClosed:
				return nil, fmt.Errorf("%w: %s", store.ErrClosed, id)
			case t.Status == store.StatusInProgress && t.Holder != holder:
				return nil, fmt.Errorf("%w: %s is held by %s", store.ErrNotHolder, id, t.Holder)
			case t.Status == store.StatusOpen && t.Holder != "" && t.Holder != holder:
				return nil, fmt.Errorf("%w: %s is assigned to %s", store.ErrNotHolder, id, t.Holder)
			case blocked:
				return nil, fmt.Errorf("%w: %s is blocked", store.ErrNoWork, id)
			}
			var d any
			c, d = s.claim(t, holder, lease)
			return d, nil
		})
		return err
	})
	return t, c, err
}

func (s *Store) blocked(ctx context.Context, q *sqlcdb.Queries, id string) (bool, error) {
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
func (s *Store) Release(ctx context.Context, c store.Claim, reason string) (t store.Ticket, err error) {
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
		t, err = s.mutate(ctx, q, c.TicketID, c.Holder, "release", func(t *store.Ticket) (any, error) {
			if err := holds(t, c); err != nil {
				return nil, err
			}
			reopen(t)
			return map[string]any{"reason": reason}, nil
		})
		return err
	})
	return t, err
}

// CloseTicket closes a ticket, fenced by c when non-nil.
func (s *Store) CloseTicket(ctx context.Context, actor, id string, c *store.Claim, resolution string) (t store.Ticket, err error) {
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
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
			return map[string]any{"resolution": resolution}, nil
		})
		return err
	})
	return t, err
}

// Reopen returns a closed ticket to open, keeping its holder unset.
func (s *Store) Reopen(ctx context.Context, actor, id string) (t store.Ticket, err error) {
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
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

func (s *Store) reopenAll(ctx context.Context, actor, op string, ids func(q *sqlcdb.Queries) ([]string, error), diff any) (out []string, err error) {
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
		list, err := ids(q)
		if err != nil {
			return err
		}
		out = list
		for _, id := range list {
			_, err := s.mutate(ctx, q, id, actor, op, func(t *store.Ticket) (any, error) {
				d := map[string]any{"holder": t.Holder, "status": t.Status}
				if m, ok := diff.(map[string]any); ok {
					for k, v := range m {
						d[k] = v
					}
				}
				reopen(t)
				return d, nil
			})
			if err != nil {
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
		func(q *sqlcdb.Queries) ([]string, error) { return q.HeldBy(ctx, holder) },
		map[string]any{"reason": reason})
}

// SweepLeases reopens tickets whose lease expired.
func (s *Store) SweepLeases(ctx context.Context, actor string) ([]string, error) {
	return s.reopenAll(ctx, actor, "lease_expired",
		func(q *sqlcdb.Queries) ([]string, error) { return q.ExpiredLeases(ctx, ms(s.now())) }, nil)
}

// Update applies u. Holder changes (pre-assignment) are allowed only while
// the ticket is open.
func (s *Store) Update(ctx context.Context, actor, id string, u store.Update) (t store.Ticket, err error) {
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
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

func (s *Store) addEdge(ctx context.Context, q *sqlcdb.Queries, actor string, e store.Edge) error {
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
		return classify(err, e.From)
	}
	t, err := s.get(ctx, q, e.From)
	if err != nil {
		return err
	}
	return history(ctx, q, t, now, actor, "edge_add", map[string]any{"to": e.To, "kind": e.Kind})
}

// AddEdge records a relation. Blocking edges that would form a cycle fail.
func (s *Store) AddEdge(ctx context.Context, actor string, e store.Edge) error {
	return s.tx(ctx, func(q *sqlcdb.Queries) error { return s.addEdge(ctx, q, actor, e) })
}

// RemoveEdge deletes a relation.
func (s *Store) RemoveEdge(ctx context.Context, actor string, e store.Edge) error {
	return s.tx(ctx, func(q *sqlcdb.Queries) error {
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

func addLabel(ctx context.Context, q *sqlcdb.Queries, id, label string) error {
	label = strings.TrimSpace(label)
	if label == "" {
		return fmt.Errorf("%w: empty label", store.ErrInvalid)
	}
	_, err := q.InsertLabel(ctx, sqlcdb.InsertLabelParams{TicketID: id, Label: label})
	return classify(err, id)
}

// AddLabel adds a user label. Adding an existing label is a no-op.
func (s *Store) AddLabel(ctx context.Context, actor, id, label string) error {
	return s.tx(ctx, func(q *sqlcdb.Queries) error {
		t, err := s.get(ctx, q, id)
		if err != nil {
			return err
		}
		if slices.Contains(t.Labels, label) {
			return nil
		}
		if err := addLabel(ctx, q, id, label); err != nil {
			return err
		}
		return history(ctx, q, t, s.now(), actor, "label_add", map[string]any{"label": label})
	})
}

// RemoveLabel removes a user label. Removing a missing label is a no-op.
func (s *Store) RemoveLabel(ctx context.Context, actor, id, label string) error {
	return s.tx(ctx, func(q *sqlcdb.Queries) error {
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

// AddComment appends a comment.
func (s *Store) AddComment(ctx context.Context, author, id, body string) (store.Comment, error) {
	if strings.TrimSpace(body) == "" {
		return store.Comment{}, fmt.Errorf("%w: empty comment", store.ErrInvalid)
	}
	r, err := s.q.InsertComment(ctx, sqlcdb.InsertCommentParams{
		TicketID: id, Author: author, Body: body, CreatedAt: ms(s.now()),
	})
	if err != nil {
		return store.Comment{}, classify(err, id)
	}
	return toComment(r), nil
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
	out := make([]store.HistoryEntry, len(rows))
	for i, r := range rows {
		out[i] = store.HistoryEntry{
			Seq: r.Seq, TicketID: r.TicketID, At: fromMS(r.At), Actor: r.Actor,
			Op: r.Op, Revision: r.Revision, Diff: json.RawMessage(r.Diff),
		}
	}
	return out, nil
}
