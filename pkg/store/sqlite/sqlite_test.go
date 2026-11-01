package sqlite_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
	"github.com/elee1766/jobshop/pkg/store/storetest"
)

func TestConformance(t *testing.T) {
	storetest.Run(t, func(t *testing.T, c *storetest.Clock) store.Store {
		s, err := sqlite.Open(t.Context(), filepath.Join(t.TempDir(), "jobshop.db"), sqlite.Options{Now: c.Now})
		require.NoError(t, err)
		return s
	})
}
