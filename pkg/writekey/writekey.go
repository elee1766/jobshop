// Package writekey stores idempotency keys for tracker writes (#280) on
// the write_keys store collection (#579). A write first claims its key
// pending, then stores the result and marks it done. A retry that finds
// a done key returns the result; one that finds a pending key reads the
// entity back instead of posting again.
package writekey

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/store"
)

// Spec is the write_keys collection. The record key is the write key.
var Spec = store.CollectionSpec{Name: "write_keys"}

// Record states.
const (
	Pending = "pending"
	Done    = "done"
)

// value is the record's JSON.
type value struct {
	Op     string `json:"op"`
	Result string `json:"result"`
	Actor  string `json:"actor"`
}

// Store is the part of the store write keys need.
type Store interface {
	Collection(ctx context.Context, spec store.CollectionSpec) (store.Collection, error)
}

// maxTries bounds the read-update loop of Finish against concurrent
// writers of the same key.
const maxTries = 10

// Claim returns the write stored under key, inserting it pending when
// new. A key reused for another op is ErrInvalid.
func Claim(ctx context.Context, st Store, key, op, actor string) (store.WriteKey, error) {
	key, op = strings.TrimSpace(key), strings.TrimSpace(op)
	if key == "" || op == "" {
		return store.WriteKey{}, fmt.Errorf("%w: write key needs a key and an op", domain.ErrInvalid)
	}
	col, err := st.Collection(ctx, Spec)
	if err != nil {
		return store.WriteKey{}, err
	}
	v, err := json.Marshal(value{Op: op, Actor: actor})
	if err != nil {
		return store.WriteKey{}, err
	}
	r, created, err := col.Put(ctx, store.Record{Key: key, State: Pending, Value: v})
	if err != nil {
		return store.WriteKey{}, err
	}
	out, err := fromRecord(r)
	if err != nil {
		return out, err
	}
	if out.Op != op {
		return out, fmt.Errorf("%w: write key %q was used for %s, not %s", domain.ErrInvalid, key, out.Op, op)
	}
	out.New = created
	return out, nil
}

// Finish stores the result of a pending write. A key already done keeps
// its first result, which is returned.
func Finish(ctx context.Context, st Store, key, result string) (store.WriteKey, error) {
	col, err := st.Collection(ctx, Spec)
	if err != nil {
		return store.WriteKey{}, err
	}
	for range maxTries {
		r, err := col.Get(ctx, key)
		if err != nil {
			return store.WriteKey{}, err
		}
		if r.State == Done {
			return fromRecord(r)
		}
		var v value
		if err := json.Unmarshal(r.Value, &v); err != nil {
			return store.WriteKey{}, err
		}
		v.Result = result
		if r.Value, err = json.Marshal(v); err != nil {
			return store.WriteKey{}, err
		}
		r.State = Done
		r, err = col.Update(ctx, r)
		if errors.Is(err, domain.ErrConflict) {
			continue // another writer finished or touched it; reread
		}
		if err != nil {
			return store.WriteKey{}, err
		}
		return fromRecord(r)
	}
	return store.WriteKey{}, fmt.Errorf("%w: write key %q: finish lost every race", domain.ErrConflict, key)
}

// Prune deletes keys created before before, done or pending (#415): the
// local collection is a cache, the remote key or marker is the record.
func Prune(ctx context.Context, st Store, before time.Time) (int64, error) {
	col, err := st.Collection(ctx, Spec)
	if err != nil {
		return 0, err
	}
	n, err := col.Delete(ctx, store.RecordFilter{Until: before})
	return int64(n), err
}

func fromRecord(r store.Record) (store.WriteKey, error) {
	var v value
	if err := json.Unmarshal(r.Value, &v); err != nil {
		return store.WriteKey{}, fmt.Errorf("write key %q: %w", r.Key, err)
	}
	return store.WriteKey{
		Key: r.Key, Op: v.Op, Result: v.Result, Done: r.State == Done, Actor: v.Actor,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}, nil
}
