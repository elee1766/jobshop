package control_test

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	host "github.com/elee1766/acphost"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/control/controltest"
	"github.com/elee1766/jobshop/pkg/hostspec"
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
	if len(os.Args) >= 2 && os.Args[1] == "fake-daemon" {
		// Record this process and the env it holds, then live on.
		line := fmt.Sprintf("%d %s\n", os.Getpid(), os.Getenv("DEPLOY"))
		if f, err := os.OpenFile("daemons", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600); err == nil {
			_, _ = f.WriteString(line)
			_ = f.Close()
		}
		awaitSignal()
		os.Exit(0)
	}
	if len(os.Args) >= 2 && os.Args[1] == "fake-agent" {
		sandboxProbe()
		isolationProbe()
		homeProbe()
		fakeAgent()
		os.Exit(0)
	}
	if len(os.Args) >= 3 && os.Args[1] == "hold-lock" {
		holdLock(os.Args[2])
		os.Exit(0)
	}
	if len(os.Args) >= 2 && filepath.Base(os.Args[0]) == "jobshop" && os.Args[1] == "version" {
		// A command step calling the controller's binary by name (#669).
		fmt.Println("jobshop test-version")
		os.Exit(0)
	}
	// Like the unit's DelegateSubgroup=controller (JOB-21): move into a
	// leaf before any test spawns hosts, or a later controller cannot
	// enable cpu/memory/pids (a cgroup with processes in it is busy).
	if info, err := host.DelegatedCgroup(); err == nil && info.Writable && filepath.Base(info.Path) != "controller" {
		_, _ = host.SetupDelegation(info.Path, []string{"cpu", "memory", "pids"})
	}
	os.Exit(m.Run())
}

// TestControllerRestart restarts the controller mid-turn: the host and
// agent survive, the new controller adopts them, and the turn's result
// still closes the ticket.
func TestControllerRestart(t *testing.T) {
	root, start := poolRootReal(t, "1m")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	stop := start()

	admin := dial(t, root.AdminSocket())
	var task store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	passEventually(t, root.Dir, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, "condition not met after supervisor passes")

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
	// The agent's JOBSHOP_SOCKET is served again by the new controller
	// (#325), so it can still comment, close or release.
	cli := dial(t, filepath.Join(sdir, "cli.sock"))
	var mine store.Ticket
	require.NoError(t, cli.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: task.ID}, &struct {
		Ticket *store.Ticket `json:"ticket"`
	}{&mine}))
	assert.Equal(t, task.ID, mine.ID)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))

	admin = dial(t, root.AdminSocket())
	show := func(id string) store.Ticket {
		var got store.Ticket
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &struct {
			Ticket *store.Ticket `json:"ticket"`
		}{&got}))
		return got
	}
	passEventually(t, root.Dir, func() bool { return show(task.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
	pid2, err := os.ReadFile(filepath.Join(sdir, host.PidFile))
	require.NoError(t, err)
	assert.Equal(t, string(pid), string(pid2), "same host adopted")
	assert.Equal(t, store.StatusOpen, show(sessions[0].ID).Status)
}

// TestEnvRefresh adopts an idle host whose spec env no longer matches what
// the controller would write: it is respawned and the session stays open.
func TestEnvRefresh(t *testing.T) {
	root, start := poolRootReal(t, "1m")
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
	passEventually(t, root.Dir, func() bool { return show(task.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	require.Len(t, sessions, 1)
	sdir := root.SessionDir(sessions[0].ID)
	pid, err := os.ReadFile(filepath.Join(sdir, host.PidFile))
	require.NoError(t, err)
	stop()

	// Pretend the host was spawned by an older deploy with another env.
	spec, err := hostspec.Read(sdir)
	require.NoError(t, err)
	spec.Env = append(spec.Env, "OLD_DEPLOY=1")
	spec.EnvHash = "old"
	require.NoError(t, hostspec.Write(sdir, spec))

	stop = start()
	defer stop()
	// The respawned agent's tool shell keeps the session env (#780).
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "shellenv"), nil, 0o600))
	admin = dial(t, root.AdminSocket())
	passEventually(t, root.Dir, func() bool {
		pid2, err := os.ReadFile(filepath.Join(sdir, host.PidFile))
		return err == nil && string(pid2) != string(pid) && host.Alive(sdir)
	}, "host respawned")
	passEventually(t, root.Dir, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "shellenv.out"))
		return err == nil
	}, "respawned agent ran its shell")
	out, err := os.ReadFile(filepath.Join(root.Dir, "shellenv.out"))
	require.NoError(t, err)
	shell := strings.TrimSpace(string(out))
	assert.True(t, strings.HasPrefix(shell, filepath.Dir(os.Args[0])+":"), "jobshop's dir first on the shell PATH: %s", shell)
	assert.True(t, strings.HasSuffix(shell, "|"+sessions[0].ID), "JOBSHOP_SESSION in the shell: %s", shell)
	assert.Contains(t, shell, os.Getenv("PATH"), "the controller PATH is kept")
	passEventually(t, root.Dir, func() bool {
		s, err := hostspec.Read(sdir)
		return err == nil && s.EnvHash != "old"
	}, "condition not met after supervisor passes")
	assert.Equal(t, store.StatusOpen, show(sessions[0].ID).Status)

	// The refreshed host resumed the conversation: a new ticket gets a turn.
	var next store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "after refresh", Pool: "dev"}, &next))
	passEventually(t, root.Dir, func() bool { return show(next.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
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
	passEventually(t, root.Dir, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, "condition not met after supervisor passes")

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

	passEventually(t, root.Dir, func() bool { return show(sessID).Status == store.StatusClosed }, "drained session closes")
	assert.Equal(t, "drained", show(sessID).Resolution)
	assert.Equal(t, store.StatusClosed, show(first.ID).Status, "running turn finished")
	assert.Equal(t, store.StatusOpen, show(second.ID).Status, "no new claim while drained")
	pass(t, root.Dir, 3)
	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	assert.Len(t, sessions, 1, "no new session while drained")

	require.NoError(t, admin.Call(t.Context(), rpc.MethodDrain, rpc.DrainParams{Undrain: true}, &dr))
	assert.Empty(t, dr.Drained)
	passEventually(t, root.Dir, func() bool { return show(second.ID).Status == store.StatusClosed }, "undrained pool runs b")
}

// TestQuarantine runs an agent that dies on every prompt: three crashed
// sessions quarantine the ticket and pause the pool.
func TestQuarantine(t *testing.T) {
	// In-process hosts: the agent still crashes as a process (#467).
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
	passEventually(t, root.Dir, func() bool { return slices.Contains(show(task.ID).Labels, "quarantined") }, "ticket quarantined")
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
	pass(t, root.Dir, 3)
	sessions = nil
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	assert.Len(t, sessions, 3, "pool paused")
}

// TestStuck runs an agent that never closes its ticket: after two unclosed
// turns the session ends as stuck instead of claiming more work.
func TestStuck(t *testing.T) {
	root, start := poolRootReal(t, "1m")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "noclose"), nil, 0o600))
	stop := start()
	defer stop()

	admin := dial(t, root.AdminSocket())
	for _, title := range []string{"a", "b", "c"} {
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: title, Pool: "dev"}, nil))
	}
	var sessions []store.Ticket
	passEventually(t, root.Dir, func() bool {
		sessions = nil
		require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession, Status: store.StatusClosed}, &sessions))
		return len(sessions) > 0
	}, "condition not met after supervisor passes")
	assert.Equal(t, "stuck", sessions[0].Resolution)
	// #301: the reminder says no wake will come, and the stuck reason
	// records how the last turn ended.
	b, _ := os.ReadFile(filepath.Join(root.Dir, "prompts"))
	assert.Contains(t, string(b), "Nothing will wake you")
	var got rpc.ShowResult
	passEventually(t, root.Dir, func() bool {
		return admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: sessions[0].ID}, &got) == nil && len(got.Comments) > 0
	}, "condition not met after supervisor passes")
	reason := got.Comments[len(got.Comments)-1].Body
	assert.Contains(t, reason, "stop_reason end_turn")
	assert.Contains(t, reason, "Waiting on the CI wake.")
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
	passEventually(t, root.Dir, func() bool { return show(first.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
	queued := send("pool:dev", "hello queued", false)
	passEventually(t, root.Dir, func() bool { return show(queued.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
	assert.Equal(t, "delivered", show(queued.ID).Resolution)
	var msg store.MessageFields
	require.NoError(t, json.Unmarshal(show(queued.ID).Fields, &msg))
	assert.Equal(t, "admin", msg.From)
	passEventually(t, root.Dir, func() bool {
		p := prompts()
		return len(p) > 0 && strings.HasPrefix(p[len(p)-1], "Messages for you:") && strings.Contains(p[len(p)-1], "hello queued")
	}, "condition not met after supervisor passes")

	// A blocked turn is cancelled by an interrupt message.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	var second store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &second))
	passEventually(t, root.Dir, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, "condition not met after supervisor passes")
	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	require.Len(t, sessions, 1)
	urgent := send(sessions[0].ID, "stop that", true)
	passEventually(t, root.Dir, func() bool { return show(urgent.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
	assert.Equal(t, store.StatusInProgress, show(second.ID).Status, "cancelled turn keeps its ticket")
	passEventually(t, root.Dir, func() bool {
		return slices.ContainsFunc(prompts(), func(p string) bool {
			return strings.HasPrefix(p, "Messages for you:") && strings.Contains(p, "stop that")
		})
	}, "condition not met after supervisor passes")

	// The reminder turn finishes the ticket once unblocked.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	passEventually(t, root.Dir, func() bool { return show(second.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
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
	pass(t, root.Dir, 1)
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
	passEventually(t, root.Dir, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, "condition not met after supervisor passes")

	var list []rpc.SessionInfo
	passEventually(t, root.Dir, func() bool {
		list = nil
		require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &list))
		return len(list) == 2 && list[0].AgentPid != 0
	}, "condition not met after supervisor passes")

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
	assert.ErrorIs(t, sess.Call(t.Context(), rpc.MethodReload, struct{}{}, nil), rpc.ErrForbidden)

	// A pool admin grant lets the session list and reload, nothing else.
	cfgPath := filepath.Join(root.Dir, config.FileName)
	b, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(cfgPath, append(b, "    admin: [session-list, config-reload]\n"...), 0o600))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReload, struct{}{}, nil))
	var granted []rpc.SessionInfo
	require.NoError(t, sess.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &granted))
	assert.Len(t, granted, 2)
	require.NoError(t, sess.Call(t.Context(), rpc.MethodReload, struct{}{}, nil))
	assert.ErrorIs(t, sess.Call(t.Context(), rpc.MethodSessionRestart, rpc.SessionRestartParams{ID: s.ID}, nil), rpc.ErrForbidden)

	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
}

// holdLock poses as a host for dir: takes its lock, writes its pid, says
// "ready" on stdout, and waits to be signalled. The test waits on that
// line, not on polling the lock, so a helper that fails says why (#480).
func holdLock(dir string) {
	f, err := os.Create(filepath.Join(dir, host.LockFile))
	if err == nil {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	}
	if err == nil {
		err = os.WriteFile(filepath.Join(dir, host.PidFile), fmt.Appendf(nil, "%d\n", os.Getpid()), 0o600)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "hold-lock:", err)
		os.Exit(1)
	}
	fmt.Println("ready")
	awaitSignal()
}

// awaitSignal blocks a helper process until it is signalled. The minute
// is a backstop so a helper whose test died does not outlive it for long;
// no test waits on it.
func awaitSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-sig:
	case <-time.After(time.Minute):
	}
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
	// The admin socket serves before that reaping; a pass is taken only
	// after it.
	pass(t, root.Dir, 1)
	odir := root.SessionDir("js-orphan")
	require.NoError(t, os.MkdirAll(odir, 0o700))
	holder := exec.Command(os.Args[0], "hold-lock", odir)
	var stderr strings.Builder
	holder.Stderr = &stderr
	// An os.Pipe, not StdoutPipe: Wait runs concurrently with the read.
	out, w, err := os.Pipe()
	require.NoError(t, err)
	defer func() { _ = out.Close() }()
	holder.Stdout = w
	require.NoError(t, holder.Start())
	_ = w.Close()
	ready := make(chan string, 1)
	go func() {
		line, _ := bufio.NewReader(out).ReadString('\n')
		ready <- line
		_, _ = io.Copy(io.Discard, out)
	}()
	exited := make(chan struct{})
	var waitErr error
	go func() { waitErr = holder.Wait(); close(exited) }()
	t.Cleanup(func() { _ = holder.Process.Kill(); <-exited })
	select {
	case line := <-ready:
		if line != "ready\n" {
			<-exited
			t.Fatalf("hold-lock helper not ready: %v, stderr: %s", waitErr, stderr.String())
		}
	case <-time.After(dialWait):
		t.Fatal("hold-lock helper never said ready")
	}
	require.True(t, host.Alive(odir), "helper holds the lock")

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
// and returns a func that stops it (leaving sessions running). Hosts run
// in-process (controltest.InProcHosts) unless the test uses poolRootReal.
func poolRoot(t *testing.T, idle string) (config.Root, func() func()) {
	return poolRootOpt(t, idle, control.Options{})
}

// realHosts in Options.Hosts asks poolRootOpt for detached host processes.
// Tests about the process tree use it: restart and adopt, crash and
// quarantine, sandbox, env refresh, stuck and busy, and any test that
// detects a respawn by a new host pid (an in-process host has the test's).
type realHosts struct{}

func (realHosts) Start(string) error { return errors.New("realHosts is a marker") }
func (realHosts) Stop(string) bool   { return false }

// poolRootReal is poolRoot with real controller -> host -> agent trees.
func poolRootReal(t *testing.T, idle string) (config.Root, func() func()) {
	return poolRootOpt(t, idle, control.Options{Hosts: realHosts{}})
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
	var inproc *controltest.InProcHosts
	switch opt.Hosts.(type) {
	case realHosts:
		opt.Hosts = nil
	case nil:
		inproc = controltest.NewInProcHosts()
		opt.Hosts = inproc
	}
	start := func() func() {
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		opt = withPass(dir, opt)
		opt.Tick = time.Hour
		if opt.HostExe == "" {
			opt.HostExe = os.Args[0]
		}
		if opt.Log == nil {
			opt.Log = testLog(t)
		}
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
		if inproc != nil {
			inproc.StopAll()
		}
		entries, _ := os.ReadDir(root.SessionDir(""))
		var pids []int
		for _, e := range entries {
			if b, err := os.ReadFile(filepath.Join(root.SessionDir(e.Name()), host.PidFile)); err == nil {
				var p int
				// An in-process host's pid is this process.
				if _, err := fmt.Sscan(string(b), &p); err == nil && p > 1 && p != selfPid {
					if syscall.Kill(p, syscall.SIGTERM) == nil {
						pids = append(pids, p)
					}
				}
			}
		}
		// Wait for the hosts to go before removing their dirs.
		assert.Eventually(t, func() bool {
			for _, p := range pids {
				b, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", p))
				if err == nil && !strings.Contains(string(b), ") Z ") {
					return false
				}
			}
			return true
		}, dialWait, 20*time.Millisecond)
		_ = os.RemoveAll(dir)
	})
	return root, start
}

// TestWorktreeDeleted removes a live session's git worktree, as happened on
// gasserver (2026-11-01): the session respawns its host with a new worktree
// and keeps working instead of failing every tool call.
func TestWorktreeDeleted(t *testing.T) {
	root, start := poolRootReal(t, "1m")
	repo := filepath.Join(root.Dir, "repo")
	for _, args := range [][]string{
		{"init", "-q", "-b", "master", repo},
		{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", "init"},
	} {
		out, err := exec.Command("git", args...).CombinedOutput()
		require.NoError(t, err, string(out))
	}
	cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: 1m\n    repo: %q\n    ref: master\n", os.Args[0], repo)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
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
	var first store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Pool: "dev"}, &first))
	passEventually(t, root.Dir, func() bool { return show(first.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")

	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	require.Len(t, sessions, 1)
	sdir := root.SessionDir(sessions[0].ID)
	work := filepath.Join(sdir, "work")
	oldPid, err := os.ReadFile(filepath.Join(sdir, host.PidFile))
	require.NoError(t, err)

	// #154: an outside `git worktree remove --force` (gh pr merge deleting
	// the branch's worktree) must not remove a live session's worktree.
	out, err := exec.Command("git", "-C", repo, "worktree", "remove", "--force", work).CombinedOutput()
	require.Error(t, err, "locked worktree removed: %s", out)
	assert.Contains(t, string(out), "locked")
	_, err = os.Stat(work)
	require.NoError(t, err, "worktree still there")

	require.NoError(t, os.RemoveAll(work))

	passEventually(t, root.Dir, func() bool {
		pid, err := os.ReadFile(filepath.Join(sdir, host.PidFile))
		_, werr := os.Stat(work)
		return err == nil && string(pid) != string(oldPid) && werr == nil
	}, "host respawned with a new worktree")

	var second store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &second))
	passEventually(t, root.Dir, func() bool { return show(second.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
	sessions = nil
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	assert.Len(t, sessions, 1, "same session, not a new one")
	assert.Equal(t, store.StatusOpen, sessions[0].Status)
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
	passEventually(t, root.Dir, func() bool { return show(task.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
	got := show(task.ID)
	assert.Equal(t, "done", got.Resolution)

	var sessions []store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	// Under load the first session can go idle and a new one spawn before
	// this list runs (#455), so check the oldest: it ran the task.
	require.NotEmpty(t, sessions)
	sess := sessions[0]
	for _, s := range sessions[1:] {
		if s.CreatedAt.Before(sess.CreatedAt) {
			sess = s
		}
	}
	assert.Equal(t, "dev", sess.Pool)

	passEventually(t, root.Dir, func() bool { return show(sess.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
	assert.Equal(t, "idle", show(sess.ID).Resolution)
	assert.False(t, host.Alive(root.SessionDir(sess.ID)), "host stopped with the session")
	_, ok := host.ReadExit(root.SessionDir(sess.ID))
	assert.True(t, ok, "host wrote exit.json")
}

// fakeAgent speaks ACP on stdio and closes each prompted ticket.
func fakeAgent() {
	a := &agent{
		known: map[acp.SessionId]bool{}, cancel: make(chan struct{}, 1),
		cur: map[string]string{"model": "default-model", "reasoning_effort": "medium"},
	}
	conn := acp.NewAgentSideConnection(a, os.Stdout, os.Stdin)
	a.conn.Store(conn)
	<-conn.Done()
}

type agent struct {
	n      int
	conn   atomic.Pointer[acp.AgentSideConnection] // set after the connection starts serving
	known  map[acp.SessionId]bool                  // created or loaded in this process
	cancel chan struct{}
	cur    map[string]string // config option values (#304)
	busy   atomic.Bool       // a prompt is running
	mu     sync.Mutex        // guards cur
}

func (a *agent) Initialize(context.Context, acp.InitializeRequest) (acp.InitializeResponse, error) {
	// A "shellenv" file asks the agent to record what its tool shell
	// sees (#780): a `bash -c` child, as agents run tools.
	if _, err := os.Stat("shellenv"); err == nil {
		out, err := exec.Command("bash", "-c", `echo "$PATH|$JOBSHOP_SESSION"`).CombinedOutput()
		if err != nil {
			out = append(out, []byte(" error: "+err.Error())...)
		}
		_ = os.WriteFile("shellenv.out", out, 0o600)
	}
	// A "daemon" file makes the agent start a detached helper that keeps
	// the agent's env, like jcode's `serve` daemon behind `acp --socket`.
	if _, err := os.Stat("daemon"); err == nil {
		cmd := exec.Command(os.Args[0], "fake-daemon")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		_ = cmd.Start()
	}
	// A "noload" file makes the agent one that cannot load sessions.
	_, err := os.Stat("noload")
	return acp.InitializeResponse{
		ProtocolVersion:   acp.ProtocolVersionNumber,
		AgentCapabilities: acp.AgentCapabilities{LoadSession: err != nil},
	}, nil
}

func (a *agent) NewSession(context.Context, acp.NewSessionRequest) (acp.NewSessionResponse, error) {
	a.n++
	id := acp.SessionId(fmt.Sprintf("s%d", a.n))
	a.known[id] = true
	return acp.NewSessionResponse{SessionId: id, ConfigOptions: a.configOptions()}, nil
}

// configOptions advertises model and reasoning_effort when a "configopts"
// file exists, like jcode (#304).
func (a *agent) configOptions() []acp.SessionConfigOption {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.configOptionsLocked()
}

func (a *agent) configOptionsLocked() []acp.SessionConfigOption {
	if _, err := os.Stat("configopts"); err != nil {
		return nil
	}
	var out []acp.SessionConfigOption
	for _, id := range []string{"model", "reasoning_effort"} {
		vals := acp.SessionConfigSelectOptionsUngrouped{{Name: a.cur[id], Value: acp.SessionConfigValueId(a.cur[id])}}
		out = append(out, acp.SessionConfigOption{Select: &acp.SessionConfigOptionSelect{
			Id: acp.SessionConfigId(id), Name: id, Type: "select", CurrentValue: acp.SessionConfigValueId(a.cur[id]),
			Options: acp.SessionConfigSelectOptions{Ungrouped: &vals},
		}})
	}
	return out
}

func (a *agent) LoadSession(_ context.Context, p acp.LoadSessionRequest) (acp.LoadSessionResponse, error) {
	// A "loadfail" file makes session/load fail, like a lost transcript.
	if _, err := os.Stat("loadfail"); err == nil {
		return acp.LoadSessionResponse{}, &acp.RequestError{Code: -32603, Message: "session not found"}
	}
	a.known[p.SessionId] = true
	return acp.LoadSessionResponse{ConfigOptions: a.configOptions()}, nil
}

func (a *agent) Prompt(ctx context.Context, p acp.PromptRequest) (acp.PromptResponse, error) {
	// Like jcode: a new process knows only sessions it created or loaded.
	if !a.known[p.SessionId] {
		return acp.PromptResponse{}, &acp.RequestError{Code: -32602, Message: "Unknown ACP session id: " + string(p.SessionId)}
	}
	a.busy.Store(true)
	res, err := a.prompt(ctx, p)
	a.busy.Store(false)
	return res, err
}

func (a *agent) prompt(ctx context.Context, p acp.PromptRequest) (acp.PromptResponse, error) {
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
	if strings.Contains(text, "Messages for you:") {
		// Reply with a tool call and an echo, for session.subscribe. A
		// message saying "hold" keeps the turn open until cancelled.
		_ = a.conn.Load().SessionUpdate(ctx, acp.SessionNotification{SessionId: p.SessionId, Update: acp.StartToolCall("t1", "Reading", acp.WithStartStatus(acp.ToolCallStatusInProgress))})
		_ = a.conn.Load().SessionUpdate(ctx, acp.SessionNotification{SessionId: p.SessionId, Update: acp.UpdateToolCall("t1", acp.WithUpdateStatus(acp.ToolCallStatusCompleted))})
		_ = a.conn.Load().SessionUpdate(ctx, acp.SessionNotification{SessionId: p.SessionId, Update: acp.UpdateAgentMessageText("echo: " + text[strings.LastIndex(text, "\n")+1:])})
		if strings.Contains(text, "hold") {
			_ = os.WriteFile("held", nil, 0o600)
			<-a.cancel
			return acp.PromptResponse{StopReason: acp.StopReasonCancelled}, nil
		}
		return acp.PromptResponse{StopReason: acp.StopReasonEndTurn}, nil
	}
	// With a "crash" file present, die mid-turn like a crashing agent.
	if _, err := os.Stat("crash"); err == nil {
		os.Exit(3)
	}
	// A "busy" file holds a count of prompts to refuse as still processing,
	// like jcode after a prompt fails mid-stream.
	if b, err := os.ReadFile("busy"); err == nil {
		var n int
		if _, err := fmt.Sscan(string(b), &n); err == nil && n > 0 {
			_ = os.WriteFile("busy", fmt.Appendf(nil, "%d", n-1), 0o600)
			return acp.PromptResponse{}, &acp.RequestError{Code: -32000, Message: "Already processing a message"}
		}
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
	// A "noclose" file makes every turn end without touching the ticket,
	// like an agent whose tools all fail.
	if _, err := os.Stat("noclose"); err == nil {
		_ = a.conn.Load().SessionUpdate(ctx, acp.SessionNotification{SessionId: p.SessionId, Update: acp.UpdateAgentMessageText("Waiting on the CI wake.")})
		return acp.PromptResponse{StopReason: acp.StopReasonEndTurn}, nil
	}
	c, err := rpc.Dial(ctx, os.Getenv("JOBSHOP_SOCKET"))
	if err != nil {
		return acp.PromptResponse{}, err
	}
	defer func() { _ = c.Close() }()
	// The current ticket is the session's claim; close it as `jobshop close` would.
	ctx = telemetry.WithTraceParent(ctx, os.Getenv("JOBSHOP_TRACEPARENT"))
	// An "outcome" file holds the outcome word to close with (#385).
	var outcome string
	if b, err := os.ReadFile("outcome"); err == nil {
		outcome = strings.TrimSpace(string(b))
	}
	_ = c.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{Outcome: outcome}, nil)
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

func (a *agent) SetSessionConfigOption(_ context.Context, p acp.SetSessionConfigOptionRequest) (acp.SetSessionConfigOptionResponse, error) {
	v := p.ValueId
	if v == nil {
		return acp.SetSessionConfigOptionResponse{}, &acp.RequestError{Code: -32602, Message: "want a value id"}
	}
	// Record each call and whether a prompt was running, one line each.
	line := fmt.Sprintf("%s=%s busy=%v\n", v.ConfigId, v.Value, a.busy.Load())
	if f, err := os.OpenFile("configsets", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600); err == nil {
		_, _ = f.WriteString(line)
		_ = f.Close()
	}
	if a.busy.Load() {
		return acp.SetSessionConfigOptionResponse{}, &acp.RequestError{Code: -32000, Message: "busy; retry when it finishes"}
	}
	// A value listed in "configreject" is refused, like an unknown route.
	if b, err := os.ReadFile("configreject"); err == nil && strings.Contains(string(b), string(v.Value)) {
		return acp.SetSessionConfigOptionResponse{}, &acp.RequestError{Code: -32602, Message: "unknown model route " + string(v.Value)}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cur[string(v.ConfigId)] = string(v.Value)
	return acp.SetSessionConfigOptionResponse{ConfigOptions: a.configOptionsLocked()}, nil
}

func (a *agent) SetSessionMode(context.Context, acp.SetSessionModeRequest) (acp.SetSessionModeResponse, error) {
	return acp.SetSessionModeResponse{}, nil
}

// TestAgentBusy has the agent refuse prompts as "Already processing a
// message", as jcode did on gasserver (2026-11-01, two dev sessions ended
// crashed). The session cancels, then respawns its host, and the ticket
// still closes without the session ending crashed.
func TestAgentBusy(t *testing.T) {
	root, start := poolRootReal(t, "1m")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "busy"), []byte("5"), 0o600))
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
	passEventually(t, root.Dir, func() bool { return show(task.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
	assert.Equal(t, "done", show(task.ID).Resolution)

	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	require.Len(t, sessions, 1)
	assert.Equal(t, store.StatusOpen, sessions[0].Status, "session did not end crashed")
}

// backoffLog counts the controller's rate-limit backoff log records. The
// backoff lasts 100-400ms, so polling session.list for BackoffUntil missed
// it under -race load (#372); a log record cannot be missed.
type backoffLog struct {
	slog.Handler
	n *atomic.Int32
}

func (h backoffLog) Handle(ctx context.Context, r slog.Record) error {
	if r.Message == "provider rate limited, backing off" {
		h.n.Add(1)
	}
	return h.Handler.Handle(ctx, r)
}

func (h backoffLog) WithAttrs(a []slog.Attr) slog.Handler {
	return backoffLog{h.Handler.WithAttrs(a), h.n}
}

func (h backoffLog) WithGroup(g string) slog.Handler {
	return backoffLog{h.Handler.WithGroup(g), h.n}
}

// TestRateLimitBackoff has the agent fail prompts with a 429 a few times:
// the session keeps its claim and host, backs off without restarting, and
// the retried turn closes the ticket.
func TestRateLimitBackoff(t *testing.T) {
	backoffs := new(atomic.Int32)
	root, start := poolRootOpt(t, "1m", control.Options{Hosts: realHosts{}, Log: slog.New(backoffLog{testLog(t).Handler(), backoffs})})
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

	hostPid := 0
	passEventually(t, root.Dir, func() bool {
		var list []rpc.SessionInfo
		require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &list))
		for _, s := range list {
			// The retry may close the ticket between reading backoff and
			// reading the claim; otherwise the claim must still be held.
			if s.BackoffUntil != nil && s.Current != task.ID {
				assert.Equal(t, store.StatusClosed, show(task.ID).Status, "claim kept while backing off")
			}
			if s.HostPid == 0 {
				continue
			}
			if hostPid == 0 {
				hostPid = s.HostPid
			}
			assert.Equal(t, hostPid, s.HostPid, "no restart")
		}
		return show(task.ID).Status == store.StatusClosed
	}, "condition not met after supervisor passes")
	assert.Equal(t, int32(3), backoffs.Load(), "one backoff per 429")
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
	addr := make(chan string, 1)
	root, start := poolRootOpt(t, "1m", control.Options{MetricsReady: func(a string) { addr <- a }})
	f, err := os.OpenFile(filepath.Join(root.Dir, config.FileName), os.O_APPEND|os.O_WRONLY, 0)
	require.NoError(t, err)
	_, err = f.WriteString("telemetry:\n  metrics_listen: 127.0.0.1:0\n")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	ctx := t.Context()
	defer start()()
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
	// The fake agent closes its ticket inside the turn, so closes_total can
	// appear a scrape before turns_total (#73). Wait for every series.
	wants := []string{
		`jobshop_closes_total{jobshop_kind="task",jobshop_pool="dev",jobshop_resolution="done"`,
		`jobshop_claims_total{jobshop_pool="dev"`,
		`jobshop_sessions{jobshop_pool="dev",jobshop_session_state="idle"`,
		`jobshop_tickets_ready{jobshop_pool="dev",jobshop_project="js"`,
		`jobshop_rpc_duration_seconds_bucket{`,
		`rpc_method="ticket.create"`,
		`jobshop_turns_total{jobshop_pool="dev",jobshop_turn_stop_reason="end_turn"`,
	}
	var body string
	passEventually(t, root.Dir, func() bool {
		body = scrape()
		for _, want := range wants {
			if !strings.Contains(body, want) {
				return false
			}
		}
		return true
	}, "metrics series")
	for _, want := range wants {
		assert.Contains(t, body, want)
	}
}

// TestTraces checks that the agent's close request is a server span in the
// session's trace, parented by the traceparent the agent got in its env.
func TestTraces(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	root, start := poolRootOpt(t, "1m", control.Options{Tracer: tp.Tracer("test")})
	ctx := t.Context()
	defer start()()

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
	passEventually(t, root.Dir, func() bool {
		s, ok := find("rpc " + rpc.MethodTicketClose)
		closeSpan = s
		return ok
	}, "close span")

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
	root, start := poolRootReal(t, "1m")
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
	passEventually(t, root.Dir, func() bool { return show(task.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
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

	// Change the agent's env in the config and reload. The idle runner
	// re-checks the env hash on every poll (#106) and refreshes the host
	// by itself, so wait for that before restarting: asserting the old
	// spec right after the reload raced the refresh (#327).
	oldSpec, err := hostspec.Read(sdir)
	require.NoError(t, err)
	require.NotContains(t, oldSpec.Env, "RESTART_MARK=1")
	cfgPath := filepath.Join(root.Dir, config.FileName)
	b, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	b = []byte(strings.Replace(string(b), "    args: [fake-agent]\n", "    args: [fake-agent]\n    env: {RESTART_MARK: \"1\"}\n", 1))
	require.NoError(t, os.WriteFile(cfgPath, b, 0o600))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReload, struct{}{}, nil))
	passEventually(t, root.Dir, func() bool {
		s, err := hostspec.Read(sdir)
		p := hostPidOf(t, sdir)
		return err == nil && slices.Contains(s.Env, "RESTART_MARK=1") && p != 0 && p != pid && host.Alive(sdir)
	}, "reload refreshed the host")
	refreshed, err := hostspec.Read(sdir)
	require.NoError(t, err)
	assert.NotEqual(t, oldSpec.EnvHash, refreshed.EnvHash, "fresh spec")
	pid = hostPidOf(t, sdir)

	var res rpc.SessionRestartResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionRestart, rpc.SessionRestartParams{ID: id}, &res))
	assert.Equal(t, pid, res.HostPid)
	passEventually(t, root.Dir, func() bool {
		p := hostPidOf(t, sdir)
		return p != 0 && p != pid && host.Alive(sdir)
	}, "host respawned")
	s, err := hostspec.Read(sdir)
	require.NoError(t, err)
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
	passEventually(t, root.Dir, func() bool { return show(next.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")

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
	root, start := poolRootReal(t, "1m")
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
		passEventually(t, root.Dir, func() bool {
			_, err := os.Stat(started)
			return err == nil
		}, "condition not met after supervisor passes")
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
	pass(t, root.Dir, 3)
	assert.Equal(t, pid, hostPidOf(t, sdir), "busy host not restarted")
	assert.Equal(t, store.StatusInProgress, show(first.ID).Status)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	passEventually(t, root.Dir, func() bool { return show(first.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
	passEventually(t, root.Dir, func() bool {
		p := hostPidOf(t, sdir)
		return p != 0 && p != pid && host.Alive(sdir)
	}, "restarted after the turn")
	pid = hostPidOf(t, sdir)

	// --now cancels a running turn; the ticket stays claimed.
	require.NoError(t, os.Remove(filepath.Join(root.Dir, "go")))
	require.NoError(t, os.Remove(started))
	var second store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &second))
	waitStarted()
	require.NoError(t, os.Remove(started))
	require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionRestart, rpc.SessionRestartParams{ID: id, Now: true}, nil))
	passEventually(t, root.Dir, func() bool {
		p := hostPidOf(t, sdir)
		return p != 0 && p != pid && host.Alive(sdir)
	}, "restarted at once")
	got := show(second.ID)
	assert.Equal(t, store.StatusInProgress, got.Status, "claim kept")
	assert.Equal(t, id, got.Holder)
	waitStarted() // the new host is prompted with the held ticket
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	passEventually(t, root.Dir, func() bool { return show(second.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
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

// TestLoadFallback respawns a session whose agent cannot reload the old
// conversation, either because it lacks LoadSession or because
// session/load fails: the session starts a new conversation and the next
// prompt opens with a summary of the journal, without crash-looping.
func TestLoadFallback(t *testing.T) {
	// nojournal: session/load fails and the journal file is gone too
	// (JOB-14); the summary comes from the recall index.
	for _, name := range []string{"noload", "loadfail", "nojournal"} {
		t.Run(name, func(t *testing.T) {
			marker := name
			if name == "nojournal" {
				marker = "loadfail"
			}
			root, start := poolRootReal(t, "1m")
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
			var first store.Ticket
			require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "first-ticket-title", Pool: "dev"}, &first))
			passEventually(t, root.Dir, func() bool { return show(first.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")
			var sessions []store.Ticket
			require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
			require.Len(t, sessions, 1)
			id := sessions[0].ID
			sdir := root.SessionDir(id)
			pid := hostPidOf(t, sdir)

			require.NoError(t, os.WriteFile(filepath.Join(root.Dir, marker), nil, 0o600))
			if name == "nojournal" {
				require.NoError(t, os.Remove(filepath.Join(sdir, host.JournalFile)))
			}
			require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionRestart, rpc.SessionRestartParams{ID: id}, nil))
			passEventually(t, root.Dir, func() bool {
				p := hostPidOf(t, sdir)
				return p != 0 && p != pid && host.Alive(sdir)
			}, "condition not met after supervisor passes")

			var second store.Ticket
			require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "second-ticket-title", Pool: "dev"}, &second))
			passEventually(t, root.Dir, func() bool { return show(second.ID).Status == store.StatusClosed }, "condition not met after supervisor passes")

			b, err := os.ReadFile(filepath.Join(root.Dir, "prompts"))
			require.NoError(t, err)
			var prompts []string
			for line := range strings.Lines(string(b)) {
				var s string
				require.NoError(t, json.Unmarshal([]byte(line), &s))
				prompts = append(prompts, s)
			}
			idx := slices.IndexFunc(prompts, func(p string) bool { return strings.Contains(p, "second-ticket-title") })
			require.GreaterOrEqual(t, idx, 0)
			p := prompts[idx]
			assert.True(t, strings.HasPrefix(p, "Summary of the earlier conversation"), "summary first: %q", p)
			before, after, ok := strings.Cut(p, host.SummaryEnd)
			require.True(t, ok)
			assert.Contains(t, before, "first-ticket-title", "summary covers the old turn")
			assert.Contains(t, after, "You have claimed Jobshop ticket "+second.ID)
			const note = "this session was resumed"
			assert.Contains(t, after, note, "resumed session gets the recovery note")
			assert.NotContains(t, prompts[0], note, "a new session gets no recovery note")
			assert.Equal(t, 1, strings.Count(string(b), note), "note once per respawn")
			assert.NotContains(t, prompts[idx-1], host.SummaryEnd, "no extra turn for the summary")

			var infos []rpc.SessionInfo
			require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &infos))
			require.Len(t, infos, 1)
			assert.Zero(t, infos[0].Crashes, "not crash-looping")
			assert.Equal(t, store.StatusOpen, show(id).Status)
		})
	}
}

// TestCancelDuringStartup cancels after the admin socket is up but before
// open sessions are listed: Run must return nil and clean up (#61).
func TestCancelDuringStartup(t *testing.T) {
	root, _ := poolRoot(t, "1m")
	f, err := os.OpenFile(filepath.Join(root.Dir, config.FileName), os.O_APPEND|os.O_WRONLY, 0)
	require.NoError(t, err)
	_, err = f.WriteString("telemetry:\n  metrics_listen: 127.0.0.1:0\n")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	ctx, cancel := context.WithCancel(t.Context())
	err = control.Run(ctx, root, control.Options{
		Tick: 50 * time.Millisecond, HostExe: os.Args[0],
		MetricsReady: func(string) { cancel() },
	})
	require.NoError(t, err)
	assert.NoFileExists(t, root.AdminSocket())
}

// TestSessionsEndedMetric ends a session crashed (agent dies) and one stuck
// (agent never closes) and checks jobshop_sessions_ended_total labels (#71).
func TestSessionsEndedMetric(t *testing.T) {
	for _, tc := range []struct{ marker, resolution string }{
		{"crash", "crashed"},
		{"noclose", "stuck"},
	} {
		t.Run(tc.resolution, func(t *testing.T) {
			addr := make(chan string, 1)
			// In-process hosts: the agent still exits as a process (#467).
			root, start := poolRootOpt(t, "1m", control.Options{MetricsReady: func(a string) { addr <- a }})
			f, err := os.OpenFile(filepath.Join(root.Dir, config.FileName), os.O_APPEND|os.O_WRONLY, 0)
			require.NoError(t, err)
			_, err = f.WriteString("telemetry:\n  metrics_listen: 127.0.0.1:0\n")
			require.NoError(t, err)
			require.NoError(t, f.Close())
			require.NoError(t, os.WriteFile(filepath.Join(root.Dir, tc.marker), nil, 0o600))
			ctx := t.Context()
			defer start()()
			var url string
			select {
			case a := <-addr:
				url = "http://" + a + "/metrics"
			case <-time.After(10 * time.Second):
				t.Fatal("metrics not served")
			}

			admin := dial(t, root.AdminSocket())
			for _, title := range []string{"a", "b", "c"} {
				require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: title, Pool: "dev"}, nil))
			}
			want := `jobshop_sessions_ended_total{jobshop_pool="dev",jobshop_resolution="` + tc.resolution + `"`
			passEventually(t, root.Dir, func() bool {
				res, err := http.Get(url) //nolint:noctx // test
				if err != nil {
					return false
				}
				defer func() { _ = res.Body.Close() }()
				b, _ := io.ReadAll(res.Body)
				return strings.Contains(string(b), want)
			}, want)
		})
	}
}

// TestRoutedBackPrompt moves a ticket refinery -> dev with a finding in a
// comment, as the refinery does, and checks the dev session sees it.
func TestRoutedBackPrompt(t *testing.T) {
	root, start := poolRoot(t, "1m")
	stop := start()
	defer stop()
	ctx := t.Context()
	admin := dial(t, root.AdminSocket())
	var tk store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "routed-back", Pool: "refinery"}, &tk))
	finding := "Refinery: not merging. Rebase on master and keep section 6 last."
	require.NoError(t, admin.Call(ctx, rpc.MethodCommentAdd, rpc.CommentParams{ID: tk.ID, Body: finding}, nil))
	dev := "dev"
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketUpdate, rpc.UpdateParams{ID: tk.ID, Pool: &dev}, nil))

	var p string
	passEventually(t, root.Dir, func() bool {
		b, _ := os.ReadFile(filepath.Join(root.Dir, "prompts"))
		for line := range strings.Lines(string(b)) {
			var s string
			if json.Unmarshal([]byte(line), &s) == nil && strings.Contains(s, "routed-back") {
				p = s
				return true
			}
		}
		return false
	}, "condition not met after supervisor passes")
	assert.Contains(t, p, "# Recent comments")
	assert.Contains(t, p, "admin:\n"+finding)
	assert.Contains(t, p, "hand the ticket back to refinery")
}

// TestScaleDownCordon scales dev to 0 during a blocked turn: the session is
// cordoned, finishes its ticket, claims nothing new and ends scaled_down.
// The target survives a controller restart, and auto takes the work.
func TestScaleDownCordon(t *testing.T) {
	root, start := poolRoot(t, "1m")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "block"), nil, 0o600))
	stop := start()

	admin := dial(t, root.AdminSocket())
	show := func(c *rpc.Client, id string) store.Ticket {
		var got store.Ticket
		require.NoError(t, c.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: id}, &struct {
			Ticket *store.Ticket `json:"ticket"`
		}{&got}))
		return got
	}
	var first, second store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Pool: "dev"}, &first))
	passEventually(t, root.Dir, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "started"))
		return err == nil
	}, "condition not met after supervisor passes")

	zero := 0
	require.NoError(t, admin.Call(t.Context(), rpc.MethodPoolScale, rpc.PoolScaleParams{Pool: "dev", Target: &zero}, nil))
	var infos []rpc.SessionInfo
	passEventually(t, root.Dir, func() bool {
		infos = nil
		require.NoError(t, admin.Call(t.Context(), rpc.MethodSessionList, rpc.SessionListParams{}, &infos))
		return len(infos) == 1 && infos[0].Cordon == store.CordonScale
	}, "session cordoned")
	sessID := infos[0].ID
	assert.Equal(t, store.StatusInProgress, show(admin, first.ID).Status, "working session is not stopped")

	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "b", Pool: "dev"}, &second))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "go"), nil, 0o600))
	passEventually(t, root.Dir, func() bool { return show(admin, sessID).Status == store.StatusClosed }, "scaled session closes")
	assert.Equal(t, "scaled_down", show(admin, sessID).Resolution)
	assert.Equal(t, store.StatusClosed, show(admin, first.ID).Status, "running turn finished")
	assert.Equal(t, store.StatusOpen, show(admin, second.ID).Status, "no new claim while scaled to 0")

	stop()
	stop = start()
	defer func() { stop() }()
	admin = dial(t, root.AdminSocket())
	var ps rpc.PoolStatusResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodPoolList, nil, &ps))
	require.Len(t, ps.Pools, 1)
	require.NotNil(t, ps.Pools[0].Target, "target survives restart")
	assert.Equal(t, 0, *ps.Pools[0].Target)
	pass(t, root.Dir, 3)
	assert.Equal(t, store.StatusOpen, show(admin, second.ID).Status)

	require.NoError(t, admin.Call(t.Context(), rpc.MethodPoolScale, rpc.PoolScaleParams{Pool: "dev", Auto: true}, nil))
	passEventually(t, root.Dir, func() bool { return show(admin, second.ID).Status == store.StatusClosed }, "auto scale runs b")
}
