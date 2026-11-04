package sqlite

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite/sqlcdb"
)

// InboxPut stores an ingested item, deduped by delivery id and natural
// key (#321, #335).
func (s *Store) InboxPut(ctx context.Context, it store.InboxItem) (out store.InboxItem, dup bool, err error) {
	if err := checkInbox(it); err != nil {
		return out, false, err
	}
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
		out, dup, err = s.inboxPut(ctx, q, it)
		return err
	})
	return out, dup, err
}

// InboxPutBatch stores its in one transaction: all or none (#335).
func (s *Store) InboxPutBatch(ctx context.Context, its []store.InboxItem) (fresh int, err error) {
	for _, it := range its {
		if err := checkInbox(it); err != nil {
			return 0, err
		}
	}
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
		fresh = 0
		for _, it := range its {
			_, dup, err := s.inboxPut(ctx, q, it)
			if err != nil {
				return err
			}
			if !dup {
				fresh++
			}
		}
		return nil
	})
	return fresh, err
}

func checkInbox(it store.InboxItem) error {
	if it.DeliveryID == "" {
		return fmt.Errorf("%w: inbox item needs a delivery id", store.ErrInvalid)
	}
	if it.Source != "github" && it.Source != "linear" {
		return fmt.Errorf("%w: inbox source %q must be github or linear", store.ErrInvalid, it.Source)
	}
	return nil
}

// inboxPut inserts it, or returns the row it clashes with: same delivery
// id first, else same natural key.
func (s *Store) inboxPut(ctx context.Context, q *sqlcdb.Queries, it store.InboxItem) (store.InboxItem, bool, error) {
	now := s.now()
	recv := it.ReceivedAt
	if recv.IsZero() {
		recv = now
	}
	payload := string(it.Payload)
	if payload == "" {
		payload = "{}"
	}
	n, err := q.InboxInsert(ctx, sqlcdb.InboxInsertParams{
		Source: it.Source, DeliveryID: it.DeliveryID, Kind: it.Kind, RepoOrTeam: it.RepoOrTeam,
		Number: it.Number, CommentID: it.CommentID, Actor: it.Actor, ReceivedAt: ms(recv),
		Payload: payload, Now: ms(now), Endpoint: it.Endpoint, NaturalKey: it.NaturalKey,
	})
	if err != nil {
		return store.InboxItem{}, false, err
	}
	r, err := q.InboxGetByDelivery(ctx, it.DeliveryID)
	if errors.Is(err, sql.ErrNoRows) && n == 0 && it.NaturalKey != "" {
		r, err = q.InboxGetByNaturalKey(ctx, it.NaturalKey)
	}
	if err != nil {
		return store.InboxItem{}, false, err
	}
	return toInbox(r), n == 0, nil
}

// InboxClaim leases the oldest claimable item in states. The take is
// conditional on the row still being claimable, so concurrent claimers
// never share a row; a lost race tries the next row.
func (s *Store) InboxClaim(ctx context.Context, states []string, holder string, lease time.Duration) (out store.InboxItem, err error) {
	if holder == "" || lease <= 0 || len(states) == 0 {
		return out, fmt.Errorf("%w: inbox claim needs states, a holder and a lease", store.ErrInvalid)
	}
	for _, st := range states {
		if !slices.Contains(store.InboxStates, st) {
			return out, fmt.Errorf("%w: unknown inbox state %q", store.ErrInvalid, st)
		}
	}
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
		for range 10 {
			now := s.now()
			r, err := q.InboxNextClaimable(ctx, sqlcdb.InboxNextClaimableParams{States: states, Now: ms(now)})
			if err != nil {
				return notFound(err, "claimable inbox item")
			}
			r, err = q.InboxTake(ctx, sqlcdb.InboxTakeParams{Holder: holder, LeaseUntil: ms(now.Add(lease)), Now: ms(now), ID: r.ID})
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			if err != nil {
				return err
			}
			out = toInbox(r)
			return nil
		}
		return fmt.Errorf("%w: inbox claim lost every race", store.ErrConflict)
	})
	return out, err
}

// InboxRenew extends holder's live lease on id.
func (s *Store) InboxRenew(ctx context.Context, id int64, holder string, lease time.Duration) (out store.InboxItem, err error) {
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
		now := s.now()
		r, err := q.InboxRenew(ctx, sqlcdb.InboxRenewParams{LeaseUntil: ms(now.Add(lease)), Now: ms(now), ID: id, Holder: holder})
		if err != nil {
			return s.inboxFenced(ctx, q, err, id, holder)
		}
		out = toInbox(r)
		return nil
	})
	return out, err
}

// InboxSetState moves id to u.State if holder still holds its lease.
func (s *Store) InboxSetState(ctx context.Context, id int64, holder string, u store.InboxUpdate) (out store.InboxItem, err error) {
	if !slices.Contains(store.InboxStates, u.State) {
		return out, fmt.Errorf("%w: unknown inbox state %q", store.ErrInvalid, u.State)
	}
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
		r, err := q.InboxSetState(ctx, sqlcdb.InboxSetStateParams{
			State: u.State, Admitted: nullJSON(u.Admitted), HookResults: nullJSON(u.HookResults),
			Now: ms(s.now()), ID: id, Holder: holder,
		})
		if err != nil {
			return s.inboxFenced(ctx, q, err, id, holder)
		}
		out = toInbox(r)
		return nil
	})
	return out, err
}

// inboxFenced turns a fenced update that matched no row into ErrNotFound
// (no such item) or ErrConflict (another holder, or the lease expired).
func (s *Store) inboxFenced(ctx context.Context, q *sqlcdb.Queries, err error, id int64, holder string) error {
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if _, err := q.InboxGet(ctx, id); err != nil {
		return notFound(err, fmt.Sprintf("inbox item %d", id))
	}
	return fmt.Errorf("%w: inbox item %d is not leased to %s", store.ErrConflict, id, holder)
}

// InboxGet returns one item.
func (s *Store) InboxGet(ctx context.Context, id int64) (store.InboxItem, error) {
	r, err := s.q.InboxGet(ctx, id)
	if err != nil {
		return store.InboxItem{}, notFound(err, fmt.Sprintf("inbox item %d", id))
	}
	return toInbox(r), nil
}

// InboxList lists items newest first.
func (s *Store) InboxList(ctx context.Context, f store.InboxFilter) ([]store.InboxItem, error) {
	rs, err := s.q.InboxList(ctx, sqlcdb.InboxListParams{State: f.State, Source: f.Source, Lim: int64(cmp.Or(f.Limit, 100))})
	if err != nil {
		return nil, err
	}
	out := make([]store.InboxItem, len(rs))
	for i, r := range rs {
		out[i] = toInbox(r)
	}
	return out, nil
}

// InboxPrune deletes terminal items received before before.
func (s *Store) InboxPrune(ctx context.Context, before time.Time) (n int64, err error) {
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
		n, err = q.InboxPrune(ctx, ms(before))
		return err
	})
	return n, err
}

func nullJSON(b json.RawMessage) sql.NullString {
	if b == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: string(b), Valid: true}
}

func toInbox(r sqlcdb.Inbox) store.InboxItem {
	it := store.InboxItem{
		ID: r.ID, Source: r.Source, DeliveryID: r.DeliveryID, Endpoint: r.Endpoint, NaturalKey: r.NaturalKey, Kind: r.Kind, RepoOrTeam: r.RepoOrTeam,
		Number: r.Number, CommentID: r.CommentID, Actor: r.Actor, ReceivedAt: fromMS(r.ReceivedAt),
		Payload: json.RawMessage(r.Payload), State: r.State, HookResults: json.RawMessage(r.HookResults),
		Holder: r.Holder, LeaseUntil: fromMS(r.LeaseUntil), Attempts: r.Attempts, UpdatedAt: fromMS(r.UpdatedAt),
	}
	if r.Admitted.Valid {
		it.Admitted = json.RawMessage(r.Admitted.String)
	}
	return it
}
