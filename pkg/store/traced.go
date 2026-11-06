package store

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/elee1766/jobshop/pkg/telemetry"
)

// SlowCall is how long a store call must take before Traced gives it a
// span. Most calls take microseconds and run thousands of times a tick;
// the slow ones are what a trace needs to show (JOB-105).
const SlowCall = 50 * time.Millisecond

// Traced wraps s so a call that takes at least SlowCall adds a child span
// to the span in its ctx, named "store <op>" with the operation and, for
// a collection, its name. Arguments never go on a span.
func Traced(s Store) Store { return &tracedStore{Store: s} }

type tracedStore struct{ Store }

// slow records the call op that began at start.
func slow(ctx context.Context, op, coll string, start time.Time, err error) {
	attrs := []attribute.KeyValue{telemetry.AttrStoreOp.String(op)}
	name := "store " + op
	if coll != "" {
		attrs = append(attrs, telemetry.AttrStoreColl.String(coll))
		name = "store " + coll + "." + op
	}
	telemetry.RecordSlow(ctx, name, start, SlowCall, err, attrs...)
}

func (t *tracedStore) Get(ctx context.Context, id string) (Ticket, error) {
	start := time.Now()
	r, err := t.Store.Get(ctx, id)
	slow(ctx, "get", "", start, err)
	return r, err
}

func (t *tracedStore) List(ctx context.Context, f Filter) ([]Ticket, error) {
	start := time.Now()
	r, err := t.Store.List(ctx, f)
	slow(ctx, "list", "", start, err)
	return r, err
}

func (t *tracedStore) Update(ctx context.Context, actor, id string, u Update) (Ticket, error) {
	start := time.Now()
	r, err := t.Store.Update(ctx, actor, id, u)
	slow(ctx, "update", "", start, err)
	return r, err
}

func (t *tracedStore) SetResources(ctx context.Context, rc ResourceConfig) error {
	start := time.Now()
	err := t.Store.SetResources(ctx, rc)
	slow(ctx, "set_resources", "", start, err)
	return err
}

func (t *tracedStore) Slot(ctx context.Context, actor string, e SlotEvent) (Slot, string, error) {
	start := time.Now()
	sl, id, err := t.Store.Slot(ctx, actor, e)
	slow(ctx, "slot", "", start, err)
	return sl, id, err
}

func (t *tracedStore) ReleaseSlots(ctx context.Context) ([]Slot, error) {
	start := time.Now()
	r, err := t.Store.ReleaseSlots(ctx)
	slow(ctx, "release_slots", "", start, err)
	return r, err
}

func (t *tracedStore) Resources(ctx context.Context, pool string) (ResourceReport, error) {
	start := time.Now()
	r, err := t.Store.Resources(ctx, pool)
	slow(ctx, "resources", "", start, err)
	return r, err
}

func (t *tracedStore) ClaimNext(ctx context.Context, req ClaimNext) (Ticket, Claim, error) {
	start := time.Now()
	tk, c, err := t.Store.ClaimNext(ctx, req)
	slow(ctx, "claim_next", "", start, err)
	return tk, c, err
}

func (t *tracedStore) Claim(ctx context.Context, id, holder string, lease time.Duration) (Ticket, Claim, error) {
	start := time.Now()
	tk, c, err := t.Store.Claim(ctx, id, holder, lease)
	slow(ctx, "claim", "", start, err)
	return tk, c, err
}

func (t *tracedStore) Heartbeat(ctx context.Context, c Claim, lease time.Duration) (Claim, error) {
	start := time.Now()
	r, err := t.Store.Heartbeat(ctx, c, lease)
	slow(ctx, "heartbeat", "", start, err)
	return r, err
}

func (t *tracedStore) Release(ctx context.Context, actor, id string, c *Claim, o Outcome) (Ticket, error) {
	start := time.Now()
	r, err := t.Store.Release(ctx, actor, id, c, o)
	slow(ctx, "release", "", start, err)
	return r, err
}

func (t *tracedStore) Expire(ctx context.Context, actor, holder, reason string) ([]string, error) {
	start := time.Now()
	r, err := t.Store.Expire(ctx, actor, holder, reason)
	slow(ctx, "expire", "", start, err)
	return r, err
}

func (t *tracedStore) Thread(ctx context.Context, id string, p Parts) (Thread, error) {
	start := time.Now()
	r, err := t.Store.Thread(ctx, id, p)
	slow(ctx, "thread", "", start, err)
	return r, err
}

func (t *tracedStore) Record(ctx context.Context, actor, id, op string, diff any) error {
	start := time.Now()
	err := t.Store.Record(ctx, actor, id, op, diff)
	slow(ctx, "record", "", start, err)
	return err
}

func (t *tracedStore) Changes(ctx context.Context, after int64, n int) (ChangePage, error) {
	start := time.Now()
	r, err := t.Store.Changes(ctx, after, n)
	slow(ctx, "changes", "", start, err)
	return r, err
}

func (t *tracedStore) Apply(ctx context.Context, actor string, g Graph) ([]Ticket, error) {
	start := time.Now()
	r, err := t.Store.Apply(ctx, actor, g)
	slow(ctx, "apply", "", start, err)
	return r, err
}

func (t *tracedStore) Recall(ctx context.Context, q RecallQuery) ([]RecallHit, error) {
	start := time.Now()
	r, err := t.Store.Recall(ctx, q)
	slow(ctx, "recall", "", start, err)
	return r, err
}

func (t *tracedStore) Collection(ctx context.Context, spec CollectionSpec) (Collection, error) {
	c, err := t.Store.Collection(ctx, spec)
	if err != nil {
		return nil, err
	}
	return &tracedCollection{Collection: c, name: spec.Name}, nil
}

type tracedCollection struct {
	Collection
	name string
}

func (t *tracedCollection) Put(ctx context.Context, r Record) (Record, bool, error) {
	start := time.Now()
	rec, created, err := t.Collection.Put(ctx, r)
	slow(ctx, "put", t.name, start, err)
	return rec, created, err
}

func (t *tracedCollection) Get(ctx context.Context, key string) (Record, error) {
	start := time.Now()
	r, err := t.Collection.Get(ctx, key)
	slow(ctx, "get", t.name, start, err)
	return r, err
}

func (t *tracedCollection) Update(ctx context.Context, r Record) (Record, error) {
	start := time.Now()
	rec, err := t.Collection.Update(ctx, r)
	slow(ctx, "update", t.name, start, err)
	return rec, err
}

func (t *tracedCollection) List(ctx context.Context, f RecordFilter) ([]Record, error) {
	start := time.Now()
	r, err := t.Collection.List(ctx, f)
	slow(ctx, "list", t.name, start, err)
	return r, err
}

func (t *tracedCollection) Count(ctx context.Context, f RecordFilter) (map[string]int, error) {
	start := time.Now()
	r, err := t.Collection.Count(ctx, f)
	slow(ctx, "count", t.name, start, err)
	return r, err
}

func (t *tracedCollection) Claim(ctx context.Context, f RecordFilter, holder string, lease time.Duration) (Record, error) {
	start := time.Now()
	r, err := t.Collection.Claim(ctx, f, holder, lease)
	slow(ctx, "claim", t.name, start, err)
	return r, err
}

func (t *tracedCollection) Renew(ctx context.Context, key, holder string, lease time.Duration) (Record, error) {
	start := time.Now()
	r, err := t.Collection.Renew(ctx, key, holder, lease)
	slow(ctx, "renew", t.name, start, err)
	return r, err
}

func (t *tracedCollection) Delete(ctx context.Context, f RecordFilter) (int, error) {
	start := time.Now()
	n, err := t.Collection.Delete(ctx, f)
	slow(ctx, "delete", t.name, start, err)
	return n, err
}

func (t *tracedCollection) Search(ctx context.Context, query string, limit int) ([]Record, error) {
	start := time.Now()
	r, err := t.Collection.Search(ctx, query, limit)
	slow(ctx, "search", t.name, start, err)
	return r, err
}
