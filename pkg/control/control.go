// Package control is the controller: it owns the store, serves the admin
// socket and per-session CLI sockets, and runs the tick (lease sweep).
// Every write in a deployment goes through it.
package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	acp "github.com/coder/acp-go-sdk"
	"golang.org/x/sys/unix"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// DefaultLease is the claim lease when the config sets none.
const DefaultLease = 30 * time.Minute

// Options tune a controller. Zero values use defaults.
type Options struct {
	Tick time.Duration    // lease sweep interval, default 30s
	Now  func() time.Time // clock, for tests
	Log  *slog.Logger
	// HostExe is the binary started as session-host. Default /proc/self/exe.
	HostExe string
	// Backoff is the first wait after a provider rate limit, doubling to
	// 15x. Default 1m; tests shorten it.
	Backoff time.Duration
	// RestartDelay is the pause before restarting an interrupted session.
	// Default 2s.
	RestartDelay time.Duration
	// MessagePoll is how often a running turn checks for interrupt
	// messages. Default 2s.
	MessagePoll time.Duration
	// MetricsReady, if set, receives the /metrics listen address once it
	// is bound (tests use ":0" in the config).
	MetricsReady func(addr string)
}

// Controller serves one Jobshop root.
type Controller struct {
	root  config.Root
	cfg   *config.Config
	store store.Store
	lease time.Duration
	opt   Options
	log   *slog.Logger

	metrics  *telemetry.Metrics
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	sessions map[string]context.CancelFunc // open session id -> stops its cli.sock
	runners  map[string]*runner
	drain    bool            // full shutdown in progress: start nothing new
	drained  map[string]bool // drained pools (drainAll for every pool); memory only
	// Pool crash-loop quarantine, memory only (see quarantine.go).
	poolCrashLog    map[string][]time.Time
	poolPausedUntil map[string]time.Time
	// gitMu serializes git commands on pool repos: concurrent fetches in
	// one repo fail on the remote-ref lock.
	gitMu sync.Mutex
	wg    sync.WaitGroup
}

// Run opens the root, serves until ctx ends or a shutdown request, and
// cleans up. It holds an exclusive lock so one root has one controller.
func Run(ctx context.Context, root config.Root, opt Options) error {
	cfg, err := config.Load(root)
	if err != nil {
		return err
	}
	lease := DefaultLease
	if cfg.Lease != "" {
		if lease, err = time.ParseDuration(cfg.Lease); err != nil || lease <= 0 {
			return fmt.Errorf("%s: lease %q: must be a positive duration", config.FileName, cfg.Lease)
		}
	}
	if opt.Tick <= 0 {
		opt.Tick = 30 * time.Second
	}
	if opt.Backoff <= 0 {
		opt.Backoff = time.Minute
	}
	if opt.RestartDelay <= 0 {
		opt.RestartDelay = 2 * time.Second
	}
	if opt.MessagePoll <= 0 {
		opt.MessagePoll = 2 * time.Second
	}
	if opt.Log == nil {
		opt.Log = slog.Default()
	}
	unlock, err := lock(root.Lock())
	if err != nil {
		return err
	}
	defer unlock()

	st, err := sqlite.Open(ctx, root.DB(), sqlite.Options{Now: opt.Now, Prefix: cfg.Prefix})
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()

	m, err := telemetry.NewMetrics()
	if err != nil {
		return err
	}
	defer func() { _ = m.Shutdown(context.WithoutCancel(ctx)) }()

	c := &Controller{
		root: root, cfg: cfg, store: st, lease: lease, opt: opt, log: opt.Log, metrics: m,
		sessions: map[string]context.CancelFunc{}, runners: map[string]*runner{},
		drained:      map[string]bool{},
		poolCrashLog: map[string][]time.Time{}, poolPausedUntil: map[string]time.Time{},
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	defer c.cancel()
	return c.serve()
}

func lock(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, unix.EWOULDBLOCK) {
			return nil, fmt.Errorf("another controller holds %s", path)
		}
		return nil, err
	}
	return func() { _ = f.Close() }, nil
}

// listen binds a unix socket, replacing a stale one. The controller lock
// guarantees no live owner exists for any socket under the root.
func listen(path string) (net.Listener, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		_ = ln.Close()
		return nil, err
	}
	return ln, nil
}

func (c *Controller) serve() error {
	admin, err := listen(c.root.AdminSocket())
	if err != nil {
		return err
	}
	c.start(admin, scope{admin: true, actor: "admin"})
	if err := c.serveMetrics(); err != nil {
		c.cancel()
		c.wg.Wait()
		return err
	}

	open, err := c.store.List(c.ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen})
	if err != nil {
		return err
	}
	for _, t := range open {
		if err := c.openSession(t); err != nil {
			c.log.Warn("session socket", "session", t.ID, "err", err)
		}
		// Adopts a live host, or restarts the agent and resumes it.
		c.drive(t)
	}
	c.reapOrphans(open)
	c.log.Info("controller serving", "root", c.root.Dir, "sessions", len(open))

	tick := time.NewTicker(c.opt.Tick)
	defer tick.Stop()
	for {
		c.sweep()
		c.supervise()
		select {
		case <-c.ctx.Done():
			c.cancel()
			c.wg.Wait()
			_ = os.Remove(c.root.AdminSocket())
			c.log.Info("controller stopped")
			return nil
		case <-tick.C:
		}
	}
}

func (c *Controller) sweep() {
	ids, err := c.store.SweepLeases(c.ctx, "controller")
	if err != nil {
		if c.ctx.Err() == nil {
			c.log.Error("lease sweep", "err", err)
		}
		return
	}
	for _, id := range ids {
		c.log.Info("lease expired", "ticket", id)
	}
	if len(ids) > 0 {
		c.metrics.LeaseExpiries.Add(c.ctx, int64(len(ids)))
	}
}

// serveMetrics serves Prometheus /metrics on telemetry.metrics_listen
// until the controller stops. Empty listen address: off.
func (c *Controller) serveMetrics() error {
	addr := c.cfg.Telemetry.MetricsListen
	if addr == "" {
		return nil
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("metrics listen: %w", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", c.metrics.Handler())
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	c.wg.Go(func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.log.Error("metrics server", "err", err)
		}
	})
	c.wg.Go(func() {
		<-c.ctx.Done()
		_ = srv.Close()
	})
	c.log.Info("serving metrics", "addr", ln.Addr().String())
	if c.opt.MetricsReady != nil {
		c.opt.MetricsReady(ln.Addr().String())
	}
	return nil
}

// start serves ln until the controller stops or the returned func is called.
func (c *Controller) start(ln net.Listener, sc scope) context.CancelFunc {
	ctx, cancel := context.WithCancel(c.ctx)
	c.wg.Go(func() {
		<-ctx.Done()
		_ = ln.Close()
	})
	c.wg.Go(func() {
		var conns sync.WaitGroup
		defer conns.Wait()
		for {
			nc, err := ln.Accept()
			if err != nil {
				if ctx.Err() == nil {
					c.log.Error("accept", "addr", ln.Addr(), "err", err)
					cancel()
				}
				return
			}
			conns.Go(func() {
				gate := rpc.NewGate(nc)
				conn := acp.NewConnection(c.handler(ctx, sc), nc, gate)
				conn.SetLogger(rpc.QuietLogger(c.log))
				gate.Open()
				select {
				case <-conn.Done(): // client hung up
				case <-ctx.Done(): // socket or controller stopping
				}
				_ = nc.Close()
			})
		}
	})
	return cancel
}

func (c *Controller) socketPath(session string) string {
	return filepath.Join(c.root.SessionDir(session), "cli.sock")
}

func (c *Controller) openSession(t store.Ticket) error {
	ln, err := listen(c.socketPath(t.ID))
	if err != nil {
		return err
	}
	sc := scope{actor: t.ID, session: t.ID, project: t.Project, pool: t.Pool}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessions[t.ID] = c.start(ln, sc)
	return nil
}

func (c *Controller) closeSession(id string) {
	c.mu.Lock()
	stop := c.sessions[id]
	delete(c.sessions, id)
	c.mu.Unlock()
	if stop != nil {
		stop()
		_ = os.Remove(c.socketPath(id))
	}
}

// scope is who is calling. Session scopes are fixed by the socket the call
// arrived on; they cannot name another holder.
type scope struct {
	admin   bool
	actor   string
	session string
	project string
	pool    string
}

// handler dispatches one socket's requests. The SDK gives each request its
// own context; it is joined with the socket's so a closing session or
// controller cancels in-flight work too.
func (c *Controller) handler(sockCtx context.Context, sc scope) acp.MethodHandler {
	return func(reqCtx context.Context, name string, params json.RawMessage) (any, *acp.RequestError) {
		ctx, cancel := context.WithCancel(reqCtx)
		defer cancel()
		stop := context.AfterFunc(sockCtx, cancel)
		defer stop()
		h, ok := methods[name]
		if !ok {
			return nil, acp.NewMethodNotFound(name)
		}
		defer c.metrics.Since(ctx, time.Now(), name)
		res, err := h(c, ctx, sc, params)
		if err != nil {
			if !errors.Is(err, store.ErrNoWork) {
				c.log.Debug("rpc error", "method", name, "actor", sc.actor, "err", err)
			}
			return nil, rpc.ToWire(err)
		}
		return res, nil
	}
}

type method func(c *Controller, ctx context.Context, sc scope, params json.RawMessage) (any, error)

// bind decodes params into P before calling fn, a (*Controller) method
// expression.
func bind[P any](fn func(c *Controller, ctx context.Context, sc scope, p P) (any, error)) method {
	return func(c *Controller, ctx context.Context, sc scope, raw json.RawMessage) (any, error) {
		var p P
		if len(raw) > 0 && string(raw) != "null" {
			dec := json.NewDecoder(bytesReader(raw))
			dec.DisallowUnknownFields()
			if err := dec.Decode(&p); err != nil {
				return nil, acp.NewInvalidParams(err.Error())
			}
		}
		return fn(c, ctx, sc, p)
	}
}

var methods map[string]method

func init() {
	methods = map[string]method{
		rpc.MethodWhoami:       bind((*Controller).whoami),
		rpc.MethodTicketCreate: bind((*Controller).create),
		rpc.MethodTicketShow:   bind((*Controller).show),
		rpc.MethodTicketList:   bind((*Controller).list),
		rpc.MethodTicketUpdate: bind((*Controller).update),
		rpc.MethodTicketClose:  bind((*Controller).closeTicket),
		rpc.MethodTicketReopen: bind((*Controller).reopen),
		rpc.MethodReady:        bind((*Controller).ready),
		rpc.MethodCommentAdd:   bind((*Controller).comment),
		rpc.MethodEdgeAdd:      bind((*Controller).edgeAdd),
		rpc.MethodEdgeRemove:   bind((*Controller).edgeRemove),
		rpc.MethodLabelAdd:     bind((*Controller).labelAdd),
		rpc.MethodLabelRemove:  bind((*Controller).labelRemove),
		rpc.MethodClaim:        bind((*Controller).claim),
		rpc.MethodHeartbeat:    bind((*Controller).heartbeat),
		rpc.MethodRelease:      bind((*Controller).release),
		rpc.MethodSessionNew:   bind((*Controller).sessionNew),
		rpc.MethodSessionList:  bind((*Controller).sessionList),
		rpc.MethodSessionReap:  bind((*Controller).sessionReap),
		rpc.MethodShutdown:     bind((*Controller).shutdown),
		rpc.MethodDrain:        bind((*Controller).drainPools),
	}
}
