package sqlite_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/storage/store/sqlite"
)

// TestNanosecondMigrate: 00026 (JOB-58) turns ms columns into ns, keeps
// the values and ticket order, clamps the old year-9999 gate hold, and
// Down puts ms back.
func TestNanosecondMigrate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	path := testPath(t)
	s, err := sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	require.NoError(t, s.Close())

	db, err := sql.Open("sqlite", "file:"+path)
	require.NoError(t, err)
	require.NoError(t, sqlite.MigrateDownTo(ctx, db, 25))
	_, err = db.ExecContext(ctx, `INSERT INTO tickets (id, project, kind, created_at, updated_at, lease_until, defer_until) VALUES
		('p-b', 'p', 'task', 2000, 2500, 0, 0),
		('p-a', 'p', 'task', 3000, 3000, 9000, 253370764800000)`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO comments (ticket_id, author, body, created_at) VALUES ('p-b', 'x', 'vendorHash mismatch', 4000)`)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	s, err = sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	b, err := s.Get(ctx, "p-b")
	require.NoError(t, err)
	assert.Equal(t, time.UnixMilli(2000).UTC(), b.CreatedAt)
	assert.Equal(t, time.UnixMilli(2500).UTC(), b.UpdatedAt)
	a, err := s.Get(ctx, "p-a")
	require.NoError(t, err)
	assert.Equal(t, time.UnixMilli(9000).UTC(), a.LeaseUntil)
	assert.Equal(t, time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC), a.DeferUntil)
	list, err := s.List(ctx, store.Filter{Project: "p", Order: store.OrderOldest})
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, "p-b", list[0].ID, "created_at order, not id order")
	hits, err := s.Recall(ctx, store.RecallQuery{Text: "vendorHash"})
	require.NoError(t, err)
	require.Len(t, hits, 1)
	assert.Equal(t, time.UnixMilli(4000).UTC(), hits[0].At)
	require.NoError(t, s.Close())

	db, err = sql.Open("sqlite", "file:"+path)
	require.NoError(t, err)
	require.NoError(t, sqlite.MigrateDownTo(ctx, db, 25))
	var created int64
	require.NoError(t, db.QueryRowContext(ctx, "SELECT created_at FROM tickets WHERE id = 'p-b'").Scan(&created))
	assert.EqualValues(t, 2000, created)
	require.NoError(t, db.Close())
}

// TestBatchCreateOrder: tickets created in one Apply with a frozen clock
// get distinct, increasing created_at and list in creation order, not id
// order (JOB-58).
func TestBatchCreateOrder(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	at := time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC)
	s := openTest(t, sqlite.Options{Now: func() time.Time { return at }})
	ids := []string{"p-c", "p-a", "p-b"}
	var g store.Graph
	for _, id := range ids {
		g.Tickets = append(g.Tickets, store.NewTicket{ID: id, Project: "p", Kind: domain.KindTask})
	}
	out, err := s.Apply(ctx, "test", g)
	require.NoError(t, err)
	require.Len(t, out, 3)
	assert.True(t, out[0].CreatedAt.Before(out[1].CreatedAt))
	assert.True(t, out[1].CreatedAt.Before(out[2].CreatedAt))
	list, err := s.List(ctx, store.Filter{Project: "p", Order: store.OrderOldest})
	require.NoError(t, err)
	got := make([]string, 0, len(list))
	for _, tk := range list {
		got = append(got, tk.ID)
	}
	assert.Equal(t, ids, got)
}
