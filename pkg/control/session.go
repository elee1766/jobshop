package control

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	acp "github.com/coder/acp-go-sdk"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/host"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
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

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	mu       sync.Mutex
	turnEnd  chan struct{} // signalled by _jobshop/turn_end
	crashes  []time.Time
	stopping bool
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
		var end errSessionEnd
		if errors.As(err, &end) {
			r.finish(end.resolution, end.reason)
			return
		}
		r.log.Warn("session interrupted, restarting", "err", err)
		if r.crashed() {
			r.finish(resolutionCrashed, fmt.Sprintf("%d restarts in %s: %v", maxCrashes, crashWindow, err))
			return
		}
		select {
		case <-r.ctx.Done():
			return
		case <-time.After(2 * time.Second):
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
	stopHost(r.dir)
	ctx := context.WithoutCancel(r.ctx)
	if _, err := r.c.store.CloseTicket(ctx, "controller", r.id, nil, resolution); err != nil && !errors.Is(err, store.ErrClosed) {
		r.log.Error("close session ticket", "err", err)
	}
	if reason != "" {
		_, _ = r.c.store.AddComment(ctx, "controller", r.id, reason)
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
		fresh = true
	}
	nc, rd, welcome, err := r.connect()
	if err != nil {
		return err
	}
	defer func() { _ = nc.Close() }()

	conn := acp.NewClientSideConnection(&client{r: r}, nc, rd)
	conn.SetLogger(rpc.QuietLogger(r.log))
	r.log.Info("connected to host", "host_pid", welcome.Pid, "agent_pid", welcome.AgentPid, "fresh", fresh, "busy", welcome.Busy)

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
	} else {
		select {
		case <-r.turnEnd: // a turn that ended while no controller was attached
		default:
		}
	}
	return r.work(ctx, conn, fields.ACPSession)
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
func (r *runner) work(ctx context.Context, conn *acp.ClientSideConnection, acpSession string) error {
	pool, havePool := r.c.cfg.Pools[r.pool]
	idle := config.Duration(pool.IdleTimeout, defaultIdle)
	idleSince := time.Now()
	for {
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
		if err := r.turn(ctx, conn, acpSession, r.ticketPrompt(t)); err != nil {
			return err
		}
		r.afterTurn(ctx, conn, acpSession, t.ID)
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
	res, err := conn.Prompt(context.WithoutCancel(ctx), acp.PromptRequest{
		SessionId: acp.SessionId(acpSession),
		Prompt:    []acp.ContentBlock{acp.TextBlock(text)},
	})
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

// spawnHost prepares the session directory and starts a detached host.
func (c *Controller) spawnHost(t store.Ticket, f store.SessionFields) error {
	agent, ok := c.cfg.Agents[f.Agent]
	if !ok {
		return fmt.Errorf("unknown agent %q", f.Agent)
	}
	dir := c.root.SessionDir(t.ID)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	cwd := c.sessionCwd(t)
	if err := c.ensureWorktree(t, cwd); err != nil {
		return err
	}
	exe, err := c.hostExe()
	if err != nil {
		return err
	}
	env := []string{
		"JOBSHOP_SOCKET=" + c.socketPath(t.ID),
		"JOBSHOP_SESSION=" + t.ID,
		"PATH=" + filepath.Dir(exe) + string(os.PathListSeparator) + os.Getenv("PATH"),
	}
	for k, v := range agent.Env {
		env = append(env, k+"="+v)
	}
	// ${SESSION_DIR} lets an agent keep per-session state, e.g.
	// `jcode acp --socket ${SESSION_DIR}/jcode.sock` for a private daemon
	// whose tool shells inherit this env.
	args := make([]string, len(agent.Args))
	for i, a := range agent.Args {
		args[i] = strings.ReplaceAll(a, "${SESSION_DIR}", dir)
	}
	if err := host.WriteSpec(dir, host.Spec{
		Protocol: host.Protocol, Session: t.ID, Command: agent.Command, Args: args,
		Env: env, Dir: cwd, LeaseMS: hostLease.Milliseconds(),
	}); err != nil {
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
	cmd.Env = withoutEnv(os.Environ(), "JOBSHOP_ROOT", "JOBSHOP_SOCKET")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
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
	if c.draining() {
		return
	}
	for name, p := range c.cfg.Pools {
		if p.MaxSessions == 0 {
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
		// Each open session takes one ready ticket; start more only for the rest.
		want := min(len(ready)-len(open), p.MaxSessions-len(open))
		for range max(want, 0) {
			if _, err := c.newSession(ctx, "controller", p.Project, name, p.Agent, "", ""); err != nil {
				c.log.Error("start session", "pool", name, "err", err)
				break
			}
		}
	}
}

func (c *Controller) draining() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.drain
}

// newSession creates a session ticket, opens its CLI socket and, when its
// agent is configured, starts its driver. Sessions whose agent is not in
// the config are external: something else (a human running an agent with
// JOBSHOP_SOCKET) drives them.
func (c *Controller) newSession(ctx context.Context, actor, project, pool, agent, key, title string) (store.Ticket, error) {
	if agent == "" {
		return store.Ticket{}, fmt.Errorf("%w: agent is required", store.ErrInvalid)
	}
	fields, err := json.Marshal(store.SessionFields{Agent: agent, Key: key})
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
