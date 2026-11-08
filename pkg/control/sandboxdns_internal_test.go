package control

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPoolSandboxDNS: sandbox dns and resolv_conf reach acphost's Sandbox
// (JOB-20), change the env hash so a reload respawns idle sessions, and
// the DNS check runs only for a pasta pool that names dns_check. The
// pasta lookup itself is acphost's CheckSandboxDNS, tested there.
func TestPoolSandboxDNS(t *testing.T) {
	c, _, dir := scanControllerYAML(t, `projects:
  js: {}
agents:
  a: {command: /bin/true}
pools:
  dns:
    agent: a
    dir: /tmp
    sandbox:
      backend: bubblewrap
      network: pasta
      deny: [10.0.0.0/8]
      dns: {forward: 169.254.1.1, host: 127.0.0.53}
      resolv_conf: /etc/hosts
  plain: {agent: a, dir: /tmp, sandbox: {backend: bubblewrap, network: pasta}}
  host: {agent: a, dir: /tmp, sandbox: {backend: bubblewrap, dns_check: example.com}}
# %s
`)
	s, err := c.poolSandbox("dns", dir, "/bin/true", "/tmp", "")
	require.NoError(t, err)
	require.NotNil(t, s.DNS)
	assert.Equal(t, "169.254.1.1", s.DNS.Forward)
	assert.Equal(t, "127.0.0.53", s.DNS.Host)
	assert.Equal(t, "/etc/hosts", s.ResolvConf)
	assert.NotEqual(t, c.sandboxKey("dns"), c.sandboxKey("plain"))

	p, err := c.poolSandbox("plain", dir, "/bin/true", "/tmp", "")
	require.NoError(t, err)
	assert.Nil(t, p.DNS)
	assert.Empty(t, p.ResolvConf)

	// No dns_check, or not pasta: nothing is spawned and nothing fails.
	require.NoError(t, c.checkSandboxDNS("dns", s))
	require.NoError(t, c.checkSandboxDNS("plain", p))
	h, err := c.poolSandbox("host", dir, "/bin/true", "/tmp", "")
	require.NoError(t, err)
	require.NoError(t, c.checkSandboxDNS("host", h))
	assert.Zero(t, c.dnsChecks.Len())
}
