package postgres_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/orchestration/ingest"
	"github.com/elee1766/jobshop/pkg/storage/store/postgres"
)

// TestInboxMigrate: 00008 (#579) copies the inbox into the collection
// with ids, state and lease, new items get ids past them, Down copies
// them back and moves the identity, and Up again keeps them.
func TestInboxMigrate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, name := openNamed(t, postgres.Options{})
	db, err := sql.Open("pgx", withDB(os.Getenv(dsnEnv), name))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, postgres.MigrateDownTo(ctx, db, 7))
	_, err = db.ExecContext(ctx, `INSERT INTO inbox (id, source, delivery_id, kind, received_at, payload, admitted, state, hook_results, holder, lease_until, attempts, updated_at, natural_key)
		VALUES (40, 'github', 'd40', 'issue_comment', 5000, '{"x":1}', '{"ok":true}', 'admitted', '[{"name":"h"}]', 'w', 9000, 2, 6000, 'nk')`)
	require.NoError(t, err)
	require.NoError(t, postgres.Migrate(ctx, db))
	check := func(wantNew int64) {
		t.Helper()
		it, err := ingest.InboxGet(ctx, s, 40)
		require.NoError(t, err)
		assert.Equal(t, domain.InboxAdmitted, it.State)
		assert.JSONEq(t, `{"x":1}`, string(it.Payload))
		assert.JSONEq(t, `{"ok":true}`, string(it.Admitted))
		assert.Equal(t, "w", it.Holder)
		assert.EqualValues(t, 2, it.Attempts)
		assert.Equal(t, time.UnixMilli(5000).UTC(), it.ReceivedAt)
		_, dup, err := ingest.InboxPut(ctx, s, domain.InboxItem{Source: "github", DeliveryID: "other", NaturalKey: "nk"})
		require.NoError(t, err)
		assert.True(t, dup)
		n, dup, err := ingest.InboxPut(ctx, s, domain.InboxItem{Source: "github", DeliveryID: fmt.Sprint("new", wantNew), Payload: json.RawMessage(`{}`)})
		require.NoError(t, err)
		assert.False(t, dup)
		assert.Equal(t, wantNew, n.ID)
	}
	check(41)

	require.NoError(t, postgres.MigrateDownTo(ctx, db, 7))
	var n int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM inbox").Scan(&n))
	assert.Equal(t, 2, n)
	var id int64
	require.NoError(t, db.QueryRowContext(ctx, "INSERT INTO inbox (source, delivery_id, received_at, updated_at) VALUES ('github', 'seq', 1, 1) RETURNING id").Scan(&id))
	assert.EqualValues(t, 42, id, "identity past the copied ids")
	_, err = db.ExecContext(ctx, "DELETE FROM inbox WHERE id = 42")
	require.NoError(t, err)
	require.NoError(t, postgres.Migrate(ctx, db))
	check(42)
}
