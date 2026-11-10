package sqlite

import (
	"context"
	"database/sql"
	"sync"

	"github.com/gfx-labs/jobshop/pkg/storage/store/sqlite/sqldb"
)

// stmtCacheMax bounds the cache. sqldb queries are a fixed set, but
// list queries (LabelsFor) expand to one text per argument count, so the cache
// stops growing once full and later texts run unprepared (#433).
const stmtCacheMax = 512

// stmtCache keeps one prepared statement per query text. Profiling the
// control tests (#433) showed about 40% of CPU in SQLite's parser:
// database/sql prepares, runs and closes every ad hoc query, so each
// supervisor pass re-parsed the same SQL.
type stmtCache struct {
	db *sql.DB
	mu sync.Mutex
	m  map[string]*sql.Stmt
	// pending are texts first seen inside a transaction. Preparing on
	// the DB there would wait for the connection the transaction holds
	// (one connection in production), so they are prepared by the next
	// call outside a transaction.
	pending map[string]bool
}

func newStmtCache(db *sql.DB) *stmtCache {
	return &stmtCache{db: db, m: map[string]*sql.Stmt{}, pending: map[string]bool{}}
}

// cached returns the statement for query if it is prepared, else
// queues it. It never takes a connection, so it is safe in a
// transaction.
func (c *stmtCache) cached(query string) *sql.Stmt {
	c.mu.Lock()
	defer c.mu.Unlock()
	st, ok := c.m[query]
	if !ok && len(c.m)+len(c.pending) < stmtCacheMax {
		c.pending[query] = true
	}
	return st
}

// get returns the statement for query, preparing it and any queued
// texts on first use. nil means run the query unprepared. Call it only
// outside a transaction.
func (c *stmtCache) get(ctx context.Context, query string) *sql.Stmt {
	c.mu.Lock()
	if st, ok := c.m[query]; ok && len(c.pending) == 0 {
		c.mu.Unlock()
		return st
	}
	todo := make([]string, 0, len(c.pending)+1)
	for q := range c.pending {
		todo = append(todo, q)
	}
	c.pending = map[string]bool{}
	if _, ok := c.m[query]; !ok {
		todo = append(todo, query)
	}
	c.mu.Unlock()
	// Prepare without the lock: PrepareContext may wait for the
	// connection a transaction holds, and that transaction calls cached.
	for _, q := range todo {
		c.prepare(ctx, q)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.m[query]
}

func (c *stmtCache) prepare(ctx context.Context, query string) {
	c.mu.Lock()
	_, have := c.m[query]
	full := len(c.m) >= stmtCacheMax
	c.mu.Unlock()
	if have || full {
		return
	}
	st, err := c.db.PrepareContext(ctx, query)
	if err != nil {
		return // the unprepared call reports the error
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.m[query]; ok || len(c.m) >= stmtCacheMax {
		_ = st.Close() // another caller prepared it first
		return
	}
	c.m[query] = st
}

// cachedDB runs queries on the database through the cache. The
// statements need no Close of their own: closing the database closes
// the connections they are prepared on.
type cachedDB struct {
	*sql.DB
	c *stmtCache
}

func (d cachedDB) ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error) {
	if st := d.c.get(ctx, q); st != nil {
		return st.ExecContext(ctx, args...)
	}
	return d.DB.ExecContext(ctx, q, args...)
}

func (d cachedDB) QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	if st := d.c.get(ctx, q); st != nil {
		return st.QueryContext(ctx, args...)
	}
	return d.DB.QueryContext(ctx, q, args...)
}

func (d cachedDB) QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row {
	if st := d.c.get(ctx, q); st != nil {
		return st.QueryRowContext(ctx, args...)
	}
	return d.DB.QueryRowContext(ctx, q, args...)
}

// cachedTx runs queries in a transaction through the cache. A statement
// from Tx.StmtContext reuses the connection's prepared statement and is
// closed by Commit or Rollback, so it is not closed here.
type cachedTx struct {
	*sql.Tx
	c *stmtCache
}

func (t cachedTx) stmt(ctx context.Context, q string) *sql.Stmt {
	if st := t.c.cached(q); st != nil {
		return t.StmtContext(ctx, st)
	}
	return nil
}

func (t cachedTx) ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error) {
	if st := t.stmt(ctx, q); st != nil {
		return st.ExecContext(ctx, args...)
	}
	return t.Tx.ExecContext(ctx, q, args...)
}

func (t cachedTx) QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	if st := t.stmt(ctx, q); st != nil {
		return st.QueryContext(ctx, args...)
	}
	return t.Tx.QueryContext(ctx, q, args...)
}

func (t cachedTx) QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row {
	if st := t.stmt(ctx, q); st != nil {
		return st.QueryRowContext(ctx, args...)
	}
	return t.Tx.QueryRowContext(ctx, q, args...)
}

// bind returns the queries over d through the cache.
func (c *stmtCache) bind(d sqldb.DBTX) sqldb.Querier {
	switch v := d.(type) {
	case *sql.DB:
		return sqldb.New(cachedDB{DB: v, c: c})
	case *sql.Tx:
		return sqldb.New(cachedTx{Tx: v, c: c})
	}
	return sqldb.New(d)
}

// conn returns d through the cache, for the store's queries built in Go.
func (c *stmtCache) conn(d sqldb.DBTX) sqldb.DBTX {
	switch v := d.(type) {
	case *sql.DB:
		return cachedDB{DB: v, c: c}
	case *sql.Tx:
		return cachedTx{Tx: v, c: c}
	}
	return d
}
