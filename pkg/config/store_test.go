package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreConfig(t *testing.T) {
	c, err := Parse(nil)
	require.NoError(t, err)
	assert.Equal(t, StoreSQLite, c.Store.DriverName())

	c, err = Parse([]byte("store:\n  driver: postgres\n  dsn_file: pg.dsn\n"))
	require.NoError(t, err)
	assert.Equal(t, StorePostgres, c.Store.DriverName())
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pg.dsn"), []byte("postgres://u@h/db\n"), 0o600))
	dsn, err := c.Store.DSN(t.Context(), dir)
	require.NoError(t, err)
	assert.Equal(t, "postgres://u@h/db", dsn)

	for _, bad := range []string{
		"store:\n  driver: postgres\n",
		"store:\n  driver: mysql\n",
		"store:\n  dsn_file: pg.dsn\n",
	} {
		_, err := Parse([]byte(bad))
		assert.Error(t, err, bad)
	}
}
