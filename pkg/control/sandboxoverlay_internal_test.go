package control

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	host "github.com/elee1766/acphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPoolSandboxOverlay: agent and pool layers merge onto one target,
// the pool's copy of a conflicting file wins, and a tmp overlay write
// does not reach the layers (JOB-55).
func TestPoolSandboxOverlay(t *testing.T) {
	h := t.TempDir()
	t.Setenv("HOME", h)
	base, top := t.TempDir(), t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(base, "f"), []byte("agent"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(base, "only-base"), []byte("b"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(top, "f"), []byte("pool"), 0o644))
	c, _, dir := stallControllerYAML(t, `projects:
  js: {}
agents:
  a: {command: /bin/sh, sandbox: {backend: bubblewrap, overlay: {"~/.agents": {layers: [`+base+`], mode: ro}}}}
pools:
  g:
    agent: a
    dir: /tmp
    sandbox:
      ro: [/nix/store, /etc, /run/current-system, /bin, /usr/bin, /lib, /lib64, /usr/lib, /usr/lib64]
      overlay: {"~/.agents": {layers: [`+top+`], mode: tmp}}
# %s
`)
	s, err := c.poolSandbox("g", dir, "/bin/sh", "/tmp", "")
	require.NoError(t, err)
	require.Len(t, s.Overlay, 1)
	assert.Equal(t, filepath.Join(h, ".agents"), s.Overlay[0].Dst)
	assert.Equal(t, []string{base, top}, s.Overlay[0].Layers)
	assert.Equal(t, "tmp", s.Overlay[0].Mode)

	if err := exec.Command("bwrap", "--ro-bind", "/", "/", "--unshare-pid", "--dev", "/dev", "--proc", "/proc", "--", "true").Run(); err != nil {
		t.Skipf("bwrap does not work here: %v", err)
	}
	t.Chdir(t.TempDir())
	target := filepath.Join(h, ".agents")
	argv, err := host.SandboxArgs(*s, "/bin/sh", []string{"-c", "read a < " + target + "/f; read b < " + target + "/only-base; echo $a$b; echo x > " + target + "/new"}, "/tmp")
	require.NoError(t, err)
	out, err := exec.Command(argv[0], argv[1:]...).CombinedOutput()
	require.NoError(t, err, "%s\n%s", out, strings.Join(argv, " "))
	assert.Equal(t, "poolb", strings.TrimSpace(string(out)))
	for _, d := range []string{base, top} {
		assert.NoFileExists(t, filepath.Join(d, "new"))
	}

	// A layer removed after load quarantines the pool at spawn.
	require.NoError(t, os.RemoveAll(top))
	_, err = c.poolSandbox("g", dir, "/bin/sh", "/tmp", "")
	require.ErrorContains(t, err, "missing")
	assert.True(t, host.Permanent(err))
}
