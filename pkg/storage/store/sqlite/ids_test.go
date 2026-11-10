package sqlite_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/ids"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/storage/store/sqlite"
)

// TestIDGenerators mints ids with every generator through the store,
// switching generator on one root: old ids stay, new ones take the new
// shape, nothing collides, and seq starts past numeric ids already
// under the prefix (JOB-57).
func TestIDGenerators(t *testing.T) {
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "db")
	spec := ids.Spec{}
	open := func() *sqlite.Store {
		s, err := sqlite.Open(ctx, path, sqlite.Options{IDs: func(string) ids.Spec { return spec }})
		require.NoError(t, err)
		return s
	}
	s := open()
	seen := map[string]bool{}
	mk := func(kind domain.Kind) string {
		t.Helper()
		nt := store.NewTicket{Project: "js", Kind: kind, Title: "x"}
		if kind == domain.KindSession {
			nt.Fields = json.RawMessage(`{"agent":"a"}`)
		}
		tk, err := store.Create(ctx, s, "t", nt)
		require.NoError(t, err)
		require.False(t, seen[tk.ID], "duplicate id %s", tk.ID)
		seen[tk.ID] = true
		return tk.ID
	}

	// An explicit numeric id, as if created earlier: seq must skip it.
	_, err := store.Create(ctx, s, "t", store.NewTicket{ID: "js-7", Project: "js", Kind: domain.KindTask, Title: "old"})
	require.NoError(t, err)
	seen["js-7"] = true

	shapes := map[string]string{
		ids.Sonyflake: `^js-[0-9a-z]{13}$`,
		ids.Seq:       `^js-[1-9][0-9]*$`,
		ids.ULID:      `^js-[0-7][0-9a-z]{25}$`,
		ids.UUIDv7:    `^js-[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
		ids.UUID:      `^js-[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
		ids.XID:       `^js-[0-9a-v]{20}$`,
		ids.Words:     `^js-[a-z]+-[a-z]+-[a-z]+$`,
	}
	for _, name := range ids.Names {
		spec = ids.Spec{Generator: name}
		for range 5 {
			assert.Regexp(t, shapes[name], mk(domain.KindTask), name)
		}
		sess := mk(domain.KindSession)
		assert.True(t, ids.IsSession(sess), "%s session id %s", name, sess)
	}
	assert.True(t, seen["js-8"], "seq starts past js-7")
	assert.False(t, seen["js-1"], "seq never goes below an existing number")

	// Reopen: seq continues from its stored counter.
	require.NoError(t, s.Close())
	s = open()
	t.Cleanup(func() { _ = s.Close() })
	spec = ids.Spec{Generator: ids.Seq}
	id := mk(domain.KindTask)
	assert.Regexp(t, `^js-1[0-9]$`, id)

	// Every id still reads back.
	for id := range seen {
		_, err := s.Get(ctx, id)
		require.NoError(t, err, id)
	}
}
