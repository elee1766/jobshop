package storetest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// TestSpec is the collection spec the conformance cases use. Backends
// reuse it in their own tests (index plans, partitions).
var TestSpec = store.CollectionSpec{
	Name:   "test_items",
	Index:  []store.Field{{Name: "source", Path: "source"}, {Name: "repo", Path: "meta.repo"}, {Name: "pr", Path: "pr"}},
	Unique: []store.Field{{Name: "natural_key", Path: "natural_key"}},
	Search: []store.Field{{Name: "summary", Path: "summary"}, {Name: "fix", Path: "fix"}},
}

func coll(t *testing.T, s store.Store) store.Collection {
	t.Helper()
	c, err := s.Collection(t.Context(), TestSpec)
	require.NoError(t, err)
	return c
}

func val(t *testing.T, v map[string]any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func fields(t *testing.T, r store.Record) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(r.Value, &m))
	return m
}

func keys(rs []store.Record) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.Key
	}
	return out
}

func testCollectionSpec(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	for _, bad := range []store.CollectionSpec{
		{Name: ""},
		{Name: "Bad"},
		{Name: "1x"},
		{Name: "x; drop table records"},
		{Name: "a234567890123456789012345678901234567890x"},
		{Name: "ok", Index: []store.Field{{Name: "f", Path: "a'b"}}},
		{Name: "ok", Index: []store.Field{{Name: "F", Path: "f"}}},
		{Name: "ok", Index: []store.Field{{Name: "f", Path: "f"}}, Unique: []store.Field{{Name: "f", Path: "g"}}},
	} {
		_, err := s.Collection(ctx, bad)
		assert.ErrorIs(t, err, domain.ErrInvalid, "%+v", bad)
	}
	// Registration is idempotent, also with a field added later.
	coll(t, s)
	coll(t, s)
	more := TestSpec
	more.Index = append([]store.Field{{Name: "kind", Path: "kind"}}, TestSpec.Index...)
	_, err := s.Collection(ctx, more)
	require.NoError(t, err)
}

func testCollectionPutGet(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	col := coll(t, s)
	r, created, err := col.Put(ctx, store.Record{Key: "a", State: "pending", Value: val(t, map[string]any{"source": "github", "n": 12})})
	require.NoError(t, err)
	assert.True(t, created)
	assert.Equal(t, TestSpec.Name, r.Collection)
	assert.Equal(t, int64(1), r.Revision)
	assert.Equal(t, c.T, r.CreatedAt)
	assert.Equal(t, c.T, r.UpdatedAt)
	assert.Equal(t, "github", fields(t, r)["source"])

	// A repeated key is a no-op returning the stored record.
	c.Advance(time.Minute)
	r2, created, err := col.Put(ctx, store.Record{Key: "a", State: "other", Value: val(t, map[string]any{"source": "linear"})})
	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, "pending", r2.State)
	assert.Equal(t, "github", fields(t, r2)["source"])

	got, err := col.Get(ctx, "a")
	require.NoError(t, err)
	assert.Equal(t, r.Key, got.Key)
	assert.InDelta(t, 12, fields(t, got)["n"], 0)

	_, err = col.Get(ctx, "missing")
	require.ErrorIs(t, err, domain.ErrNotFound)

	// Created time may be given, for migrated or received-at times.
	at := c.T.Add(-time.Hour)
	r3, _, err := col.Put(ctx, store.Record{Key: "b", CreatedAt: at})
	require.NoError(t, err)
	assert.Equal(t, at, r3.CreatedAt)
	assert.JSONEq(t, `{}`, string(r3.Value))

	_, _, err = col.Put(ctx, store.Record{Key: ""})
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, _, err = col.Put(ctx, store.Record{Key: "c", Value: json.RawMessage(`[1]`)})
	require.ErrorIs(t, err, domain.ErrInvalid)
}

func testCollectionUnique(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	col := coll(t, s)
	first, created, err := col.Put(ctx, store.Record{Key: "d1", Value: val(t, map[string]any{"natural_key": "github:comment:1"})})
	require.NoError(t, err)
	require.True(t, created)
	// Another key with the same unique value dedupes to the first.
	got, created, err := col.Put(ctx, store.Record{Key: "d2", Value: val(t, map[string]any{"natural_key": "github:comment:1"})})
	require.NoError(t, err)
	assert.False(t, created)
	assert.Equal(t, first.Key, got.Key)
	_, err = col.Get(ctx, "d2")
	require.ErrorIs(t, err, domain.ErrNotFound)
	// Empty unique values do not collide.
	for _, k := range []string{"e1", "e2"} {
		_, created, err := col.Put(ctx, store.Record{Key: k, Value: val(t, map[string]any{"natural_key": ""})})
		require.NoError(t, err)
		assert.True(t, created, k)
	}
	rs, err := col.List(ctx, store.RecordFilter{Where: map[string]string{"natural_key": "github:comment:1"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"d1"}, keys(rs))
	// Updating another record onto a taken unique value fails.
	e1, err := col.Get(ctx, "e1")
	require.NoError(t, err)
	e1.Value = val(t, map[string]any{"natural_key": "github:comment:1"})
	_, err = col.Update(ctx, e1)
	require.ErrorIs(t, err, domain.ErrAlreadyExist)
}

func testCollectionUpdate(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	col := coll(t, s)
	r, _, err := col.Put(ctx, store.Record{Key: "u", State: "pending"})
	require.NoError(t, err)
	c.Advance(time.Second)
	r.State = "done"
	r.Value = val(t, map[string]any{"x": "y"})
	r2, err := col.Update(ctx, r)
	require.NoError(t, err)
	assert.Equal(t, int64(2), r2.Revision)
	assert.Equal(t, "done", r2.State)
	assert.Equal(t, c.T, r2.UpdatedAt)
	assert.Equal(t, r.CreatedAt, r2.CreatedAt)
	// The stale revision conflicts.
	_, err = col.Update(ctx, r)
	require.ErrorIs(t, err, domain.ErrConflict)
	_, err = col.Update(ctx, store.Record{Key: "nope", Revision: 1})
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func testCollectionClaim(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	col := coll(t, s)
	for i, k := range []string{"c3", "c1", "c2"} {
		_, _, err := col.Put(ctx, store.Record{Key: k, State: "pending", CreatedAt: c.T.Add(time.Duration(i) * time.Second)})
		require.NoError(t, err)
	}
	_, _, err := col.Put(ctx, store.Record{Key: "c0", State: "done", CreatedAt: c.T.Add(-time.Hour)})
	require.NoError(t, err)
	pending := store.RecordFilter{States: []string{"pending"}}

	_, err = col.Claim(ctx, store.RecordFilter{}, "w", time.Minute)
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = col.Claim(ctx, pending, "", time.Minute)
	require.ErrorIs(t, err, domain.ErrInvalid)

	// Oldest created first, not key order.
	a, err := col.Claim(ctx, pending, "w1", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, "c3", a.Key)
	assert.Equal(t, "w1", a.Holder)
	assert.Equal(t, int64(1), a.Attempts)
	assert.Equal(t, int64(2), a.Revision)
	assert.Equal(t, c.T.Add(time.Minute), a.LeaseUntil)
	b, err := col.Claim(ctx, pending, "w2", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, "c1", b.Key)

	// The holder writes under its lease; others may not.
	other := a
	other.Holder = "w2"
	_, err = col.Update(ctx, other)
	require.ErrorIs(t, err, domain.ErrNotHolder)
	unheld := a
	unheld.Holder = ""
	_, err = col.Update(ctx, unheld)
	require.ErrorIs(t, err, domain.ErrNotHolder)

	// Renew extends a live lease without bumping the revision.
	c.Advance(30 * time.Second)
	rn, err := col.Renew(ctx, "c3", "w1", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, c.T.Add(time.Minute), rn.LeaseUntil)
	assert.Equal(t, a.Revision, rn.Revision)
	_, err = col.Renew(ctx, "c3", "w2", time.Minute)
	require.ErrorIs(t, err, domain.ErrNotHolder)
	_, err = col.Renew(ctx, "missing", "w1", time.Minute)
	require.ErrorIs(t, err, domain.ErrNotFound)

	// The holder finishes, keeping the lease fields as stored.
	rn.State = "processing"
	kept, err := col.Update(ctx, rn)
	require.NoError(t, err)
	assert.Equal(t, "w1", kept.Holder)
	// Then releases: a zero lease clears the holder.
	kept.State = "done"
	kept.LeaseUntil = time.Time{}
	done, err := col.Update(ctx, kept)
	require.NoError(t, err)
	assert.Empty(t, done.Holder)
	assert.True(t, done.LeaseUntil.IsZero())

	// An expired lease is claimable again, and its holder is fenced.
	c.Advance(2 * time.Minute)
	again, err := col.Claim(ctx, pending, "w3", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, "c1", again.Key)
	assert.Equal(t, int64(2), again.Attempts)
	_, err = col.Renew(ctx, "c1", "w2", time.Minute)
	require.ErrorIs(t, err, domain.ErrNotHolder)

	next, err := col.Claim(ctx, pending, "w4", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, "c2", next.Key)
	_, err = col.Claim(ctx, pending, "w5", time.Minute)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func testCollectionListCount(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	col := coll(t, s)
	t0 := c.T
	put := func(k, st string, at time.Duration, v map[string]any) {
		t.Helper()
		_, _, err := col.Put(ctx, store.Record{Key: k, State: st, CreatedAt: t0.Add(at), Value: val(t, v)})
		require.NoError(t, err)
	}
	put("l1", "pending", 0, map[string]any{"source": "github", "pr": 12, "meta": map[string]any{"repo": "a/b"}})
	put("l2", "done", time.Second, map[string]any{"source": "linear", "pr": "12"})
	put("l3", "pending", 2*time.Second, map[string]any{"source": "github", "meta": map[string]any{"repo": "c/d"}})
	put("l4", "failed", 3*time.Second, map[string]any{"source": "github"})

	list := func(f store.RecordFilter) []string {
		t.Helper()
		rs, err := col.List(ctx, f)
		require.NoError(t, err)
		return keys(rs)
	}
	assert.Equal(t, []string{"l1", "l2", "l3", "l4"}, list(store.RecordFilter{}))
	assert.Equal(t, []string{"l4", "l3", "l2", "l1"}, list(store.RecordFilter{Order: store.OrderCreatedDesc}))
	assert.Equal(t, []string{"l1", "l2"}, list(store.RecordFilter{Limit: 2}))
	assert.Equal(t, []string{"l1", "l3"}, list(store.RecordFilter{States: []string{"pending"}}))
	assert.Equal(t, []string{"l3"}, list(store.RecordFilter{Keys: []string{"l3", "l2"}, States: []string{"pending"}}))
	assert.Equal(t, []string{"l1", "l3", "l4"}, list(store.RecordFilter{Where: map[string]string{"source": "github"}}))
	assert.Equal(t, []string{"l3"}, list(store.RecordFilter{Where: map[string]string{"repo": "c/d"}}))
	// Numbers and strings compare as text.
	assert.Equal(t, []string{"l1", "l2"}, list(store.RecordFilter{Where: map[string]string{"pr": "12"}}))
	assert.Equal(t, []string{"l2", "l3"}, list(store.RecordFilter{Since: t0.Add(time.Second), Until: t0.Add(3 * time.Second)}))

	_, err := col.List(ctx, store.RecordFilter{Where: map[string]string{"undeclared": "x"}})
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = col.List(ctx, store.RecordFilter{Order: "key"})
	require.ErrorIs(t, err, domain.ErrInvalid)

	n, err := col.Count(ctx, store.RecordFilter{})
	require.NoError(t, err)
	assert.Equal(t, map[string]int{"pending": 2, "done": 1, "failed": 1}, n)
	n, err = col.Count(ctx, store.RecordFilter{States: []string{"pending"}, Until: t0.Add(time.Second)})
	require.NoError(t, err)
	assert.Equal(t, map[string]int{"pending": 1}, n)

	// Collections do not see each other's records.
	other, err := s.Collection(ctx, store.CollectionSpec{Name: "test_other"})
	require.NoError(t, err)
	_, _, err = other.Put(ctx, store.Record{Key: "l1", State: "x"})
	require.NoError(t, err)
	rs, err := other.List(ctx, store.RecordFilter{})
	require.NoError(t, err)
	assert.Equal(t, []string{"l1"}, keys(rs))
	assert.Equal(t, "x", rs[0].State)
	assert.Len(t, list(store.RecordFilter{}), 4)
}

func testCollectionDelete(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	col := coll(t, s)
	_, _, err := col.Put(ctx, store.Record{Key: "old", State: "done", Value: val(t, map[string]any{"summary": "vendor hash mismatch"})})
	require.NoError(t, err)
	c.Advance(time.Hour)
	_, _, err = col.Put(ctx, store.Record{Key: "new", State: "done", Value: val(t, map[string]any{"summary": "vendor hash again"})})
	require.NoError(t, err)
	_, _, err = col.Put(ctx, store.Record{Key: "live", State: "pending"})
	require.NoError(t, err)

	_, err = col.Delete(ctx, store.RecordFilter{})
	require.ErrorIs(t, err, domain.ErrInvalid)
	n, err := col.Delete(ctx, store.RecordFilter{States: []string{"done"}, UpdatedBefore: c.T.Add(-time.Minute)})
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	rs, err := col.List(ctx, store.RecordFilter{})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"new", "live"}, keys(rs))
	// The deleted record left the search index too.
	hits, err := col.Search(ctx, "vendor", 10)
	require.NoError(t, err)
	assert.Equal(t, []string{"new"}, keys(hits))
}

func testCollectionSearch(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	col := coll(t, s)
	_, _, err := col.Put(ctx, store.Record{Key: "s1", Value: val(t, map[string]any{
		"summary": "vendor hash mismatch once, then lots of other text about reviewing and merging and waiting",
	})})
	require.NoError(t, err)
	r2, _, err := col.Put(ctx, store.Record{Key: "s2", Value: val(t, map[string]any{
		"summary": "nix vendor hash mismatch", "fix": "set vendorHash to the got value",
	})})
	require.NoError(t, err)
	_, _, err = col.Put(ctx, store.Record{Key: "s3", Value: val(t, map[string]any{"summary": "unrelated weather"})})
	require.NoError(t, err)

	hits, err := col.Search(ctx, "vendor hash mismatch", 10)
	require.NoError(t, err)
	assert.Equal(t, []string{"s2", "s1"}, keys(hits))
	// The fix field is searched too, and query syntax is text.
	hits, err = col.Search(ctx, `vendorHash "got" value: & |`, 10)
	require.NoError(t, err)
	assert.Equal(t, []string{"s2"}, keys(hits))
	// Any word, when no record has all.
	hits, err = col.Search(ctx, "weather zebra", 10)
	require.NoError(t, err)
	assert.Equal(t, []string{"s3"}, keys(hits))

	// An update reindexes.
	r2.Value = val(t, map[string]any{"summary": "something else"})
	_, err = col.Update(ctx, r2)
	require.NoError(t, err)
	hits, err = col.Search(ctx, "vendor", 10)
	require.NoError(t, err)
	assert.Equal(t, []string{"s1"}, keys(hits))

	_, err = col.Search(ctx, " ", 10)
	require.ErrorIs(t, err, domain.ErrInvalid)
	plain, err := s.Collection(ctx, store.CollectionSpec{Name: "test_plain"})
	require.NoError(t, err)
	_, err = plain.Search(ctx, "x", 10)
	require.ErrorIs(t, err, domain.ErrInvalid)
}

// testCollectionConcurrentClaims checks no record is claimed twice when
// claimers race (#406).
func testCollectionConcurrentClaims(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	col := coll(t, s)
	const n = 20
	for i := range n {
		_, _, err := col.Put(ctx, store.Record{Key: fmt.Sprintf("k%02d", i), State: "pending"})
		require.NoError(t, err)
	}
	var (
		mu  sync.Mutex
		got = map[string]string{}
		wg  sync.WaitGroup
	)
	for w := range 8 {
		wg.Go(func() {
			holder := fmt.Sprintf("w%d", w)
			for {
				r, err := col.Claim(context.WithoutCancel(ctx), store.RecordFilter{States: []string{"pending"}}, holder, time.Hour)
				if errors.Is(err, domain.ErrNotFound) {
					return
				}
				if errors.Is(err, domain.ErrConflict) {
					continue
				}
				if !assert.NoError(t, err) {
					return
				}
				mu.Lock()
				prev, dup := got[r.Key]
				got[r.Key] = holder
				mu.Unlock()
				assert.False(t, dup, "%s claimed by %s and %s", r.Key, prev, holder)
			}
		})
	}
	wg.Wait()
	assert.Len(t, got, n)
}
