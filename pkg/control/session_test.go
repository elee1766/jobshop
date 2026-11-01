package control_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/host"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// The test binary is also the session host and the agent, so the test runs
// the real process tree: controller -> host -> agent.
func TestMain(m *testing.M) {
	if len(os.Args) >= 3 && os.Args[1] == "session-host" {
		if err := host.Main(os.Args[2]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if len(os.Args) >= 2 && os.Args[1] == "fake-agent" {
		fakeAgent()
		os.Exit(0)
	}
	if len(os.Args) >= 3 && os.Args[1] == "hold-lock" {
		holdLock(os.Args[2])
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// TestControllerRestart restarts the controller mid-turn: the host and
// agent survive, the new controller adopts them, and the turn's result
// still closes the ticket.
func TestControllerRestart(t *testing.T) {
	root, start := poolRoot(t, "1m")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	stop := start()

	admin := dial(t, root.AdminSocket())
	var task store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	require.Eventually(t, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, 20*time.Second, 20*time.Millisecond)

	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	require.Len(t, sessions, 1)
	sdir := root.SessionDir(sessions[0].ID)
	pid, err := os.ReadFile(filepath.Join(sdir, host.PidFile))
	require.NoError(t, err)

	stop()
	require.True(t, host.Alive(sdir), "host outlives the controller")
	stop = start()
	defer stop()
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))

	admin = dial(t, root.AdminSocket())
	show := func(id string) store.Ticket {
		var got store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &struct {
			Ticket *store.Ticket `json:"ticket"`
		}{&got}))
		return got
	}
	require.Eventually(t, func() bool { return show(task.ID).Status == store.StatusClosed }, 20*time.Second, 50*time.Millisecond)
	pid2, err := os.ReadFile(filepath.Join(sdir, host.PidFile))
	require.NoError(t, err)
	assert.Equal(t, string(pid), string(pid2), "same host adopted")
	assert.Equal(t, store.StatusOpen, show(sessions[0].ID).Status)
}

// TestEnvRefresh adopts an idle host whose spec env no longer matches what
// the controller would write: it is respawned and the session stays open.
func TestEnvRefresh(t *testing.T) {
	root, start := poolRoot(t, "1m")
	stop := start()

	admin := dial(t, root.AdminSocket())
	var task store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	show := func(id string) store.Ticket {
		var got store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &struct {
			Ticket *store.Ticket `json:"ticket"`
		}{&got}))
		return got
	}
	require.Eventually(t, func() bool { return show(task.ID).Status == store.StatusClosed }, 20*time.Second, 50*time.Millisecond)
	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	require.Len(t, sessions, 1)
	sdir := root.SessionDir(sessions[0].ID)
	pid, err := os.ReadFile(filepath.Join(sdir, host.PidFile))
	require.NoError(t, err)
	stop()

	// Pretend the host was spawned by an older deploy with another env.
	spec, err := host.ReadSpec(sdir)
	require.NoError(t, err)
	spec.Env = append(spec.Env, "OLD_DEPLOY=1")
	spec.EnvHash = "old"
	require.NoError(t, host.WriteSpec(sdir, spec))

	stop = start()
	defer stop()
	admin = dial(t, root.AdminSocket())
	require.Eventually(t, func() bool {
		pid2, err := os.ReadFile(filepath.Join(sdir, host.PidFile))
		return err == nil && string(pid2) != string(pid) && host.Alive(sdir)
	}, 20*time.Second, 50*time.Millisecond, "host respawned")
	require.Eventually(t, func() bool {
		s, err := host.ReadSpec(sdir)
		return err == nil && s.EnvHash != "old"
	}, 5*time.Second, 50*time.Millisecond)
	assert.Equal(t, store.StatusOpen, show(sessions[0].ID).Status)
}

// TestDrain drains during a blocked turn: the turn finishes and closes its
// ticket, no other ticket is claimed, the session ends drained, and
// undrain lets a new session take the waiting work.
func TestDrain(t *testing.T) {
	root, start := poolRoot(t, "1m")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	stop := start()
	defer stop()

	admin := dial(t, root.AdminSocket())
	show := func(id string) store.Ticket {
		var got store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &struct {
			Ticket *store.Ticket `json:"ticket"`
		}{&got}))
		return got
	}
	var first, second store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Pool: "dev"}, &first))
	require.Eventually(t, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, 20*time.Second, 20*time.Millisecond)

	var dr rpc.DrainResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodDrain, rpc.DrainParams{Pool: "dev"}, &dr))
	assert.Equal(t, []string{"dev"}, dr.Drained)
	var infos []rpc.SessionInfo
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &infos))
	require.Len(t, infos, 1)
	assert.True(t, infos[0].Drained)
	sessID := infos[0].ID

	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &second))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))

	require.Eventually(t, func() bool { return show(sessID).Status == store.StatusClosed }, 20*time.Second, 50*time.Millisecond)
	assert.Equal(t, "drained", show(sessID).Resolution)
	assert.Equal(t, store.StatusClosed, show(first.ID).Status, "running turn finished")
	assert.Equal(t, store.StatusOpen, show(second.ID).Status, "no new claim while drained")
	time.Sleep(300 * time.Millisecond) // several supervisor ticks
	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	assert.Len(t, sessions, 1, "no new session while drained")

	require.NoError(t, admin.Call(t.Context(), rpc.MethodDrain, rpc.DrainParams{Undrain: true}, &dr))
	assert.Empty(t, dr.Drained)
	require.Eventually(t, func() bool { return show(second.ID).Status == store.StatusClosed }, 20*time.Second, 50*time.Millisecond)
}

// TestSessionList joins a running session's ticket with its live host
// state and reports a locked session directory with no ticket as an orphan.
func TestSessionList(t *testing.T) {
	root, start := poolRoot(t, "1m")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	// An orphan: a session directory whose lock is held (here by the test
	// process) but that no session ticket names. Made after startup so the
	// controller's startup reaping does not stop it.
	odir := root.SessionDir("js-orphan")
	require.NoError(t, os.MkdirAll(odir, 0o700))
	lf, err := os.Create(filepath.Join(odir, host.LockFile))
	require.NoError(t, err)
	defer func() { _ = lf.Close() }()
	require.NoError(t, syscall.Flock(int(lf.Fd()), syscall.LOCK_EX))
	// The pid is a throwaway child: reconciliation may signal it.
	sleeper := exec.Command("sleep", "60")
	require.NoError(t, sleeper.Start())
	t.Cleanup(func() { _ = sleeper.Process.Kill(); _ = sleeper.Wait() })
	require.NoError(t, os.WriteFile(filepath.Join(odir, host.PidFile), fmt.Appendf(nil, "%d\n", sleeper.Process.Pid), 0o600))

	var task store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	require.Eventually(t, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, 20*time.Second, 20*time.Millisecond)

	var list []rpc.SessionInfo
	require.Eventually(t, func() bool {
		list = nil
		require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &list))
		return len(list) == 2 && list[0].AgentPid != 0
	}, 10*time.Second, 50*time.Millisecond)

	s := list[0]
	require.NotNil(t, s.Ticket)
	assert.False(t, s.Orphan)
	assert.True(t, s.Alive)
	assert.Equal(t, task.ID, s.Current)
	assert.Positive(t, s.HostPid)
	assert.Positive(t, s.Seq, "journal has frames")
	assert.NotNil(t, s.SeqTime)
	assert.Empty(t, s.Exit)

	o := list[1]
	assert.Equal(t, "js-orphan", o.ID)
	assert.True(t, o.Orphan)
	assert.True(t, o.Alive)
	assert.Nil(t, o.Ticket)
	assert.Equal(t, sleeper.Process.Pid, o.HostPid)

	// Sessions cannot list sessions.
	sess := dial(t, filepath.Join(root.SessionDir(s.ID), "cli.sock"))
	err = sess.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, nil)
	assert.ErrorIs(t, err, rpc.ErrForbidden)

	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
}

// holdLock poses as a host for dir: takes its lock, writes its pid, and
// waits to be signalled.
func holdLock(dir string) {
	f, err := os.Create(filepath.Join(dir, host.LockFile))
	if err != nil || syscall.Flock(int(f.Fd()), syscall.LOCK_EX) != nil {
		os.Exit(1)
	}
	_ = os.WriteFile(filepath.Join(dir, host.PidFile), fmt.Appendf(nil, "%d\n", os.Getpid()), 0o600)
	time.Sleep(time.Minute)
}

// TestSessionReap stops a host with no session ticket and, with Closed,
// removes a closed session's directory while leaving an open one alone.
func TestSessionReap(t *testing.T) {
	root, start := poolRoot(t, "1m")
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	// External sessions (agent not configured) get a directory but no host.
	newSession := func() store.Ticket {
		var r rpc.SessionNewResult
		require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "ext"}, &r))
		return r.Ticket
	}
	open, closed := newSession(), newSession()
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: closed.ID}, nil))
	require.DirExists(t, root.SessionDir(closed.ID))

	// Made after startup so the controller's startup reaping leaves it.
	odir := root.SessionDir("js-orphan")
	require.NoError(t, os.MkdirAll(odir, 0o700))
	holder := exec.Command(os.Args[0], "hold-lock", odir)
	require.NoError(t, holder.Start())
	exited := make(chan struct{})
	go func() { _ = holder.Wait(); close(exited) }()
	t.Cleanup(func() { _ = holder.Process.Kill(); <-exited })
	require.Eventually(t, func() bool { return host.Alive(odir) }, 5*time.Second, 10*time.Millisecond)

	params := rpc.SessionReapParams{Closed: true, OlderThan: time.Nanosecond, DryRun: true}
	var res rpc.SessionReapResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionReap, params, &res))
	assert.Equal(t, []string{"js-orphan"}, res.Stopped)
	assert.Equal(t, []string{closed.ID}, res.Removed)
	assert.True(t, host.Alive(odir), "dry run stops nothing")
	assert.DirExists(t, root.SessionDir(closed.ID), "dry run removes nothing")

	params.DryRun = false
	res = rpc.SessionReapResult{}
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionReap, params, &res))
	assert.Equal(t, []string{"js-orphan"}, res.Stopped)
	assert.Equal(t, []string{closed.ID}, res.Removed)
	select {
	case <-exited:
	case <-time.After(15 * time.Second):
		t.Fatal("orphan host not stopped")
	}
	assert.NoDirExists(t, root.SessionDir(closed.ID))
	assert.DirExists(t, root.SessionDir(open.ID), "open session untouched")

	// Sessions cannot reap.
	sess := dial(t, filepath.Join(root.SessionDir(open.ID), "cli.sock"))
	assert.ErrorIs(t, sess.Call(t.Context(), rpc.MethodSessionReap, rpc.SessionReapParams{DryRun: true}, nil), rpc.ErrForbidden)
}

// poolRoot writes a root with one fake-agent pool. start runs a controller
// and returns a func that stops it (leaving sessions running).
func poolRoot(t *testing.T, idle string) (config.Root, func() func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	root := config.Root{Dir: dir}
	cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: %s\n    dir: %q\n", os.Args[0], idle, dir)
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte(cfg), 0o600))
	start := func() func() {
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		go func() {
			done <- control.Run(ctx, root, control.Options{Tick: 50 * time.Millisecond, HostExe: os.Args[0]})
		}()
		return func() { cancel(); assert.NoError(t, <-done) }
	}
	t.Cleanup(func() {
		// Stop leftover hosts so the test leaves no processes behind.
		entries, _ := os.ReadDir(root.SessionDir(""))
		for _, e := range entries {
			if b, err := os.ReadFile(filepath.Join(root.SessionDir(e.Name()), host.PidFile)); err == nil {
				var p int
				if _, err := fmt.Sscan(string(b), &p); err == nil && p > 1 {
					_ = syscall.Kill(p, syscall.SIGTERM)
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
		_ = os.RemoveAll(dir)
	})
	return root, start
}

// TestPoolSession runs a pool end to end: ready work starts a session, the
// agent closes the ticket through its CLI socket, and the idle session ends
// with its host.
func TestPoolSession(t *testing.T) {
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
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
    idle_timeout: 1s
    dir: %q
`, os.Args[0], dir)
	cfgPath := filepath.Join(dir, config.FileName)
	require.NoError(t, os.WriteFile(cfgPath, []byte(cfg), 0o600))
	root := config.Root{Dir: dir}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		done <- control.Run(ctx, root, control.Options{Tick: 50 * time.Millisecond, HostExe: os.Args[0]})
	}()
	t.Cleanup(func() { cancel(); assert.NoError(t, <-done) })

	admin := dial(t, root.AdminSocket())
	var task store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))

	show := func(id string) store.Ticket {
		var got store.Ticket
		require.NoError(t, admin.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: id}, &struct {
			Ticket *store.Ticket `json:"ticket"`
		}{&got}))
		return got
	}
	require.Eventually(t, func() bool { return show(task.ID).Status == store.StatusClosed }, 20*time.Second, 50*time.Millisecond)
	got := show(task.ID)
	assert.Equal(t, "done", got.Resolution)

	var sessions []store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	require.Len(t, sessions, 1)
	sess := sessions[0]
	assert.Equal(t, "dev", sess.Pool)

	require.Eventually(t, func() bool { return show(sess.ID).Status == store.StatusClosed }, 20*time.Second, 50*time.Millisecond)
	assert.Equal(t, "idle", show(sess.ID).Resolution)
	assert.False(t, host.Alive(root.SessionDir(sess.ID)), "host stopped with the session")
	_, ok := host.ReadExit(root.SessionDir(sess.ID))
	assert.True(t, ok, "host wrote exit.json")
}

// fakeAgent speaks ACP on stdio and closes each prompted ticket.
func fakeAgent() {
	a := &agent{}
	conn := acp.NewAgentSideConnection(a, os.Stdout, os.Stdin)
	<-conn.Done()
}

type agent struct{ n int }

func (a *agent) Initialize(context.Context, acp.InitializeRequest) (acp.InitializeResponse, error) {
	return acp.InitializeResponse{
		ProtocolVersion:   acp.ProtocolVersionNumber,
		AgentCapabilities: acp.AgentCapabilities{LoadSession: true},
	}, nil
}

func (a *agent) NewSession(context.Context, acp.NewSessionRequest) (acp.NewSessionResponse, error) {
	a.n++
	return acp.NewSessionResponse{SessionId: acp.SessionId(fmt.Sprintf("s%d", a.n))}, nil
}

func (a *agent) LoadSession(context.Context, acp.LoadSessionRequest) (acp.LoadSessionResponse, error) {
	return acp.LoadSessionResponse{}, nil
}

func (a *agent) Prompt(ctx context.Context, _ acp.PromptRequest) (acp.PromptResponse, error) {
	// With a "block" file present, hold the turn open until "go" appears.
	if _, err := os.Stat("block"); err == nil {
		_ = os.WriteFile("started", nil, 0o600)
		for {
			if _, err := os.Stat("go"); err == nil {
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	c, err := rpc.Dial(ctx, os.Getenv("JOBSHOP_SOCKET"))
	if err != nil {
		return acp.PromptResponse{}, err
	}
	defer func() { _ = c.Close() }()
	// The current ticket is the session's claim; close it as `jobshop close` would.
	_ = c.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{}, nil)
	return acp.PromptResponse{StopReason: acp.StopReasonEndTurn}, nil
}

func (a *agent) Cancel(context.Context, acp.CancelNotification) error { return nil }

func (a *agent) Authenticate(context.Context, acp.AuthenticateRequest) (acp.AuthenticateResponse, error) {
	return acp.AuthenticateResponse{}, nil
}

func (a *agent) Logout(context.Context, acp.LogoutRequest) (acp.LogoutResponse, error) {
	return acp.LogoutResponse{}, nil
}

func (a *agent) CloseSession(context.Context, acp.CloseSessionRequest) (acp.CloseSessionResponse, error) {
	return acp.CloseSessionResponse{}, nil
}

func (a *agent) ListSessions(context.Context, acp.ListSessionsRequest) (acp.ListSessionsResponse, error) {
	return acp.ListSessionsResponse{}, nil
}

func (a *agent) ResumeSession(context.Context, acp.ResumeSessionRequest) (acp.ResumeSessionResponse, error) {
	return acp.ResumeSessionResponse{}, nil
}

func (a *agent) SetSessionConfigOption(context.Context, acp.SetSessionConfigOptionRequest) (acp.SetSessionConfigOptionResponse, error) {
	return acp.SetSessionConfigOptionResponse{}, nil
}

func (a *agent) SetSessionMode(context.Context, acp.SetSessionModeRequest) (acp.SetSessionModeResponse, error) {
	return acp.SetSessionModeResponse{}, nil
}
