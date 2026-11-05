package sqlstore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nullism/bqb"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqlcdb"
)

// Ticket reads built in Go (#579 step 3): List builds its WHERE from the
// Filter fields with bqb, and ClaimNext shares the readiness fragments
// below, so the rule that makes a ticket ready is written once. The
// backend renders placeholders (Driver.Render). Every value is bound;
// only fixed SQL text is concatenated.

const ticketCols = "t.id, t.project, t.kind, t.title, t.body, t.status, t.resolution, t.priority, t.pool, " +
	"t.holder, t.lease_until, t.fence, t.parent_id, t.defer_until, t.ephemeral, t.fields, t.meta, " +
	"t.revision, t.created_by, t.created_at, t.updated_at, t.closed_at, t.f_key, t.f_to, t.f_state, " +
	"t.f_workflow, t.f_index, t.f_waiting"

// Readiness fragments over tickets aliased t.
const (
	// A blocks or waits_for target that is open, or a step that closed
	// other than done or skipped, blocks t (#90, #385).
	notBlocked = `NOT EXISTS (
    SELECT 1 FROM edges e JOIN tickets b ON b.id = e.to_id
    WHERE e.from_id = t.id AND e.kind IN ('blocks','waits_for')
      AND (b.status <> 'closed' OR (b.kind = 'step' AND b.resolution NOT IN ('done','skipped'))))`
	// Every resource t needs has a free slot (#400). Occupancy is
	// in-progress tickets naming the same key; a key "name:..." takes
	// the cap of "name"; a key with no cap is unlimited. h.id <> t.id
	// leaves out a holder's own in-progress ticket.
	slotsFree = `NOT EXISTS (
    SELECT 1 FROM ticket_resources r
    JOIN resource_caps c ON c.name = r.rkey OR c.name = r.rname
    WHERE r.ticket_id = t.id AND c.capacity <= (
      SELECT count(*) FROM ticket_resources hr JOIN tickets h ON h.id = hr.ticket_id
      WHERE h.status = 'in_progress' AND h.id <> t.id AND hr.rkey = r.rkey))`
	// A provider resource needs a ready slot (#407).
	providerReady = `NOT EXISTS (
    SELECT 1 FROM ticket_resources r
    JOIN resource_providers p ON p.name = r.rkey
    WHERE r.ticket_id = t.id
      AND NOT EXISTS (SELECT 1 FROM resource_slots s WHERE s.pool = p.name AND s.state = 'ready'))`
	// fields.wait parks a ticket until the controller clears it (#381).
	notWaiting = "t.f_waiting = 0"
)

func strs[T ~string](xs []T) []string {
	out := make([]string, len(xs))
	for i, x := range xs {
		out[i] = string(x)
	}
	return out
}

// filterKinds is f.Kinds plus f.Kind; Ready with none means the
// claimable kinds.
func filterKinds(f store.Filter) []string {
	ks := strs(f.Kinds)
	if f.Kind != "" {
		ks = append(ks, string(f.Kind))
	}
	if len(ks) == 0 && f.Ready {
		ks = strs(store.ClaimableKinds)
	}
	return ks
}

func filterStatuses(f store.Filter) []string {
	ss := strs(f.Statuses)
	if f.Status != "" {
		ss = append(ss, string(f.Status))
	}
	return ss
}

// listQuery builds the List statement for f at now (ms).
func listQuery(f store.Filter, now int64) (*bqb.Query, error) {
	metas := f.Meta
	if f.MetaPack != "" && f.MetaKey != "" {
		metas = append([]store.MetaMatch{{Pack: f.MetaPack, Key: f.MetaKey, Value: f.MetaValue}}, metas...)
	}
	if len(metas) > store.MaxMetaFilters {
		return nil, fmt.Errorf("%w: at most %d meta filters", store.ErrInvalid, store.MaxMetaFilters)
	}
	order := f.Order
	w := bqb.Optional("WHERE")
	eq := func(col, v string) {
		if v != "" {
			w.And("t."+col+" = ?", v)
		}
	}
	eq("project", f.Project)
	eq("pool", f.Pool)
	eq("holder", f.Holder)
	eq("parent_id", f.ParentID)
	eq("created_by", f.CreatedBy)
	eq("f_key", f.Key)
	eq("f_workflow", f.Workflow)
	eq("f_state", f.State)
	if ks := filterKinds(f); len(ks) > 0 {
		w.And("t.kind IN (?)", ks)
	}
	ss := filterStatuses(f)
	switch {
	case len(ss) == 2 && ((ss[0] == "open" && ss[1] == "in_progress") || (ss[0] == "in_progress" && ss[1] == "open")):
		// The same set, written as the partial indexes on open work
		// (tickets_run_key) are, so SQLite can use them.
		w.And("t.status <> 'closed'")
	case len(ss) > 0:
		w.And("t.status IN (?)", ss)
	}
	if f.Unheld {
		w.And("t.holder = ''")
	}
	if len(f.PoolNotIn) > 0 {
		w.And("(t.pool = '' OR t.pool NOT IN (?))", f.PoolNotIn)
	}
	if !f.Since.IsZero() {
		w.And("t.updated_at >= ?", ms(f.Since))
	}
	if f.Created != nil {
		if !f.Created.Since.IsZero() {
			w.And("t.created_at >= ?", ms(f.Created.Since))
		}
		if !f.Created.Until.IsZero() {
			w.And("t.created_at < ?", ms(f.Created.Until))
		}
	}
	if f.Label != "" {
		w.And("EXISTS (SELECT 1 FROM labels l WHERE l.ticket_id = t.id AND l.label = ?)", f.Label)
	}
	for _, m := range metas {
		w.And("EXISTS (SELECT 1 FROM ticket_meta m WHERE m.ticket_id = t.id AND m.pack = ? AND m.mkey = ? AND m.value = ?)",
			m.Pack, m.Key, m.Value)
	}
	if f.Ready {
		w.And("t.status = 'open'")
		w.And("t.defer_until <= ?", now)
		w.And(notWaiting)
		w.And(notBlocked)
		w.And(slotsFree)
		w.And(providerReady)
	}
	var by string
	switch order {
	case "", store.OrderPriority:
		by = "t.priority, t.created_at, t.id"
		if f.After != "" {
			w.And(`EXISTS (SELECT 1 FROM tickets a WHERE a.id = ?
      AND (t.priority > a.priority OR (t.priority = a.priority AND (t.created_at > a.created_at
        OR (t.created_at = a.created_at AND t.id > a.id)))))`, f.After)
		}
	case store.OrderNewest:
		by = "t.created_at DESC, t.id DESC"
		if f.After != "" {
			w.And(`EXISTS (SELECT 1 FROM tickets a WHERE a.id = ?
      AND (t.created_at < a.created_at OR (t.created_at = a.created_at AND t.id < a.id)))`, f.After)
		}
	case store.OrderOldest:
		by = "t.created_at, t.id"
		if f.After != "" {
			w.And(`EXISTS (SELECT 1 FROM tickets a WHERE a.id = ?
      AND (t.created_at > a.created_at OR (t.created_at = a.created_at AND t.id > a.id)))`, f.After)
		}
	case store.OrderIndex:
		if f.After != "" {
			return nil, fmt.Errorf("%w: order %q has no cursor", store.ErrInvalid, order)
		}
		by = "t.f_index, t.created_at, t.id"
	default:
		return nil, fmt.Errorf("%w: order %q", store.ErrInvalid, order)
	}
	q := bqb.New("SELECT "+ticketCols+" FROM tickets t ?", w)
	q.Space("ORDER BY " + by)
	if f.Limit >= 0 {
		// CAST keeps the statement text the same for every limit: a bare
		// bound LIMIT makes SQLite re-prepare a cached statement (#433).
		q.Space("LIMIT CAST(? AS BIGINT)", limit(f.Limit))
	}
	return q, nil
}

// candidateQuery picks ClaimNext's ticket: the holder's in-progress
// work, then tickets pre-assigned to it, then the pool queue. It uses
// List's readiness fragments; the holder's own in-progress ticket
// already holds its provider slots.
func candidateQuery(req store.ClaimNext, now int64) *bqb.Query {
	w := bqb.New("t.project = ?", req.Project)
	w.And("t.kind IN (?)", filterKinds(store.Filter{Kinds: req.Kinds, Ready: true}))
	w.And(`(
    (t.status = 'in_progress' AND t.holder = ?)
    OR (t.status = 'open' AND t.holder = ?)
    OR (t.status = 'open' AND t.holder = '' AND t.pool = ?))`, req.Holder, req.Holder, req.Pool)
	w.And("t.defer_until <= ?", now)
	w.And(notWaiting)
	w.And(notBlocked)
	w.And(slotsFree)
	w.And("(t.status = 'in_progress' OR " + providerReady + ")")
	return bqb.New("SELECT "+ticketCols+" FROM tickets t WHERE ? ORDER BY "+
		"CASE WHEN t.status = 'in_progress' THEN 0 WHEN t.holder <> '' THEN 1 ELSE 2 END, "+
		"t.priority, t.created_at, t.id LIMIT 1", w)
}

type scanner interface{ Scan(dest ...any) error }

func scanTicket(r scanner) (sqlcdb.Ticket, error) {
	var i sqlcdb.Ticket
	err := r.Scan(&i.ID, &i.Project, &i.Kind, &i.Title, &i.Body, &i.Status, &i.Resolution, &i.Priority,
		&i.Pool, &i.Holder, &i.LeaseUntil, &i.Fence, &i.ParentID, &i.DeferUntil, &i.Ephemeral, &i.Fields,
		&i.Meta, &i.Revision, &i.CreatedBy, &i.CreatedAt, &i.UpdatedAt, &i.ClosedAt, &i.FKey, &i.FTo,
		&i.FState, &i.FWorkflow, &i.FIndex, &i.FWaiting)
	return i, err
}

// conn is d with the backend's statement handling (Driver.Conn).
func (s *Store) conn(d sqlcdb.DBTX) sqlcdb.DBTX {
	if s.drv.Conn == nil {
		return d
	}
	return s.drv.Conn(d)
}

func (s *Store) render(q *bqb.Query) (string, []any, error) {
	if s.drv.Render == nil {
		return q.ToSql()
	}
	return s.drv.Render(q)
}

func (s *Store) queryTickets(ctx context.Context, d sqlcdb.DBTX, q *bqb.Query) ([]sqlcdb.Ticket, error) {
	text, args, err := s.render(q)
	if err != nil {
		return nil, err
	}
	rows, err := s.conn(d).QueryContext(ctx, text, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []sqlcdb.Ticket{}
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// List returns tickets matching f in f.Order.
func (s *Store) List(ctx context.Context, f store.Filter) ([]store.Ticket, error) {
	q, err := listQuery(f, ms(s.now()))
	if err != nil {
		return nil, err
	}
	rows, err := s.queryTickets(ctx, s.db, q)
	if err != nil {
		return nil, err
	}
	return s.withLabels(ctx, s.q, toTickets(rows))
}

// claimCandidate reads ClaimNext's candidate in tx; sql.ErrNoRows means
// none.
func (s *Store) claimCandidate(ctx context.Context, tx *sql.Tx, req store.ClaimNext) (string, error) {
	rows, err := s.queryTickets(ctx, tx, candidateQuery(req, ms(s.now())))
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", sql.ErrNoRows
	}
	return rows[0].ID, nil
}
