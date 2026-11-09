package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
)

// serve must reach the controller when only a pack fails and a last good
// config exists: the controller serves the snapshot (#227, #1128).
func TestServeBadPackWithLastGood(t *testing.T) {
	dir := t.TempDir()
	const host = "projects:\n  js: {}\nagents:\n  a:\n    command: x\n"
	write := func(name, body string) {
		t.Helper()
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600))
	}
	write(config.FileName, host)
	r := config.Root{Dir: dir}
	c, err := config.Load(r)
	require.NoError(t, err)
	require.NoError(t, config.SaveLastGood(r.LastGood(), c))

	write(config.FileName, host+"pack: missing-pack.cue\n")
	var served bool
	controller := func(context.Context, config.Root, ServeOptions) error {
		served = true
		return nil
	}
	assert.Equal(t, 0, Main([]string{"--root", dir, "serve"}, controller))
	assert.True(t, served, "serve stopped before the controller")

	// Without a snapshot the pack error still stops serve.
	require.NoError(t, os.Remove(r.LastGood()))
	served = false
	assert.NotEqual(t, 0, Main([]string{"--root", dir, "serve"}, controller))
	assert.False(t, served)
}
