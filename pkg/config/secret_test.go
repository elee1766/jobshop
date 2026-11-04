package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadSecret(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	f := filepath.Join(dir, "tok")
	require.NoError(t, os.WriteFile(f, []byte("from-file\n"), 0o600))
	r := Root{Dir: dir}

	b, err := ReadSecret(ctx, r.SecretPath("tok"))
	require.NoError(t, err)
	assert.Equal(t, "from-file\n", string(b))

	// A command is not joined to the root and its stdout is the secret.
	b, err = ReadSecret(ctx, r.SecretPath("!printf from-cmd"))
	require.NoError(t, err)
	assert.Equal(t, "from-cmd", string(b))

	// A failing command reports stderr's last line, not the command.
	_, err = ReadSecret(ctx, "!echo noise >&2; echo vault sealed >&2; exit 3 # tok-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vault sealed")
	assert.NotContains(t, err.Error(), "noise")
	assert.NotContains(t, err.Error(), "tok-123")
}
