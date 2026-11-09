package control

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/elee1766/acphost/pkg/sandbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPoolSandboxGuidance: sandbox guidance paths are bound read-only at
// the same path, after (over) a mapped user's rw ~/.jcode, change the env
// hash, and a guidance path holding a credentials file or equal to an rw
// path is refused (JOB-51).
func TestPoolSandboxGuidance(t *testing.T) {
	h := t.TempDir()
	t.Setenv("HOME", h)
	require.NoError(t, os.MkdirAll(filepath.Join(h, ".agents"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(h, ".jcode"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(h, ".jcode", "prompt-overlay.md"), []byte("x"), 0o644))
	c, _, dir := scanControllerYAML(t, `projects:
  js: {}
agents:
  a: {command: /bin/true, credentials: ["~/.jcode/auth.json"]}
pools:
  g:
    agent: a
    dir: /tmp
    sandbox:
      backend: bubblewrap
      rw: ["~/.jcode"]
      guidance: ["~/.agents", "~/.jcode/prompt-overlay.md", "~/.missing"]
  plain: {agent: a, dir: /tmp, sandbox: {backend: bubblewrap, rw: ["~/.jcode"]}}
  creds: {agent: a, dir: /tmp, sandbox: {backend: bubblewrap, guidance: ["~/.jcode"]}}
  same: {agent: a, dir: /tmp, sandbox: {backend: bubblewrap, rw: ["~/.agents"], guidance: ["~/.agents"]}}
# %s
`)
	home := t.TempDir()
	s, err := c.poolSandbox("g", dir, "/bin/true", "/tmp", home)
	require.NoError(t, err)
	for _, p := range []string{".agents", ".jcode/prompt-overlay.md", ".missing"} {
		assert.Contains(t, s.RO, sandbox.Mount{Src: filepath.Join(h, p), SkipMissing: true})
	}
	assert.Contains(t, s.RW, sandbox.Mount{Src: filepath.Join(home, ".jcode"), Dst: filepath.Join(h, ".jcode")})
	assert.NotEqual(t, c.sandboxKey("g"), c.sandboxKey("plain"))

	_, err = c.poolSandbox("creds", dir, "/bin/true", "/tmp", home)
	require.ErrorContains(t, err, "credentials")
	_, err = c.poolSandbox("same", dir, "/bin/true", "/tmp", home)
	require.ErrorContains(t, err, "also an rw path")

	// SandboxArgs resolves bwrap on PATH; CI runners have none.
	if _, err := exec.LookPath("bwrap"); err != nil {
		t.Skip("bwrap not on PATH: argv order not checked")
	}

	// bwrap gets the file bind after the rw dir it sits in, and never
	// the controller's auth.json.
	argv, err := sandbox.Args(*s, "/bin/true", nil, "/tmp")
	require.NoError(t, err)
	joined := strings.Join(argv, " ")
	rw := strings.Index(joined, "--bind "+filepath.Join(home, ".jcode")+" ")
	ro := strings.Index(joined, "--ro-bind "+filepath.Join(h, ".jcode", "prompt-overlay.md"))
	require.GreaterOrEqual(t, rw, 0, joined)
	require.Greater(t, ro, rw, joined)
	assert.NotContains(t, joined, filepath.Join(h, ".jcode", "auth.json"))
	assert.NotContains(t, joined, "--ro-bind "+filepath.Join(h, ".missing"))
}
