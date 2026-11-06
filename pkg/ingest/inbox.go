package ingest

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/elee1766/jobshop/pkg/store"
)

// The inbox (#321) lives in the inbox store collection (#579). Records
// are keyed by the item id as a decimal string: the ingest RPC and logs
// name items by id. Delivery id and natural key are unique fields, so a
// Put that repeats either returns the stored item. The record state is
// the inbox state, the record lease is the admission or dispatch lease,
// and created_at is received_at, so claims, prune and the backlog count
// order and filter by it as the table did. Ids come from a counter in
// the inbox_ids collection: received times are not in id order, so the
// newest records do not give the highest id.

// InboxSpec is the inbox collection.
var InboxSpec = store.CollectionSpec{
	Name:   "inbox",
	Index:  []store.Field{{Name: "source", Path: "source"}},
	Unique: []store.Field{{Name: "delivery_id", Path: "delivery_id"}, {Name: "natural_key", Path: "natural_key"}},
}

// InboxStore is the part of the store the inbox needs.
type InboxStore interface {
	Collection(ctx context.Context, spec store.CollectionSpec) (store.Collection, error)
}

type inboxValue struct {
	ID          int64           `json:"id"`
	Source      string          `json:"source"`
	DeliveryID  string          `json:"delivery_id"`
	Endpoint    string          `json:"endpoint"`
	NaturalKey  string          `json:"natural_key"`
	Kind        string          `json:"kind"`
	RepoOrTeam  string          `json:"repo_or_team"`
	Number      int64           `json:"number"`
	CommentID   string          `json:"comment_id"`
	Actor       string          `json:"actor"`
	Payload     json.RawMessage `json:"payload"`
	Admitted    json.RawMessage `json:"admitted,omitempty"`
	HookResults json.RawMessage `json:"hook_results"`
}

// IDSpec is the collection holding the inbox id counter: one record,
// key "next", value {"next": n}.
var IDSpec = store.CollectionSpec{Name: "inbox_ids"}

type idValue struct {
	Next int64 `json:"next"`
}

const (
	// inboxTries bounds the id and compare-and-set loops under races.
	inboxTries = 10
	// inboxListLimit is List's default limit.
	inboxListLimit = 100
)

// terminal states are pruned.
var terminal = []string{store.InboxDispatched, store.InboxDenied, store.InboxFailed}

func inboxCol(ctx context.Context, st InboxStore) (store.Collection, error) {
	return st.Collection(ctx, InboxSpec)
}

func inboxKey(id int64) string { return strconv.FormatInt(id, 10) }

// InboxPut stores an ingested item, deduped by delivery id and then natural
// key (#321, #335). A repeat returns the stored item and dup true.
func InboxPut(ctx context.Context, st InboxStore, it store.InboxItem) (store.InboxItem, bool, error) {
	if err := checkInbox(it); err != nil {
		return store.InboxItem{}, false, err
	}
	col, ids, err := cols(ctx, st)
	if err != nil {
		return store.InboxItem{}, false, err
	}
	return put(ctx, col, ids, it)
}

func cols(ctx context.Context, st InboxStore) (col, ids store.Collection, err error) {
	if col, err = inboxCol(ctx, st); err != nil {
		return nil, nil, err
	}
	ids, err = st.Collection(ctx, IDSpec)
	return col, ids, err
}

// InboxPutBatch stores items, each deduped like InboxPut, and returns the
// new ones as stored (with their ids): the targeted wait wake reads
// their subjects (#679). It is not atomic: a failure part way leaves the
// earlier items stored, and a retry dedupes them.
func InboxPutBatch(ctx context.Context, st InboxStore, its []store.InboxItem) ([]store.InboxItem, error) {
	for _, it := range its {
		if err := checkInbox(it); err != nil {
			return nil, err
		}
	}
	col, ids, err := cols(ctx, st)
	if err != nil {
		return nil, err
	}
	var fresh []store.InboxItem
	for _, it := range its {
		got, dup, err := put(ctx, col, ids, it)
		if err != nil {
			return fresh, err
		}
		if !dup {
			fresh = append(fresh, got)
		}
	}
	return fresh, nil
}

func checkInbox(it store.InboxItem) error {
	if it.DeliveryID == "" {
		return fmt.Errorf("%w: inbox item needs a delivery id", store.ErrInvalid)
	}
	if it.Source == "" {
		return fmt.Errorf("%w: inbox item needs a source", store.ErrInvalid)
	}
	return nil
}

// put returns the stored copy of a known delivery id or natural key,
// else inserts it under a new id. A put that still clashes lost a race
// to a concurrent put of the same item, or found its id taken: the
// record it returns says which.
func put(ctx context.Context, col, ids store.Collection, it store.InboxItem) (store.InboxItem, bool, error) {
	if got, ok, err := existing(ctx, col, it); err != nil || ok {
		return got, ok, err
	}
	v := toInboxValue(it)
	for range inboxTries {
		id, err := nextID(ctx, ids)
		if err != nil {
			return store.InboxItem{}, false, err
		}
		v.ID = id
		b, err := json.Marshal(v)
		if err != nil {
			return store.InboxItem{}, false, err
		}
		r, created, err := col.Put(ctx, store.Record{Key: inboxKey(id), State: store.InboxPending, Value: b, CreatedAt: it.ReceivedAt})
		if err != nil {
			return store.InboxItem{}, false, err
		}
		got, err := fromInbox(r)
		if err != nil {
			return store.InboxItem{}, false, err
		}
		if created {
			return got, false, nil
		}
		if got.DeliveryID == it.DeliveryID || (it.NaturalKey != "" && got.NaturalKey == it.NaturalKey) {
			return got, true, nil
		}
	}
	return store.InboxItem{}, false, fmt.Errorf("%w: inbox: lost every id race", store.ErrConflict)
}

// existing finds the stored item with its delivery id, else its
// natural key.
func existing(ctx context.Context, col store.Collection, it store.InboxItem) (store.InboxItem, bool, error) {
	wheres := []map[string]string{{"delivery_id": it.DeliveryID}}
	if it.NaturalKey != "" {
		wheres = append(wheres, map[string]string{"natural_key": it.NaturalKey})
	}
	for _, w := range wheres {
		rs, err := col.List(ctx, store.RecordFilter{Where: w, Limit: 1})
		if err != nil {
			return store.InboxItem{}, false, err
		}
		if len(rs) > 0 {
			got, err := fromInbox(rs[0])
			return got, err == nil, err
		}
	}
	return store.InboxItem{}, false, nil
}

// nextID takes the next id from the counter.
func nextID(ctx context.Context, ids store.Collection) (int64, error) {
	for range inboxTries {
		r, err := ids.Get(ctx, "next")
		if errors.Is(err, store.ErrNotFound) {
			if _, _, err := ids.Put(ctx, store.Record{Key: "next", Value: json.RawMessage(`{"next":1}`)}); err != nil {
				return 0, err
			}
			continue
		}
		if err != nil {
			return 0, err
		}
		var v idValue
		if err := json.Unmarshal(r.Value, &v); err != nil {
			return 0, fmt.Errorf("inbox id counter: %w", err)
		}
		id := max(v.Next, 1)
		if r.Value, err = json.Marshal(idValue{Next: id + 1}); err != nil {
			return 0, err
		}
		_, err = ids.Update(ctx, r)
		if errors.Is(err, store.ErrConflict) {
			continue
		}
		return id, err
	}
	return 0, fmt.Errorf("%w: inbox id counter: lost every race", store.ErrConflict)
}

// InboxClaim leases the oldest item in one of states whose lease is empty or
// expired, adding one to attempts. ErrNotFound if none.
func InboxClaim(ctx context.Context, st InboxStore, states []string, holder string, lease time.Duration) (store.InboxItem, error) {
	if holder == "" || lease <= 0 || len(states) == 0 {
		return store.InboxItem{}, fmt.Errorf("%w: inbox claim needs states, a holder and a lease", store.ErrInvalid)
	}
	if err := checkStates(states...); err != nil {
		return store.InboxItem{}, err
	}
	col, err := inboxCol(ctx, st)
	if err != nil {
		return store.InboxItem{}, err
	}
	r, err := col.Claim(ctx, store.RecordFilter{States: states}, holder, lease)
	if err != nil {
		return store.InboxItem{}, err
	}
	return fromInbox(r)
}

// InboxRenew extends holder's live lease; ErrConflict if not held.
func InboxRenew(ctx context.Context, st InboxStore, id int64, holder string, lease time.Duration) (store.InboxItem, error) {
	col, err := inboxCol(ctx, st)
	if err != nil {
		return store.InboxItem{}, err
	}
	r, err := col.Renew(ctx, inboxKey(id), holder, lease)
	if err != nil {
		return store.InboxItem{}, conflict(err)
	}
	return fromInbox(r)
}

// InboxSetState moves a held item to u.State and drops the lease. Nil
// Admitted or HookResults keep the stored value. ErrConflict if holder
// no longer holds a live lease.
func InboxSetState(ctx context.Context, st InboxStore, id int64, holder string, u store.InboxUpdate) (store.InboxItem, error) {
	if err := checkStates(u.State); err != nil {
		return store.InboxItem{}, err
	}
	if holder == "" {
		return store.InboxItem{}, fmt.Errorf("%w: inbox set state needs a holder", store.ErrInvalid)
	}
	return update(ctx, st, id, func(r *store.Record, v *inboxValue) error {
		r.State, r.Holder = u.State, holder
		if u.Admitted != nil {
			v.Admitted = u.Admitted
		}
		if u.HookResults != nil {
			v.HookResults = u.HookResults
		}
		return nil
	})
}

// InboxRequeue moves an item no one leases from state from to state to
// (held -> pending once a human answered, #353). ErrConflict if it is
// not in from or is leased.
func InboxRequeue(ctx context.Context, st InboxStore, id int64, from, to string) (store.InboxItem, error) {
	if err := checkStates(from, to); err != nil {
		return store.InboxItem{}, err
	}
	return update(ctx, st, id, func(r *store.Record, _ *inboxValue) error {
		if r.State != from {
			return fmt.Errorf("%w: inbox item %d is not %s", store.ErrConflict, id, from)
		}
		// No holder: the update fails with a conflict while a live
		// lease is on the item.
		r.State, r.Holder = to, ""
		return nil
	})
}

// update reads id, applies fn and writes it back with the lease
// cleared, retrying a lost compare-and-set.
func update(ctx context.Context, st InboxStore, id int64, fn func(*store.Record, *inboxValue) error) (store.InboxItem, error) {
	col, err := inboxCol(ctx, st)
	if err != nil {
		return store.InboxItem{}, err
	}
	for range inboxTries {
		r, err := col.Get(ctx, inboxKey(id))
		if err != nil {
			return store.InboxItem{}, err
		}
		var v inboxValue
		if err := json.Unmarshal(r.Value, &v); err != nil {
			return store.InboxItem{}, fmt.Errorf("inbox item %d: %w", id, err)
		}
		if err := fn(&r, &v); err != nil {
			return store.InboxItem{}, err
		}
		r.LeaseUntil = time.Time{}
		if r.Value, err = json.Marshal(v); err != nil {
			return store.InboxItem{}, err
		}
		r, err = col.Update(ctx, r)
		if errors.Is(err, store.ErrConflict) {
			continue
		}
		if err != nil {
			return store.InboxItem{}, conflict(err)
		}
		return fromInbox(r)
	}
	return store.InboxItem{}, fmt.Errorf("%w: inbox item %d: lost every race", store.ErrConflict, id)
}

// conflict reports a lost lease as ErrConflict, as the inbox always has.
func conflict(err error) error {
	if errors.Is(err, store.ErrNotHolder) {
		return fmt.Errorf("%w: %w", store.ErrConflict, err)
	}
	return err
}

func checkStates(states ...string) error {
	for _, s := range states {
		if !slices.Contains(store.InboxStates, s) {
			return fmt.Errorf("%w: unknown inbox state %q", store.ErrInvalid, s)
		}
	}
	return nil
}

// InboxGet returns one item.
func InboxGet(ctx context.Context, st InboxStore, id int64) (store.InboxItem, error) {
	col, err := inboxCol(ctx, st)
	if err != nil {
		return store.InboxItem{}, err
	}
	r, err := col.Get(ctx, inboxKey(id))
	if err != nil {
		return store.InboxItem{}, err
	}
	return fromInbox(r)
}

// InboxList lists items newest first; Limit 0 is 100.
func InboxList(ctx context.Context, st InboxStore, f store.InboxFilter) ([]store.InboxItem, error) {
	col, err := inboxCol(ctx, st)
	if err != nil {
		return nil, err
	}
	rf := store.RecordFilter{Order: store.OrderCreatedDesc, Limit: cmp.Or(f.Limit, inboxListLimit)}
	if f.State != "" {
		rf.States = []string{f.State}
	}
	if f.Source != "" {
		rf.Where = map[string]string{"source": f.Source}
	}
	rs, err := col.List(ctx, rf)
	if err != nil {
		return nil, err
	}
	out := make([]store.InboxItem, 0, len(rs))
	for _, r := range rs {
		it, err := fromInbox(r)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	// Ties in received_at: newest id first, as the table ordered. Keys
	// are decimal text, so the store's tie order is not numeric.
	slices.SortStableFunc(out, func(a, b store.InboxItem) int {
		return cmp.Or(b.ReceivedAt.Compare(a.ReceivedAt), cmp.Compare(b.ID, a.ID))
	})
	return out, nil
}

// inboxPruneBatch bounds one prune delete, so a large backlog (JOB-63:
// 7k CI records a day) is removed in short write transactions instead
// of one that holds the lock.
const inboxPruneBatch = 200

// InboxPrune deletes terminal items last updated before before and
// returns how many. Pending, admitted and held items are never pruned.
// Dedup by delivery id and natural key covers only items still stored,
// so a redelivery older than the retention is taken as new (JOB-63).
func InboxPrune(ctx context.Context, st InboxStore, before time.Time) (int64, error) {
	col, err := inboxCol(ctx, st)
	if err != nil {
		return 0, err
	}
	f := store.RecordFilter{States: terminal, UpdatedBefore: before, Order: store.OrderUpdated, Limit: inboxPruneBatch}
	var total int64
	for {
		recs, err := col.List(ctx, f)
		if err != nil || len(recs) == 0 {
			return total, err
		}
		keys := make([]string, len(recs))
		for i, r := range recs {
			keys[i] = r.Key
		}
		// The state and age filters stay on the delete: a record that
		// changed since the list is kept.
		n, err := col.Delete(ctx, store.RecordFilter{Keys: keys, States: terminal, UpdatedBefore: before})
		total += int64(n)
		if err != nil || len(recs) < inboxPruneBatch || n == 0 {
			return total, err
		}
	}
}

// InboxCounts counts items by state, every state present (#443).
func InboxCounts(ctx context.Context, st InboxStore) (map[string]int64, error) {
	col, err := inboxCol(ctx, st)
	if err != nil {
		return nil, err
	}
	m, err := col.Count(ctx, store.RecordFilter{})
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(store.InboxStates))
	for _, s := range store.InboxStates {
		out[s] = 0
	}
	for s, n := range m {
		out[s] = int64(n)
	}
	return out, nil
}

// InboxCountOlder counts items in state received before before (#443).
func InboxCountOlder(ctx context.Context, st InboxStore, state string, before time.Time) (int64, error) {
	col, err := inboxCol(ctx, st)
	if err != nil {
		return 0, err
	}
	m, err := col.Count(ctx, store.RecordFilter{States: []string{state}, Until: before})
	if err != nil {
		return 0, err
	}
	return int64(m[state]), nil
}

func toInboxValue(it store.InboxItem) inboxValue {
	v := inboxValue{
		Source: it.Source, DeliveryID: it.DeliveryID, Endpoint: it.Endpoint, NaturalKey: it.NaturalKey,
		Kind: it.Kind, RepoOrTeam: it.RepoOrTeam, Number: it.Number, CommentID: it.CommentID, Actor: it.Actor,
		Payload: it.Payload, HookResults: json.RawMessage("[]"),
	}
	if len(v.Payload) == 0 {
		v.Payload = json.RawMessage("{}")
	}
	return v
}

func fromInbox(r store.Record) (store.InboxItem, error) {
	var v inboxValue
	if err := json.Unmarshal(r.Value, &v); err != nil {
		return store.InboxItem{}, fmt.Errorf("inbox item %s: %w", r.Key, err)
	}
	it := store.InboxItem{
		ID: v.ID, Source: v.Source, DeliveryID: v.DeliveryID, Endpoint: v.Endpoint, NaturalKey: v.NaturalKey,
		Kind: v.Kind, RepoOrTeam: v.RepoOrTeam, Number: v.Number, CommentID: v.CommentID, Actor: v.Actor,
		ReceivedAt: r.CreatedAt, Payload: v.Payload, State: r.State, HookResults: v.HookResults,
		Holder: r.Holder, LeaseUntil: r.LeaseUntil, Attempts: r.Attempts, UpdatedAt: r.UpdatedAt,
	}
	if len(v.Admitted) > 0 && string(v.Admitted) != "null" {
		it.Admitted = v.Admitted
	}
	return it, nil
}
