package postgres_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/ids"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/storage/store/postgres"
)

// TestIDGenerators: every generator mints through the Postgres store,
// and seq starts past numeric ids already under the prefix (JOB-57).
func TestIDGenerators(t *testing.T) {
	t.Parallel()
	if os.Getenv(dsnEnv) == "" {
		t.Skip(dsnEnv + " is not set")
	}
	ctx := t.Context()
	spec := ids.Spec{}
	s, _ := openNamed(t, postgres.Options{IDs: func(string) ids.Spec { return spec }})
	_, err := store.Create(ctx, s, "t", store.NewTicket{ID: "js-7", Project: "js", Kind: domain.KindTask, Title: "old"})
	require.NoError(t, err)
	seen := map[string]bool{"js-7": true}
	for _, name := range ids.Names {
		spec = ids.Spec{Generator: name}
		for range 3 {
			tk, err := store.Create(ctx, s, "t", store.NewTicket{Project: "js", Kind: domain.KindTask, Title: "x"})
			require.NoError(t, err, name)
			require.False(t, seen[tk.ID], tk.ID)
			seen[tk.ID] = true
		}
	}
	assert.True(t, seen["js-8"], "seq starts past js-7")
}
