package control

import (
	"os"
	"testing"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/acphost/pkg/cgroup"
	"github.com/elee1766/acphost/pkg/userns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/sessionrt"
)

// TestApplyIsolationNoDelegation: a pool with limits and no delegated
// cgroup fails with a permanent error (pool quarantine); a pool without
// isolation gets nothing (#388).
func TestApplyIsolationNoDelegation(t *testing.T) {
	c, _, _ := scanControllerYAML(t, `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  plain: {agent: fake, dir: %[1]q}
  limited: {agent: fake, dir: %[1]q, isolation: {limits: {memory: 1G}}}
  mapped: {agent: fake, dir: %[1]q, isolation: {user: {uid: 1000, map_uid: 100000, map_gid: 100000, groups: ["0"]}}}
`)
	c.cgroupOnce.Do(func() {})
	c.cgroup = sessionrt.Delegation{Err: os.ErrPermission}

	var s host.Spec
	require.NoError(t, c.applyIsolation(&s, "plain"))
	assert.Nil(t, s.User)
	assert.Nil(t, s.Cgroup)
	assert.Empty(t, c.isolationKey("plain"), "no isolation: env hash unchanged")

	err := c.applyIsolation(&s, "limited")
	require.Error(t, err)
	assert.True(t, host.Permanent(err))
	assert.Contains(t, err.Error(), "no delegated cgroup")

	s = host.Spec{}
	require.NoError(t, c.applyIsolation(&s, "mapped"))
	require.NotNil(t, s.User)
	assert.Equal(t, userns.User{UID: 1000, GID: 1000, MapUID: 100000, MapGID: 100000, Groups: []int{0}}, *s.User)
	assert.NotEmpty(t, c.isolationKey("mapped"))
}

// TestPoolCgroupSystemd (#944): with the systemd driver a limited pool
// gets a scope in the sessions slice and nothing touches the controller's
// cgroup; limits.cgroup alone also gives unlimited pools a scope, so
// their agents count against the slice.
func TestPoolCgroupSystemd(t *testing.T) {
	c, _, _ := scanControllerYAML(t, `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  plain: {agent: fake, dir: %[1]q}
  limited: {agent: fake, dir: %[1]q, isolation: {limits: {memory: 1G, cpu: "2"}}}
limits:
  cgroup_driver: systemd
`)
	c.cgroupOnce.Do(func() {})
	c.cgroup = sessionrt.Delegation{Err: os.ErrPermission}
	cfg := c.config()
	assert.False(t, wantsCgroup(cfg), "no delegated cgroup set up")
	assert.Empty(t, c.hostsCgroup())

	var s host.Spec
	require.NoError(t, c.applyIsolation(&s, "limited"), "no delegation needed")
	require.NotNil(t, s.Cgroup)
	assert.Equal(t, cgroup.Cgroup{
		Driver: cgroup.DriverSystemd, Parent: config.DefaultSessionsSlice,
		CPUMax: "200000 100000", MemoryMax: "1073741824", SwapMax: "0",
	}, *s.Cgroup)
	assert.Contains(t, c.isolationKey("limited"), "JOBSHOP_ISOLATION_SCOPE="+config.DefaultSessionsSlice)

	s = host.Spec{}
	require.NoError(t, c.applyIsolation(&s, "plain"))
	assert.Nil(t, s.Cgroup, "no limits and no limits.cgroup: no scope")
	assert.Empty(t, c.isolationKey("plain"))

	cfg.Limits.Cgroup = &config.HostCgroup{}
	hc, err := c.poolCgroup(cfg, "plain")
	require.NoError(t, err)
	assert.Equal(t, &cgroup.Cgroup{Driver: cgroup.DriverSystemd, Parent: config.DefaultSessionsSlice}, hc)

	spec, ok := c.preflightSpec(cfg, "limited")
	require.True(t, ok)
	assert.Equal(t, cgroup.DriverSystemd, spec.Cgroup.Driver, "preflight runs the systemd checks")
}

// TestHostPlacement (#992): with systemd a session host starts in a scope
// in the sessions slice (InScope), with cgroupfs in the hosts cgroup
// (InCgroup) or the controller's.
func TestHostPlacement(t *testing.T) {
	for _, tc := range []struct {
		name   string
		limits string
		setup  sessionrt.Delegation
		want   sessionrt.Placement
	}{
		{"systemd default slice", "limits:\n  cgroup_driver: systemd\n", sessionrt.Delegation{}, sessionrt.Placement{Systemd: true, Slice: config.DefaultSessionsSlice}},
		{"systemd slice", "limits:\n  cgroup_driver: systemd\n  sessions_slice: x.slice\n  cgroup: {}\n", sessionrt.Delegation{Hosts: "/h"}, sessionrt.Placement{Systemd: true, Slice: "x.slice"}},
		{"cgroupfs split", "limits:\n  cgroup: {}\n", sessionrt.Delegation{Hosts: "/h"}, sessionrt.Placement{Cgroup: "/h"}},
		{"cgroupfs no split", "", sessionrt.Delegation{Hosts: "/h"}, sessionrt.Placement{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _, _ := scanControllerYAML(t, "projects:\n  js: {}\nagents:\n  fake: {command: /bin/false}\npools:\n  plain: {agent: fake, dir: %[1]q}\n"+tc.limits)
			c.cgroupOnce.Do(func() { c.cgroup = tc.setup })
			assert.Equal(t, tc.want, c.hostPlacement())
		})
	}
}
