package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
)

// TestVersionServer: `version --server` prints the controller's commit,
// not the CLI binary's (#180).
func TestVersionServer(t *testing.T) {
	// Short dir: unix socket paths are limited to ~100 bytes.
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte("projects:\n  js: {}\n"), 0o600))
	root := config.Root{Dir: dir}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		done <- control.Run(ctx, root, control.Options{Tick: time.Hour, Version: "v9", Commit: "abc123server"})
	}()
	t.Cleanup(func() { cancel(); require.NoError(t, <-done) })

	var buf bytes.Buffer
	old := stdout
	stdout = &buf
	t.Cleanup(func() { stdout = old })

	g := &Globals{Socket: root.AdminSocket()}
	require.Eventually(t, func() bool {
		buf.Reset()
		return versionCmd{Server: true}.Run(g) == nil
	}, 5*time.Second, 10*time.Millisecond)
	assert.Equal(t, "jobshop v9 (abc123server)\n", buf.String())
}
