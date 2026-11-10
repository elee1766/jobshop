package control_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	host "github.com/gfx-labs/acphost/pkg/acphost"
	"github.com/gfx-labs/acphost/pkg/cgroup"
	"github.com/gfx-labs/acphost/pkg/proc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// TestSystemdScopeRestart (#944): with limits.cgroup_driver systemd a
// limited pool's agent runs in a transient scope in the sessions slice.
// The controller restarts twice mid-turn (JOB-99); the host is adopted
// and the scope keeps its memory.max, pids.max and cpu.max. Skipped
// without a systemd manager this process can reach (a user manager, or
// root).
func TestSystemdScopeRestart(t *testing.T) {
	if err := cgroup.CheckManager(); err != nil {
		t.Skipf("no systemd manager: %v", err)
	}
	// A non-root caller must be in its user manager's tree to hand it a
	// process (acphost note 226); a CI runner's job usually is not.
	if b, err := os.ReadFile("/proc/self/cgroup"); os.Geteuid() != 0 && (err != nil || !strings.Contains(string(b), fmt.Sprintf("/user@%d.service/", os.Geteuid()))) {
		t.Skip("not in the user manager's cgroup tree (run under systemd-run --user --scope)")
	}
	root, start := poolRootReal(t, "1m")
	// No dash: a dash nests the slice under its prefix's slice.
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	slice := "jstest" + hex.EncodeToString(b) + ".slice"
	// systemd makes the slice on first use and keeps it; stopping it also
	// ends any scope a failed run left in it.
	t.Cleanup(func() {
		args := []string{"stop", slice}
		if os.Geteuid() != 0 {
			args = append([]string{"--user"}, args...)
		}
		_ = exec.Command("systemctl", args...).Run()
	})
	cfg := fmt.Sprintf(`projects:
  js: {}
agents:
  fake:
    command: %q
    args: [fake-agent]
pools:
  dev:
    project: js
    agent: fake
    max_sessions: 1
    idle_timeout: 1m
    dir: %q
    isolation:
      limits: {memory: 256M, pids: 512, cpu: "1.5"}
limits:
  cgroup_driver: systemd
  sessions_slice: %s
`, os.Args[0], root.Dir, slice)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	stop := start()
	admin := dial(t, root.AdminSocket())
	var task domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	passEventually(t, root.Dir, func() bool {
		quarantined(t, admin)
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, "limited agent started its turn")

	var sessions []domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: domain.KindSession}, &sessions))
	require.Len(t, sessions, 1)
	sdir := root.SessionDir(sessions[0].ID)
	iso, ok, err := host.ReadIsolation(sdir)
	require.NoError(t, err)
	require.True(t, ok)
	require.NotNil(t, iso.Cgroup)
	assert.Equal(t, cgroup.DriverSystemd, iso.Cgroup.Driver)
	unit, err := cgroup.ScopeUnit(sessions[0].ID, 1)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(iso.Cgroup.Unit, strings.TrimSuffix(unit, "1.scope")), iso.Cgroup.Unit)
	assert.Contains(t, iso.Cgroup.Dir, "/"+slice+"/", "the scope is in the sessions slice")
	limits := func() map[string]string {
		m := map[string]string{}
		for _, f := range []string{"memory.max", "pids.max", "cpu.max"} {
			b, err := os.ReadFile(filepath.Join(iso.Cgroup.Dir, f))
			require.NoError(t, err, f)
			m[f] = strings.TrimSpace(string(b))
		}
		return m
	}
	before := limits()
	assert.Equal(t, map[string]string{"memory.max": "268435456", "pids.max": "512", "cpu.max": "150000 100000"}, before)
	procs, err := os.ReadFile(filepath.Join(iso.Cgroup.Dir, "cmd", "cgroup.procs"))
	require.NoError(t, err)
	assert.NotEmpty(t, strings.TrimSpace(string(procs)), "the agent runs in the scope's leaf")
	pid, err := os.ReadFile(filepath.Join(sdir, proc.SupervisorPidFile))
	require.NoError(t, err)
	// The host is in its own scope in the slice, outside the controller's
	// unit (#992).
	hcg, err := os.ReadFile(filepath.Join("/proc", strings.TrimSpace(string(pid)), "cgroup"))
	require.NoError(t, err)
	hostUnit, err := cgroup.HostScopeUnit(sessions[0].ID, 1)
	require.NoError(t, err)
	assert.Contains(t, string(hcg), "/"+slice+"/"+strings.TrimSuffix(hostUnit, "1.scope"), "host in an acphost-host scope in the slice")
	assert.True(t, strings.HasSuffix(strings.TrimSpace(string(hcg)), ".scope"), string(hcg))

	// Two restarts (JOB-99): the cgroupfs EBUSY of JOB-97 showed on the
	// restart after the first.
	for i := range 2 {
		stop()
		require.True(t, hostRunning(sdir), "host outlives the controller (restart %d)", i+1)
		stop = start()
		// Serving: a stop during startup cancels the migration.
		admin = dial(t, root.AdminSocket())
		assert.Equal(t, before, limits(), "the scope keeps its limits across restart %d", i+1)
	}
	defer func() { stop() }()

	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	passEventually(t, root.Dir, func() bool {
		var got domain.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: task.ID}, &struct {
			Ticket *domain.Ticket `json:"ticket"`
		}{&got}))
		return got.Status == domain.StatusClosed
	}, "the adopted turn closes the ticket")
	pid2, err := os.ReadFile(filepath.Join(sdir, proc.SupervisorPidFile))
	require.NoError(t, err)
	assert.Equal(t, string(pid), string(pid2), "same host adopted")
	assert.Equal(t, before, limits(), "limits unchanged after the turn")
}
