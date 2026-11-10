package sandbox

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elee1766/acphost/pkg/proc"
	"github.com/elee1766/acphost/pkg/xerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
)

func TestKeepEnv(t *testing.T) {
	keep := []string{"HOME", "JOBSHOP_*"}
	assert.True(t, KeepEnv(keep, "HOME"))
	assert.True(t, KeepEnv(keep, "JOBSHOP_SOCKET"))
	assert.False(t, KeepEnv(keep, "HOMEX"))
	assert.False(t, KeepEnv(keep, "PATH"))
	assert.False(t, KeepEnv(nil, "HOME"))
}

func TestWithin(t *testing.T) {
	assert.True(t, Within("/a/b", "/a/b"))
	assert.True(t, Within("/a/b/c", "/a/b"))
	assert.False(t, Within("/a/bc", "/a/b"))
	assert.False(t, Within("/a", "/a/b"))
}

func TestExpandHome(t *testing.T) {
	h := t.TempDir()
	t.Setenv("HOME", h)
	assert.Equal(t, filepath.Join(h, ".x"), ExpandHome("~/.x"))
	assert.Equal(t, "/etc", ExpandHome("/etc"))
	assert.Equal(t, "~x", ExpandHome("~x"))
}

func TestHasEnv(t *testing.T) {
	assert.True(t, HasEnv([]string{"A=1", "PATH=/bin"}, "PATH"))
	assert.False(t, HasEnv([]string{"PATHX=1"}, "PATH"))
}

// TestOverlayMountsPersist: a persist overlay keeps its upper dir under
// own and recreates an empty work dir; a missing layer is permanent.
func TestOverlayMountsPersist(t *testing.T) {
	own, layer := t.TempDir(), t.TempDir()
	ov := map[string]config.SandboxOverlay{"/opt/x": {Layers: []string{layer}, Mode: config.OverlayPersist}}
	got, err := overlayMounts(ov, own)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, filepath.Join(own, "overlay", "opt_x", "upper"), got[0].Upper)
	assert.DirExists(t, got[0].Work)

	// The work dir must be empty on each spawn.
	require.NoError(t, os.WriteFile(filepath.Join(got[0].Work, "stale"), nil, 0o600))
	_, err = overlayMounts(ov, own)
	require.NoError(t, err)
	assert.NoFileExists(t, filepath.Join(got[0].Work, "stale"))

	require.NoError(t, os.RemoveAll(layer))
	_, err = overlayMounts(ov, own)
	require.ErrorContains(t, err, "missing")
	assert.ErrorAs(t, err, new(*xerr.PermanentError))
}

// TestIsolateSpecNoConfig: a command with a pool and no config is an
// error, not an unisolated run.
func TestIsolateSpecNoConfig(t *testing.T) {
	_, err := IsolateSpec(&proc.Spec{}, CommandSetup{Pool: "p"})
	require.ErrorContains(t, err, "no config")
}

// TestIsolateSpecPlainPool: a pool with no sandbox, user or limits
// leaves the spec alone and asks the controller for nothing.
func TestIsolateSpecPlainPool(t *testing.T) {
	cfg := &config.Config{Pools: map[string]config.Pool{"p": {}}}
	spec := proc.Spec{Command: "true", Dir: "/tmp"}
	cleanup, err := IsolateSpec(&spec, CommandSetup{Cfg: cfg, Pool: "p"})
	require.NoError(t, err)
	assert.Nil(t, cleanup)
	assert.Equal(t, proc.Spec{Command: "true", Dir: "/tmp"}, spec)
}

func TestKeyUnsandboxed(t *testing.T) {
	cfg := &config.Config{Pools: map[string]config.Pool{"p": {}}}
	assert.Nil(t, Key(cfg, "p"))
}

func TestDNSChecksSkipsWithoutCheck(t *testing.T) {
	cfg := &config.Config{Pools: map[string]config.Pool{"p": {}}}
	var d DNSChecks
	require.NoError(t, d.Check(cfg, "p", nil))
	assert.Zero(t, d.Len())
}
