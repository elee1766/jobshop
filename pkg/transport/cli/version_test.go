package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/orchestration/control/controltest"
)

// TestVersionServer: `version --server` prints the controller's commit,
// not the CLI binary's (#180).
func TestVersionServer(t *testing.T) {
	root := controltest.NewRoot(t, "projects:\n  js: {}\n")
	controltest.StartWith(t, root, controltest.WithVersion("v9", "abc123server"))

	var buf bytes.Buffer
	old := stdout
	stdout = &buf
	t.Cleanup(func() { stdout = old })

	require.NoError(t, versionCmd{Server: true}.Run(&Globals{Socket: root.AdminSocket()}))
	assert.Equal(t, "jobshop v9 (abc123server)\n", buf.String())
}
