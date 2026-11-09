package store

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/domain"
)

// Collections (#579, docs/notes/579-collections.md) are named buckets of
// records with one generic schema. The store implements them once;
// domain packages (ingest, learnings, memory) build their logic on top.

// CollectionSpec declares a collection. A domain keeps it as a package
// var and passes it to Store.Collection at startup; specs are not stored.
type CollectionSpec struct {
	// Name is [a-z][a-z0-9_]*, at most 40 chars: it names a Postgres
	// partition and the collection's indexes.
	Name string
	// Index fields get an index each and may be used in RecordFilter.Where.
	Index []Field
	// Unique fields are unique within the collection when non-empty, and
	// may be used in RecordFilter.Where.
	Unique []Field
	// Search fields feed the collection's full-text index. Empty means
	// the collection has no Search.
	Search []Field
}

// Field is a JSON field of Record.Value.
type Field struct {
	Name string // the name filters use, [a-z_][a-z0-9_]*
	Path string // JSON path in Value, dotted, e.g. "pr" or "meta.repo"
}

// Record is one entry of a collection.
type Record struct {
	Collection string
	// Key is the domain's natural key, unique in the collection. Key
	// order means nothing; order by created time or a declared field.
	Key   string
	State string // free text; the domain owns the values
	// Value is the domain's struct as a JSON object. Postgres stores it
	// as jsonb, so key order and spacing are not kept.
	Value      json.RawMessage
	Revision   int64 // 1 on insert, +1 on every write and claim
	Holder     string
	LeaseUntil time.Time
	Attempts   int64 // +1 on every Claim
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Record orders for RecordFilter.Order.
const (
	OrderCreated     = "created"  // created_at, then key (the default)
	OrderCreatedDesc = "-created" // newest first
	OrderUpdated     = "updated"  // updated_at, then key
)

// RecordFilter selects records. Empty fields match all.
type RecordFilter struct {
	Keys   []string
	States []string
	// Where matches declared Index or Unique fields by value. Values
	// compare as text, so a JSON number 12 matches "12". Undeclared
	// fields are ErrInvalid, so no query scans a bucket by accident.
	Where map[string]string
	// Since and Until bound created_at: Since <= created_at < Until.
	Since, Until time.Time
	// UpdatedBefore matches updated_at < UpdatedBefore (prune by age).
	UpdatedBefore time.Time
	Order         string
	Limit         int // List only; 0 means DefaultLimit
}

// Empty reports whether f selects every record.
func (f RecordFilter) Empty() bool {
	return len(f.Keys) == 0 && len(f.States) == 0 && len(f.Where) == 0 &&
		f.Since.IsZero() && f.Until.IsZero() && f.UpdatedBefore.IsZero()
}

// Collection is the handle Store.Collection returns.
type Collection interface {
	// Put inserts r (key, state, value, and created_at if set), or does
	// nothing if the key or a non-empty Unique field already exists. It
	// returns the stored record and whether it was created.
	Put(ctx context.Context, r Record) (rec Record, created bool, err error)
	// Get returns the record with key, or ErrNotFound.
	Get(ctx context.Context, key string) (Record, error)
	// Update writes r.State and r.Value if r.Revision is the stored
	// revision, else ErrConflict. A record under a live lease may only be
	// updated by its holder (r.Holder), else ErrNotHolder; a non-empty
	// r.Holder must hold a live lease. A zero r.LeaseUntil clears the
	// lease; otherwise the lease is kept as stored (Renew extends it).
	Update(ctx context.Context, r Record) (Record, error)
	// List returns matching records in f.Order, at most f.Limit.
	List(ctx context.Context, f RecordFilter) ([]Record, error)
	// Count counts matching records by state.
	Count(ctx context.Context, f RecordFilter) (map[string]int, error)
	// Claim leases the oldest (created_at, key) matching record that is
	// unleased or whose lease ran out, bumping attempts and revision. It
	// returns ErrNotFound when none matches. An empty filter is ErrInvalid.
	Claim(ctx context.Context, f RecordFilter, holder string, lease time.Duration) (Record, error)
	// Renew extends holder's live lease on key, else ErrNotHolder. It
	// does not bump the revision.
	Renew(ctx context.Context, key, holder string, lease time.Duration) (Record, error)
	// Delete removes matching records and returns how many. Order and
	// Limit are ignored; an empty filter is ErrInvalid.
	Delete(ctx context.Context, f RecordFilter) (int, error)
	// Search runs a full-text query over the Search fields, best match
	// first. ErrInvalid if the collection has no Search fields.
	Search(ctx context.Context, query string, limit int) ([]Record, error)
}

var (
	collectionName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	fieldName      = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)
	fieldPath      = regexp.MustCompile(`^[a-z_][a-z0-9_]*(\.[a-z_][a-z0-9_]*)*$`)
)

// maxCollectionName keeps index names (ru_<name>_<field>) under
// Postgres's 63-byte identifier limit.
const (
	maxCollectionName = 40
	maxIdent          = 63
)

// Validate checks names and paths. They are written into SQL as
// identifiers and literals, so only this pattern is allowed.
func (s CollectionSpec) Validate() error {
	if !collectionName.MatchString(s.Name) || len(s.Name) > maxCollectionName {
		return fmt.Errorf("%w: collection name %q: want [a-z][a-z0-9_]*, at most %d chars", domain.ErrInvalid, s.Name, maxCollectionName)
	}
	seen := map[string]bool{}
	for _, fs := range [][]Field{s.Index, s.Unique} {
		for _, f := range fs {
			if seen[f.Name] {
				return fmt.Errorf("%w: collection %s: field %q declared twice", domain.ErrInvalid, s.Name, f.Name)
			}
			seen[f.Name] = true
		}
	}
	for _, fs := range [][]Field{s.Index, s.Unique, s.Search} {
		for _, f := range fs {
			if !fieldName.MatchString(f.Name) || !fieldPath.MatchString(f.Path) {
				return fmt.Errorf("%w: collection %s: field %q path %q", domain.ErrInvalid, s.Name, f.Name, f.Path)
			}
			if len("ru_"+s.Name+"_"+f.Name) > maxIdent {
				return fmt.Errorf("%w: collection %s: field name %q too long", domain.ErrInvalid, s.Name, f.Name)
			}
		}
	}
	return nil
}

// Lookup returns the declared Index or Unique field name, and whether it
// is unique.
func (s CollectionSpec) Lookup(name string) (f Field, unique, ok bool) {
	for _, f := range s.Index {
		if f.Name == name {
			return f, false, true
		}
	}
	for _, f := range s.Unique {
		if f.Name == name {
			return f, true, true
		}
	}
	return Field{}, false, false
}

// CheckValue returns v if it is a JSON object, {} if empty, else
// ErrInvalid. Backends call it before every write.
func (s CollectionSpec) CheckValue(v json.RawMessage) (json.RawMessage, error) {
	if len(v) == 0 {
		return json.RawMessage("{}"), nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(v, &m); err != nil || m == nil {
		return nil, fmt.Errorf("%w: %s: value must be a JSON object", domain.ErrInvalid, s.Name)
	}
	return v, nil
}

// FieldText reads the dotted path of a JSON object as text, as the
// backends' field expressions do: a string as is, other values as JSON.
// It reports false for a missing path or null.
func FieldText(v json.RawMessage, path string) (string, bool) {
	var cur any
	if err := json.Unmarshal(v, &cur); err != nil {
		return "", false
	}
	for p := range strings.SplitSeq(path, ".") {
		obj, ok := cur.(map[string]any)
		if !ok {
			return "", false
		}
		if cur, ok = obj[p]; !ok {
			return "", false
		}
	}
	switch x := cur.(type) {
	case nil:
		return "", false
	case string:
		return x, true
	default:
		b, err := json.Marshal(x)
		return string(b), err == nil
	}
}

// RecordOrder checks a RecordFilter.Order and returns its column order.
func RecordOrder(o string) (string, error) {
	switch o {
	case "", OrderCreated:
		return "created_at, key", nil
	case OrderCreatedDesc:
		return "created_at DESC, key DESC", nil
	case OrderUpdated:
		return "updated_at, key", nil
	}
	return "", fmt.Errorf("%w: record order %q", domain.ErrInvalid, o)
}
