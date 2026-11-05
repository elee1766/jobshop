package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
)

func TestInclude(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600))
	}
	write("jobshop.module.yaml", `
shared_tools: [/opt/mise]
attention: {rollback_file: /var/lib/jobshop-deploy/failed}
defaults:
  sandbox: {backend: bubblewrap}
projects:
  js: {prefix: js}
`)
	write("jobshop.yaml", `
include: [jobshop.module.yaml]
shared_tools: [/opt/other]
projects:
  js: {prefix: jx}
  ot: {}
`)
	c, err := config.LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	assert.Equal(t, []string{"/opt/mise", "/opt/other"}, c.SharedTools, "lists append, include first")
	assert.Equal(t, "/var/lib/jobshop-deploy/failed", c.Attention.RollbackFile)
	assert.Equal(t, "bubblewrap", c.Defaults.Sandbox.Backend)
	assert.Equal(t, "jx", c.Projects["js"].Prefix, "the including file wins")
	assert.Contains(t, c.Projects, "ot")

	write("jobshop.yaml", "include: [missing.yaml]\n")
	_, err = config.LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.ErrorContains(t, err, "include[0]")

	write("jobshop.module.yaml", "include: [x.yaml]\n")
	write("jobshop.yaml", "include: [jobshop.module.yaml]\n")
	_, err = config.LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.ErrorContains(t, err, "cannot include")

	write("jobshop.yaml", "include: jobshop.module.yaml\n")
	_, err = config.LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.ErrorContains(t, err, "must be a list")
}
