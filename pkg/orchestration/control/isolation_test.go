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

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/acphost/pkg/cgroup"
	"github.com/elee1766/acphost/pkg/sandbox"
	"github.com/elee1766/acphost/pkg/userns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/orchestration/control"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// isolationEnv skips unless this host can run an isolated pool: working
// bwrap, host.Preflight (#724) for newuidmap/newgidmap and pasta, whose
// report (with fix hints) is the skip reason, a subuid/subgid range for
// this user, and a delegated cgroup with cpu, memory and pids. It
// returns the first subuid and subgid.
func isolationEnv(t *testing.T) (int, int) {
	t.Helper()
	bwrapWorks(t)
	// Test roots live under os.TempDir(); the mapped uid must search every
	// dir above it (a TMPDIR in a 0700 home fails the spawn pre-check).
	// Lifted by JOB-53, which checks only what the agent opens.
	for d := os.TempDir(); ; d = filepath.Dir(d) {
		if st, err := os.Stat(d); err != nil || st.Mode().Perm()&0o001 == 0 {
			t.Skipf("isolation: %s is not searchable by other uids; set TMPDIR to a dir like /tmp", d)
		}
		if d == filepath.Dir(d) {
			break
		}
	}
	// Mapped to the caller's own ids, which need no subid range, so the
	// tool checks report with their fix hints on any host. The cgroup is
	// left out: Preflight wants the controllers already enabled in
	// subtree_control, which the controller does at startup.
	rep, err := host.Preflight(host.Spec{
		User:    &userns.User{UID: 1000, GID: 1000, MapUID: os.Geteuid(), MapGID: os.Getegid()},
		Sandbox: &sandbox.Sandbox{Network: "pasta"},
	})
	require.NoError(t, err)
	if !rep.OK() {
		t.Skipf("isolation preflight:\n%s", rep)
	}
	first := func(path string) int {
		f, err := os.Open(path)
		if err != nil {
			t.Skipf("%s: %v", path, err)
		}
		defer func() { _ = f.Close() }()
		rs, err := userns.ParseSubIDs(f, os.Getenv("USER"), os.Geteuid())
		if err != nil || len(rs) == 0 {
			t.Skipf("no %s range for this user", path)
		}
		return rs[0].Start
	}
	uid, gid := first("/etc/subuid"), first("/etc/subgid")
	info, err := cgroup.Delegated()
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

// insideID is the uid, and so the primary gid, agents get in these
// pools. The pools also pass the caller's gid through as a group, which
// cannot be the primary gid too; a login user's gid is often 1000.
func insideID() int {
	if os.Getegid() == 1000 {
		return 1001
	}
	return 1000
}

// isolatedRoot is poolRootReal with the test binary, which is both the
// session host and the fake agent, copied to a 0755 temp dir: go test
// runs it from a 0700 build dir the mapped uid cannot search. It returns
// the copy for the agent command.
func isolatedRoot(t *testing.T) (config.Root, func() func(), string) {
	t.Helper()
	dir, err := os.MkdirTemp("", "jsagent")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.Chmod(dir, 0o755))
	b, err := os.ReadFile(os.Args[0])
	require.NoError(t, err)
	p := filepath.Join(dir, "agent.test")
	require.NoError(t, os.WriteFile(p, b, 0o755))
	root, start := poolRootOpt(t, "1m", control.Options{Hosts: realHosts{}, HostExe: p})
	return root, start, p
}

// quarantined fails the test at once when pool dev is quarantined: the
// spawn error is permanent, so waiting for the session only burns the
// deadline.
func quarantined(t *testing.T, admin *rpc.Client) {
	t.Helper()
	var st rpc.PoolStatusResult
	if admin.Call(t.Context(), rpc.MethodPoolList, nil, &st) == nil {
		if msg, ok := st.Quarantined["dev"]; ok {
			t.Fatalf("pool dev quarantined: %s", msg)
		}
	}
}

// TestIsolatedPool (#388) runs one pool as a mapped uid with a memory
// limit and a pasta network that denies a CIDR. The fake agent writes a
// file (owned by the mapped host uid), tries the denied CIDR (refused),
// then allocates past the limit: the session ends oom.
// Skipped without the environment; docs/verify/isolation.md runs it on
// gasserver.
func TestIsolatedPool(t *testing.T) {
	mapUID, mapGID := isolationEnv(t)
	root, start, exe := isolatedRoot(t)
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
    sandbox: {backend: bubblewrap, network: pasta, deny: [127.0.0.0/8, 10.0.0.0/8], ro: %s}
    isolation:
      user: {uid: %d, map_uid: %d, map_gid: %d, groups: [%q]}
      limits: {memory: 64M, pids: 256}
`, exe, ln.Addr().String(), work, testSystemRO, insideID(), mapUID, mapGID, strconv.Itoa(os.Getegid()))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	var task domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	var ended domain.Ticket
	passEventually(t, root.Dir, func() bool {
		// The first session ends oom; a retry may then quarantine the pool
		// (the agent's files are the mapped uid's), so check ended first.
		var got []domain.Ticket
		if admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Status: domain.StatusClosed, Kind: domain.KindSession}, &got) == nil {
			for _, s := range got {
				if s.Resolution == "oom" {
					ended = s
					return true
				}
			}
		}
		quarantined(t, admin)
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
	if os.Getenv("PEER_PROBE") != "" {
		peerProbe()
		return
	}
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

// TestPerSessionIsolation (#540) runs two sessions of one pool with
// per_session ids. Each gets its own host uid, and neither can read the
// other's worktree. Skipped without the environment, like
// TestIsolatedPool; docs/verify/per-session-isolation.md runs it on
// gasserver.
func TestPerSessionIsolation(t *testing.T) {
	mapUID, mapGID := isolationEnv(t)
	root, start, exe := isolatedRoot(t)
	// The mapped uids traverse the root to reach their session dirs.
	require.NoError(t, os.MkdirAll(filepath.Join(root.Dir, "sessions"), 0o711))
	require.NoError(t, os.Chmod(root.Dir, 0o711))
	require.NoError(t, os.Chmod(filepath.Join(root.Dir, "sessions"), 0o711))
	repo := filepath.Join(root.Dir, "repo")
	for _, args := range [][]string{
		{"init", "-q", "-b", "master", repo},
		{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "init"},
	} {
		out, err := exec.Command("git", args...).CombinedOutput()
		require.NoError(t, err, string(out))
	}
	cfg := fmt.Sprintf(`projects:
  js: {}
agents:
  fake:
    command: %q
    args: [fake-agent]
    env: {PEER_PROBE: "1"}
pools:
  dev:
    project: js
    agent: fake
    max_sessions: 2
    idle_timeout: 1m
    repo: %q
    ref: master
    isolation:
      user: {uid: %d, map_uid: %d, map_gid: %d, groups: [%q], per_session: {range: 2}}
`, exe, repo, insideID(), mapUID, mapGID, strconv.Itoa(os.Getegid()))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	for range 2 {
		var task domain.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	}

	var work, homes []string
	passEventually(t, root.Dir, func() bool {
		quarantined(t, admin)
		var got []domain.Ticket
		if admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Status: domain.StatusOpen, Kind: domain.KindSession}, &got) != nil || len(got) < 2 {
			return false
		}
		work, homes = work[:0], homes[:0]
		for _, s := range got[:2] {
			w := filepath.Join(root.SessionDir(s.ID), "work")
			f, err := domain.DecodeFields[domain.SessionFields](s)
			if err != nil || f.Slot == nil {
				return false
			}
			h := filepath.Join(root.Dir, "homes", "dev-"+strconv.Itoa(*f.Slot))
			if _, err := os.Stat(w); err != nil {
				return false
			}
			if _, err := os.Stat(h); err != nil {
				return false
			}
			work, homes = append(work, w), append(homes, h)
		}
		return true
	}, "two sessions with worktrees and homes")

	// Each per-slot home (JOB-28) gets a secret too; the probe reads the
	// peer's through peer-home, written before peer-go.
	for i, h := range homes {
		require.NoError(t, os.WriteFile(filepath.Join(h, "secret"), []byte("s"), 0o660))
		require.NoError(t, os.WriteFile(filepath.Join(work[i], "peer-home"), []byte(filepath.Join(homes[1-i], "secret")), 0o660))
	}

	// Each worktree gets a secret, then a go file naming the other's
	// secret; the agents wait for it before probing.
	// 0660: with a default ACL the create mode's group bits are the mask,
	// so 0600 would shut out the session's own uid.
	for i, w := range work {
		require.NoError(t, os.WriteFile(filepath.Join(w, "secret"), []byte("s"), 0o660))
		require.NoError(t, os.WriteFile(filepath.Join(w, "peer-go"), []byte(filepath.Join(work[1-i], "secret")), 0o660))
	}
	uids := map[uint32]bool{}
	for _, w := range work {
		res := filepath.Join(w, "peer-probe")
		passEventually(t, root.Dir, func() bool {
			_, err := os.Stat(res)
			return err == nil
		}, "peer probe result")
		b, err := os.ReadFile(res)
		require.NoError(t, err)
		assert.Equal(t, "own: ok\npeer: denied\nown home: ok\npeer home: denied\n", string(b), w)
		st, err := os.Stat(res)
		require.NoError(t, err)
		uids[st.Sys().(*syscall.Stat_t).Uid] = true
	}
	assert.Len(t, uids, 2, "each session writes as its own host uid")

	// Closing the sessions frees their slots and deletes their homes, so
	// the next session in a slot cannot read what the last one left.
	var open []domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Status: domain.StatusOpen, Kind: domain.KindSession}, &open))
	for _, s := range open {
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: s.ID}, nil))
	}
	passEventually(t, root.Dir, func() bool {
		for _, h := range homes {
			if _, err := os.Stat(h); err == nil {
				return false
			}
		}
		return true
	}, "per-session homes removed")
}

// peerProbe runs in the fake agent for TestPerSessionIsolation: wait for
// peer-go, read the own secret and the peer's, and record the result.
func peerProbe() {
	deadline := time.Now().Add(dialWait)
	var peer []byte
	for {
		b, err := os.ReadFile("peer-go")
		if err == nil {
			peer = b
			break
		}
		if time.Now().After(deadline) {
			return
		}
		<-time.After(50 * time.Millisecond)
	}
	res := "own: ok\n"
	if _, err := os.ReadFile("secret"); err != nil {
		res = "own: " + err.Error() + "\n"
	}
	if _, err := os.ReadFile(string(peer)); err == nil {
		res += "peer: read\n"
	} else {
		res += "peer: denied\n"
	}
	if home, err := os.ReadFile("peer-home"); err == nil {
		if _, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), "secret")); err == nil {
			res += "own home: ok\n"
		} else {
			res += "own home: " + err.Error() + "\n"
		}
		if _, err := os.ReadFile(string(home)); err == nil {
			res += "peer home: read\n"
		} else {
			res += "peer home: denied\n"
		}
	}
	_ = os.WriteFile("peer-probe", []byte(res), 0o660)
}

// TestMappedUserHome (JOB-19) runs a pool with isolation.user and a
// "~/" rw path from a 0700 controller root, as the NixOS module sets
// up. The session starts (the mapped uid gets its own home under
// <root>/homes), the agent sees its copied login file at the usual path
// and writes a cache file owned by the mapped uid.
func TestMappedUserHome(t *testing.T) {
	mapUID, mapGID := isolationEnv(t)
	root, start, exe := isolatedRoot(t)
	require.NoError(t, os.Chmod(root.Dir, 0o700))
	// The controller's HOME is shared by parallel tests: use a unique
	// subdir under it.
	h, err := os.UserHomeDir()
	require.NoError(t, err)
	sub := ".jobshop-test-" + filepath.Base(root.Dir)
	home := filepath.Join(h, sub)
	t.Cleanup(func() { _ = os.RemoveAll(home) })
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".agent"), 0o700))
	// The credentials' parent is 0700, as a home-manager home is: the
	// sandbox binds as the controller (JOB-53), so the mapped uid needs
	// no way through it.
	require.NoError(t, os.Chmod(home, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".agent", "auth.json"), []byte("login"), 0o600))
	work := filepath.Join(root.Dir, "work")
	require.NoError(t, os.MkdirAll(work, 0o770))
	cfg := fmt.Sprintf(`projects:
  js: {}
agents:
  fake:
    command: %q
    args: [fake-agent]
    env: {HOME_PROBE: %q}
    credentials: ["~/%s/.agent/auth.json"]
pools:
  dev:
    project: js
    agent: fake
    max_sessions: 1
    idle_timeout: 1m
    dir: %q
    sandbox: {backend: bubblewrap, ro: %s, rw: ["~/%s/.cache/probe"]}
    isolation:
      user: {uid: %d, map_uid: %d, map_gid: %d, groups: [%q]}
`, exe, sub, sub, work, testSystemRO, sub, insideID(), mapUID, mapGID, strconv.Itoa(os.Getegid()))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	require.NoError(t, shareWith(work, os.Getegid()))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	var task domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))

	cache := filepath.Join(root.Dir, "homes", "dev", sub, ".cache", "probe", "home-probe")
	passEventually(t, root.Dir, func() bool {
		quarantined(t, admin)
		_, err := os.Stat(filepath.Join(work, "home-probe"))
		return err == nil
	}, "agent wrote its cache file")
	st, err := os.Stat(cache)
	require.NoError(t, err)
	assert.Equal(t, uint32(mapUID), st.Sys().(*syscall.Stat_t).Uid, "cache file owned by the mapped uid")
	// The cache file is the mapped uid's own; the copy in the shared work
	// dir (setgid, the caller's group) is what the test can read.
	b, err := os.ReadFile(filepath.Join(work, "home-probe"))
	require.NoError(t, err)
	assert.Equal(t, "login", string(b), "agent read its copied login")
	_, err = os.Stat(filepath.Join(home, ".cache", "probe", "home-probe"))
	assert.ErrorIs(t, err, os.ErrNotExist, "nothing written to the controller's home")
}

// shareWith makes dir group gid rwx with setgid, as the module's
// workspace dirs are.
func shareWith(dir string, gid int) error {
	if err := os.Chown(dir, -1, gid); err != nil {
		return err
	}
	return os.Chmod(dir, 0o2770)
}

// homeProbe runs in the fake agent for TestMappedUserHome: copy the
// login it can see at ~/.agent/auth.json into ~/.cache/probe.
func homeProbe() {
	if os.Getenv("HOME_PROBE") == "" {
		return
	}
	h := filepath.Join(os.Getenv("HOME"), os.Getenv("HOME_PROBE"))
	b, err := os.ReadFile(filepath.Join(h, ".agent", "auth.json"))
	if err != nil {
		b = []byte("error: " + err.Error())
	}
	_ = os.WriteFile(filepath.Join(h, ".cache", "probe", "home-probe"), b, 0o660)
	_ = os.WriteFile("home-probe", b, 0o660)
}
