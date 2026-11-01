package control

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	acp "github.com/coder/acp-go-sdk"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/host"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

const (
	defaultIdle       = 5 * time.Minute
	defaultTurn       = 2 * time.Hour
	hostLease         = 30 * time.Minute
	pollInterval      = 5 * time.Second
	crashWindow       = 10 * time.Minute
	maxCrashes        = 3
	unclosedDeferral  = 10 * time.Minute
	resolutionIdle    = "idle"
	resolutionCrashed = "crashed"
)

// runner drives one session: keeps its host alive, speaks ACP to the agent
// through it, and feeds it tickets.
type runner struct {
	c    *Controller
	id   string
	pool string // "" for manual sessions
	dir  string
	log  *slog.Logger
	// traceParent is the session trace (SessionFields.TraceParent).
	traceParent string

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	mu       sync.Mutex
	turnEnd  chan struct{} // signalled by _jobshop/turn_end
	crashes  []time.Time
	stopping bool
	agentPid int // from the host's welcome; 0 until connected
	// backoffUntil is set while the agent's provider is rate limited and the
	// runner waits to retry the same turn.
	backoffUntil time.Time
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
	}
	if f, err := store.DecodeFields[store.SessionFields](t); err == nil {
		r.traceParent = f.TraceParent
	}
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
		// Pause first: the release makes the ticket ready, and a supervisor
		// tick in between would start another session.
		r.c.poolCrashed(r.pool)
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

type errSessionEnd struct{ resolution, reason string }

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
		if err := r.c.spawnHost(t, fields); err != nil {
			return errSessionEnd{resolution: resolutionCrashed, reason: "spawn: " + err.Error()}
		}
		r.lifecycle("spawn")
		fresh = true
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
		if err := r.saveFields(fields); err != nil {
			return err
		}
	case fresh && init.AgentCapabilities.LoadSession:
		_, err := conn.LoadSession(ctx, acp.LoadSessionRequest{
			SessionId: acp.SessionId(fields.ACPSession), Cwd: cwd, McpServers: []acp.McpServer{},
		})
		if err != nil {
			return fmt.Errorf("session/load: %w", err)
		}
		r.log.Info("resumed conversation", "acp_session", fields.ACPSession)
	case fresh:
		ns, err := conn.NewSession(ctx, acp.NewSessionRequest{Cwd: cwd, McpServers: []acp.McpServer{}})
		if err != nil {
			return fmt.Errorf("session/new: %w", err)
		}
		r.log.Warn("agent cannot load sessions; started a new conversation", "old", fields.ACPSession)
		fields.ACPSession = string(ns.SessionId)
		if err := r.saveFields(fields); err != nil {
			return err
		}
	}

	if welcome.Busy {
		// A turn started by a previous controller is still running.
		select {
		case <-r.turnEnd:
		case <-ctx.Done():
			return r.dropped(conn)
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
	spec, err := host.ReadSpec(r.dir)
	if err != nil {
		return false
	}
	want, _, err := r.c.hostSpec(t, f)
	if err != nil {
		return false
	}
	return spec.EnvHash != want.EnvHash
}

// idle: the session holds no in_progress ticket.
func (r *runner) idle() bool {
	held, err := r.c.store.List(r.ctx, store.Filter{Holder: r.id, Status: store.StatusInProgress})
	return err == nil && len(held) == 0
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
	pool, havePool := r.c.cfg.Pools[r.pool]
	idle := config.Duration(pool.IdleTimeout, defaultIdle)
	idleSince := time.Now()
	for {
		// Queued messages go before the next claim, so also after each turn.
		if err := r.deliver(ctx, conn, acpSession); err != nil {
			return err
		}
		if r.c.poolDrained(r.pool) && r.idle() {
			return errSessionEnd{resolution: resolutionDrained}
		}
		t, _, err := r.c.store.ClaimNext(ctx, store.ClaimNext{
			Project: r.project(), Pool: r.pool, Holder: r.id, Lease: r.c.lease,
		})
		switch {
		case errors.Is(err, store.ErrNoWork):
			if havePool && time.Since(idleSince) > idle {
				return errSessionEnd{resolution: resolutionIdle}
			}
			select {
			case <-ctx.Done():
				return r.dropped(conn)
			case <-time.After(pollInterval):
			}
			continue
		case err != nil:
			if ctx.Err() != nil {
				return r.dropped(conn)
			}
			return err
		}
		r.c.metrics.Claims.Add(ctx, 1, metric.WithAttributes(telemetry.AttrPool.String(r.pool)))
		interrupted, err := r.interruptible(ctx, conn, acpSession, r.ticketPrompt(t))
		if err != nil {
			return err
		}
		if interrupted {
			if err := r.deliver(ctx, conn, acpSession); err != nil {
				return err
			}
		}
		r.afterTurn(ctx, conn, acpSession, t.ID)
		if stale && r.idle() {
			return r.refresh(nc)
		}
		idleSince = time.Now()
	}
}

func (r *runner) project() string {
	if p, ok := r.c.cfg.Pools[r.pool]; ok {
		return p.Project
	}
	t, err := r.c.store.Get(r.ctx, r.id)
	if err != nil {
		return r.c.cfg.DefaultProject
	}
	return t.Project
}

// turn sends one prompt and waits for the agent to finish it.
func (r *runner) turn(ctx context.Context, conn *acp.ClientSideConnection, acpSession, text string) error {
	timeout := config.Duration(r.c.cfg.Pools[r.pool].TurnTimeout, defaultTurn)
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	stop := context.AfterFunc(tctx, func() {
		if ctx.Err() == nil { // turn timeout, not a lost connection
			_ = conn.Cancel(context.Background(), acp.CancelNotification{SessionId: acp.SessionId(acpSession)})
		}
	})
	defer stop()
	// A rate limit is not a crash: keep the session and its claims (the
	// supervisor keeps heartbeating) and retry the same prompt later.
	wait := r.c.opt.Backoff
	var res acp.PromptResponse
	var err error
	for {
		res, err = conn.Prompt(context.WithoutCancel(ctx), acp.PromptRequest{
			SessionId: acp.SessionId(acpSession),
			Prompt:    []acp.ContentBlock{acp.TextBlock(text)},
		})
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
	held, err := r.c.store.List(ctx, store.Filter{Holder: r.id, Status: store.StatusInProgress})
	if err != nil || len(held) == 0 {
		return
	}
	if ticketID != "" {
		reminder := "Your turn ended but you still hold " + ids(held) + ". " +
			"If the work is finished run `jobshop close`; if you cannot finish it run " +
			"`jobshop release -r \"<why>\"`; otherwise continue working."
		if err := r.turn(ctx, conn, acpSession, reminder); err != nil {
			return
		}
		if held, err = r.c.store.List(ctx, store.Filter{Holder: r.id, Status: store.StatusInProgress}); err != nil {
			return
		}
	}
	for _, t := range held {
		cl := store.Claim{TicketID: t.ID, Holder: r.id, Fence: t.Fence}
		if _, err := r.c.store.Release(ctx, cl, "agent turn ended without closing or releasing"); err != nil {
			continue
		}
		until := time.Now().Add(unclosedDeferral)
		_, _ = r.c.store.Update(ctx, "controller", t.ID, store.Update{DeferUntil: &until})
		r.log.Warn("released unclosed ticket", "ticket", t.ID)
	}
}

func ids(ts []store.Ticket) string {
	s := make([]string, len(ts))
	for i, t := range ts {
		s[i] = t.ID
	}
	return strings.Join(s, ", ")
}

func (r *runner) ticketPrompt(t store.Ticket) string {
	var b strings.Builder
	if p := r.c.cfg.Pools[r.pool].Prompt; p != "" {
		b.WriteString(strings.TrimSpace(p))
		b.WriteString("\n\n")
	}
	fmt.Fprintf(&b, "You have claimed Jobshop ticket %s.\n\n# %s\n\n", t.ID, t.Title)
	if t.Body != "" {
		b.WriteString(t.Body)
		b.WriteString("\n\n")
	}
	b.WriteString(`The jobshop CLI is on your PATH and acts as your session; with no id it uses your current ticket.
- jobshop ticket show                       ticket details, comments and history
- jobshop comment "<text>"                  record progress or findings
- jobshop close                             when the work is done (-r failed if it cannot be done)
- jobshop release -r "<why>"                hand the ticket back without finishing it
- jobshop ticket create "<title>" -b "<body>"   file follow-up work
Close or release the ticket before you end your turn.`)
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

func (c *client) SessionUpdate(context.Context, acp.SessionNotification) error { return nil }

func (c *client) HandleExtensionMethod(_ context.Context, method string, _ json.RawMessage) (any, error) {
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
	p, ok := c.cfg.Pools[t.Pool]
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
func (c *Controller) hostSpec(t store.Ticket, f store.SessionFields) (host.Spec, string, error) {
	agent, ok := c.cfg.Agents[f.Agent]
	if !ok {
		return host.Spec{}, "", fmt.Errorf("unknown agent %q", f.Agent)
	}
	exe, err := c.hostExe()
	if err != nil {
		return host.Spec{}, "", err
	}
	dir := c.root.SessionDir(t.ID)
	env := []string{
		"JOBSHOP_SOCKET=" + c.socketPath(t.ID),
		"JOBSHOP_SESSION=" + t.ID,
		"PATH=" + filepath.Dir(exe) + string(os.PathListSeparator) + os.Getenv("PATH"),
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
	return host.Spec{
		Protocol: host.Protocol, Session: t.ID, Command: agent.Command, Args: args,
		Env: env, EnvHash: envHash(hostInheritedEnv(), env), Dir: c.sessionCwd(t),
		LeaseMS: hostLease.Milliseconds(),
	}, exe, nil
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
	if err := c.ensureWorktree(t, spec.Dir); err != nil {
		return err
	}
	if err := host.WriteSpec(dir, spec); err != nil {
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

func (c *Controller) ensureWorktree(t store.Ticket, cwd string) error {
	p, ok := c.cfg.Pools[t.Pool]
	if !ok || p.Repo == "" {
		return nil
	}
	c.gitMu.Lock()
	defer c.gitMu.Unlock()
	if _, err := os.Stat(cwd); err == nil {
		return nil
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
	out, err := exec.Command("git", "-C", p.Repo, "worktree", "add", "--detach", cwd, ref).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (c *Controller) removeWorktree(t store.Ticket) {
	p, ok := c.cfg.Pools[t.Pool]
	if !ok || p.Repo == "" {
		return
	}
	cwd := c.sessionCwd(t)
	if _, err := os.Stat(cwd); err != nil {
		return
	}
	c.gitMu.Lock()
	defer c.gitMu.Unlock()
	if out, err := exec.Command("git", "-C", p.Repo, "worktree", "remove", "--force", cwd).CombinedOutput(); err != nil {
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
			if _, err := c.store.Heartbeat(ctx, cl, c.lease); err != nil && !errors.Is(err, store.ErrNotHolder) {
				c.log.Warn("heartbeat", "ticket", t.ID, "err", err)
			}
		}
	}
	c.recordGauges()
	if c.draining() {
		return
	}
	backingOff := c.backingOffPools()
	for name, p := range c.cfg.Pools {
		if p.MaxSessions == 0 || c.poolDrained(name) || backingOff[name] || c.poolPaused(name) {
			// A rate-limited pool would only start more sessions that hit
			// the same limit; a crash-looping one is paused.
			continue
		}
		open, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen, Pool: name})
		if err != nil {
			continue
		}
		ready, err := c.store.Ready(ctx, p.Project, name, p.MaxSessions)
		if err != nil {
			continue
		}
		// Idle sessions (holding nothing) will claim ready work themselves;
		// start more only for the rest, up to the cap.
		idle := 0
		for _, s := range open {
			held, err := c.store.List(ctx, store.Filter{Holder: s.ID, Status: store.StatusInProgress})
			if err == nil && len(held) == 0 {
				idle++
			}
		}
		want := min(len(ready)-idle, p.MaxSessions-len(open))
		for range max(want, 0) {
			if _, err := c.newSession(ctx, "controller", p.Project, name, p.Agent, "", ""); err != nil {
				c.log.Error("start session", "pool", name, "err", err)
				break
			}
		}
	}
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
	for name := range c.cfg.Pools {
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
	for name, p := range c.cfg.Pools {
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
	_, span := r.c.tracer.Start(ctx, "session "+event, trace.WithAttributes(append(attrs,
		telemetry.AttrSession.String(r.id), telemetry.AttrPool.String(r.pool))...))
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
	if _, ok := c.cfg.Agents[f.Agent]; ok {
		c.startRunner(t)
	}
}
