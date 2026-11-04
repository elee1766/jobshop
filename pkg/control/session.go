package control

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	acp "github.com/coder/acp-go-sdk"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	host "github.com/elee1766/acphost"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/hostspec"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

const (
	defaultIdle       = 5 * time.Minute
	defaultChatIdle   = time.Hour
	defaultTurn       = 2 * time.Hour
	hostLease         = 30 * time.Minute
	crashWindow       = 10 * time.Minute
	maxCrashes        = 3
	unclosedDeferral  = 10 * time.Minute
	resolutionIdle    = "idle"
	resolutionCrashed = "crashed"
	// resolutionStuck ends a session whose turns keep leaving tickets unclosed
	// (e.g. a host whose tools all fail), so a fresh host replaces it.
	resolutionStuck = "stuck"
	maxUnclosed     = 2
)

// runner drives one session: keeps its host alive, speaks ACP to the agent
// through it, and feeds it tickets.
type runner struct {
	c    *Controller
	id   string
	pool string // "" for manual sessions
	// unclosed counts consecutive turns that ended still holding a ticket.
	unclosed int
	dir      string
	log      *slog.Logger
	// traceParent is the session trace (SessionFields.TraceParent).
	traceParent string
	// agent and projectName are the session's, for prompt rule matching.
	agent, projectName string
	// systemDue is set when a host starts (new or respawned): its next
	// turn carries the system prompt rules. Guarded by mu.
	systemDue bool
	// turnRules are the prompt rules in the current prompt, sent in its
	// _meta so the host journal records which applied (#205).
	turnRules []string

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	mu       sync.Mutex
	turnEnd  chan struct{} // signalled by _jobshop/turn_end
	crashes  []time.Time
	stopping bool
	agentPid int // from the host's welcome; 0 until connected

	// lastStop and lastSaid are the stop reason and agent text of the
	// last turn, recorded when a session ends stuck (#301).
	lastStop string
	lastSaid strings.Builder
	// backoffUntil is set while the agent's provider is rate limited and the
	// runner waits to retry the same turn.
	backoffUntil time.Time
	// restartPending is set by session.restart: the host is stopped and
	// respawned at the next point where no turn runs. restartWake wakes an
	// idle wait; restartNow cancels the running turn (--now).
	restartPending bool
	restartWake    chan struct{}
	restartNow     chan struct{}
	// seed is a journal summary for a conversation that could not be
	// reloaded; the next turn's prompt starts with it. Memory only.
	seed string
	// resumed is set when a new host takes over an existing conversation
	// (restart, crash, respawn, self-deploy); the next prompt starts with
	// recoveryNote. Memory only.
	resumed bool
	// turnTicket is the ticket the current prompt works on, sent in the
	// prompt's _meta so conversation export can tag records with it.
	turnTicket string
	// delivered is when deliver last ran a message turn; it resets the
	// idle timer, so a chat session lives while its thread is active.
	// Written under mu (session.prompt reads it); the runner goroutine
	// reads it without.
	delivered time.Time
	// lastClaim is when the runner last claimed a ticket; the scaler
	// cordons the least recently used first (#281). Guarded by mu.
	lastClaim time.Time
	// skipped are config option ids already logged as not advertised;
	// model and effort the agent's current values (#304). Guarded by mu.
	skipped       map[string]bool
	model, effort string
}

func (r *runner) lastClaimAt() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lastClaim
}

func (r *runner) setTurnTicket(id string) {
	r.mu.Lock()
	r.turnTicket = id
	r.mu.Unlock()
}

// promptMeta is the session/prompt _meta: the ticket and the prompt rules
// in the prompt, if any.
func (r *runner) promptMeta() map[string]any {
	r.mu.Lock()
	defer r.mu.Unlock()
	m := map[string]any{}
	if r.turnTicket != "" {
		m[telemetry.MetaTicket] = r.turnTicket
	}
	if len(r.turnRules) > 0 {
		m[metaRules] = r.turnRules
	}
	if r.model != "" {
		m[telemetry.MetaModel] = r.model
	}
	if r.effort != "" {
		m[telemetry.MetaEffort] = r.effort
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// metaRules is the session/prompt _meta key listing the prompt rules the
// prompt carries; the host journals it with the prompt.
const metaRules = "jobshop/rules"

func (r *runner) setTurnRules(names []string) {
	r.mu.Lock()
	r.turnRules = names
	r.mu.Unlock()
}

// newConversation starts a new ACP session in place of one the agent could
// not reload, and queues a summary of the old one from the host journal.
func (r *runner) newConversation(ctx context.Context, conn *acp.ClientSideConnection, cwd string, fields *store.SessionFields) error {
	ns, err := conn.NewSession(ctx, acp.NewSessionRequest{Cwd: cwd, McpServers: []acp.McpServer{}})
	if err != nil {
		return fmt.Errorf("session/new: %w", err)
	}
	frames, _, err := host.ReadJournal(filepath.Join(r.dir, host.JournalFile), 0)
	if err != nil {
		r.log.Warn("read journal for summary", "err", err)
	}
	summary := host.Summarize(frames, host.DefaultSummaryBytes)
	r.mu.Lock()
	r.seed = summary
	r.mu.Unlock()
	r.log.Warn("started a new conversation", "old", fields.ACPSession, "new", ns.SessionId,
		"path", "new+summary", "summary_bytes", len(summary))
	fields.ACPSession = string(ns.SessionId)
	noteAdvertised(fields, ns.ConfigOptions)
	return r.saveFields(*fields)
}

// recoveryNote opens the first prompt after a host respawn. The agent may
// have pushed, commented or closed before it stopped, and repeating that
// (a second PR, a duplicate comment) is worse than checking first.
const recoveryNote = "Note: this session was resumed after a restart, crash, respawn or deploy. " +
	"Before repeating any outside action (push, PR, issue or ticket comment, close), " +
	"check your worktree (`git status`, `git log`), the remote branch and GitHub, " +
	"and `jobshop ticket show` for what already happened."

func (r *runner) takeResumed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	v := r.resumed
	r.resumed = false
	return v
}

func (r *runner) takeSeed() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.seed
	r.seed = ""
	return s
}

// rateLimited reports whether a prompt error is the provider refusing
// work for now (account limit) rather than a broken agent. Agents report
// it as a generic ACP error, so match the message.
func rateLimited(err error) bool {
	var re *acp.RequestError
	if !errors.As(err, &re) {
		return false
	}
	s := strings.ToLower(fmt.Sprint(re.Message, " ", re.Data))
	for _, k := range []string{"429", "rate_limit", "rate limit", "is unavailable"} {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

// agentBusy reports whether the agent refused a prompt because it still
// considers an earlier one in flight. jcode does this after a prompt fails
// mid-stream (seen 2026-11-01); a session/cancel clears it, a host restart
// does not, so restarting only burns the crash budget.
func agentBusy(err error) bool {
	var re *acp.RequestError
	if !errors.As(err, &re) {
		return false
	}
	return strings.Contains(strings.ToLower(fmt.Sprint(re.Message, " ", re.Data)), "already processing")
}

func (r *runner) setBackoff(until time.Time) {
	r.mu.Lock()
	r.backoffUntil = until
	r.mu.Unlock()
}

func (r *runner) backoff() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.backoffUntil
}

func (c *Controller) startRunner(t store.Ticket) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.runners[t.ID]; ok {
		return
	}
	ctx, cancel := context.WithCancel(c.ctx)
	r := &runner{
		c: c, id: t.ID, pool: t.Pool, dir: c.root.SessionDir(t.ID),
		log: c.log.With("session", t.ID, "pool", t.Pool), ctx: ctx, cancel: cancel,
		done: make(chan struct{}), turnEnd: make(chan struct{}, 1),
		restartWake: make(chan struct{}, 1), restartNow: make(chan struct{}, 1),
	}
	if f, err := store.DecodeFields[store.SessionFields](t); err == nil {
		r.traceParent = f.TraceParent
		r.agent = f.Agent
	}
	r.projectName = t.Project
	c.runners[t.ID] = r
	c.wg.Go(func() {
		defer close(r.done)
		r.run()
	})
}

// liveRunners returns the ids of sessions with a running driver.
func (c *Controller) liveRunners() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	ids := make([]string, 0, len(c.runners))
	for id := range c.runners {
		ids = append(ids, id)
	}
	return ids
}

// stopRunner ends a session's driver, agent and host, and waits for them.
func (c *Controller) stopRunner(id string) {
	c.mu.Lock()
	r := c.runners[id]
	c.mu.Unlock()
	if r != nil {
		r.mu.Lock()
		r.stopping = true
		r.mu.Unlock()
		r.cancel()
		<-r.done
	}
	stopHost(c.root.SessionDir(id))
}

func (r *runner) run() {
	defer func() {
		r.c.mu.Lock()
		delete(r.c.runners, r.id)
		r.c.mu.Unlock()
	}()
	for r.ctx.Err() == nil {
		err := r.session()
		if r.ctx.Err() != nil {
			return // controller restart or stop: the host keeps running
		}
		r.mu.Lock()
		stopping := r.stopping
		r.mu.Unlock()
		if stopping {
			return
		}
		if errors.Is(err, errRefresh) {
			continue
		}
		var end errSessionEnd
		if errors.As(err, &end) {
			if end.permanent {
				// Before finish, which releases the ticket: no tick may start
				// another session into the same error.
				r.c.quarantinePool(context.WithoutCancel(r.ctx), r.pool, end.reason)
			}
			r.finish(end.resolution, end.reason)
			return
		}
		r.log.Warn("session interrupted, restarting", "err", err)
		r.lifecycle("restart", attribute.String("error", err.Error()))
		r.c.metrics.Restarts.Add(r.ctx, 1, metric.WithAttributes(telemetry.AttrPool.String(r.pool)))
		if r.crashed() {
			r.finish(resolutionCrashed, fmt.Sprintf("%d restarts in %s: %v", maxCrashes, crashWindow, err))
			return
		}
		select {
		case <-r.ctx.Done():
			return
		case <-time.After(r.c.opt.RestartDelay):
		}
	}
}

func (r *runner) crashed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	keep := r.crashes[:0]
	for _, t := range r.crashes {
		if now.Sub(t) < crashWindow {
			keep = append(keep, t)
		}
	}
	keep = append(keep, now)
	r.crashes = keep
	return len(r.crashes) > maxCrashes
}

// finish closes the session ticket, which releases its claims and socket.
func (r *runner) finish(resolution, reason string) {
	r.log.Info("session ended", "resolution", resolution, "reason", reason)
	r.lifecycle("end", telemetry.AttrResolution.String(resolution))
	stopHost(r.dir)
	ctx := context.WithoutCancel(r.ctx)
	if resolution == resolutionCrashed || resolution == resolutionConfigFailed {
		// Pause before closing: a closed session frees a pool slot, and a
		// supervisor tick before the pause would start another (#237).
		// A rejected model repeats on every new session the same way.
		r.c.poolCrashed(r.pool)
	}
	r.c.metrics.SessionsEnded.Add(ctx, 1, metric.WithAttributes(telemetry.AttrPool.String(r.pool),
		telemetry.AttrResolution.String(resolution)))
	if _, err := r.c.store.CloseTicket(ctx, "controller", r.id, nil, resolution); err != nil && !errors.Is(err, store.ErrClosed) {
		r.log.Error("close session ticket", "err", err)
	} else if err == nil {
		r.c.metrics.Closes.Add(ctx, 1, metric.WithAttributes(telemetry.AttrPool.String(r.pool),
			telemetry.AttrKind.String(string(store.KindSession)), telemetry.AttrResolution.String(resolution)))
	}
	if reason != "" {
		_, _ = r.c.store.AddComment(ctx, "controller", r.id, reason)
	}
	if resolution == resolutionCrashed {
		r.c.crashRelease(ctx, r.id)
	}
	r.c.endSession(ctx, r.id, "session "+resolution)
}

// stopHost asks a host to stop (it forwards SIGTERM to the agent's group)
// and kills it if it is still alive after 10 seconds.
func stopHost(dir string) {
	pid := hostPid(dir)
	if pid <= 1 || !host.Alive(dir) {
		return
	}
	_ = syscall.Kill(pid, syscall.SIGTERM)
	deadline := time.Now().Add(10 * time.Second)
	for host.Alive(dir) && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if host.Alive(dir) {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}
	stopSessionProcs(dir)
}

// stopSessionProcs stops processes that still carry the session's env
// after its host is gone. Agents may start a detached daemon (jcode
// `acp --socket` runs tools in a `jcode serve` outside the host's process
// group); it keeps the env it started with, so without this a refreshed
// session still runs tools with the old PATH (#111).
func stopSessionProcs(dir string) {
	if dir == "" {
		return
	}
	pids := sessionProcs(dir)
	for _, p := range pids {
		_ = syscall.Kill(p, syscall.SIGTERM)
	}
	deadline := time.Now().Add(5 * time.Second)
	for len(pids) > 0 && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
		pids = slices.DeleteFunc(pids, func(p int) bool { return syscall.Kill(p, 0) != nil })
	}
	for _, p := range pids {
		_ = syscall.Kill(p, syscall.SIGKILL)
	}
}

// sessionProcs lists processes whose environ has the JOBSHOP_SOCKET the
// controller gives agents of the session in dir (unique per root, unlike
// the session id). Linux only; elsewhere it finds none.
func sessionProcs(dir string) []int {
	want := []byte("JOBSHOP_SOCKET=" + filepath.Join(dir, "cli.sock"))
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}
	self := os.Getpid()
	var out []int
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil || pid <= 1 || pid == self {
			continue
		}
		env, err := os.ReadFile(filepath.Join("/proc", e.Name(), "environ"))
		if err != nil {
			continue
		}
		for _, kv := range bytes.Split(env, []byte{0}) {
			if bytes.Equal(kv, want) {
				out = append(out, pid)
				break
			}
		}
	}
	return out
}

func hostPid(dir string) int {
	b, err := os.ReadFile(filepath.Join(dir, host.PidFile))
	if err != nil {
		return 0
	}
	var pid int
	if _, err := fmt.Sscan(string(b), &pid); err != nil {
		return 0
	}
	return pid
}

type errSessionEnd struct {
	resolution, reason string
	// permanent marks a spawn error that will repeat (host.Permanent).
	permanent bool
}

func (e errSessionEnd) Error() string { return "session ended: " + e.resolution }

// session runs one host connection until it drops or the session ends.
func (r *runner) session() error {
	t, err := r.c.store.Get(r.ctx, r.id)
	if err != nil {
		return err
	}
	if t.Status == store.StatusClosed {
		return errSessionEnd{resolution: t.Resolution}
	}
	fields, err := store.DecodeFields[store.SessionFields](t)
	if err != nil {
		return err
	}
	fresh := false
	if !host.Alive(r.dir) {
		r.clearRestart()
		if err := r.c.spawnHost(t, fields); err != nil {
			return errSessionEnd{resolution: resolutionCrashed, reason: "spawn: " + err.Error(), permanent: host.Permanent(err)}
		}
		r.lifecycle("spawn")
		fresh = true
		r.mu.Lock()
		r.systemDue = true
		r.mu.Unlock()
	} else {
		r.lifecycle("adopt")
	}
	nc, rd, welcome, err := r.connect()
	if err != nil {
		return err
	}
	defer func() { _ = nc.Close() }()
	r.mu.Lock()
	r.agentPid = welcome.AgentPid
	r.mu.Unlock()
	// A host whose agent was never initialized has not loaded the
	// conversation either, even when an earlier attempt spawned it.
	if welcome.Initialize == nil {
		fresh = true
	}

	gate := rpc.NewGate(rd)
	conn := acp.NewClientSideConnection(&client{r: r}, nc, gate)
	conn.SetLogger(rpc.QuietLogger(r.log))
	gate.Open()
	r.log.Info("connected to host", "host_pid", welcome.Pid, "agent_pid", welcome.AgentPid, "fresh", fresh, "busy", welcome.Busy)
	stale := !fresh && r.staleEnv(t, fields)
	if r.wantsRestart() && !welcome.Busy {
		return r.restart(nc)
	}
	if stale && !welcome.Busy && r.idle() {
		return r.refresh(nc)
	}

	ctx, cancel := context.WithCancel(r.ctx)
	defer cancel()
	// Stopping the controller closes the connection rather than cancelling
	// requests: a cancelled request makes the SDK send session/cancel, which
	// would end a turn that should keep running in the host.
	stopClose := context.AfterFunc(ctx, func() { _ = nc.Close() })
	defer stopClose()
	go func() {
		select {
		case <-conn.Done():
			cancel()
		case <-ctx.Done():
		}
	}()

	var init acp.InitializeResponse
	if welcome.Initialize != nil {
		if err := json.Unmarshal(welcome.Initialize, &init); err != nil {
			return err
		}
	} else {
		init, err = conn.Initialize(ctx, acp.InitializeRequest{
			ProtocolVersion: acp.ProtocolVersionNumber,
			ClientInfo:      &acp.Implementation{Name: "jobshop", Version: "1"},
		})
		if err != nil {
			return fmt.Errorf("initialize: %w", err)
		}
	}

	cwd := r.c.sessionCwd(t)
	switch {
	case fields.ACPSession == "":
		ns, err := conn.NewSession(ctx, acp.NewSessionRequest{Cwd: cwd, McpServers: []acp.McpServer{}})
		if err != nil {
			return fmt.Errorf("session/new: %w", err)
		}
		fields.ACPSession = string(ns.SessionId)
		noteAdvertised(&fields, ns.ConfigOptions)
		if err := r.saveFields(fields); err != nil {
			return err
		}
	case fresh:
		r.mu.Lock()
		r.resumed = true
		r.mu.Unlock()
		loaded := false
		if init.AgentCapabilities.LoadSession {
			ls, err := conn.LoadSession(ctx, acp.LoadSessionRequest{
				SessionId: acp.SessionId(fields.ACPSession), Cwd: cwd, McpServers: []acp.McpServer{},
			})
			switch {
			case err == nil:
				loaded = true
				if err := r.c.updateSessionFields(ctx, r.id, func(f *store.SessionFields) { noteAdvertised(f, ls.ConfigOptions) }); err != nil {
					return err
				}
				r.log.Info("resumed conversation", "acp_session", fields.ACPSession, "path", "load")
			case ctx.Err() != nil:
				return r.dropped(conn)
			default:
				// A failed load would fail again on every respawn and end
				// the session as crashed; a new conversation keeps going.
				r.log.Warn("session/load failed; falling back to a new conversation", "acp_session", fields.ACPSession, "err", err)
			}
		}
		if !loaded {
			if err := r.newConversation(ctx, conn, cwd, &fields); err != nil {
				return err
			}
		}
	}

	if welcome.Busy {
		// A turn started by a previous controller is still running.
		select {
		case <-r.turnEnd:
		case <-ctx.Done():
			return r.dropped(conn)
		case <-r.restartNow:
			return r.restart(nc)
		}
		if r.wantsRestart() {
			return r.restart(nc)
		}
		r.afterTurn(ctx, conn, fields.ACPSession, "")
		if stale && r.idle() {
			return r.refresh(nc)
		}
	} else {
		select {
		case <-r.turnEnd: // a turn that ended while no controller was attached
		default:
		}
	}
	return r.work(ctx, conn, nc, fields.ACPSession, stale)
}

// errRefresh: the host was stopped so the next session() spawns it with the
// current env. run() restarts at once and does not count it as a crash.
var errRefresh = errors.New("host env refresh")

// staleEnv reports whether the adopted host's spec env differs from what a
// new host would get, e.g. after a deploy changed PATH or credentials.
func (r *runner) staleEnv(t store.Ticket, f store.SessionFields) bool {
	spec, err := hostspec.Read(r.dir)
	if err != nil {
		return false
	}
	want, _, err := r.c.hostSpec(t, f)
	if err != nil {
		return false
	}
	return spec.EnvHash != want.EnvHash
}

// currentEnvStale is staleEnv for the session as stored now.
func (r *runner) currentEnvStale() bool {
	t, err := r.c.store.Get(r.ctx, r.id)
	if err != nil {
		return false
	}
	f, err := store.DecodeFields[store.SessionFields](t)
	if err != nil {
		return false
	}
	return r.staleEnv(t, f)
}

// idle: the session holds no in_progress ticket.
func (r *runner) idle() bool {
	held, err := r.c.store.List(r.ctx, store.Filter{Holder: r.id, Status: store.StatusInProgress})
	return err == nil && len(held) == 0
}

// requestRestart asks for a host restart. Without now it waits for the
// running turn, if any, to end.
func (r *runner) requestRestart(now bool) {
	r.mu.Lock()
	r.restartPending = true
	r.mu.Unlock()
	select {
	case r.restartWake <- struct{}{}:
	default:
	}
	if now {
		select {
		case r.restartNow <- struct{}{}:
		default:
		}
	}
}

func (r *runner) wantsRestart() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.restartPending
}

// clearRestart drops a pending request once a new host is about to spawn.
func (r *runner) clearRestart() {
	r.mu.Lock()
	r.restartPending = false
	r.mu.Unlock()
	for _, ch := range []chan struct{}{r.restartWake, r.restartNow} {
		select {
		case <-ch:
		default:
		}
	}
}

// restart stops the host on request. Like refresh it returns errRefresh,
// so run() respawns at once with a fresh spec and counts no crash. It skips
// afterTurn: held tickets stay claimed (the supervisor keeps heartbeating
// them) and the new host is prompted with them again.
func (r *runner) restart(nc net.Conn) error {
	r.log.Info("restarting host on request", "host_pid", hostPid(r.dir))
	r.lifecycle("restart_requested")
	_ = nc.Close()
	stopHost(r.dir)
	return errRefresh
}

func (r *runner) crashCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, t := range r.crashes {
		if time.Since(t) < crashWindow {
			n++
		}
	}
	return n
}

// worktreeMissing reports whether the session's git worktree was deleted
// under a live host (only pools with a repo get one).
func (r *runner) worktreeMissing() bool {
	t, err := r.c.store.Get(r.ctx, r.id)
	if err != nil {
		return false
	}
	if p, ok := r.c.config().Pools[t.Pool]; !ok || p.Repo == "" {
		return false
	}
	_, err = os.Stat(r.c.sessionCwd(t))
	return errors.Is(err, fs.ErrNotExist)
}

func (r *runner) refresh(nc net.Conn) error {
	r.log.Info("refreshing host env", "host_pid", hostPid(r.dir))
	_ = nc.Close()
	stopHost(r.dir)
	return errRefresh
}

func (r *runner) dropped(conn *acp.ClientSideConnection) error {
	if r.ctx.Err() != nil {
		return r.ctx.Err()
	}
	select {
	case <-conn.Done():
		if ex, ok := host.ReadExit(r.dir); ok {
			return fmt.Errorf("host exited: %s (code %d %s)", ex.Reason, ex.ExitCode, ex.Signal)
		}
		return errors.New("host connection lost")
	default:
		return errors.New("session interrupted")
	}
}

func (r *runner) saveFields(f store.SessionFields) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	_, err = r.c.store.Update(context.WithoutCancel(r.ctx), "controller", r.id, store.Update{Fields: b})
	return err
}

func (r *runner) connect() (net.Conn, *bufio.Reader, host.Welcome, error) {
	deadline := time.Now().Add(15 * time.Second)
	for {
		nc, br, wel, err := host.Connect(r.ctx, r.dir)
		if err == nil {
			return nc, br, wel, nil
		}
		if time.Now().After(deadline) || r.ctx.Err() != nil {
			return nil, nil, wel, fmt.Errorf("connect to host: %w", err)
		}
		if ex, exited := host.ReadExit(r.dir); exited && !host.Alive(r.dir) {
			return nil, nil, wel, fmt.Errorf("host exited: %s", ex.Reason)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// work claims and prompts tickets until the pool has none for idle_timeout.
// A stale host is refreshed after the first turn that leaves it idle.
func (r *runner) work(ctx context.Context, conn *acp.ClientSideConnection, nc net.Conn, acpSession string, stale bool) error {
	pool, havePool := r.c.config().Pools[r.pool]
	def := defaultIdle
	if pool.Chat() {
		def = defaultChatIdle
	}
	idle := config.Duration(pool.IdleTimeout, def)
	idleSince := time.Now()
	for {
		if r.wantsRestart() {
			return r.restart(nc)
		}
		if r.worktreeMissing() {
			// Every tool call fails in a deleted cwd, so turns would only
			// claim and drop tickets. Respawning recreates the worktree
			// (spawnHost) and keeps the conversation and claims.
			r.log.Warn("session worktree missing, respawning host", "session", r.id,
				"caller", "runner.worktreeMissing", "host_pid", hostPid(r.dir))
			r.lifecycle("worktree_missing")
			_ = nc.Close()
			stopHost(r.dir)
			return errRefresh
		}
		// Config options (#304) change only between turns: jcode refuses
		// a change while a prompt runs. A reload or `session config`
		// shows up here before the next message or claim.
		if err := r.applyConfig(ctx, conn, acpSession); err != nil {
			return err
		}
		// Queued messages go before the next claim, so also after each turn.
		if err := r.deliver(ctx, conn, acpSession); err != nil {
			return err
		}
		// The env a new host would get can change while this one is live
		// (a deploy restarts the controller with a new binary, a reload
		// changes agent env). Re-check between turns so the next turn runs
		// in the current env; refresh keeps the conversation and claims.
		if r.currentEnvStale() {
			return r.refresh(nc)
		}
		if r.c.poolDrained(r.pool) && r.idle() {
			return errSessionEnd{resolution: resolutionDrained}
		}
		// A cordoned session (#281) finishes what it holds and claims
		// nothing new. Scale and drain cordons then end it; a manual one
		// keeps it open and out of rotation.
		cordon := r.c.sessionCordon(ctx, r.id)
		var cordonWork string
		if cordon != nil {
			cordonWork = r.c.cordonedWork(ctx, r.id)
		}
		if cordon != nil && cordonWork == "" {
			switch cordon.Reason {
			case store.CordonScale:
				return errSessionEnd{resolution: resolutionScaledDown}
			case store.CordonDrain:
				return errSessionEnd{resolution: resolutionDrained}
			}
		}
		var t store.Ticket
		err := store.ErrNoWork
		switch {
		case pool.Chat(): // chat sessions answer messages only
		case cordon != nil:
			if cordonWork != "" {
				t, _, err = r.c.store.Claim(ctx, cordonWork, r.id, r.c.leaseDur())
				if err != nil && ctx.Err() == nil {
					err = store.ErrNoWork // closed or reassigned meanwhile; look again
				}
			}
		default:
			t, _, err = r.c.store.ClaimNext(ctx, store.ClaimNext{
				Project: r.project(), Pool: r.pool, Holder: r.id, Lease: r.c.leaseDur(),
			})
		}
		switch {
		case errors.Is(err, store.ErrNoWork):
			if r.delivered.After(idleSince) {
				idleSince = r.delivered
			}
			if havePool && time.Since(idleSince) > idle {
				// A manually cordoned session stays for whoever cordoned
				// it; a warm pool keeps min_sessions (or the manual
				// target) open with no work.
				if cordon != nil || r.c.keepWarm(ctx, r.pool) {
					idleSince = time.Now()
				} else {
					return errSessionEnd{resolution: resolutionIdle}
				}
			}
			select {
			case <-ctx.Done():
				return r.dropped(conn)
			case <-r.restartWake:
			case <-time.After(r.c.opt.Poll):
			}
			continue
		case err != nil:
			if ctx.Err() != nil {
				return r.dropped(conn)
			}
			return err
		}
		r.c.metrics.Claims.Add(ctx, 1, metric.WithAttributes(telemetry.AttrPool.String(r.pool)))
		r.mu.Lock()
		r.lastClaim = time.Now()
		r.mu.Unlock()
		r.setTurnTicket(t.ID)
		comments, err := r.c.store.Comments(ctx, t.ID)
		if err != nil {
			r.setTurnTicket("")
			return err
		}
		interrupted, err := r.interruptible(ctx, conn, acpSession, r.ticketPrompt(t, comments))
		if err != nil {
			r.setTurnTicket("")
			return err
		}
		if r.wantsRestart() {
			// Keep the claim: the restarted host gets the ticket again.
			return r.restart(nc)
		}
		if interrupted {
			if err := r.deliver(ctx, conn, acpSession); err != nil {
				return err
			}
		}
		r.afterTurn(ctx, conn, acpSession, t.ID)
		if r.unclosed >= maxUnclosed {
			return errSessionEnd{resolution: resolutionStuck, reason: fmt.Sprintf("%d turns in a row ended without closing or releasing; %s", r.unclosed, r.turnEndReason())}
		}
		if stale && r.idle() {
			return r.refresh(nc)
		}
		idleSince = time.Now()
	}
}

func (r *runner) project() string {
	if p, ok := r.c.config().Pools[r.pool]; ok {
		return p.Project
	}
	t, err := r.c.store.Get(r.ctx, r.id)
	if err != nil {
		return r.c.config().DefaultProject
	}
	return t.Project
}

// turn sends one prompt and waits for the agent to finish it.
func (r *runner) turn(ctx context.Context, conn *acp.ClientSideConnection, acpSession, text string) error {
	timeout := config.Duration(r.c.config().Pools[r.pool].TurnTimeout, defaultTurn)
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	stop := context.AfterFunc(tctx, func() {
		if ctx.Err() == nil { // turn timeout, not a lost connection
			_ = conn.Cancel(context.Background(), acp.CancelNotification{SessionId: acp.SessionId(acpSession)})
		}
	})
	defer stop()
	if r.takeResumed() {
		text = recoveryNote + "\n\n" + text
	}
	if seed := r.takeSeed(); seed != "" {
		// No extra turn: the summary rides on the next prompt.
		text = seed + "\n\n" + host.SummaryEnd + "\n\n" + text
	}
	// A rate limit is not a crash: keep the session and its claims (the
	// supervisor keeps heartbeating) and retry the same prompt later.
	wait := r.c.opt.Backoff
	var res acp.PromptResponse
	var err error
	// Counted once per turn, after rate-limit retries; a shutdown is not a
	// turn outcome.
	defer func() {
		if ctx.Err() != nil {
			return
		}
		reason := string(res.StopReason)
		if err != nil {
			reason = "error"
		}
		r.c.metrics.Turns.Add(context.WithoutCancel(ctx), 1, metric.WithAttributes(
			telemetry.AttrPool.String(r.pool), telemetry.AttrStopReason.String(reason)))
	}()
	r.mu.Lock()
	r.lastStop = ""
	r.lastSaid.Reset()
	r.mu.Unlock()
	defer func() {
		r.mu.Lock()
		r.lastStop = string(res.StopReason)
		if err != nil {
			r.lastStop = "error"
		}
		r.mu.Unlock()
	}()
	busyRetries := 0
	for {
		res, err = conn.Prompt(context.WithoutCancel(ctx), acp.PromptRequest{
			SessionId: acp.SessionId(acpSession),
			Prompt:    []acp.ContentBlock{acp.TextBlock(text)},
			Meta:      r.promptMeta(),
		})
		if err != nil && ctx.Err() == nil && agentBusy(err) && busyRetries < 3 {
			busyRetries++
			r.log.Warn("agent still busy with an earlier prompt, cancelling it", "attempt", busyRetries, "err", err)
			_ = conn.Cancel(context.WithoutCancel(ctx), acp.CancelNotification{SessionId: acp.SessionId(acpSession)})
			select {
			case <-ctx.Done():
				return r.dropped(conn)
			case <-time.After(r.c.opt.Backoff):
			}
			continue
		}
		if err != nil && ctx.Err() == nil && agentBusy(err) {
			// Reconnecting keeps the same stuck agent (the crash loop seen
			// on gasserver); a new host starts a new agent and reloads the
			// conversation, and the claim survives for the next prompt.
			r.log.Warn("agent stuck busy, respawning host", "host_pid", hostPid(r.dir))
			r.lifecycle("agent_stuck_busy")
			stopHost(r.dir)
			return errRefresh
		}
		if err == nil || ctx.Err() != nil || !rateLimited(err) {
			break
		}
		r.log.Warn("provider rate limited, backing off", "wait", wait, "err", err)
		r.setBackoff(time.Now().Add(wait))
		select {
		case <-ctx.Done():
			r.setBackoff(time.Time{})
			return r.dropped(conn)
		case <-time.After(wait):
		}
		wait = min(wait*2, 15*r.c.opt.Backoff)
	}
	r.setBackoff(time.Time{})
	if err != nil {
		if ctx.Err() != nil {
			return r.dropped(conn)
		}
		return fmt.Errorf("session/prompt: %w", err)
	}
	r.log.Info("turn finished", "stop_reason", res.StopReason)
	return nil
}

// afterTurn makes sure a ticket does not stay claimed by a session that has
// stopped working on it. One reminder, then release with a deferral so a
// confused agent cannot spin on the same ticket.
func (r *runner) afterTurn(ctx context.Context, conn *acp.ClientSideConnection, acpSession, ticketID string) {
	defer r.setTurnTicket("")
	r.indexJournal(ctx, ticketID)
	held, err := r.c.store.List(ctx, store.Filter{Holder: r.id, Status: store.StatusInProgress})
	if err != nil {
		return
	}
	if len(held) == 0 {
		r.unclosed = 0
		return
	}
	if ticketID != "" {
		if err := r.turn(ctx, conn, acpSession, unclosedReminder(ids(held))); err != nil {
			return
		}
		if held, err = r.c.store.List(ctx, store.Filter{Holder: r.id, Status: store.StatusInProgress}); err != nil {
			return
		}
	}
	for _, t := range held {
		cl := store.Claim{TicketID: t.ID, Holder: r.id, Fence: t.Fence}
		until := time.Now().Add(unclosedDeferral)
		// Agents end turns to wait (a harness that backgrounds a long CI
		// watch, #241); the first such release of a ticket is a deferral,
		// not a failed attempt. A repeat counts, so a ticket no agent
		// finishes still reaches its retry limit.
		repeat := r.releasedUnclosedBefore(ctx, t.ID)
		release := until
		if repeat {
			release = time.Time{}
		}
		if _, err := r.c.store.Release(ctx, cl, unclosedReason, release); err != nil {
			continue
		}
		r.unclosed++
		if repeat {
			_, _ = r.c.store.Update(ctx, "controller", t.ID, store.Update{DeferUntil: &until})
		}
		r.log.Warn("released unclosed ticket", "ticket", t.ID, "failure", repeat)
		r.c.stepEvent(ctx, t)
	}
}

const unclosedReason = "agent turn ended without closing or releasing"

// unclosedReminder is the prompt after a turn ends holding tickets. The
// sessions in #301 ended turns "waiting for the wake" of a backgrounded
// command; no wake prompt reaches an ACP session, so say so.
func unclosedReminder(held string) string {
	return "You still hold " + held + " and your turn ended. Nothing will wake you: " +
		"a background task's wake or notification does not start a new turn in this session, " +
		"and the next time a turn ends holding the ticket it is released and the session ends. " +
		"Do one of these now, in this turn: run `jobshop close` if the work is done; " +
		"keep waiting in the foreground with short commands (`sleep 60; gh pr checks <n>`) until it finishes; " +
		"or `jobshop release --until <duration> -r \"<what you wait for>\"` for a long wait; " +
		"or `jobshop release -r \"<why>\"` if you cannot finish it."
}

// turnEndReason describes how the last turn ended, for the stuck reason.
func (r *runner) turnEndReason() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	said := strings.Join(strings.Fields(r.lastSaid.String()), " ")
	if n := len(said); n > 200 {
		said = "..." + said[n-200:]
	}
	stop := r.lastStop
	if stop == "" {
		stop = "unknown"
	}
	if said == "" {
		return "last turn stop_reason " + stop
	}
	return fmt.Sprintf("last turn stop_reason %s, agent said: %q", stop, said)
}

// releasedUnclosedBefore reports whether the controller already released
// id for an unclosed turn.
func (r *runner) releasedUnclosedBefore(ctx context.Context, id string) bool {
	hist, err := r.c.store.History(ctx, id)
	if err != nil {
		return false
	}
	for _, h := range hist {
		if h.Op == "release" && strings.Contains(string(h.Diff), unclosedReason) {
			return true
		}
	}
	return false
}

func ids(ts []store.Ticket) string {
	s := make([]string, len(ts))
	for i, t := range ts {
		s[i] = t.ID
	}
	return strings.Join(s, ", ")
}

func (r *runner) ticketPrompt(t store.Ticket, comments []store.Comment) string {
	parts := r.c.assemble(r.ctx, r.turnSpec(turnTicket, &t, true), PromptPart{Source: "ticket", Text: ticketPromptText("", t, comments)})
	return r.useParts(parts)
}

// turnSpec describes this session's next turn. It takes the pending
// system flag: system rules go out once per host.
func (r *runner) turnSpec(kind string, t *store.Ticket, poolPrompt bool) turnSpec {
	r.mu.Lock()
	sys := r.systemDue
	r.systemDue = false
	r.mu.Unlock()
	return turnSpec{
		session: r.id, pool: r.pool, agent: r.agent, project: r.projectName,
		kind: kind, ticket: t, system: sys, poolPrompt: poolPrompt,
	}
}

// useParts records which rules the prompt carries and returns its text.
func (r *runner) useParts(parts []PromptPart) string {
	names := ruleNames(parts)
	r.setTurnRules(names)
	if len(names) > 0 {
		r.log.Info("prompt rules applied", "rules", strings.Join(names, ","))
	}
	return joinParts(parts)
}

// Recent comments in a ticket prompt: a routed-back ticket's finding lives
// in a comment, and an agent that only sees the body redoes the work or
// stalls (#67, #70).
const (
	promptComments    = 3
	promptCommentSize = 2000
)

func ticketPromptText(poolPrompt string, t store.Ticket, comments []store.Comment) string {
	var b strings.Builder
	if p := poolPrompt; p != "" {
		b.WriteString(strings.TrimSpace(p))
		b.WriteString("\n\n")
	}
	fmt.Fprintf(&b, "You have claimed Jobshop ticket %s.\n\n# %s\n\n", t.ID, t.Title)
	if t.Body != "" {
		b.WriteString(t.Body)
		b.WriteString("\n\n")
	}
	if len(comments) > 0 {
		b.WriteString("# Recent comments\n\n")
		for _, c := range comments[max(0, len(comments)-promptComments):] {
			body := c.Body
			if len(body) > promptCommentSize {
				body = strings.ToValidUTF8(body[:promptCommentSize], "") + " [truncated]"
			}
			fmt.Fprintf(&b, "[%s] %s:\n%s\n\n", c.CreatedAt.UTC().Format(time.RFC3339), c.Author, body)
		}
		b.WriteString("If the latest comment is a refinery finding, address it, push, and hand the ticket back to refinery.\n\n")
	}
	b.WriteString(`The jobshop CLI is on your PATH and acts as your session; with no id it uses your current ticket.
- jobshop ticket show                       ticket details, comments and history
- jobshop comment "<text>"                  record progress or findings
- jobshop close                             when the work is done (-r failed if it cannot be done)
- jobshop release -r "<why>"                hand the ticket back without finishing it
- jobshop release --until <2h|RFC3339> -r "<why>"   not workable yet; deferred, not a failure
- jobshop ticket create "<title>" -b "<body>"   file follow-up work
Close or release the ticket before you end your turn. Ending a turn while
you hold it releases it, and a repeat counts as a failed attempt. To wait (CI, a
rebase, a deploy), stay in the turn: if a long command is moved to the
background, keep polling it (` + "`gh pr checks <n>`, `sleep 60`" + `) until it
finishes: a background task's wake does not start a new turn here. For
longer waits, ` + "`jobshop release --until <duration>`" + ` instead.`)
	return b.String()
}

// client is the ACP client side the agent talks to. It auto-approves
// permission prompts: agents run unattended in their own slice and worktree.
type client struct{ r *runner }

func (c *client) RequestPermission(_ context.Context, p acp.RequestPermissionRequest) (acp.RequestPermissionResponse, error) {
	for _, kind := range []acp.PermissionOptionKind{acp.PermissionOptionKindAllowAlways, acp.PermissionOptionKindAllowOnce} {
		for _, o := range p.Options {
			if o.Kind == kind {
				return acp.RequestPermissionResponse{Outcome: acp.NewRequestPermissionOutcomeSelected(o.OptionId)}, nil
			}
		}
	}
	return acp.RequestPermissionResponse{Outcome: acp.NewRequestPermissionOutcomeCancelled()}, nil
}

func (c *client) SessionUpdate(_ context.Context, n acp.SessionNotification) error {
	if m := n.Update.AgentMessageChunk; m != nil && m.Content.Text != nil {
		c.r.mu.Lock()
		c.r.lastSaid.WriteString(m.Content.Text.Text)
		if s := c.r.lastSaid.String(); len(s) > 4096 { // only the tail is reported
			c.r.lastSaid.Reset()
			c.r.lastSaid.WriteString(s[len(s)-1024:])
		}
		c.r.mu.Unlock()
	}
	return nil
}

func (c *client) HandleExtensionMethod(_ context.Context, method string, params json.RawMessage) (any, error) {
	if method == host.MethodExportStats {
		var s host.ExportStats
		if json.Unmarshal(params, &s) == nil {
			ctx := c.r.c.ctx
			attrs := metric.WithAttributes(telemetry.AttrPool.String(c.r.pool))
			c.r.c.metrics.ConvDropped.Add(ctx, s.Dropped, attrs)
			c.r.c.metrics.ConvTruncated.Add(ctx, s.Truncated, attrs)
		}
		return nil, nil
	}
	if method == host.MethodTurnEnd {
		select {
		case c.r.turnEnd <- struct{}{}:
		default:
		}
		return nil, nil
	}
	return nil, acp.NewMethodNotFound(method)
}

var errUnsupported = acp.NewMethodNotFound("not supported by jobshop")

func (c *client) ReadTextFile(context.Context, acp.ReadTextFileRequest) (acp.ReadTextFileResponse, error) {
	return acp.ReadTextFileResponse{}, errUnsupported
}

func (c *client) WriteTextFile(context.Context, acp.WriteTextFileRequest) (acp.WriteTextFileResponse, error) {
	return acp.WriteTextFileResponse{}, errUnsupported
}

func (c *client) CreateTerminal(context.Context, acp.CreateTerminalRequest) (acp.CreateTerminalResponse, error) {
	return acp.CreateTerminalResponse{}, errUnsupported
}

func (c *client) KillTerminal(context.Context, acp.KillTerminalRequest) (acp.KillTerminalResponse, error) {
	return acp.KillTerminalResponse{}, errUnsupported
}

func (c *client) TerminalOutput(context.Context, acp.TerminalOutputRequest) (acp.TerminalOutputResponse, error) {
	return acp.TerminalOutputResponse{}, errUnsupported
}

func (c *client) ReleaseTerminal(context.Context, acp.ReleaseTerminalRequest) (acp.ReleaseTerminalResponse, error) {
	return acp.ReleaseTerminalResponse{}, errUnsupported
}

func (c *client) WaitForTerminalExit(context.Context, acp.WaitForTerminalExitRequest) (acp.WaitForTerminalExitResponse, error) {
	return acp.WaitForTerminalExitResponse{}, errUnsupported
}

// sessionCwd is where the agent works: the session worktree or pool dir.
func (c *Controller) sessionCwd(t store.Ticket) string {
	p, ok := c.config().Pools[t.Pool]
	if ok && p.Repo != "" {
		return filepath.Join(c.root.SessionDir(t.ID), "work")
	}
	if ok {
		return p.Dir
	}
	return c.root.SessionDir(t.ID)
}

// hostSpec builds the spec a new host for t would get. EnvHash also covers
// the env the host inherits from this controller, so a deploy that changes
// either is noticed when a live host is adopted.
func (c *Controller) hostSpec(t store.Ticket, f store.SessionFields) (hostspec.Spec, string, error) {
	agent, ok := c.config().Agents[f.Agent]
	if !ok {
		return hostspec.Spec{}, "", fmt.Errorf("unknown agent %q", f.Agent)
	}
	exe, err := c.hostExe()
	if err != nil {
		return hostspec.Spec{}, "", err
	}
	dir := c.root.SessionDir(t.ID)
	env := []string{
		"JOBSHOP_SOCKET=" + c.socketPath(t.ID),
		"JOBSHOP_SESSION=" + t.ID,
		"PATH=" + filepath.Dir(exe) + string(os.PathListSeparator) + os.Getenv("PATH"),
		// Agents may run tools in a login shell (jcode does after a resume:
		// `bash -lc`). On NixOS /etc/profile then sources set-environment,
		// which overwrites PATH with the system one, so jobshop and gh
		// vanish (#253). This marker makes it skip that and keep ours.
		"__NIXOS_SET_ENVIRONMENT_DONE=1",
	}
	for k, v := range agent.Env {
		env = append(env, k+"="+v)
	}
	if f.TraceParent != "" {
		// The agent's `jobshop` calls join the session's trace.
		env = append(env, "JOBSHOP_TRACEPARENT="+f.TraceParent)
	}
	// ${SESSION_DIR} lets an agent keep per-session state, e.g.
	// `jcode acp --socket ${SESSION_DIR}/jcode.sock` for a private daemon
	// whose tool shells inherit this env.
	args := make([]string, len(agent.Args))
	for i, a := range agent.Args {
		args[i] = strings.ReplaceAll(a, "${SESSION_DIR}", dir)
	}
	spec := hostspec.Spec{Spec: host.Spec{
		Protocol: host.Protocol, Session: t.ID, Command: agent.Command, Args: args,
		Env: env, EnvHash: envHash(hostInheritedEnv(), env), Dir: c.sessionCwd(t),
		LeaseMS: hostLease.Milliseconds(),
	}}
	// Config validation guarantees a logs endpoint when a pool exports.
	if p, ok := c.config().Pools[t.Pool]; ok && p.ConversationExport {
		tel := c.config().Telemetry
		spec.ConvExport = &hostspec.ConvExport{
			LogsURL: tel.LogsURL(), Headers: tel.Headers(os.Getenv), Pool: t.Pool, Project: t.Project,
			TraceParent: f.TraceParent,
		}
	}
	return spec, exe, nil
}

// hostInheritedEnv is the controller env a spawned host starts with.
func hostInheritedEnv() []string {
	// systemd sets these per service start; hashing them would refresh
	// every session on each controller restart.
	return withoutEnv(os.Environ(), "JOBSHOP_ROOT", "JOBSHOP_SOCKET",
		"INVOCATION_ID", "JOURNAL_STREAM", "SYSTEMD_EXEC_PID", "MEMORY_PRESSURE_WATCH", "MEMORY_PRESSURE_WRITE")
}

// envHash hashes the agent's final env (inherited, overridden by env),
// sorted so agent.Env map order does not matter.
func envHash(inherited, env []string) string {
	m := map[string]string{}
	for _, kv := range append(slices.Clone(inherited), env...) {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	h := sha256.New()
	for _, k := range slices.Sorted(maps.Keys(m)) {
		_, _ = fmt.Fprintf(h, "%s=%s\x00", k, m[k])
	}
	return hex.EncodeToString(h.Sum(nil))
}

// spawnHost prepares the session directory and starts a detached host.
func (c *Controller) spawnHost(t store.Ticket, f store.SessionFields) error {
	spec, exe, err := c.hostSpec(t, f)
	if err != nil {
		return err
	}
	dir := c.root.SessionDir(t.ID)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	worktree, err := c.ensureWorktree(t, spec.Dir)
	if err != nil {
		return err
	}
	// A worktree was just made; any other cwd must already exist.
	if err := host.CheckSpawn(spec.Spec, worktree); err != nil {
		return err
	}
	if err := hostspec.Write(dir, spec); err != nil {
		return err
	}
	logf, err := os.OpenFile(filepath.Join(dir, host.LogFile), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = logf.Close() }()
	cmd := exec.Command(exe, "session-host", dir)
	cmd.Args[0] = "jobshop"
	cmd.Dir = dir
	cmd.Stdout, cmd.Stderr = logf, logf
	cmd.Env = hostInheritedEnv()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	// The previous host's exit record would make connect() report this
	// host as exited before it has taken the lock.
	_ = os.Remove(filepath.Join(dir, host.ExitFile))
	if err := cmd.Start(); err != nil {
		return err
	}
	// Reap it if it exits while this controller lives; after a controller
	// restart init reaps it instead.
	go func() { _ = cmd.Wait() }()
	c.log.Info("spawned host", "session", t.ID, "pid", cmd.Process.Pid, "agent", f.Agent)
	return nil
}

func (c *Controller) hostExe() (string, error) {
	if c.opt.HostExe != "" {
		return c.opt.HostExe, nil
	}
	// The host runs the exact build of this controller even if the release
	// symlink moves later.
	return os.Readlink("/proc/self/exe")
}

func withoutEnv(env []string, keys ...string) []string {
	out := env[:0:0]
	for _, kv := range env {
		drop := false
		for _, k := range keys {
			if strings.HasPrefix(kv, k+"=") {
				drop = true
			}
		}
		if !drop {
			out = append(out, kv)
		}
	}
	return out
}

// ensureWorktree makes the session's worktree of the pool repo. It reports
// whether the pool uses a worktree. A missing repo, or a ref or path git
// rejects, is a permanent error (#146).
func (c *Controller) ensureWorktree(t store.Ticket, cwd string) (bool, error) {
	p, ok := c.config().Pools[t.Pool]
	if !ok || p.Repo == "" {
		return false, nil
	}
	c.gitMu.Lock()
	defer c.gitMu.Unlock()
	if _, err := os.Stat(cwd); err == nil {
		return true, nil
	}
	if st, err := os.Stat(p.Repo); err != nil {
		return true, host.Permanentf("pool %s repo %s: %w", t.Pool, p.Repo, err)
	} else if !st.IsDir() {
		return true, host.Permanentf("pool %s repo %s is not a directory", t.Pool, p.Repo)
	}
	ref := p.Ref
	if ref == "" {
		ref = "HEAD"
	}
	// Start from the latest remote state; a failed fetch (offline, no
	// remote) still allows a worktree from what is there.
	if out, err := exec.Command("git", "-C", p.Repo, "fetch", "--quiet", "--prune").CombinedOutput(); err != nil {
		c.log.Warn("git fetch", "repo", p.Repo, "err", err, "out", strings.TrimSpace(string(out)))
	}
	// A worktree deleted without git (rm -rf) stays registered and makes
	// `worktree add` refuse the same path. It is locked (below) and prune
	// skips locked entries, so unlock it first.
	_ = exec.Command("git", "-C", p.Repo, "worktree", "unlock", cwd).Run()
	if out, err := exec.Command("git", "-C", p.Repo, "worktree", "prune").CombinedOutput(); err != nil {
		c.log.Warn("git worktree prune", "repo", p.Repo, "err", err, "out", strings.TrimSpace(string(out)))
	}
	out, err := exec.Command("git", "-C", p.Repo, "worktree", "add", "--detach",
		"--lock", "--reason", "jobshop session "+t.ID, cwd, ref).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("git worktree add: %w: %s", err, strings.TrimSpace(string(out)))
		if worktreeAddPermanent(string(out)) {
			err = &host.PermanentError{Err: err}
		}
		return true, err
	}
	return true, nil
}

// worktreeAddPermanent matches git's messages for a repo or ref that is
// not there; lock and transient I/O failures are not matched.
func worktreeAddPermanent(out string) bool {
	for _, s := range []string{"not a git repository", "invalid reference", "not a valid object name", "Permission denied", "No such file or directory"} {
		if strings.Contains(out, s) {
			return true
		}
	}
	return false
}

// removeWorktree deletes a session's worktree. caller names the code path
// for the WARN (#154). It refuses while the ticket is open or the host is
// alive: a live session's worktree must never go.
func (c *Controller) removeWorktree(t store.Ticket, caller string) {
	p, ok := c.config().Pools[t.Pool]
	if !ok || p.Repo == "" {
		return
	}
	cwd := c.sessionCwd(t)
	if _, err := os.Stat(cwd); err != nil {
		return
	}
	if t.Status != store.StatusClosed || host.Alive(c.root.SessionDir(t.ID)) {
		c.log.Warn("refusing to remove worktree of live session", "session", t.ID, "caller", caller,
			"status", t.Status, "cwd", cwd)
		return
	}
	c.log.Warn("removing session worktree", "session", t.ID, "caller", caller, "cwd", cwd)
	c.gitMu.Lock()
	defer c.gitMu.Unlock()
	// The worktree is locked so outside `git worktree remove` (gh pr merge
	// removing the worktree of a deleted branch) fails; two --force
	// override the lock.
	if out, err := exec.Command("git", "-C", p.Repo, "worktree", "remove", "--force", "--force", cwd).CombinedOutput(); err != nil {
		c.log.Warn("git worktree remove", "session", t.ID, "err", err, "out", strings.TrimSpace(string(out)))
	}
}

// supervise runs once per tick: heartbeats claims held by live sessions and
// starts sessions for pools with waiting work.
func (c *Controller) supervise() {
	ctx := c.ctx
	live := c.liveRunners()
	for _, id := range live {
		held, err := c.store.List(ctx, store.Filter{Holder: id, Status: store.StatusInProgress})
		if err != nil {
			continue
		}
		for _, t := range held {
			cl := store.Claim{TicketID: t.ID, Holder: id, Fence: t.Fence}
			if _, err := c.store.Heartbeat(ctx, cl, c.leaseDur()); err != nil && !errors.Is(err, store.ErrNotHolder) {
				c.log.Warn("heartbeat", "ticket", t.ID, "err", err)
			}
		}
	}
	c.recordGauges()
	if c.draining() {
		return
	}
	c.reconcile()
}

// recordGauges sets the session and ready-ticket gauges for each pool.
// Session state is backoff, working (holds a ticket) or idle.
func (c *Controller) recordGauges() {
	ctx := c.ctx
	backoff := map[string]bool{}
	c.mu.Lock()
	for id, r := range c.runners {
		if !r.backoff().IsZero() {
			backoff[id] = true
		}
	}
	c.mu.Unlock()
	open, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen})
	if err != nil {
		return
	}
	counts := map[[2]string]int64{}
	for name := range c.config().Pools {
		for _, st := range []string{"idle", "working", "backoff"} {
			counts[[2]string{name, st}] = 0
		}
	}
	for _, s := range open {
		state := "idle"
		if backoff[s.ID] {
			state = "backoff"
		} else if held, err := c.store.List(ctx, store.Filter{Holder: s.ID, Status: store.StatusInProgress, Limit: 1}); err == nil && len(held) > 0 {
			state = "working"
		}
		counts[[2]string{s.Pool, state}]++
	}
	for k, n := range counts {
		c.metrics.Sessions.Record(ctx, n, metric.WithAttributes(telemetry.AttrPool.String(k[0]), telemetry.AttrState.String(k[1])))
	}
	for name, p := range c.config().Pools {
		// Ready is capped by the store's default limit; enough for a gauge.
		if ready, err := c.store.Ready(ctx, p.Project, name, 0); err == nil {
			c.metrics.Ready.Record(ctx, int64(len(ready)), metric.WithAttributes(
				telemetry.AttrProject.String(p.Project), telemetry.AttrPool.String(name)))
		}
	}
}

// backingOffPools returns pools with a session waiting out a rate limit.
func (c *Controller) backingOffPools() map[string]bool {
	c.mu.Lock()
	rs := make([]*runner, 0, len(c.runners))
	for _, r := range c.runners {
		rs = append(rs, r)
	}
	c.mu.Unlock()
	out := map[string]bool{}
	for _, r := range rs {
		if !r.backoff().IsZero() {
			out[r.pool] = true
		}
	}
	return out
}

func (c *Controller) draining() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.drain
}

// sessionTrace starts the trace a new session's lifecycle spans and its
// agent's CLI calls join, returning its traceparent. The root span ends at
// once: a span open for the whole session would be lost on restart.
func (c *Controller) sessionTrace(ctx context.Context, project, pool, agent string) string {
	ctx, span := c.tracer.Start(trace.ContextWithSpanContext(ctx, trace.SpanContext{}), "session",
		trace.WithNewRoot(), trace.WithAttributes(telemetry.AttrProject.String(project),
			telemetry.AttrPool.String(pool), attribute.String("jobshop.agent", agent)))
	defer span.End()
	return telemetry.TraceParent(ctx)
}

// lifecycle records one session lifecycle event (spawn, adopt, restart,
// end) as a span in the session's trace.
func (r *runner) lifecycle(event string, attrs ...attribute.KeyValue) {
	ctx := telemetry.WithTraceParent(r.ctx, r.traceParent)
	r.mu.Lock()
	model, effort := r.model, r.effort
	r.mu.Unlock()
	_, span := r.c.tracer.Start(ctx, "session "+event, trace.WithAttributes(append(attrs,
		telemetry.AttrSession.String(r.id), telemetry.AttrPool.String(r.pool),
		telemetry.AttrModel.String(model), telemetry.AttrEffort.String(effort))...))
	span.End()
}

// newSession creates a session ticket, opens its CLI socket and, when its
// agent is configured, starts its driver. Sessions whose agent is not in
// the config are external: something else (a human running an agent with
// JOBSHOP_SOCKET) drives them.
func (c *Controller) newSession(ctx context.Context, actor, project, pool, agent, key, title string) (store.Ticket, error) {
	if agent == "" {
		return store.Ticket{}, fmt.Errorf("%w: agent is required", store.ErrInvalid)
	}
	fields, err := json.Marshal(store.SessionFields{Agent: agent, Key: key, TraceParent: c.sessionTrace(ctx, project, pool, agent)})
	if err != nil {
		return store.Ticket{}, err
	}
	if title == "" {
		title = agent + " session"
		if pool != "" {
			title = pool + " " + title
		}
	}
	t, err := c.store.Create(ctx, actor, store.NewTicket{
		Project: project, Kind: store.KindSession, Title: title, Pool: pool, Fields: fields,
	})
	if err != nil {
		return t, err
	}
	if err := c.openSession(t); err != nil {
		return t, err
	}
	c.drive(t)
	return t, nil
}

// drive starts the driver for a session whose agent is configured.
func (c *Controller) drive(t store.Ticket) {
	f, err := store.DecodeFields[store.SessionFields](t)
	if err != nil {
		return
	}
	if _, ok := c.config().Agents[f.Agent]; ok {
		c.startRunner(t)
	}
}
