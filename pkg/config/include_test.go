package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
)

func TestInclude(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600))
	}
	write("jobshop.module.yaml", `
shared_tools: [/opt/mise]
attention: {resource_wait: 1h}
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
	assert.Equal(t, "1h", c.Attention.ResourceWait)
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

// TestLinkedConfigPacks: a root jobshop.yaml linked from a checkout
// resolves its packs against the checkout and its include against the
// root (#532).
func TestLinkedConfigPacks(t *testing.T) {
	root, repo := t.TempDir(), t.TempDir()
	write := func(path, body string) {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	}
	write(filepath.Join(root, "jobshop.module.yaml"), "agents:\n  a:\n    command: x\n")
	write(filepath.Join(repo, "packs/p/pack.yaml"), "pools:\n  dev:\n    agent: a\n    dir: /w\n    prompt_file: dev.md\n")
	write(filepath.Join(repo, "packs/p/dev.md"), "do work")
	write(filepath.Join(repo, "jobshop.yaml"), "include: [jobshop.module.yaml]\npack: pack.yaml\n")
	write(filepath.Join(repo, "pack.yaml"), "projects:\n  js: {}\npacks:\n  - path: packs/p\n")
	require.NoError(t, os.Symlink(filepath.Join(repo, "jobshop.yaml"), filepath.Join(root, "jobshop.yaml")))
	// The include is beside the link, so the checkout copy cannot load it.
	write(filepath.Join(root, "jobshop.module.yaml"), "agents:\n  a:\n    command: x\n")

	c, err := config.LoadFile(filepath.Join(root, "jobshop.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "do work", c.Pools["dev"].Prompt)
	assert.Equal(t, "x", c.Agents["a"].Command)
}
