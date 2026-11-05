package config_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
)

// TestDeployPack loads packs/deploy (#577) and runs its deploy-state
// desk command against a state dir: JSON the Deploys plugin reads, with
// empty values when a file is missing.
func TestDeployPack(t *testing.T) {
	pack, err := filepath.Abs("../../packs/deploy")
	require.NoError(t, err)
	dir := t.TempDir()
	state := filepath.Join(dir, "state")
	require.NoError(t, os.MkdirAll(state, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "jobshop.yaml"), []byte("projects:\n  js: {}\npacks:\n  - path: "+pack+"\n    vars: {state_dir: "+state+"}\n"), 0o600))
	c, err := config.LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	dc, ok := c.DeskCommands["deploy-state"]
	require.True(t, ok)

	run := func() map[string]any {
		t.Helper()
		out, err := exec.Command("sh", "-c", dc.Command).Output()
		require.NoError(t, err)
		var m map[string]any
		require.NoError(t, json.Unmarshal(out, &m), string(out))
		return m
	}
	m := run()
	assert.Equal(t, "", m["deployed"])
	assert.InDelta(t, 0, m["failed_at"], 0)

	require.NoError(t, os.WriteFile(filepath.Join(state, "deployed"), []byte("abc123\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(state, "failed"), []byte("def456\n"), 0o600))
	m = run()
	assert.Equal(t, "abc123", m["deployed"])
	assert.Equal(t, "def456", m["failed"])
	assert.Greater(t, m["deployed_at"], float64(0))
}

// TestAttentionDeployKeysRemoved: a config that still sets the removed
// deploy attention keys loads with a warning naming the replacement
// (#577). Self-deploy checks the new binary against the old module's
// yaml, which still sets them, so an error would block the deploy.
func TestAttentionDeployKeysRemoved(t *testing.T) {
	for _, key := range []string{"rollback_file", "config_repo_file"} {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "jobshop.yaml"), []byte("projects:\n  js: {}\nattention: {"+key+": /x}\n"), 0o600))
		c, err := config.LoadFile(filepath.Join(dir, "jobshop.yaml"))
		require.NoError(t, err)
		require.Len(t, c.Warnings, 1)
		assert.Contains(t, c.Warnings[0], "attention."+key+" is removed")
		assert.Contains(t, c.Warnings[0], "ticket upsert")
	}
}
