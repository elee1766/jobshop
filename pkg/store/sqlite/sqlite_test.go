package sqlite_test

import (
	"testing"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
	"github.com/elee1766/jobshop/pkg/store/storetest"
)

func TestConformance(t *testing.T) {
	t.Parallel()
	storetest.Run(t, func(t *testing.T, c *storetest.Clock) store.Store {
		return openTest(t, sqlite.Options{Now: c.Now})
	})
}

// TestConformanceConcurrent runs the suite with deferred transactions on
// several connections (#406): claims, sweeps and their guards must not
// rely on SQLite serializing writers.
func TestConformanceConcurrent(t *testing.T) {
	t.Parallel()
	storetest.Run(t, func(t *testing.T, c *storetest.Clock) store.Store {
		return openTest(t, sqlite.Options{Now: c.Now, Concurrent: true})
	})
}
