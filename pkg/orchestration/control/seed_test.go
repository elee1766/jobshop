package control_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control/controltest"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/storage/store/sqlite"
)

// seedStore opens root's database while no controller runs on it and
// passes it to fn. Tests that need more rows than one page write them
// here in one transaction: one RPC per row cost seconds under -race
// (JOB-117).
func seedStore(t *testing.T, root config.Root, fn func(st store.Store)) {
	t.Helper()
	st, err := sqlite.Open(t.Context(), root.DB(), sqlite.Options{Now: func() time.Time { return controltest.Epoch }})
	require.NoError(t, err)
	fn(st)
	require.NoError(t, st.Close())
}
