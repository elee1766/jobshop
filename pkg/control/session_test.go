package control_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/host"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
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

	// The refreshed host resumed the conversation: a new ticket gets a turn.
	var next store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "after refresh", Pool: "dev"}, &next))
	require.Eventually(t, func() bool { return show(next.ID).Status == store.StatusClosed }, 20*time.Second, 50*time.Millisecond)
	assert.Equal(t, store.StatusOpen, show(sessions[0].ID).Status, "same session, not crashed")
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

// TestQuarantine runs an agent that dies on every prompt: three crashed
// sessions quarantine the ticket and pause the pool.
func TestQuarantine(t *testing.T) {
	root, start := poolRootOpt(t, "1m", control.Options{RestartDelay: 10 * time.Millisecond})
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "crash"), nil, 0o600))
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
	var task store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	require.Eventually(t, func() bool { return slices.Contains(show(task.ID).Labels, "quarantined") }, 60*time.Second, 100*time.Millisecond)
	got := show(task.ID)
	assert.Equal(t, store.StatusOpen, got.Status)
	assert.True(t, got.DeferUntil.After(time.Now().Add(50*time.Minute)), "deferred about an hour")

	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	require.Len(t, sessions, 3)
	for _, s := range sessions {
		assert.Equal(t, "crashed", s.Resolution)
	}

	// The pool is paused: ready work starts no session.
	var other store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "u", Pool: "dev"}, &other))
	time.Sleep(500 * time.Millisecond)
	sessions = nil
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	assert.Len(t, sessions, 3, "pool paused")
}

// TestMessages delivers a queued message to an idle pool session, then an
// interrupt message that cancels a blocked turn and is delivered before the
// ticket is finished.
func TestMessages(t *testing.T) {
	root, start := poolRootOpt(t, "1m", control.Options{MessagePoll: 50 * time.Millisecond})
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
	prompts := func() []string {
		b, _ := os.ReadFile(filepath.Join(root.Dir, "prompts"))
		var out []string
		for line := range strings.Lines(string(b)) {
			var s string
			if json.Unmarshal([]byte(line), &s) == nil {
				out = append(out, s)
			}
		}
		return out
	}
	send := func(to, text string, interrupt bool) store.Ticket {
		f, err := json.Marshal(store.MessageFields{To: to, Interrupt: interrupt})
		require.NoError(t, err)
		var m store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Kind: store.KindMessage, Title: text, Fields: f}, &m))
		return m
	}

	// Start a session with a first ticket, then queue a message to the pool.
	var first store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Pool: "dev"}, &first))
	require.Eventually(t, func() bool { return show(first.ID).Status == store.StatusClosed }, 20*time.Second, 50*time.Millisecond)
	queued := send("pool:dev", "hello queued", false)
	require.Eventually(t, func() bool { return show(queued.ID).Status == store.StatusClosed }, 20*time.Second, 50*time.Millisecond)
	assert.Equal(t, "delivered", show(queued.ID).Resolution)
	var msg store.MessageFields
	require.NoError(t, json.Unmarshal(show(queued.ID).Fields, &msg))
	assert.Equal(t, "admin", msg.From)
	require.Eventually(t, func() bool {
		p := prompts()
		return len(p) > 0 && strings.HasPrefix(p[len(p)-1], "Messages for you:") && strings.Contains(p[len(p)-1], "hello queued")
	}, 5*time.Second, 50*time.Millisecond)

	// A blocked turn is cancelled by an interrupt message.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	var second store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &second))
	require.Eventually(t, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, 20*time.Second, 20*time.Millisecond)
	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	require.Len(t, sessions, 1)
	urgent := send(sessions[0].ID, "stop that", true)
	require.Eventually(t, func() bool { return show(urgent.ID).Status == store.StatusClosed }, 20*time.Second, 50*time.Millisecond)
	assert.Equal(t, store.StatusInProgress, show(second.ID).Status, "cancelled turn keeps its ticket")
	require.Eventually(t, func() bool {
		return slices.ContainsFunc(prompts(), func(p string) bool {
			return strings.HasPrefix(p, "Messages for you:") && strings.Contains(p, "stop that")
		})
	}, 5*time.Second, 50*time.Millisecond)

	// The reminder turn finishes the ticket once unblocked.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
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
	return poolRootOpt(t, idle, control.Options{})
}

func poolRootOpt(t *testing.T, idle string, opt control.Options) (config.Root, func() func()) {
	// Each root has its own dir, sockets and hosts; the race build spends
	// ~1s per host spawn, so serial runs are mostly waiting.
	t.Parallel()
	t.Helper()
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	root := config.Root{Dir: dir}
	cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: %s\n    dir: %q\n", os.Args[0], idle, dir)
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte(cfg), 0o600))
	start := func() func() {
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		opt.Tick, opt.HostExe = 50*time.Millisecond, os.Args[0]
		if opt.Backoff == 0 {
			opt.Backoff = 100 * time.Millisecond
		}
		if opt.Poll == 0 {
			opt.Poll = 50 * time.Millisecond
		}
		if opt.RestartDelay == 0 {
			opt.RestartDelay = 50 * time.Millisecond
		}
		if opt.MessagePoll == 0 {
			opt.MessagePoll = 50 * time.Millisecond
		}
		go func() {
			done <- control.Run(ctx, root, opt)
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
	root, start := poolRoot(t, "300ms")
	stop := start()
	defer stop()
	ctx := t.Context()
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
	a := &agent{known: map[acp.SessionId]bool{}, cancel: make(chan struct{}, 1)}
	conn := acp.NewAgentSideConnection(a, os.Stdout, os.Stdin)
	<-conn.Done()
}

type agent struct {
	n      int
	known  map[acp.SessionId]bool // created or loaded in this process
	cancel chan struct{}
}

func (a *agent) Initialize(context.Context, acp.InitializeRequest) (acp.InitializeResponse, error) {
	return acp.InitializeResponse{
		ProtocolVersion:   acp.ProtocolVersionNumber,
		AgentCapabilities: acp.AgentCapabilities{LoadSession: true},
	}, nil
}

func (a *agent) NewSession(context.Context, acp.NewSessionRequest) (acp.NewSessionResponse, error) {
	a.n++
	id := acp.SessionId(fmt.Sprintf("s%d", a.n))
	a.known[id] = true
	return acp.NewSessionResponse{SessionId: id}, nil
}

func (a *agent) LoadSession(_ context.Context, p acp.LoadSessionRequest) (acp.LoadSessionResponse, error) {
	a.known[p.SessionId] = true
	return acp.LoadSessionResponse{}, nil
}

func (a *agent) Prompt(ctx context.Context, p acp.PromptRequest) (acp.PromptResponse, error) {
	// Like jcode: a new process knows only sessions it created or loaded.
	if !a.known[p.SessionId] {
		return acp.PromptResponse{}, &acp.RequestError{Code: -32602, Message: "Unknown ACP session id: " + string(p.SessionId)}
	}
	// Record every prompt (one JSON string per line) for tests to inspect.
	var text string
	for _, b := range p.Prompt {
		if b.Text != nil {
			text += b.Text.Text
		}
	}
	if f, err := os.OpenFile("prompts", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600); err == nil {
		line, _ := json.Marshal(text)
		_, _ = f.Write(append(line, '\n'))
		_ = f.Close()
	}
	if strings.HasPrefix(text, "Messages for you:") {
		return acp.PromptResponse{StopReason: acp.StopReasonEndTurn}, nil
	}
	// With a "crash" file present, die mid-turn like a crashing agent.
	if _, err := os.Stat("crash"); err == nil {
		os.Exit(3)
	}
	// A "ratelimit" file holding N fails the next N prompts the way a
	// provider account limit does.
	if b, err := os.ReadFile("ratelimit"); err == nil {
		var n int
		if _, err := fmt.Sscan(string(b), &n); err == nil && n > 0 {
			_ = os.WriteFile("ratelimit", fmt.Appendf(nil, "%d", n-1), 0o600)
			return acp.PromptResponse{}, &acp.RequestError{Code: -32000, Message: "API Error: 429 rate_limit_error"}
		}
	}
	// With a "block" file present, hold the turn open until "go" appears.
	if _, err := os.Stat("block"); err == nil {
		_ = os.WriteFile("started", nil, 0o600)
		for {
			if _, err := os.Stat("go"); err == nil {
				break
			}
			select {
			case <-a.cancel:
				// Cancelled turns leave the ticket held, like a real agent.
				return acp.PromptResponse{StopReason: acp.StopReasonCancelled}, nil
			case <-time.After(20 * time.Millisecond):
			}
		}
	}
	c, err := rpc.Dial(ctx, os.Getenv("JOBSHOP_SOCKET"))
	if err != nil {
		return acp.PromptResponse{}, err
	}
	defer func() { _ = c.Close() }()
	// The current ticket is the session's claim; close it as `jobshop close` would.
	ctx = telemetry.WithTraceParent(ctx, os.Getenv("JOBSHOP_TRACEPARENT"))
	_ = c.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{}, nil)
	return acp.PromptResponse{StopReason: acp.StopReasonEndTurn}, nil
}

func (a *agent) Cancel(context.Context, acp.CancelNotification) error {
	select {
	case a.cancel <- struct{}{}:
	default:
	}
	return nil
}

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

// TestRateLimitBackoff has the agent fail prompts with a 429 a few times:
// the session keeps its claim and host, backs off without restarting, and
// the retried turn closes the ticket.
func TestRateLimitBackoff(t *testing.T) {
	root, start := poolRoot(t, "1m")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "ratelimit"), []byte("3"), 0o600))
	stop := start()
	defer stop()
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

	sawBackoff := false
	hostPid := 0
	require.Eventually(t, func() bool {
		var list []rpc.SessionInfo
		require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &list))
		for _, s := range list {
			if s.BackoffUntil == nil {
				continue
			}
			sawBackoff = true
			// The retry may close the ticket between reading backoff and
			// reading the claim; otherwise the claim must still be held.
			if s.Current != task.ID {
				assert.Equal(t, store.StatusClosed, show(task.ID).Status, "claim kept while backing off")
			}
			if hostPid == 0 {
				hostPid = s.HostPid
			}
			assert.Equal(t, hostPid, s.HostPid, "no restart")
		}
		return show(task.ID).Status == store.StatusClosed
	}, 30*time.Second, 20*time.Millisecond)
	assert.True(t, sawBackoff)
	assert.Equal(t, "done", show(task.ID).Resolution)

	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	require.Len(t, sessions, 1, "no extra sessions started")
	assert.Equal(t, store.StatusOpen, sessions[0].Status)
	pid, err := os.ReadFile(filepath.Join(root.SessionDir(sessions[0].ID), host.PidFile))
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprint(hostPid), strings.TrimSpace(string(pid)), "same host")
	b, err := os.ReadFile(filepath.Join(root.Dir, "ratelimit"))
	require.NoError(t, err)
	assert.Equal(t, "0", string(b))
}

// TestMetrics serves /metrics from the controller and checks that a pool
// run produces the starting series.
func TestMetrics(t *testing.T) {
	root, _ := poolRoot(t, "1m")
	f, err := os.OpenFile(filepath.Join(root.Dir, config.FileName), os.O_APPEND|os.O_WRONLY, 0)
	require.NoError(t, err)
	_, err = f.WriteString("telemetry:\n  metrics_listen: 127.0.0.1:0\n")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	addr := make(chan string, 1)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		done <- control.Run(ctx, root, control.Options{
			Tick: 50 * time.Millisecond, HostExe: os.Args[0],
			MetricsReady: func(a string) { addr <- a },
		})
	}()
	defer func() { cancel(); assert.NoError(t, <-done) }()
	var url string
	select {
	case a := <-addr:
		url = "http://" + a + "/metrics"
	case <-time.After(10 * time.Second):
		t.Fatal("metrics not served")
	}

	admin := dial(t, root.AdminSocket())
	var task store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))

	scrape := func() string {
		res, err := http.Get(url) //nolint:noctx // test
		require.NoError(t, err)
		defer func() { _ = res.Body.Close() }()
		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		return string(b)
	}
	var body string
	require.Eventually(t, func() bool {
		body = scrape()
		return strings.Contains(body, `jobshop_closes_total{jobshop_kind="task",jobshop_pool="dev",jobshop_resolution="done"`)
	}, 20*time.Second, 100*time.Millisecond, "closes series")
	for _, want := range []string{
		`jobshop_claims_total{jobshop_pool="dev"`,
		`jobshop_sessions{jobshop_pool="dev",jobshop_session_state="idle"`,
		`jobshop_tickets_ready{jobshop_pool="dev",jobshop_project="js"`,
		`jobshop_rpc_duration_seconds_bucket{`,
		`rpc_method="ticket.create"`,
	} {
		assert.Contains(t, body, want)
	}
}

// TestTraces checks that the agent's close request is a server span in the
// session's trace, parented by the traceparent the agent got in its env.
func TestTraces(t *testing.T) {
	root, _ := poolRoot(t, "1m")
	exp := tracetest.NewInMemoryExporter()
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		done <- control.Run(ctx, root, control.Options{
			Tick: 50 * time.Millisecond, HostExe: os.Args[0],
			SpanProcessor: sdktrace.NewSimpleSpanProcessor(exp),
		})
	}()
	defer func() { cancel(); assert.NoError(t, <-done) }()

	admin := dial(t, root.AdminSocket())
	var task store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))

	find := func(name string) (tracetest.SpanStub, bool) {
		for _, s := range exp.GetSpans() {
			if s.Name == name {
				return s, true
			}
		}
		return tracetest.SpanStub{}, false
	}
	var closeSpan tracetest.SpanStub
	require.Eventually(t, func() bool {
		s, ok := find("rpc " + rpc.MethodTicketClose)
		closeSpan = s
		return ok
	}, 20*time.Second, 50*time.Millisecond)

	root2, ok := find("session")
	require.True(t, ok, "session root span")
	assert.Equal(t, root2.SpanContext.TraceID(), closeSpan.SpanContext.TraceID(), "close joins the session trace")
	assert.Equal(t, root2.SpanContext.SpanID(), closeSpan.Parent.SpanID(), "parent is the session span")
	assert.True(t, closeSpan.Parent.IsRemote())
	attrs := map[attribute.Key]string{}
	for _, kv := range closeSpan.Attributes {
		attrs[kv.Key] = kv.Value.String()
	}
	assert.Equal(t, rpc.MethodTicketClose, attrs[telemetry.AttrMethod])
	assert.Equal(t, "dev", attrs[telemetry.AttrPool])
	assert.NotEmpty(t, attrs[telemetry.AttrSession])

	spawn, ok := find("session spawn")
	require.True(t, ok, "spawn lifecycle span")
	assert.Equal(t, root2.SpanContext.TraceID(), spawn.SpanContext.TraceID())
	_, ok = find("supervisor tick")
	assert.True(t, ok, "tick span")
}

// TestSessionRestartIdle restarts an idle session: the host is replaced
// with a freshly built spec, a ticket it holds stays claimed, and the crash
// counter does not move.
func TestSessionRestartIdle(t *testing.T) {
	root, start := poolRoot(t, "1m")
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
	var task store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	require.Eventually(t, func() bool { return show(task.ID).Status == store.StatusClosed }, 20*time.Second, 50*time.Millisecond)
	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	require.Len(t, sessions, 1)
	id := sessions[0].ID
	sdir := root.SessionDir(id)
	pid := hostPidOf(t, sdir)

	// A ticket the session holds but will not be prompted with (deferred).
	// Defer before claiming: the idle runner polls ClaimNext, which resumes
	// any undeferred ticket the session holds and the agent would close it.
	// Claim by id ignores defer_until.
	var held store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "held", Pool: "other", Project: "js"}, &held))
	later := time.Now().Add(time.Hour)
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketUpdate, rpc.UpdateParams{ID: held.ID, DeferUntil: &later}, nil))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodClaim, rpc.ClaimParams{ID: held.ID, Holder: id}, nil))

	// Change the agent's env in the config and reload: the running host
	// keeps its old spec until the restart.
	cfgPath := filepath.Join(root.Dir, config.FileName)
	b, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	b = []byte(strings.Replace(string(b), "    args: [fake-agent]\n", "    args: [fake-agent]\n    env: {RESTART_MARK: \"1\"}\n", 1))
	require.NoError(t, os.WriteFile(cfgPath, b, 0o600))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReload, struct{}{}, nil))
	spec, err := host.ReadSpec(sdir)
	require.NoError(t, err)
	require.NotContains(t, spec.Env, "RESTART_MARK=1")
	oldHash := spec.EnvHash

	var res rpc.SessionRestartResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionRestart, rpc.SessionRestartParams{ID: id}, &res))
	assert.Equal(t, pid, res.HostPid)
	require.Eventually(t, func() bool {
		p := hostPidOf(t, sdir)
		return p != 0 && p != pid && host.Alive(sdir)
	}, 20*time.Second, 50*time.Millisecond, "host respawned")
	s, err := host.ReadSpec(sdir)
	require.NoError(t, err)
	assert.NotEqual(t, oldHash, s.EnvHash, "fresh spec")
	assert.Contains(t, s.Env, "RESTART_MARK=1", "spec reflects the reloaded config")

	got := show(held.ID)
	assert.Equal(t, store.StatusInProgress, got.Status, "claim kept")
	assert.Equal(t, id, got.Holder)
	assert.Equal(t, store.StatusOpen, show(id).Status)
	var infos []rpc.SessionInfo
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &infos))
	require.Len(t, infos, 1)
	assert.Zero(t, infos[0].Crashes, "not a crash")

	// The resumed conversation takes new work. Release the held ticket
	// first: the fake agent closes "the current ticket", which needs one.
	require.NoError(t, admin.Call(t.Context(), rpc.MethodRelease, rpc.ReleaseParams{ID: held.ID, Reason: "test"}, nil))
	var next store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "after", Pool: "dev"}, &next))
	require.Eventually(t, func() bool { return show(next.ID).Status == store.StatusClosed }, 20*time.Second, 50*time.Millisecond)

	// Unknown, closed and non-session ids are errors.
	assert.ErrorIs(t, admin.Call(t.Context(), rpc.MethodSessionRestart, rpc.SessionRestartParams{ID: "js-nope"}, nil), store.ErrNotFound)
	assert.Error(t, admin.Call(t.Context(), rpc.MethodSessionRestart, rpc.SessionRestartParams{ID: next.ID}, nil))
	var ext rpc.SessionNewResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "ext"}, &ext))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketClose, rpc.CloseParams{ID: ext.Ticket.ID}, nil))
	assert.Error(t, admin.Call(t.Context(), rpc.MethodSessionRestart, rpc.SessionRestartParams{ID: ext.Ticket.ID}, nil))
	sess := dial(t, filepath.Join(sdir, "cli.sock"))
	assert.ErrorIs(t, sess.Call(t.Context(), rpc.MethodSessionRestart, rpc.SessionRestartParams{ID: id}, nil), rpc.ErrForbidden)
}

// TestSessionRestartBusy: a restart waits for the running turn; --now
// cancels it and the restarted host gets the still-claimed ticket again.
func TestSessionRestartBusy(t *testing.T) {
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
	started := filepath.Join(root.Dir, "started")
	waitStarted := func() {
		require.Eventually(t, func() bool {
			_, err := os.Stat(started)
			return err == nil
		}, 20*time.Second, 20*time.Millisecond)
	}
	var first store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Pool: "dev"}, &first))
	waitStarted()
	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	require.Len(t, sessions, 1)
	id := sessions[0].ID
	sdir := root.SessionDir(id)
	pid := hostPidOf(t, sdir)

	// Without --now the turn keeps running.
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionRestart, rpc.SessionRestartParams{ID: id}, nil))
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, pid, hostPidOf(t, sdir), "busy host not restarted")
	assert.Equal(t, store.StatusInProgress, show(first.ID).Status)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	require.Eventually(t, func() bool { return show(first.ID).Status == store.StatusClosed }, 20*time.Second, 50*time.Millisecond)
	require.Eventually(t, func() bool {
		p := hostPidOf(t, sdir)
		return p != 0 && p != pid && host.Alive(sdir)
	}, 20*time.Second, 50*time.Millisecond, "restarted after the turn")
	pid = hostPidOf(t, sdir)

	// --now cancels a running turn; the ticket stays claimed.
	require.NoError(t, os.Remove(filepath.Join(root.Dir, "go")))
	require.NoError(t, os.Remove(started))
	var second store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &second))
	waitStarted()
	require.NoError(t, os.Remove(started))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionRestart, rpc.SessionRestartParams{ID: id, Now: true}, nil))
	require.Eventually(t, func() bool {
		p := hostPidOf(t, sdir)
		return p != 0 && p != pid && host.Alive(sdir)
	}, 20*time.Second, 50*time.Millisecond, "restarted at once")
	got := show(second.ID)
	assert.Equal(t, store.StatusInProgress, got.Status, "claim kept")
	assert.Equal(t, id, got.Holder)
	waitStarted() // the new host is prompted with the held ticket
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	require.Eventually(t, func() bool { return show(second.ID).Status == store.StatusClosed }, 20*time.Second, 50*time.Millisecond)
	assert.Equal(t, store.StatusOpen, show(id).Status)
	var infos []rpc.SessionInfo
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &infos))
	require.Len(t, infos, 1)
	assert.Zero(t, infos[0].Crashes)
}

func hostPidOf(t *testing.T, dir string) int {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, host.PidFile))
	if err != nil {
		return 0
	}
	var p int
	_, _ = fmt.Sscan(string(b), &p)
	return p
}
