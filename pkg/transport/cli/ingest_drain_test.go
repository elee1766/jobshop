package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/orchestration/control/controltest"
)

// TestIngestDrainOnce: `ingest drain --once` against a running controller
// exits 0 with an empty spool and reports nothing moved (#335).
func TestIngestDrainOnce(t *testing.T) {
	root := controltest.NewRoot(t, "projects:\n  js: {}\n")
	controltest.StartWith(t, root)

	var buf bytes.Buffer
	old := stdout
	stdout = &buf
	t.Cleanup(func() { stdout = old })

	require.NoError(t, ingestDrainCmd{Once: true}.Run(&Globals{Socket: root.AdminSocket()}))
	assert.Equal(t, "moved 0 (new 0), bad 0\n", buf.String())
}
