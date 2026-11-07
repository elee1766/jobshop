package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
)

const profilesBase = "projects: {js: {}}\nagents: {fake: {command: /bin/true}}\n"

// TestSandboxMergeOrder (JOB-83): defaults, label rules in config order,
// agent, pool, each level's profiles before its own fields.
func TestSandboxMergeOrder(t *testing.T) {
	c, err := config.Parse([]byte("projects: {js: {}}\n" + `
agents:
  fake:
    command: /bin/true
    sandbox: {profiles: [ag], ro: [/agent]}
sandbox_profiles:
  base: {ro: [/base-profile], network: host}
  ag: {ro: [/agent-profile]}
  r1: {ro: [/rule1-profile]}
  pl: {ro: [/pool-profile], network: none}
defaults:
  sandbox:
    backend: bubblewrap
    profiles: [base]
    ro: [/default]
  sandbox_by_label:
    - labels: ["team:a"]
      ro: [/rule1]
      profiles: [r1]
    - labels: ["team:a", "tier=prod"]
      ro: [/rule2]
      network: none
    - labels: ["team:b"]
      ro: [/other]
pools:
  p: {agent: fake, dir: /tmp, labels: ["team:a", "tier:prod"], sandbox: {profiles: [pl], ro: [/pool]}}
  q: {agent: fake, dir: /tmp, labels: ["team:b"]}
`))
	require.NoError(t, err)
	p := c.SandboxFor("p")
	assert.Equal(t, []string{
		"/base-profile", "/default",
		"/rule1-profile", "/rule1", "/rule2",
		"/agent-profile", "/agent",
		"/pool-profile", "/pool",
	}, p.RO)
	assert.Equal(t, "none", p.Network)
	assert.Equal(t, "bubblewrap", p.Backend)
	assert.Equal(t, []string{"/base-profile", "/default", "/other", "/agent-profile", "/agent"}, c.SandboxFor("q").RO)
	assert.Empty(t, c.Warnings)

	// The explanation names the level of each entry, in merge order.
	var got []string
	for _, e := range c.ExplainSandbox("p") {
		if e.Field == "ro" {
			got = append(got, e.Value+" <- "+e.Source)
		}
	}
	assert.Equal(t, []string{
		"/base-profile <- profile base (defaults)",
		"/default <- defaults",
		"/rule1-profile <- profile r1 (label rule 1 [team:a])",
		"/rule1 <- label rule 1 [team:a]",
		"/rule2 <- label rule 2 [team:a tier=prod]",
		"/agent-profile <- profile ag (agent fake)",
		"/agent <- agent fake",
		"/pool-profile <- profile pl (pool)",
		"/pool <- pool",
	}, got)
	for _, e := range c.ExplainSandbox("p") {
		if e.Field == "network" {
			assert.Equal(t, "none", e.Value)
			assert.Equal(t, "profile pl (pool)", e.Source, "the last level to set a scalar wins")
		}
	}
}

// TestLabelRuleMatchesPackInstanceLabels: a rule scoped by instance:<name>
// reaches the pools a packs entry declares, and not the host's own.
func TestLabelRuleMatchesPackInstanceLabels(t *testing.T) {
	dir := t.TempDir()
	pk := filepath.Join(dir, "pk")
	require.NoError(t, os.MkdirAll(pk, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(pk, config.PackFileName), []byte("pools:\n  dev: {agent: fake, dir: /tmp}\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte(profilesBase+`
sandbox_profiles:
  host: {ro: [/nix/store]}
defaults:
  sandbox_by_label:
    - {labels: ["instance:sap"], profiles: [host], backend: bubblewrap}
pools:
  own: {agent: fake, dir: /tmp}
packs:
  - {path: pk, prefix: sap-}
`), 0o600))
	c, err := config.LoadFile(filepath.Join(dir, config.FileName))
	require.NoError(t, err)
	assert.Equal(t, []string{"/nix/store"}, c.SandboxFor("sap-dev").RO)
	assert.Equal(t, "bubblewrap", c.SandboxFor("sap-dev").Backend)
	assert.Empty(t, c.SandboxFor("own").RO)
	require.Len(t, c.Warnings, 1)
	assert.Contains(t, c.Warnings[0], "YAML packs are deprecated")
}

// TestHostAttachesProfilesToPackPool: a host pool entry adds profiles to
// a pack pool by name, with no pack var, and keeps the pack's own.
func TestHostAttachesProfilesToPackPool(t *testing.T) {
	dir := t.TempDir()
	pk := filepath.Join(dir, "pk")
	require.NoError(t, os.MkdirAll(pk, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(pk, config.PackFileName), []byte(`sandbox_profiles:
  sys: {ro: [/usr/lib]}
isolation_profiles:
  mapped: {user: {uid: 1000, map_uid: 100000, map_gid: 100000, groups: [g]}}
pools:
  dev: {agent: fake, dir: /tmp, sandbox: {backend: bubblewrap, profiles: [sys]}, isolation: {profiles: [mapped]}}
`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte(profilesBase+`
sandbox_profiles:
  host: {ro: [/nix/store]}
isolation_profiles:
  big: {limits: {memory: 8G}}
pools:
  x-dev: {sandbox: {profiles: [host]}, isolation: {profiles: [big]}}
packs:
  - {path: pk, prefix: x-}
`), 0o600))
	c, err := config.LoadFile(filepath.Join(dir, config.FileName))
	require.NoError(t, err)
	assert.Equal(t, []string{"/usr/lib", "/nix/store"}, c.SandboxFor("x-dev").RO)
	assert.Equal(t, "bubblewrap", c.SandboxFor("x-dev").Backend)
	iso := c.IsolationFor("x-dev")
	require.NotNil(t, iso.User)
	require.NotNil(t, iso.Limits)
	assert.Equal(t, 100000, iso.User.MapUID)
	assert.Equal(t, "8G", iso.Limits.Memory)
}

// TestIsolationProfiles: user and limits come whole from the last level
// that sets them, profiles before the level's own fields.
func TestIsolationProfiles(t *testing.T) {
	c, err := config.Parse([]byte(profilesBase + `
isolation_profiles:
  mapped: {user: {uid: 1000, map_uid: 100000, map_gid: 100000, groups: [jobshop]}}
  small: {limits: {memory: 2G, pids: 100}}
  big: {limits: {memory: 8G}}
defaults:
  isolation: {profiles: [mapped]}
  isolation_by_label:
    - {labels: [heavy], profiles: [big]}
pools:
  a: {agent: fake, dir: /tmp, isolation: {profiles: [small]}}
  b: {agent: fake, dir: /tmp, labels: [heavy], isolation: {limits: {cpu: "2"}}}
  c: {agent: fake, dir: /tmp, labels: [heavy], isolation: {profiles: [small]}}
`))
	require.NoError(t, err)
	a := c.IsolationFor("a")
	require.NotNil(t, a.User)
	assert.Equal(t, 100000, a.User.MapUID)
	assert.Equal(t, "2G", a.Limits.Memory)
	assert.Equal(t, int64(100), a.Limits.Pids)
	b := c.IsolationFor("b")
	assert.Equal(t, "2", b.Limits.CPU)
	assert.Empty(t, b.Limits.Memory, "limits are taken whole from the pool, not merged with the rule's profile")
	assert.Equal(t, "2G", c.IsolationFor("c").Limits.Memory, "the pool's profile beats the label rule's")

	var src string
	for _, e := range c.ExplainIsolation("c") {
		if e.Field == "limits" {
			src = e.Source
		}
	}
	assert.Equal(t, "profile small (pool)", src)
}

// TestProfileErrors: unknown names, profiles naming profiles, rules with
// no labels and bad fields fail at load.
func TestProfileErrors(t *testing.T) {
	for name, bad := range map[string]string{
		"unknown sandbox profile on a rule":   "defaults:\n  sandbox_by_label: [{labels: [a], profiles: [nope]}]\n",
		"unknown isolation profile":           "pools:\n  p: {agent: fake, dir: /tmp, isolation: {profiles: [nope]}}\n",
		"unknown isolation profile on a rule": "defaults:\n  isolation_by_label: [{labels: [a], profiles: [nope]}]\n",
		"isolation profile names profiles":    "isolation_profiles:\n  x: {profiles: [y]}\n  y: {}\n",
		"isolation profile bad user":          "isolation_profiles:\n  x: {user: {uid: 1, map_uid: 0, map_gid: 1}}\n",
		"isolation profile bad limits":        "isolation_profiles:\n  x: {limits: {cpu: lots}}\n",
		"isolation profile bad name":          "isolation_profiles:\n  Bad_Name: {}\n",
		"rule without labels":                 "defaults:\n  sandbox_by_label: [{ro: [/x]}]\n",
		"isolation rule without labels":       "defaults:\n  isolation_by_label: [{limits: {pids: 5}}]\n",
		"rule with a bad label":               "defaults:\n  sandbox_by_label: [{labels: [\"a b\"], ro: [/x]}]\n",
		"rule with a relative path":           "defaults:\n  sandbox_by_label: [{labels: [a], ro: [x]}]\n",
		"unknown rule key":                    "defaults:\n  sandbox_by_label: [{labels: [a], nope: 1}]\n",
	} {
		_, err := config.Parse([]byte(profilesBase + bad))
		assert.Error(t, err, name)
	}
}

// TestLabelRuleMatchesNothingWarns: a typo in a rule's label would
// silently drop its profile, so config check says so.
func TestLabelRuleMatchesNothingWarns(t *testing.T) {
	c, err := config.Parse([]byte(profilesBase + `
defaults:
  sandbox_by_label: [{labels: ["team:typo"], ro: [/x]}]
  isolation_by_label: [{labels: ["team:a"], limits: {pids: 5}}]
pools:
  p: {agent: fake, dir: /tmp, labels: ["team:a"]}
`))
	require.NoError(t, err)
	require.Len(t, c.Warnings, 1)
	assert.Contains(t, c.Warnings[0], "sandbox_by_label[0]")
	assert.Contains(t, c.Warnings[0], "team:typo")
}

// TestPackRemovedVarErrors: packs/agent no longer takes sandbox or
// isolation vars; a host that still sets one fails with the replacement.
func TestPackRemovedVarErrors(t *testing.T) {
	pack, err := filepath.Abs("../../packs/agent")
	require.NoError(t, err)
	for _, v := range []string{"sandbox", "isolation"} {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte(profilesBase+
			"packs:\n  - path: "+pack+"\n    prefix: app-\n    vars: {agent: fake, checkout: /tmp, repo: a/b, "+v+": {}}\n"), 0o600))
		_, err := config.LoadFile(filepath.Join(dir, config.FileName))
		require.Error(t, err, v)
		assert.ErrorIs(t, err, config.ErrPack)
		assert.Contains(t, err.Error(), `var "`+v+`" is removed`)
		assert.Contains(t, err.Error(), "profiles")
	}
	// Without them it loads, and the pools carry no sandbox.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte(profilesBase+
		"sandbox_profiles:\n  host: {ro: [/nix/store]}\npools:\n  app-primary: {sandbox: {profiles: [host]}}\npacks:\n  - path: "+pack+"\n    prefix: app-\n    vars: {agent: fake, checkout: /tmp, repo: a/b}\n"), 0o600))
	c, err := config.LoadFile(filepath.Join(dir, config.FileName))
	require.NoError(t, err)
	assert.Equal(t, []string{"/nix/store"}, c.SandboxFor("app-primary").RO)
	assert.Empty(t, c.SandboxFor("app-review").RO)
}

// TestIsolatedPoolAsProfiles: gasserver's isolated pool written inline and
// as profiles resolves to the same sandbox and isolation, so the same
// bwrap argv.
func TestIsolatedPoolAsProfiles(t *testing.T) {
	inline, err := config.Parse([]byte(profilesBase + `
pools:
  isolated:
    agent: fake
    dir: /tmp
    sandbox:
      backend: bubblewrap
      ro: [/nix/store, /etc/ssl, "~/.config/git"]
      rw: ["~/go", "~/.cache"]
      env_keep: [HOME, PATH, "GIT_*"]
      hide: [/proc/sys]
      network: pasta
      deny: [10.0.0.0/8, 192.168.0.0/16]
      dns: {forward: 169.254.1.1, host: 10.0.0.2}
      dns_check: example.com
    isolation:
      user: {uid: 1000, map_uid: 100000, map_gid: 100000, groups: [jobshop], per_session: {range: 8}}
      limits: {memory: 8G, pids: 4096, cpu: "4"}
`))
	require.NoError(t, err)
	profiles, err := config.Parse([]byte(profilesBase + `
sandbox_profiles:
  host:
    backend: bubblewrap
    ro: [/nix/store, /etc/ssl, "~/.config/git"]
    rw: ["~/go", "~/.cache"]
    env_keep: [HOME, PATH, "GIT_*"]
    hide: [/proc/sys]
  egress-deny:
    network: pasta
    deny: [10.0.0.0/8, 192.168.0.0/16]
    dns: {forward: 169.254.1.1, host: 10.0.0.2}
    dns_check: example.com
isolation_profiles:
  mapped:
    user: {uid: 1000, map_uid: 100000, map_gid: 100000, groups: [jobshop], per_session: {range: 8}}
    limits: {memory: 8G, pids: 4096, cpu: "4"}
pools:
  isolated:
    agent: fake
    dir: /tmp
    sandbox: {profiles: [host, egress-deny]}
    isolation: {profiles: [mapped]}
`))
	require.NoError(t, err)
	assert.Equal(t, inline.SandboxFor("isolated"), profiles.SandboxFor("isolated"))
	assert.Equal(t, inline.IsolationFor("isolated"), profiles.IsolationFor("isolated"))
}

// TestProfilesSurviveLastGood: the saved merged config keeps profiles and
// label rules, so a controller started from it merges the same sandbox.
func TestProfilesSurviveLastGood(t *testing.T) {
	c, err := config.Parse([]byte(profilesBase + `
sandbox_profiles:
  host: {backend: bubblewrap, ro: [/nix/store]}
isolation_profiles:
  mapped: {user: {uid: 1000, map_uid: 100000, map_gid: 100000, groups: [g]}}
defaults:
  sandbox_by_label: [{labels: ["tier:iso"], profiles: [host]}]
  isolation_by_label: [{labels: ["tier:iso"], profiles: [mapped]}]
pools:
  p: {agent: fake, dir: /tmp, labels: ["tier:iso"], isolation: {profiles: [mapped]}}
`))
	require.NoError(t, err)
	r := config.Root{Dir: t.TempDir()}
	require.NoError(t, config.SaveLastGood(r.LastGood(), c))
	lg, err := config.LoadLastGood(r.LastGood())
	require.NoError(t, err)
	assert.Equal(t, c.SandboxFor("p"), lg.SandboxFor("p"))
	assert.Equal(t, c.IsolationFor("p"), lg.IsolationFor("p"))
	assert.Equal(t, []string{"/nix/store"}, lg.SandboxFor("p").RO)
}
