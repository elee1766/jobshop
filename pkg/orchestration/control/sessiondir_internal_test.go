package control

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
)

// TestSessionDirArgExpands (#566): ${SESSION_DIR} in a pack's agent args
// loads without an escape and the controller expands it at session start.
func TestSessionDirArgExpands(t *testing.T) {
	dir := t.TempDir()
	pack := filepath.Join(dir, "pk")
	require.NoError(t, os.MkdirAll(pack, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(pack, config.PackFileName),
		[]byte("agents:\n  j: {command: /bin/true, args: [acp, --socket, \"${SESSION_DIR}/jcode.sock\"]}\npools:\n  dev: {agent: j, dir: /tmp, prompt: \"x {{.n}}\"}\n"), 0o600))
	c, _, _ := scanControllerYAML(t, "projects:\n  js: {}\npacks:\n  - path: "+pack+"\n    vars: {n: 1}\n# %[1]s\n")
	spec, _, err := c.hostSpec(domain.Ticket{ID: "js-s1", Pool: "dev"}, domain.SessionFields{Agent: "j"})
	require.NoError(t, err)
	assert.Equal(t, []string{"acp", "--socket", filepath.Join(c.root.SessionDir("js-s1"), "jcode.sock")}, spec.Args)
}
