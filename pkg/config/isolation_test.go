package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
)

const isoBase = `
projects: {js: {}}
agents:
  fake:
    command: /bin/true
    isolation: {limits: {cpu: "2", memory: 4G}}
pools:
`

// TestIsolationMerge: user and limits each come whole from the most
// specific level; pasta deny CIDRs append like sandbox paths (#388).
func TestIsolationMerge(t *testing.T) {
	c, err := config.Parse([]byte(`
projects: {js: {}}
defaults:
  isolation: {user: {uid: 1000, map_uid: 231072, map_gid: 231072, groups: [jobshop, "100"]}}
  sandbox: {backend: bubblewrap, network: pasta, deny: [10.0.0.0/8]}
agents:
  fake:
    command: /bin/true
    isolation: {limits: {cpu: "2", memory: 4G}}
pools:
  a: {agent: fake, dir: /tmp}
  b:
    agent: fake
    dir: /tmp
    isolation: {limits: {memory: 1G, pids: 64}}
    sandbox: {deny: [192.168.0.0/16]}
  c:
    agent: fake
    dir: /tmp
    sandbox: {network: host}
`))
	require.NoError(t, err)
	a := c.IsolationFor("a")
	require.NotNil(t, a.User)
	assert.Equal(t, 231072, a.User.MapUID)
	assert.Equal(t, 1000, a.User.GIDOrUID(), "gid defaults to uid")
	assert.Equal(t, []string{"jobshop", "100"}, a.User.Groups)
	require.NotNil(t, a.Limits)
	assert.Equal(t, "4G", a.Limits.Memory)

	b := c.IsolationFor("b")
	assert.Equal(t, "1G", b.Limits.Memory)
	assert.Empty(t, b.Limits.CPU, "pool limits replace the agent's whole")
	assert.Equal(t, []string{"10.0.0.0/8", "192.168.0.0/16"}, c.SandboxFor("b").Deny)
	assert.Equal(t, "pasta", c.SandboxFor("b").Network)
	assert.Empty(t, c.SandboxFor("c").Deny, "deny is dropped off pasta")

	// No isolation anywhere: zero value, nothing changes (#388).
	plain, err := config.Parse([]byte("projects: {js: {}}\nagents: {f: {command: /bin/true}}\npools: {p: {agent: f, dir: /tmp}}\n"))
	require.NoError(t, err)
	assert.Equal(t, config.Isolation{}, plain.IsolationFor("p"))
}

func TestIsolationErrors(t *testing.T) {
	for _, bad := range []string{
		"  p: {agent: fake, dir: /tmp, isolation: {limits: {cpu: lots}}}\n",
		"  p: {agent: fake, dir: /tmp, isolation: {limits: {cpu: \"0\"}}}\n",
		"  p: {agent: fake, dir: /tmp, isolation: {limits: {memory: 4X}}}\n",
		"  p: {agent: fake, dir: /tmp, isolation: {limits: {pids: -1}}}\n",
		"  p: {agent: fake, dir: /tmp, isolation: {user: {uid: 0, map_uid: 0, map_gid: 1}}}\n",
		"  p: {agent: fake, dir: /tmp, isolation: {user: {uid: 0, map_uid: 1, map_gid: 1, groups: [\"-3\"]}}}\n",
		"  p: {agent: fake, dir: /tmp, sandbox: {network: pasta, deny: [not-a-cidr]}}\n",
		"  p: {agent: fake, dir: /tmp, sandbox: {network: host, deny: [10.0.0.0/8]}}\n",
		"  p: {agent: fake, dir: /tmp, sandbox: {network: slirp}}\n",
		`  p: {agent: fake, dir: /tmp, isolation: {user: {uid: 0, map_uid: 1, map_gid: 1, groups: ["100"], per_session: {range: 0}}}}` + "\n",
		`  p: {agent: fake, dir: /tmp, isolation: {user: {uid: 0, map_uid: 1, map_gid: 1, per_session: {range: 4}}}}` + "\n",
	} {
		_, err := config.Parse([]byte(isoBase + bad))
		assert.Error(t, err, bad)
	}
}

func TestPerSession(t *testing.T) {
	c, err := config.Parse([]byte(isoBase + `  p: {agent: fake, dir: /tmp, isolation: {user: {uid: 0, map_uid: 100000, map_gid: 100000, groups: ["100"], per_session: {range: 8}}}}` + "\n"))
	require.NoError(t, err)
	assert.Equal(t, 8, c.IsolationFor("p").User.Slots())
	assert.Equal(t, 0, config.IsolationUser{}.Slots())
}

func TestLimitsValues(t *testing.T) {
	l := config.CgroupLimits{CPU: "0.5"}
	got, err := l.CPUMax()
	require.NoError(t, err)
	assert.Equal(t, "50000 100000", got)
	for in, want := range map[string]string{"": "", "max": "max", "1024": "1024", "8G": "8589934592", "512m": "536870912", "2K": "2048"} {
		got, err := config.ParseBytes(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	_, err = config.ParseBytes("99999999999999999999T")
	assert.Error(t, err)
}
