package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
)

// TestSandboxProfiles: system mounts and env come from config (#625).
// A level's profiles apply before its own fields, so a pool's network
// beats its profile's, and lists append in order.
func TestSandboxProfiles(t *testing.T) {
	c, err := config.Parse([]byte(`
projects: {js: {}}
sandbox_profiles:
  nixos: {ro: [/nix/store, /etc], env_keep: [HOME, "GIT_*"]}
  offline: {network: none}
defaults:
  sandbox: {backend: bubblewrap, profiles: [nixos]}
agents:
  fake: {command: /bin/true, sandbox: {env_keep: [GOPATH]}}
pools:
  a: {agent: fake, dir: /tmp}
  b: {agent: fake, dir: /tmp, sandbox: {profiles: [offline], ro: [/opt/x]}}
  c: {agent: fake, dir: /tmp, sandbox: {profiles: [offline], network: host}}
`))
	require.NoError(t, err)
	a := c.SandboxFor("a")
	assert.Equal(t, []string{"/nix/store", "/etc"}, a.RO)
	assert.Equal(t, []string{"HOME", "GIT_*", "GOPATH"}, a.EnvKeep)
	assert.Empty(t, a.Network)
	b := c.SandboxFor("b")
	assert.Equal(t, "none", b.Network)
	assert.Equal(t, []string{"/nix/store", "/etc", "/opt/x"}, b.RO)
	assert.Equal(t, "host", c.SandboxFor("c").Network, "the level's own field beats its profile")
	assert.Empty(t, c.Warnings)
}

func TestSandboxProfileErrors(t *testing.T) {
	for _, bad := range []string{
		"pools:\n  p: {agent: fake, dir: /tmp, sandbox: {profiles: [nope]}}\n",
		"sandbox_profiles:\n  x: {profiles: [y]}\n  y: {}\n",
		"sandbox_profiles:\n  x: {ro: [relative]}\n",
		"sandbox_profiles:\n  x: {env_keep: [\"A=B\"]}\n",
		"sandbox_profiles:\n  x: {env_keep: [\"*\"]}\n",
	} {
		_, err := config.Parse([]byte("projects: {js: {}}\nagents: {fake: {command: /bin/true}}\n" + bad))
		assert.Error(t, err, bad)
	}
}

// TestSandboxNoROWarns: a sandboxed pool without system mounts cannot
// exec its agent, so loading says so.
func TestSandboxNoROWarns(t *testing.T) {
	c, err := config.Parse([]byte("projects: {js: {}}\nagents: {fake: {command: /bin/true}}\npools:\n  p: {agent: fake, dir: /tmp, sandbox: {backend: bubblewrap}}\n"))
	require.NoError(t, err)
	require.Len(t, c.Warnings, 1)
	assert.Contains(t, c.Warnings[0], "no ro mounts")
}

// TestPackSandboxProfile: a pack supplies a profile; a prefixed pack's
// profile and the references to it are renamed together.
func TestPackSandboxProfile(t *testing.T) {
	dir := t.TempDir()
	pk := filepath.Join(dir, "pk")
	require.NoError(t, os.MkdirAll(pk, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(pk, config.PackFileName), []byte(`sandbox_profiles:
  sys: {ro: [/usr/lib]}
pools:
  dev: {agent: fake, dir: /tmp, sandbox: {backend: bubblewrap, profiles: [sys]}}
`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte(`projects: {js: {}}
agents: {fake: {command: /bin/true}}
packs:
  - {path: pk, prefix: x-}
`), 0o600))
	c, err := config.LoadFile(filepath.Join(dir, config.FileName))
	require.NoError(t, err)
	assert.Contains(t, c.SandboxProfiles, "x-sys")
	assert.Equal(t, []string{"/usr/lib"}, c.SandboxFor("x-dev").RO)
}
