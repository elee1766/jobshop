package sqlite_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/storage/store/sqlite"
)

// TestInboxMigrate: rows in the old inbox table are in the inbox
// collection after 00023 (#579) with their ids, state, lease and
// received time; new items get ids past them and dedupe against them;
// Down puts them back.
func TestInboxMigrate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	path := testPath(t)
	s, err := sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	require.NoError(t, s.Close())

	db, err := openAt21(t, path)
	require.NoError(t, err)
	for _, q := range []string{
		`INSERT INTO inbox (id, source, delivery_id, kind, repo_or_team, number, comment_id, actor, received_at, payload, admitted, state, hook_results, holder, lease_until, attempts, updated_at, endpoint, natural_key)
		 VALUES (40, 'github', 'd40', 'issue_comment', 'a/b', 7, 'c1', 'x', 5000, '{"x":1}', '{"ok":true}', 'admitted', '[{"name":"h"}]', 'w', 9000, 2, 6000, 'gh-app', 'github:comment:a/b:c1:t')`,
		`INSERT INTO inbox (id, source, delivery_id, received_at, updated_at) VALUES (12, 'linear', 'd12', 8000, 8000)`,
	} {
		_, err = db.ExecContext(ctx, q)
		require.NoError(t, err, q)
	}
	require.NoError(t, db.Close())

	check := func() {
		t.Helper()
		s, err := sqlite.Open(ctx, path, sqlite.Options{})
		require.NoError(t, err)
		defer func() { _ = s.Close() }()
		it, err := ingest.InboxGet(ctx, s, 40)
		require.NoError(t, err)
		assert.Equal(t, "d40", it.DeliveryID)
		assert.Equal(t, "gh-app", it.Endpoint)
		assert.Equal(t, "github:comment:a/b:c1:t", it.NaturalKey)
		assert.Equal(t, "issue_comment", it.Kind)
		assert.Equal(t, "a/b", it.RepoOrTeam)
		assert.EqualValues(t, 7, it.Number)
		assert.Equal(t, "c1", it.CommentID)
		assert.Equal(t, domain.InboxAdmitted, it.State)
		assert.JSONEq(t, `{"x":1}`, string(it.Payload))
		assert.JSONEq(t, `{"ok":true}`, string(it.Admitted))
		assert.JSONEq(t, `[{"name":"h"}]`, string(it.HookResults))
		assert.Equal(t, "w", it.Holder)
		assert.Equal(t, time.UnixMilli(9000).UTC(), it.LeaseUntil)
		assert.EqualValues(t, 2, it.Attempts)
		assert.Equal(t, time.UnixMilli(5000).UTC(), it.ReceivedAt)
		other, err := ingest.InboxGet(ctx, s, 12)
		require.NoError(t, err)
		assert.Nil(t, other.Admitted)
		assert.JSONEq(t, `[]`, string(other.HookResults))

		_, dup, err := ingest.InboxPut(ctx, s, domain.InboxItem{Source: "github", DeliveryID: "poll:x", NaturalKey: "github:comment:a/b:c1:t"})
		require.NoError(t, err)
		assert.True(t, dup, "natural key of a migrated row")
		n, dup, err := ingest.InboxPut(ctx, s, domain.InboxItem{Source: "github", DeliveryID: "new", Payload: json.RawMessage(`{}`)})
		require.NoError(t, err)
		assert.False(t, dup)
		assert.EqualValues(t, 41, n.ID, "past the highest id")
		pend, err := ingest.InboxList(ctx, s, store.InboxFilter{State: domain.InboxPending})
		require.NoError(t, err)
		assert.Len(t, pend, 2)
	}
	check()

	// Down and up again: the rows survive the round trip.
	db, err = openAt21(t, path)
	require.NoError(t, err)
	var cnt int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM inbox").Scan(&cnt))
	assert.Equal(t, 3, cnt)
	var admitted, hooks string
	require.NoError(t, db.QueryRowContext(ctx, "SELECT admitted, hook_results FROM inbox WHERE id = 40").Scan(&admitted, &hooks))
	assert.JSONEq(t, `{"ok":true}`, admitted)
	assert.JSONEq(t, `[{"name":"h"}]`, hooks)
	var null bool
	require.NoError(t, db.QueryRowContext(ctx, "SELECT admitted IS NULL FROM inbox WHERE id = 12").Scan(&null))
	assert.True(t, null, "no admitted stays NULL")
	_, err = db.ExecContext(ctx, "DELETE FROM inbox WHERE id = 41")
	require.NoError(t, err)
	require.NoError(t, db.Close())
	check()
}
