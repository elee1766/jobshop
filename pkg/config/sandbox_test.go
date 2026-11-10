package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	packpkg "github.com/gfx-labs/jobshop/pkg/pack"
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
		"sandbox_profiles:\n  x: {network: host, dns: {forward: 169.254.1.1}}\n",
		"sandbox_profiles:\n  x: {dns: {forward: resolver}}\n",
		"sandbox_profiles:\n  x: {dns: {forward: 169.254.1.1, host: localhost}}\n",
		"sandbox_profiles:\n  x: {resolv_conf: resolv.conf}\n",
		"sandbox_profiles:\n  x: {dns_check: \"a b\"}\n",
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
	require.NoError(t, os.WriteFile(filepath.Join(pk, packpkg.FileName), []byte(`sandbox_profiles:
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

// TestSandboxOverlayMerge: layers for one target append profile, agent,
// pool; mode takes the most specific level (JOB-55).
func TestSandboxOverlayMerge(t *testing.T) {
	d := t.TempDir()
	for _, n := range []string{"base", "agent", "pool"} {
		require.NoError(t, os.MkdirAll(filepath.Join(d, n), 0o700))
	}
	c, err := config.Parse(fmt.Appendf(nil, `
projects: {js: {}}
sandbox_profiles:
  skills: {overlay: {"~/.agents": {layers: [%[1]s/base], mode: tmp}}}
agents:
  fake: {command: /bin/true, sandbox: {backend: bubblewrap, profiles: [skills], overlay: {"~/.agents": {layers: [%[1]s/agent]}}}}
pools:
  p: {agent: fake, dir: /tmp, sandbox: {ro: [/usr], overlay: {"~/.agents": {layers: [%[1]s/pool], mode: ro}, /opt/x: {layers: [%[1]s/base]}}}}
`, d))
	require.NoError(t, err)
	ov := c.SandboxFor("p").Overlay
	assert.Equal(t, config.SandboxOverlay{Layers: []string{d + "/base", d + "/agent", d + "/pool"}, Mode: "ro"}, ov["~/.agents"])
	assert.Equal(t, []string{d + "/base"}, ov["/opt/x"].Layers)
}

func TestSandboxOverlayErrors(t *testing.T) {
	d := t.TempDir()
	for _, bad := range []string{
		"pools:\n  p: {agent: fake, dir: /tmp, sandbox: {overlay: {/x: {layers: [" + d + "]}}}}\n",
		"pools:\n  p: {agent: fake, dir: /tmp, sandbox: {backend: bubblewrap, overlay: {/x: {layers: [" + d + "/missing]}}}}\n",
		"pools:\n  p: {agent: fake, dir: /tmp, sandbox: {backend: bubblewrap, overlay: {/x: {layers: [rel]}}}}\n",
		"pools:\n  p: {agent: fake, dir: /tmp, sandbox: {backend: bubblewrap, overlay: {x: {layers: [" + d + "]}}}}\n",
		"pools:\n  p: {agent: fake, dir: /tmp, sandbox: {backend: bubblewrap, overlay: {/x: {layers: [" + d + "], mode: rw}}}}\n",
		"pools:\n  p: {agent: fake, dir: /tmp, sandbox: {backend: bubblewrap, overlay: {/x: {mode: tmp}}}}\n",
		"pools:\n  p: {agent: fake, dir: /tmp, sandbox: {backend: bubblewrap, overlay: {/x: {layers: [" + strings.Repeat(d+", ", 500) + d + "]}}}}\n",
	} {
		_, err := config.Parse([]byte("projects: {js: {}}\nagents: {fake: {command: /bin/true}}\n" + bad))
		assert.Error(t, err, bad)
	}
}
