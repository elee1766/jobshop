// Package knowledge keeps what Jobshop learns and what it writes down
// about it (JOB-170): learnings (#264) and memory edits (#274, #592) on
// store collections (#579), the curator gate's judgement of an edit, and
// the mechanics of recording, evaluating and restoring an edit.
//
// What stays in pkg/orchestration/control is the decision to curate: who may write
// memory, which cap and thresholds apply, when an evaluation is due, and
// the coordination with tickets and runs. This package is given those
// values and does the work.
//
// A local memory write touches a file and then SQL records. They are not
// one transaction: WriteLocal puts the file back when the edit record
// fails, and a failure marking a cited learning later leaves the edit
// recorded (the next call can finish it). The memory files themselves are
// in the memory subpackage, which knows nothing of this storage.
package knowledge

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Store is the part of the store learnings and memory edits need.
type Store interface {
	Collection(ctx context.Context, spec store.CollectionSpec) (store.Collection, error)
}

const (
	// maxTries bounds the id and compare-and-set loops under races.
	maxTries = 10
	// newest is how many of the newest records an add reads for the
	// highest id: records created in the same millisecond tie on
	// created_at and key order is text order, so the newest one alone may
	// not hold it.
	newest = 20
)

// putNext puts a record under the next free id, one past the highest of
// the newest records. A put that finds its id taken lost a race: dup is
// asked about the record there, and when it says yes that record is the
// result (created false), else the next id is tried. dup may be nil.
func putNext(ctx context.Context, col store.Collection, state, what string, value func(id int64) ([]byte, error), dup func(store.Record) (bool, error)) (store.Record, bool, error) {
	last, err := col.List(ctx, store.RecordFilter{Order: store.OrderCreatedDesc, Limit: newest})
	if err != nil {
		return store.Record{}, false, err
	}
	id := int64(1)
	for _, r := range last {
		if n, err := strconv.ParseInt(r.Key, 10, 64); err == nil && n >= id {
			id = n + 1
		}
	}
	for range maxTries {
		b, err := value(id)
		if err != nil {
			return store.Record{}, false, err
		}
		r, created, err := col.Put(ctx, store.Record{Key: strconv.FormatInt(id, 10), State: state, Value: b})
		if err != nil {
			return store.Record{}, false, err
		}
		if created {
			return r, true, nil
		}
		if dup != nil {
			same, err := dup(r)
			if err != nil || same {
				return r, false, err
			}
		}
		id++ // taken by a concurrent add
	}
	return store.Record{}, false, fmt.Errorf("%w: %s: lost every id race", domain.ErrConflict, what)
}

// casUpdate reads the record at key, lets fn change it and writes it back,
// again from the read when another writer got there first.
func casUpdate(ctx context.Context, col store.Collection, key, what string, fn func(*store.Record) error) (store.Record, error) {
	for range maxTries {
		r, err := col.Get(ctx, key)
		if err != nil {
			return store.Record{}, err
		}
		if err := fn(&r); err != nil {
			return store.Record{}, err
		}
		r, err = col.Update(ctx, r)
		if errors.Is(err, domain.ErrConflict) {
			continue
		}
		return r, err
	}
	return store.Record{}, fmt.Errorf("%w: %s: lost every race", domain.ErrConflict, what)
}
