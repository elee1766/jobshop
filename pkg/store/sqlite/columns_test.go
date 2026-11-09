package sqlite_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

// undo00013 returns a database to its schema before 00013 (#405), as on a
// database from an older release. Later migrations are undone first:
// goose refuses a gap below the top.
func undo00013() []string {
	return append(undo00015(), []string{
		"DROP TABLE resource_slots", "DROP TABLE resource_providers",
		"DELETE FROM goose_db_version WHERE version_id = 14",
		"DROP TABLE ticket_meta", "DROP INDEX tickets_children",
		"DROP INDEX tickets_message_to", "DROP INDEX tickets_session_key",
		"DROP INDEX tickets_run_key", "DROP INDEX tickets_run_state",
		"ALTER TABLE tickets DROP COLUMN f_waiting", "ALTER TABLE tickets DROP COLUMN f_index",
		"ALTER TABLE tickets DROP COLUMN f_workflow", "ALTER TABLE tickets DROP COLUMN f_state",
		"ALTER TABLE tickets DROP COLUMN f_to", "ALTER TABLE tickets DROP COLUMN f_key",
		"CREATE INDEX tickets_message_to ON tickets(json_extract(fields, '$.to')) WHERE kind = 'message'",
		"CREATE UNIQUE INDEX tickets_session_key ON tickets(json_extract(fields, '$.key')) WHERE kind = 'session' AND status <> 'closed' AND json_extract(fields, '$.key') IS NOT NULL",
		"CREATE INDEX tickets_run_key ON tickets(json_extract(fields, '$.key')) WHERE kind = 'run' AND status <> 'closed'",
		"CREATE INDEX tickets_run_state ON tickets(json_extract(fields, '$.state')) WHERE kind = 'run'",
		"DELETE FROM goose_db_version WHERE version_id = 13",
	}...)
}

// undo00015 returns a database to its schema before 00015 (#426).
func undo00015() []string {
	return []string{
		// 00021 (#579) moved memory_edits into records; 00018 below
		// drops and recreates the table, so any shape will do.
		"CREATE TABLE memory_edits (id INTEGER PRIMARY KEY)",
		"DROP INDEX IF EXISTS ru_memory_edits_pr",
		"DELETE FROM goose_db_version WHERE version_id = 21",
		// 00020 (#579) moved write_keys and poll_cursors into records.
		"CREATE TABLE write_keys (key TEXT PRIMARY KEY, op TEXT NOT NULL, result TEXT NOT NULL DEFAULT '', done INTEGER NOT NULL DEFAULT 0, actor TEXT NOT NULL DEFAULT '', created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL)",
		"CREATE TABLE poll_cursors (poller TEXT NOT NULL, repo TEXT NOT NULL, stream TEXT NOT NULL, since INTEGER NOT NULL DEFAULT 0, etag TEXT NOT NULL DEFAULT '', updated_at INTEGER NOT NULL, PRIMARY KEY (poller, repo, stream))",
		"DELETE FROM goose_db_version WHERE version_id = 20",
		// 00019 (#579) only adds records.
		"DROP TABLE records", "DELETE FROM goose_db_version WHERE version_id = 19",
		// 00018 (#592) rebuilt memory_edits; put back the 00006 shape.
		"DROP TABLE memory_edits",
		"CREATE TABLE memory_edits (id INTEGER PRIMARY KEY, pr TEXT NOT NULL UNIQUE, learnings TEXT NOT NULL DEFAULT '[]', pools TEXT NOT NULL DEFAULT '[]', merged_at INTEGER NOT NULL, baseline TEXT NOT NULL DEFAULT '{}', state TEXT NOT NULL DEFAULT 'pending' CHECK (state IN ('pending','helped','not_helped')), after TEXT NOT NULL DEFAULT '{}', reason TEXT NOT NULL DEFAULT '', evaluated_at INTEGER NOT NULL DEFAULT 0, created_at INTEGER NOT NULL)",
		"CREATE INDEX memory_edits_state ON memory_edits(state, merged_at)",
		"DELETE FROM goose_db_version WHERE version_id = 18",
		// 00017 (#524) only relaxes a CHECK, so the newer table stays.
		"DELETE FROM goose_db_version WHERE version_id = 17",
		// 00016 (#341) is undone first.
		"DROP TABLE poll_cursors", "DELETE FROM goose_db_version WHERE version_id = 16",
		"DROP TABLE ticket_resources", "DELETE FROM goose_db_version WHERE version_id = 15",
	}
}

// TestPortableColumnsBackfill: rows written before 00013 get their
// promoted columns and meta rows from the migration, so run lookup by
// key, run state lists, step order, waits and meta filters still work.
func TestPortableColumnsBackfill(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	path := testPath(t)
	s, err := sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	rf, _ := json.Marshal(domain.RunFields{Workflow: "w", Key: "k1", State: domain.RunRunning})
	meta := json.RawMessage(`{"trigger":{"name":"nightly","n":3}}`)
	run, err := store.Create(ctx, s, "admin", store.NewTicket{Project: "p", Kind: domain.KindRun, Title: "r", Fields: rf, Meta: meta})
	require.NoError(t, err)
	var steps []string
	for _, i := range []int{1, 0} {
		sf, _ := json.Marshal(domain.StepFields{Run: run.ID, Step: "s", Index: i})
		st, err := store.Create(ctx, s, "admin", store.NewTicket{Project: "p", Kind: domain.KindStep, Title: "s", Pool: "dev", ParentID: run.ID, Fields: sf})
		require.NoError(t, err)
		steps = append(steps, st.ID)
	}
	cond := domain.WaitCond{Type: domain.WaitTime, At: run.CreatedAt.Add(time.Hour)}
	wf, _ := json.Marshal(domain.TaskFields{Wait: &domain.Wait{Reason: "later", Since: run.CreatedAt, Conds: []domain.WaitCond{cond}}})
	waiting, err := store.Create(ctx, s, "admin", store.NewTicket{Project: "p", Kind: domain.KindTask, Title: "w", Pool: "dev", Fields: wf})
	require.NoError(t, err)
	require.NoError(t, s.Close())

	db, err := openAt21(t, path)
	require.NoError(t, err)
	for _, q := range undo00013() {
		_, err = db.ExecContext(ctx, q)
		require.NoError(t, err, q)
	}
	require.NoError(t, db.Close())

	s, err = sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	defer func() { _ = s.Close() }()
	got, err := store.OpenRun(ctx, s, "w", "k1")
	require.NoError(t, err)
	assert.Equal(t, run.ID, got.ID)
	runs, err := s.List(ctx, store.Filter{Kind: domain.KindRun, State: domain.RunRunning, Order: store.OrderNewest})
	require.NoError(t, err)
	require.Len(t, runs, 1)
	kids, err := store.Children(ctx, s, run.ID)
	require.NoError(t, err)
	require.Len(t, kids, 2)
	assert.Equal(t, []string{steps[1], steps[0]}, []string{kids[0].ID, kids[1].ID}, "ordered by index")
	ready, err := s.List(ctx, store.Filter{Project: "p", Pool: "dev", Ready: true})
	require.NoError(t, err)
	for _, r := range ready {
		assert.NotEqual(t, waiting.ID, r.ID, "a waiting ticket stays parked")
	}
	byMeta, err := s.List(ctx, store.Filter{MetaPack: "trigger", MetaKey: "name", MetaValue: "nightly"})
	require.NoError(t, err)
	require.Len(t, byMeta, 1)
	assert.Equal(t, run.ID, byMeta[0].ID)
}

// TestTicketResourcesBackfill: steps written before 00015 get their
// ticket_resources rows from the migration, so a full resource still
// holds them out of Ready (#426).
func TestTicketResourcesBackfill(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	path := testPath(t)
	s, err := sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	require.NoError(t, s.SetResources(ctx, domain.ResourceConfig{Caps: map[string]int{"pr": 1}}))
	rf, _ := json.Marshal(domain.RunFields{Workflow: "w", State: domain.RunRunning})
	run, err := store.Create(ctx, s, "admin", store.NewTicket{Project: "p", Kind: domain.KindRun, Title: "r", Fields: rf})
	require.NoError(t, err)
	var ids []string
	for range 2 {
		sf, _ := json.Marshal(domain.StepFields{Run: run.ID, Step: "s", Resources: []string{"pr:x#1"}})
		st, err := store.Create(ctx, s, "admin", store.NewTicket{Project: "p", Kind: domain.KindStep, Title: "s", Pool: "dev", ParentID: run.ID, Fields: sf})
		require.NoError(t, err)
		ids = append(ids, st.ID)
	}
	_, _, err = s.Claim(ctx, ids[0], "h", time.Hour)
	require.NoError(t, err)
	require.NoError(t, s.Close())

	db, err := openAt21(t, path)
	require.NoError(t, err)
	for _, q := range undo00015() {
		_, err = db.ExecContext(ctx, q)
		require.NoError(t, err, q)
	}
	require.NoError(t, db.Close())

	s, err = sqlite.Open(ctx, path, sqlite.Options{})
	require.NoError(t, err)
	defer func() { _ = s.Close() }()
	ready, err := s.List(ctx, store.Filter{Project: "p", Pool: "dev", Ready: true})
	require.NoError(t, err)
	assert.Empty(t, ready, "pr:x#1 is held by the backfilled claim")
	rep, err := s.Resources(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []domain.ResourceUse{{Key: "pr:x#1", Capacity: 1, Held: 1, Waiting: 1}}, rep.Use)
}
