package control_test

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	host "github.com/elee1766/acphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// isolationEnv skips unless this host can run an isolated pool: a
// subuid/subgid range for this user, newuidmap and pasta on PATH,
// working bwrap and a delegated cgroup with cpu, memory and pids. It
// returns the first subuid and subgid.
func isolationEnv(t *testing.T) (int, int) {
	t.Helper()
	bwrapWorks(t)
	for _, c := range []string{"newuidmap", "newgidmap", "pasta"} {
		if _, err := exec.LookPath(c); err != nil {
			t.Skipf("%s not on PATH", c)
		}
	}
	first := func(path string) int {
		f, err := os.Open(path)
		if err != nil {
			t.Skipf("%s: %v", path, err)
		}
		defer func() { _ = f.Close() }()
		rs, err := host.ParseSubIDs(f, os.Getenv("USER"), os.Geteuid())
		if err != nil || len(rs) == 0 {
			t.Skipf("no %s range for this user", path)
		}
		return rs[0].Start
	}
	uid, gid := first("/etc/subuid"), first("/etc/subgid")
	info, err := host.DelegatedCgroup()
	if err != nil || !info.Writable {
		t.Skip("no delegated cgroup")
	}
	for _, c := range []string{"cpu", "memory", "pids"} {
		if !strings.Contains(" "+strings.Join(info.Controllers, " ")+" ", " "+c+" ") {
			t.Skipf("cgroup controller %s not delegated", c)
		}
	}
	return uid, gid
}

// TestIsolatedPool (#388) runs one pool as a mapped uid with a memory
// limit and a pasta network that denies a CIDR. The fake agent writes a
// file (owned by the mapped host uid), tries the denied CIDR (refused),
// then allocates past the limit: the session ends oom.
// Skipped without the environment; docs/verify/isolation.md runs it on
// gasserver.
func TestIsolatedPool(t *testing.T) {
	mapUID, mapGID := isolationEnv(t)
	root, start := poolRootReal(t, "1m")
	work := filepath.Join(root.Dir, "work")
	require.NoError(t, os.MkdirAll(work, 0o777))
	require.NoError(t, os.Chmod(work, 0o777))
	// 127.0.0.0/8 holds a listener the agent must not reach.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()
	cfg := fmt.Sprintf(`projects:
  js: {}
agents:
  fake:
    command: %q
    args: [fake-agent]
    env: {ISO_PROBE: %q}
pools:
  dev:
    project: js
    agent: fake
    max_sessions: 1
    idle_timeout: 1m
    dir: %q
    sandbox: {backend: bubblewrap, network: pasta, deny: [127.0.0.0/8, 10.0.0.0/8]}
    isolation:
      user: {uid: 1000, map_uid: %d, map_gid: %d, groups: [%q]}
      limits: {memory: 64M, pids: 256}
`, os.Args[0], ln.Addr().String(), work, mapUID, mapGID, strconv.Itoa(os.Getegid()))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	var task store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	var ended store.Ticket
	passEventually(t, root.Dir, func() bool {
		var got []store.Ticket
		if admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Status: store.StatusClosed, Kind: store.KindSession}, &got) == nil && len(got) > 0 {
			ended = got[0]
			return true
		}
		return false
	}, "isolated session ended")
	assert.Equal(t, "oom", ended.Resolution, "the memory limit killed the agent")

	st, err := os.Stat(filepath.Join(work, "iso-probe"))
	require.NoError(t, err)
	assert.Equal(t, uint32(mapUID), st.Sys().(*syscall.Stat_t).Uid, "files show the mapped host uid")
	b, err := os.ReadFile(filepath.Join(work, "iso-probe"))
	require.NoError(t, err)
	assert.Contains(t, string(b), "denied: unreachable")
}

// isolationProbe runs in the fake agent for TestIsolatedPool: write a
// file, dial the denied address, then allocate until killed.
// ISO_PROBE is the denied host:port.
func isolationProbe() {
	addr := os.Getenv("ISO_PROBE")
	if addr == "" {
		return
	}
	res := "denied: unreachable"
	if c, err := net.DialTimeout("tcp", addr, 2*time.Second); err == nil {
		_ = c.Close()
		res = "denied: reachable"
	}
	_ = os.WriteFile("iso-probe", []byte(res+"\n"), 0o666)
	var hold [][]byte
	for {
		b := make([]byte, 16<<20)
		for i := range b {
			b[i] = 1
		}
		hold = append(hold, b)
		_ = len(hold)
	}
}
