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

// TestIngestDrainOnce: `ingest drain --once` against a running controller
// exits 0 with an empty spool and reports nothing moved (#335).
func TestIngestDrainOnce(t *testing.T) {
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte("projects:\n  js: {}\n"), 0o600))
	root := config.Root{Dir: dir}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- control.Run(ctx, root, control.Options{Tick: time.Hour}) }()
	t.Cleanup(func() { cancel(); require.NoError(t, <-done) })

	var buf bytes.Buffer
	old := stdout
	stdout = &buf
	t.Cleanup(func() { stdout = old })

	g := &Globals{Socket: root.AdminSocket()}
	require.Eventually(t, func() bool {
		buf.Reset()
		return ingestDrainCmd{Once: true}.Run(g) == nil
	}, 5*time.Second, 10*time.Millisecond)
	assert.Equal(t, "moved 0 (new 0), bad 0\n", buf.String())
}
