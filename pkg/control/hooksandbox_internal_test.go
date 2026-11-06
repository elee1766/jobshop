package control

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
)

// TestHookSandboxed (#810): a hook for an event of a sandboxed pool runs
// in that pool's bubblewrap sandbox through proc.Start, so it cannot read
// a file outside its mounts; the same hook with no pool can.
func TestHookSandboxed(t *testing.T) {
	if err := exec.Command("bwrap", "--ro-bind", "/", "/", "--unshare-pid", "--dev", "/dev", "--proc", "/proc", "--", "true").Run(); err != nil {
		t.Skipf("bwrap does not work here: %v", err)
	}
	t.Parallel()
	dir := t.TempDir()
	secret := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, os.WriteFile(secret, []byte("s3cret"), 0o600))
	cfgYAML := fmt.Sprintf(`projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  dev: {agent: fake, dir: %q, sandbox: {backend: bubblewrap, ro: %s}}
hooks:
  - on: action
    command: 'if cat %s >/dev/null 2>&1; then r=read; else r=hidden; fi; echo "{\"decision\":\"deny\",\"reason\":\"$r\"}"'
`, dir, "[/nix/store, /etc, /run/current-system, /bin, /usr/bin, /lib, /lib64, /usr/lib, /usr/lib64]", secret)
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte(cfgYAML), 0o600))
	root := config.Root{Dir: dir}
	cfg, err := config.Load(root)
	require.NoError(t, err)
	lc, err := newLiveConfig(cfg)
	require.NoError(t, err)
	live := new(atomic.Pointer[liveConfig])
	live.Store(lc)
	exe, err := os.Executable()
	require.NoError(t, err)
	c := &Controller{root: root, live: live, opt: Options{HostExe: exe}, log: slog.New(slog.DiscardHandler)}
	c.ctx = t.Context()

	run := func(pool string) hookAnswer {
		return c.runHooks(context.Background(), hookEvent{On: config.HookAction, Action: "push", Pool: pool})
	}
	a := run("dev")
	assert.Equal(t, "hidden", a.Reason, "sandboxed hook cannot read the secret")
	a = run("")
	assert.Equal(t, "read", a.Reason, "no pool: unsandboxed, reads it")
}
