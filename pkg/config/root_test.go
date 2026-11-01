package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindRoot(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, FileName), nil, 0o600))
	deep := filepath.Join(dir, "a", "b")
	require.NoError(t, os.MkdirAll(deep, 0o700))

	r, err := FindRoot("", deep)
	require.NoError(t, err)
	assert.Equal(t, dir, r.Dir)
	assert.Equal(t, filepath.Join(dir, "sessions", "x"), r.SessionDir("x"))

	r, err = FindRoot(dir, "/")
	require.NoError(t, err)
	assert.Equal(t, dir, r.Dir)

	_, err = FindRoot(deep, "/")
	assert.Error(t, err, "explicit root without jobshop.yaml is rejected")
	_, err = FindRoot("", t.TempDir())
	assert.ErrorIs(t, err, ErrNoRoot)
}
