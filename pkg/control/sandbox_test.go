package control_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// bwrapWorks skips the test unless unprivileged bwrap runs here.
func bwrapWorks(t *testing.T) {
	t.Helper()
	if err := exec.Command("bwrap", "--ro-bind", "/", "/", "--unshare-pid", "--dev", "/dev", "--proc", "/proc", "--", "true").Run(); err != nil {
		t.Skipf("bwrap does not work here: %v", err)
	}
}

// TestSandboxedSession runs the fake agent under bubblewrap (#262): it
// writes its probe results, which show it cannot read admin.sock,
// jobshop.db or another session's dir, can write its own dir, and
// still closes its ticket over cli.sock.
func TestSandboxedSession(t *testing.T) {
	bwrapWorks(t)
	root, start := poolRoot(t, "1m")
	other := filepath.Join(root.Dir, "sessions", "js-other")
	require.NoError(t, os.MkdirAll(other, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(other, "secret"), []byte("x"), 0o600))
	work := filepath.Join(root.Dir, "work")
	require.NoError(t, os.MkdirAll(work, 0o700))
	// The test binary (host and agent) lives in a temp dir that the
	// sandbox must see; it is bound through the agent command's dir.
	cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\n    env: {SANDBOX_PROBE: %q}\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: 1m\n    dir: %q\n    sandbox: {backend: bubblewrap}\n",
		os.Args[0], root.Dir+"|"+other, work)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	var task store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	var endedWhy string
	require.Eventually(t, func() bool {
		var got []store.Ticket
		err := admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Status: store.StatusClosed, Kind: store.KindTask}, &got)
		if err == nil && len(got) == 1 && got[0].ID == task.ID {
			return true
		}
		// A session that failed to spawn or ended fails the test now,
		// with its reason, instead of waiting out the timeout (#359).
		var ended []store.Ticket
		if admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Status: store.StatusClosed, Kind: store.KindSession}, &ended) == nil && len(ended) > 0 {
			var res rpc.SessionSandboxResult
			_ = admin.Call(t.Context(), rpc.MethodSessionSandbox, rpc.SessionSandboxParams{ID: ended[0].ID}, &res)
			endedWhy = fmt.Sprintf("session %s ended %q before the ticket closed: %s", ended[0].ID, ended[0].Resolution, res.Error)
			return true
		}
		return false
	}, dialWait, 50*time.Millisecond, "ticket closed over cli.sock from inside the sandbox")
	require.Empty(t, endedWhy)

	b, err := os.ReadFile(filepath.Join(work, "probe"))
	require.NoError(t, err)
	probe := string(b)
	assert.Contains(t, probe, "admin.sock: hidden")
	assert.Contains(t, probe, "jobshop.db: hidden")
	assert.Contains(t, probe, "other: hidden")
	assert.Contains(t, probe, "pid: 2")

	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	require.NotEmpty(t, sessions)
	var res rpc.SessionSandboxResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionSandbox, rpc.SessionSandboxParams{ID: sessions[0].ID}, &res))
	require.NotNil(t, res.Sandbox)
	assert.Contains(t, strings.Join(res.Argv, " "), "--bind "+work+" "+work)
	assert.Empty(t, res.Error)
}

// sandboxProbe records what the agent can see, for TestSandboxedSession.
// SANDBOX_PROBE is "<root>|<other session dir>".
func sandboxProbe() {
	spec := os.Getenv("SANDBOX_PROBE")
	if spec == "" {
		return
	}
	root, other, _ := strings.Cut(spec, "|")
	var out strings.Builder
	seen := func(name, p string) {
		if _, err := os.Stat(p); err != nil {
			fmt.Fprintf(&out, "%s: hidden\n", name)
		} else {
			fmt.Fprintf(&out, "%s: visible\n", name)
		}
	}
	seen("admin.sock", filepath.Join(root, "admin.sock"))
	seen("jobshop.db", filepath.Join(root, "jobshop.db"))
	seen("other", filepath.Join(other, "secret"))
	fmt.Fprintf(&out, "pid: %d\n", os.Getpid())
	_ = os.WriteFile("probe", []byte(out.String()), 0o600)
}
