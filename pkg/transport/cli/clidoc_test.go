package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const cliDocPath = "../../../llm/cli.md"

// TestCLIDoc fails when llm/cli.md is stale. Regenerate with
// go generate ./pkg/transport/cli.
func TestCLIDoc(t *testing.T) {
	t.Setenv("COLUMNS", "80")
	t.Setenv("JOBSHOP_ROOT", "")
	t.Setenv("JOBSHOP_SOCKET", "")
	got, err := cliDoc()
	require.NoError(t, err)
	if os.Getenv("JOBSHOP_UPDATE_CLI_DOC") != "" {
		require.NoError(t, os.WriteFile(cliDocPath, []byte(got), 0o600))
		return
	}
	want, err := os.ReadFile(cliDocPath)
	require.NoError(t, err)
	require.Equal(t, string(want), got, "llm/cli.md is stale: run go generate ./pkg/transport/cli")
}
