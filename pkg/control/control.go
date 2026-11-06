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
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	acp "github.com/coder/acp-go-sdk"
	host "github.com/elee1766/acphost/pkg/acphost"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sys/unix"

	"github.com/elee1766/jobshop/pkg/clock"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/identity"
	idgen "github.com/elee1766/jobshop/pkg/ids"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/logctx"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/postgres"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
	"github.com/elee1766/jobshop/pkg/telemetry"
	"github.com/elee1766/jobshop/pkg/writekey"
)

// DefaultLease is the claim lease when the config sets none.
const DefaultLease = 30 * time.Minute

// HostRunner runs session hosts for Options.Hosts. Start begins a host
// for a session directory whose spec is written; it must not wait for the
// host to listen. Stop stops the host for dir and waits for it, reporting
// false for a dir it does not run so the controller signals the pid in
// the dir as usual.
type HostRunner interface {
	Start(dir string) error
	Stop(dir string) bool
}

// HostExits is an optional HostRunner method. A host process that exits
// closes the controller's connection with it; an in-process host's
// accepted connection outlives acphost.Main, so the controller closes
// it when Exited's channel closes, or it never sees the crash (#467).
// Exited returns nil for a dir the runner does not run.
type HostExits interface {
	Exited(dir string) <-chan struct{}
}

// Options tune a controller. Zero values use defaults.
type Options struct {
	Tick time.Duration // lease sweep interval, default 30s
	// Clock is the time source for every controller timer and timestamp.
	// Default the wall clock; tests pass a clock.Fake (#362).
	Clock clock.Clock
	Log   *slog.Logger
	// HostExe is the jobshop binary sessions and command steps see: first
	// on their PATH, JOBSHOP_BIN, bound into sandboxes, and left out of
	// background jobs. Default /proc/self/exe. Session hosts always run
	// this binary (acphost.StartDetached re-execs it, #904).
	HostExe string
	// Hosts, if set, starts and stops session hosts instead of a detached
	// process per session. Tests run hosts in-process with it so a test of
	// controller logic does not pay a process spawn per host (#361).
	Hosts HostRunner
	// Preflight checks an isolated pool's host prerequisites at start and
	// reload (#728). Default acphost.Preflight; tests set a fake.
	Preflight func(host.Spec, ...host.PreflightOption) (host.PreflightReport, error)
	// Backoff is the first wait after a provider rate limit, doubling to
	// 15x. Default 1m; tests shorten it.
	Backoff time.Duration
	// RestartDelay is the pause before restarting an interrupted session.
	// Default 2s.
	RestartDelay time.Duration
	// MessagePoll is how often a running turn checks for interrupt
	// messages. Default 2s.
	MessagePoll time.Duration
	// Poll is how often an idle session looks for ready work. Default 5s.
	Poll time.Duration
	// MetricsReady, if set, receives the /metrics listen address once it
	// is bound (tests use ":0" in the config).
	MetricsReady func(addr string)
	// Tracer makes the controller's spans. The serve command passes the
	// process tracer from telemetry.Start, with Log from the same call.
	// Default: spans with no exporter, so trace context still propagates.
	Tracer trace.Tracer
	// Resource is the process resource from telemetry.Start; /metrics
	// serves it as target_info (JOB-106). Nil: service.name=jobshop only.
	Resource *resource.Resource
	// Reload, if set, triggers a config reload per receive. The serve
	// command sends on SIGHUP; tests send directly.
	Reload <-chan struct{}
	// Pass, if set, runs one supervisor pass per receive and closes the
	// received channel when it is done. Tests use it instead of sleeping
	// through ticks (#359).
	Pass <-chan chan struct{}
	// EventBuffer is the per-subscriber events.subscribe buffer. Default
	// DefaultEventBuffer; tests shrink it to force lagged.
	EventBuffer int
	// Version and Commit identify the running binary for the version
	// method (#180).
	Version, Commit string
	// ShellProbe lists the commands an agent's tool shell must find on
	// its PATH before a host is spawned (JOB-67); a miss quarantines the
	// pool. The serve command passes DefaultShellProbe; nil skips it, as
	// test hosts are not named jobshop.
	ShellProbe []string
}

// Controller serves one Jobshop root.
type Controller struct {
	root  config.Root
	live  *atomic.Pointer[liveConfig] // swapped by reload
	store store.Store
	opt   Options
	log   *slog.Logger
	// incidents keeps recent WARN/ERROR lines for report.incidents.
	incidents *incidentLog

	metrics  *telemetry.Metrics
	tracer   trace.Tracer
	started  time.Time // base for cron triggers that never fired
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	sessions map[string]context.CancelFunc // open session id -> stops its cli.sock
	runners  map[string]*runner
	// readOnly are sessions limited to read methods until a human
	// writes (#606); mirrors SessionFields.ReadOnly.
	readOnly map[string]bool
	// watchers counts Desk streams per session: a watched question chat
	// stays open (JOB-52).
	watchers map[string]int
	drain    bool            // full shutdown in progress: start nothing new
	drained  map[string]bool // drained pools (drainAll for every pool); persisted in scale.json
	// Pool scaling (#281): manual targets (persisted with drains), the
	// stabilization samples and each pool's last desired count.
	scaleTargets map[string]int
	scaleHist    map[string][]scaleSample
	scaleViews   map[string]poolScaleView
	// Pool crash-loop quarantine, memory only (see quarantine.go).
	poolCrashLog    map[string][]time.Time
	poolPausedUntil map[string]time.Time
	// poolPauseReason is the last session error and failed preflight
	// checks behind a crash-loop pause (JOB-27); nil until first set.
	poolPauseReason map[string]string
	// inboxPrunedAt is the last inbox prune (#321); tick-only state.
	inboxPrunedAt time.Time
	// triggerPolled is each trigger's last poll, for fallback: (JOB-62);
	// tick-only state.
	triggerPolled map[string]time.Time
	// artifactsPrunedAt is the last attachment prune (#508); tick-only state.
	artifactsPrunedAt time.Time
	// journalsCompressedAt is the last journal compression sweep start
	// (#631); journalCompressing is set while one runs.
	journalsCompressedAt time.Time
	journalCompressing   atomic.Bool
	// prov is resource provider demand and idle state (#407); tick-only.
	prov providerState
	// writeKeysPrunedAt is the last write key prune (#415); tick-only.
	writeKeysPrunedAt time.Time
	// poolQuarantined maps a pool to the permanent spawn error that
	// stopped it; cleared by reload or pool unquarantine (#146).
	poolQuarantined map[string]string
	// dnsChecked holds the pool and DNS settings whose sandbox DNS check
	// passed (JOB-20); guarded by mu.
	dnsChecked map[string]bool
	// ingestWorkers is each external ingest worker's last heartbeat (#443).
	ingestWorkers map[string]*rpc.IngestWorker
	// configErr is the pack load error the controller started past,
	// serving the last good config (#227); cleared by a good reload.
	configErr   string
	configErrAt time.Time
	// watchErr is the last failed watch reload (#533), cleared by a good
	// reload.
	watchErr   string
	watchErrAt time.Time
	// configSwapped tells the config watcher a reload swapped the config.
	configSwapped chan struct{}
	// watcher is the config watcher (#533); nil until Run starts it.
	watcher *configWatcher
	// orphanWarned maps orphaned ticket id -> reason already logged.
	orphanMu     sync.Mutex
	orphanWarned map[string]string
	// sessOrphans is the last supervisor pass's live hosts with no open
	// session ticket (JOB-32); endedFilled is set once every closed
	// session has its Ended state.
	sessScanMu  sync.Mutex
	sessOrphans []string
	sessScanned bool
	endedFilled bool
	// gitMu serializes git commands on pool repos: concurrent fetches in
	// one repo fail on the remote-ref lock.
	gitMu sync.Mutex
	// memMu serializes local memory writes and restores (#592), so a
	// write's before text is the file it replaced.
	memMu sync.Mutex
	// cgroup is the delegated sessions cgroup for isolation limits (#388),
	// set up once on first need.
	cgroupOnce sync.Once
	cgroup     cgroupSetup
	// hostsMoved is set once live hosts were moved out of the controller's
	// cgroup into the sessions cgroup (JOB-91).
	hostsMoved atomic.Bool
	// slotMu makes picking a per-session id and creating the session one
	// step (#540).
	slotMu sync.Mutex
	// minter caches identity tokens (#200).
	minter *identity.Minter
	// selfLoginErr is the last self-login error logged per identity, so a
	// refused lookup warns once, not on every poll (JOB-108).
	selfLoginErr sync.Map
	// ingest is the spool drain (#335).
	ingest ingestDrain
	// reconcileAt is the last GitHub App delivery reconcile (#335).
	reconcileAt time.Time
	// linearURL is the Linear GraphQL endpoint for identity check; "" is
	// linear.DefaultAPIURL (tests set a fake).
	linearURL string
	// ghIssueOpen, when set, replaces the GitHub issue state read of the
	// intake dependency check (JOB-68; tests).
	ghIssueOpen func(context.Context, linear.GitHubDep) (bool, error)
	// linIntake caches linear_issue_no_run reads between polls (#287).
	linIntake linearIntake
	// linSessions throttles the Linear agent-session sweep.
	linSessions linearSessions
	// reloadMu serializes reloads so the last file read wins.
	reloadMu sync.Mutex
	// tickMu serializes supervisor passes: the tick and the event paths
	// that run part of one (trigger events, a Linear issue webhook). A
	// pass reads the config once at its start (JOB-104), so reload does
	// not take it.
	tickMu sync.Mutex
	// holder is who holds tickMu, for the lock wait spans (tracing.go).
	// lockSpanMin is the shortest wait that gets a span; zero is
	// lockSpanThreshold, tests lower it.
	holderMu    sync.Mutex
	holder      lockHolder
	lockSpanMin time.Duration
	// lockWaitHook, if set, runs after a lock waiter has read the holder
	// and before it blocks; tests use it to know the waiter is committed.
	lockWaitHook func()
	// runMu serializes run start (key dedup) and run state changes.
	runMu sync.Mutex
	// autoMu guards auto, the command and http step attempts running in
	// this process (#396).
	autoMu sync.Mutex
	auto   autoSteps
	// events fans store changes out to events.subscribe.
	events *hub
	// subSource reads GitHub threads; nil means gh (tests set a fake).
	subSource threadSource
	// subPollMu serializes subscription passes: the pump and webhook
	// passes for one thread (#300).
	subPollMu sync.Mutex
	// dispatchMu serializes inbox dispatch passes (#442): the pump and
	// poll-triggered passes, so a thread's news is never delivered out
	// of order.
	dispatchMu sync.Mutex
	// waitSource reads PR state for waits; nil means gh (#381). waitWake
	// asks for a pass now; waitMu serializes passes.
	waitSource waitSource
	// linearWait reads Linear issues for linear.* waits; nil means live.
	linearWait linearWaitSource
	waitWake   chan struct{}
	waitMu     sync.Mutex
	// waitPending is what a wake asked of the wait pump (#679).
	waitWakeMu  sync.Mutex
	waitPending waitWakes
	// inbox is the inbox change signal (#679); see inboxNotifier.
	inbox     *ingest.Notifier
	inboxOnce sync.Once
	// activitySrc reads GitHub and Linear activity; nil means live APIs
	// (tests set a fake). activity is the cache, loaded on first poll.
	activitySrc activitySource
	activityMu  sync.Mutex
	activity    *activityState
	// ship is journal shipping to VictoriaLogs (#530).
	ship shipper
	wg   sync.WaitGroup
	// Trigger checks run outside the tick when backgroundChecks is set, by
	// Run (JOB-109, triggercheck.go). checkJobs is under tickMu.
	backgroundChecks bool
	checkJobs        checkJobs
	checkSlots       chan struct{}
	checkSlotsOnce   sync.Once
}

// Run opens the root, serves until ctx ends or a shutdown request, and
// cleans up. It holds an exclusive lock so one root has one controller.
func Run(ctx context.Context, root config.Root, opt Options) error {
	// A live pack that needs a newer binary must not crash-loop the
	// controller: serve the last good config and report it (#227).
	cfg, loadErr := config.LoadOrLastGood(root)
	if cfg == nil {
		return loadErr
	}
	lc, err := newLiveConfig(cfg)
	if err != nil {
		return err
	}
	live := new(atomic.Pointer[liveConfig])
	live.Store(lc)
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
	if opt.Poll <= 0 {
		opt.Poll = 5 * time.Second
	}
	if opt.Log == nil {
		// Libraries never log to the process default (JOB-49).
		opt.Log = slog.New(slog.DiscardHandler)
	}
	if opt.Clock == nil {
		opt.Clock = clock.Real{}
	}
	unlock, err := lock(root.Lock())
	if err != nil {
		return err
	}
	defer unlock()

	ev := newHub(opt.EventBuffer)
	st, err := openStore(ctx, root, cfg.Store, storeOptions{Now: opt.Clock.Now, OnCommit: ev.kick, Prefix: func(p string) string {
		// Follows reloads, so a project added later gets its prefix.
		return live.Load().cfg.Prefix(p)
	}, IDs: func(p string) idgen.Spec {
		return live.Load().cfg.IDSpec(p)
	}})
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()
	// Calls over store.SlowCall add a span under the step that made them
	// (JOB-105).
	st = store.Traced(st)
	// Resource caps gate claims (#400); reload re-syncs them.
	if err := st.SetResources(ctx, store.ResourceConfig{Caps: cfg.ResourceCaps(), Providers: cfg.ResourceProviders()}); err != nil {
		return err
	}

	m, err := telemetry.NewMetricsFor(opt.Resource)
	if err != nil {
		return err
	}
	defer func() { _ = m.Shutdown(context.WithoutCancel(ctx)) }()

	if opt.Tracer == nil {
		tr, err := telemetry.NewTracing(ctx, telemetry.Export{})
		if err != nil {
			return err
		}
		defer func() { _ = tr.Shutdown(context.WithoutCancel(ctx)) }()
		opt.Tracer = tr.Tracer()
	}
	incidents := &incidentLog{}
	// logctx outermost: the caller's handler (with the OTel bridge, see
	// telemetry.Start) and the incident log get the ids in ctx (JOB-49).
	log := slog.New(logctx.NewHandler(incidentHandler{Handler: logctx.Unwrap(opt.Log.Handler()), log: incidents}))

	c := &Controller{
		root: root, live: live, store: st, opt: opt, log: log, metrics: m, incidents: incidents,
		tracer:   opt.Tracer,
		sessions: map[string]context.CancelFunc{}, runners: map[string]*runner{}, readOnly: map[string]bool{},
		minter:       identity.NewMinter(),
		ingest:       ingestDrain{wake: make(chan struct{}, 1)},
		waitWake:     make(chan struct{}, 1),
		drained:      map[string]bool{},
		scaleTargets: map[string]int{}, scaleHist: map[string][]scaleSample{}, scaleViews: map[string]poolScaleView{},
		poolCrashLog: map[string][]time.Time{}, poolPausedUntil: map[string]time.Time{},
		poolQuarantined: map[string]string{},
		events:          ev,
		configSwapped:   make(chan struct{}, 1),
	}
	// The stream starts at the current end of history; older rows are
	// only sent as replay.
	if ev.last, err = store.LastSeq(ctx, st); err != nil {
		return err
	}
	// Metrics ride on the ctx so outbound HTTP below any path records its
	// timing without a parameter (JOB-105).
	c.backgroundChecks = true
	c.ctx, c.cancel = context.WithCancel(telemetry.WithMetrics(ctx, m))
	c.started = c.now()
	if loadErr != nil {
		c.configErr, c.configErrAt = loadErr.Error(), c.started
		c.log.ErrorContext(ctx, "config load failed in a pack, serving the last good config", "err", loadErr, "last_good", root.LastGood())
	} else {
		c.saveLastGood(cfg)
	}
	c.loadScale()
	if wantsCgroup(cfg) {
		// Before any session spawns, so the controller is in its leaf
		// cgroup when it enables controllers for sessions (#388).
		c.ensureCgroup()
		c.tuneCgroup(cfg)
	}
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
	// Unlock before close (#601): a flock belongs to the open file
	// description, and a child forked by another goroutine holds a copy
	// until its exec closes it, so close alone can leave the lock held and
	// a restarted controller in the same process fails to take it.
	return func() {
		_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
		_ = f.Close()
	}, nil
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
	c.watcher = newConfigWatcher(c)
	c.wg.Go(c.watcher.run)
	c.wg.Go(c.pumpEvents)
	c.wg.Go(c.pumpSubscriptions)
	c.wg.Go(c.pumpWaits)
	c.wg.Go(c.pumpActivity)
	c.wg.Go(c.pumpDrain)
	c.wg.Go(c.pumpReconcile)
	c.wg.Go(c.pumpPollers)
	c.wg.Go(c.pumpAdmission)
	c.wg.Go(c.pumpJournal)
	c.start(admin, scope{admin: true, actor: "admin"})
	if err := c.serveMetrics(); err != nil {
		c.cancel()
		c.wg.Wait()
		return err
	}
	if err := c.serveLinearWebhook(); err != nil {
		c.cancel()
		c.wg.Wait()
		return err
	}

	open, err := c.store.List(c.ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen})
	if err != nil {
		// Canceled during startup is a clean stop, not a failure (#61).
		// Checked before stop, which cancels c.ctx itself.
		canceled := c.ctx.Err() != nil
		c.stop()
		if canceled {
			return nil
		}
		return err
	}
	startCfg := c.config()
	for _, t := range open {
		if f, err := store.DecodeFields[store.SessionFields](t); err == nil && f.ReadOnly {
			c.setReadOnly(t.ID, true)
		}
		if err := c.openSession(t); err != nil {
			c.log.Warn("session socket", "session", t.ID, "err", err)
		}
		// Adopts a live host, or restarts the agent and resumes it.
		c.drive(startCfg, t)
	}
	c.reapOrphans(open)
	// After adoption, before the first pass spawns anything (#728).
	c.preflightPools(c.ctx)
	if err := c.backfillGitHubRepo(c.ctx); err != nil {
		c.log.Warn("backfill github repo", "err", err)
	}
	c.log.Info("controller serving", "root", c.root.Dir, "sessions", len(open))

	tick := c.opt.Clock.NewTicker(c.opt.Tick)
	defer tick.Stop()
	var passDone chan struct{}
	for {
		c.tick()
		if passDone != nil {
			close(passDone)
			passDone = nil
		}
		select {
		case <-c.ctx.Done():
			c.stop()
			return nil
		case <-tick.C():
		case <-c.opt.Reload:
			// Errors are logged and the old config stays active. The
			// tick below applies new pool limits right away.
			_ = c.reload()
		case passDone = <-c.opt.Pass:
		}
	}
}

// stop cancels the controller, waits for its goroutines and removes the
// admin socket.
func (c *Controller) stop() {
	c.cancel()
	c.wg.Wait()
	c.closeSpool()
	_ = os.Remove(c.root.AdminSocket())
	c.log.Info("controller stopped")
}

// tick runs one supervisor pass under a span, with one child span and one
// jobshop.tick.step.duration sample per step, so a slow pass names its
// slow step (JOB-105). Steps get the step span's ctx: the store, HTTP and
// exec spans below nest under it.
func (c *Controller) tick() {
	ctx, span := c.tracer.Start(c.ctx, "supervisor tick")
	defer span.End()
	defer c.lockTick(ctx, "tick")()
	// One config for the whole pass (JOB-104): a reload swaps the pointer
	// without waiting, and applies from the next pass. Pass code takes
	// cfg, never c.config() (TestPassReadsConfigOnce).
	lc := c.live.Load()
	cfg := lc.cfg
	c.passStep(ctx, cfg, "sweep", c.sweep)
	// Right after the sweep, so slots of expired claims reset this pass.
	c.passStep(ctx, cfg, "provider_pass", c.providerPass)
	// Before supervise, so a created ticket can get a session this tick.
	if !c.draining() {
		c.passStep(ctx, cfg, "run_triggers", c.runTriggers)
	}
	// Placement moves steps between pools before sessions are sized (#409).
	c.passStep(ctx, cfg, "place_steps", c.placeSteps)
	c.step(ctx, "supervise", func(ctx context.Context) { c.supervise(ctx, cfg, lc.lease) })
	c.passStep(ctx, cfg, "run_auto_steps", c.runAutoSteps)
	c.passStep(ctx, cfg, "report_orphans", c.reportOrphans)
	c.step(ctx, "scan_sessions", c.scanSessions)
	c.passStep(ctx, cfg, "sweep_runs", c.sweepRuns)
	c.passStep(ctx, cfg, "sweep_escalation_chats", c.sweepEscalationChats)
	c.passStep(ctx, cfg, "sweep_attention_chats", c.sweepAttentionChats)
	c.step(ctx, "sweep_batches", c.sweepBatches)
	c.passStep(ctx, cfg, "sweep_linear_asks", c.sweepLinearAsks)
	c.step(ctx, "sweep_linear_agent_sessions", func(ctx context.Context) { c.sweepLinearAgentSessions(ctx, cfg, false) })
	c.passStep(ctx, cfg, "prune_inbox", c.pruneInbox)
	c.passStep(ctx, cfg, "prune_write_keys", c.pruneWriteKeys)
	c.passStep(ctx, cfg, "prune_artifacts", c.pruneArtifacts)
	c.passStep(ctx, cfg, "compress_journals", c.maybeCompressJournals)
}

// inboxPruneEvery spaces inbox prunes; retention is days, so the 50ms
// tick would only add writes.
const inboxPruneEvery = time.Hour

// pruneInbox deletes terminal inbox items older than ingest.retention.
func (c *Controller) pruneInbox(ctx context.Context, cfg *config.Config) {
	now := c.now()
	if now.Sub(c.inboxPrunedAt) < inboxPruneEvery {
		return
	}
	c.inboxPrunedAt = now
	n, err := ingest.InboxPrune(ctx, c.store, now.Add(-cfg.Ingest.RetentionDuration()))
	if err != nil {
		if ctx.Err() == nil {
			c.log.ErrorContext(ctx, "inbox prune", "err", err)
		}
		return
	}
	if n > 0 {
		c.log.InfoContext(ctx, "inbox pruned", "items", n)
	}
}

// writeKeyPruneEvery spaces write key prunes; retention is hours.
const writeKeyPruneEvery = 10 * time.Minute

// pruneWriteKeys deletes write keys older than write_keys.retention,
// done or pending (#415). A pending key's retry then checks the remote
// key or marker instead of the local row.
func (c *Controller) pruneWriteKeys(ctx context.Context, cfg *config.Config) {
	now := c.now()
	if now.Sub(c.writeKeysPrunedAt) < writeKeyPruneEvery {
		return
	}
	c.writeKeysPrunedAt = now
	n, err := writekey.Prune(ctx, c.store, now.Add(-cfg.WriteKeys.RetentionDuration()))
	if err != nil {
		if ctx.Err() == nil {
			c.log.ErrorContext(ctx, "write key prune", "err", err)
		}
		return
	}
	if n > 0 {
		c.log.InfoContext(ctx, "write keys pruned", "keys", n)
	}
}

func (c *Controller) sweep(ctx context.Context, cfg *config.Config) {
	ids, err := c.store.Expire(ctx, "controller", "", "")
	if err != nil {
		if ctx.Err() == nil {
			c.log.ErrorContext(ctx, "lease sweep", "err", err)
		}
		return
	}
	for _, id := range ids {
		c.log.InfoContext(ctx, "lease expired", "ticket", id)
	}
	c.stepEvents(ctx, cfg, ids)
	if len(ids) > 0 {
		c.metrics.LeaseExpiries.Add(ctx, int64(len(ids)))
	}
}

// serveMetrics serves Prometheus /metrics on telemetry.metrics_listen
// until the controller stops. Empty listen address: off.
func (c *Controller) serveMetrics() error {
	addr := c.config().Telemetry.MetricsListen
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
				// The handler reaches its connection to send subscribe events;
				// p.conn is set before the gate opens, so before any request.
				csc := sc
				csc.peer = &peer{}
				conn := acp.NewConnection(c.handler(ctx, csc), nc, gate)
				csc.peer.conn = conn
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

// openSession binds a session's cli.sock once. The admin socket serves
// before startup reopens the open sessions, so a session created in that
// window reaches here twice; binding under c.mu and skipping an open
// session keeps the second call from removing or rebinding the live
// socket (#479).
func (c *Controller) openSession(t store.Ticket) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.sessions[t.ID]; ok {
		return nil
	}
	ln, err := listen(c.socketPath(t.ID))
	if err != nil {
		return err
	}
	sc := scope{actor: t.ID, session: t.ID, project: t.Project, pool: t.Pool}
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
	peer    *peer
}

// peer is the connection a request arrived on.
type peer struct{ conn *acp.Connection }

// handler dispatches one socket's requests. The SDK gives each request its
// own context; it is joined with the socket's so a closing session or
// controller cancels in-flight work too.
func (c *Controller) handler(sockCtx context.Context, sockScope scope) acp.MethodHandler {
	return func(reqCtx context.Context, name string, params json.RawMessage) (any, *acp.RequestError) {
		// Per call: _meta.actor must not leak into later calls on the socket.
		sc := sockScope
		ctx, cancel := context.WithCancel(reqCtx)
		defer cancel()
		stop := context.AfterFunc(sockCtx, cancel)
		defer stop()
		h, ok := methods[name]
		if !ok {
			return nil, acp.NewMethodNotFound(name)
		}
		if err := c.checkReadOnly(sc, name); err != nil {
			return nil, rpc.ToWire(err)
		}
		params, meta := rpc.SplitMeta(params)
		if meta.Actor != "" {
			// Only the admin socket may name who is behind a call (#515).
			if !sc.admin {
				return nil, rpc.ToWire(fmt.Errorf("%w: only the admin socket can set _meta.actor", rpc.ErrForbidden))
			}
			if err := rpc.ValidActor(meta.Actor); err != nil {
				return nil, rpc.ToWire(err)
			}
			sc.actor = meta.Actor
		}
		ctx = telemetry.WithTraceParent(telemetry.WithMetrics(ctx, c.metrics), meta.TraceParent)
		ctx, span := c.tracer.Start(ctx, "rpc "+name, trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(telemetry.AttrMethod.String(name), attribute.String("jobshop.actor", sc.actor)))
		defer span.End()
		if sc.session != "" {
			span.SetAttributes(telemetry.AttrSession.String(sc.session), telemetry.AttrPool.String(sc.pool),
				telemetry.AttrProject.String(sc.project))
		}
		defer c.metrics.Since(ctx, time.Now(), name)
		res, err := h.call(c, ctx, sc, params)
		if err != nil {
			if !errors.Is(err, store.ErrNoWork) {
				span.SetStatus(codes.Error, err.Error())
				c.log.DebugContext(ctx, "rpc error", "method", name, "actor", sc.actor, "err", err)
			}
			return nil, rpc.ToWire(err)
		}
		return res, nil
	}
}

// method is one handler. params is its parameter type, which a test
// checks against rpc.Specs (#542).
type method struct {
	call   func(c *Controller, ctx context.Context, sc scope, params json.RawMessage) (any, error)
	params reflect.Type
}

// bind decodes params into P before calling fn, a (*Controller) method
// expression.
func bind[P any](fn func(c *Controller, ctx context.Context, sc scope, p P) (any, error)) method {
	return method{params: reflect.TypeFor[P](), call: func(c *Controller, ctx context.Context, sc scope, raw json.RawMessage) (any, error) {
		var p P
		if len(raw) > 0 && string(raw) != "null" {
			dec := json.NewDecoder(bytesReader(raw))
			dec.DisallowUnknownFields()
			if err := dec.Decode(&p); err != nil {
				return nil, acp.NewInvalidParams(err.Error())
			}
		}
		return fn(c, ctx, sc, p)
	}}
}

var methods map[string]method

func init() {
	methods = map[string]method{
		rpc.MethodWhoami:            bind((*Controller).whoami),
		rpc.MethodVersion:           bind((*Controller).version),
		rpc.MethodIdentityList:      bind((*Controller).identityList),
		rpc.MethodIdentityCheck:     bind((*Controller).identityCheck),
		rpc.MethodIdentityToken:     bind((*Controller).identityToken),
		rpc.MethodIngestDrain:       bind((*Controller).ingestDrainMethod),
		rpc.MethodIngestClaim:       bind((*Controller).ingestClaim),
		rpc.MethodIngestAdmit:       bind((*Controller).ingestAdmit),
		rpc.MethodIngestWorkers:     bind((*Controller).ingestWorkersMethod),
		rpc.MethodIngestDispatch:    bind((*Controller).ingestDispatchMethod),
		rpc.MethodIngestNext:        bind((*Controller).ingestNext),
		rpc.MethodIngestNotify:      bind((*Controller).ingestNotify),
		rpc.MethodTicketCreate:      bind((*Controller).create),
		rpc.MethodTicketShow:        bind((*Controller).show),
		rpc.MethodTicketTrace:       bind((*Controller).trace),
		rpc.MethodRecall:            bind((*Controller).recall),
		rpc.MethodTicketList:        bind((*Controller).list),
		rpc.MethodTicketUpdate:      bind((*Controller).update),
		rpc.MethodTicketClose:       bind((*Controller).closeTicket),
		rpc.MethodTicketReopen:      bind((*Controller).reopen),
		rpc.MethodTicketUpsert:      bind((*Controller).upsert),
		rpc.MethodReady:             bind((*Controller).ready),
		rpc.MethodCommentAdd:        bind((*Controller).comment),
		rpc.MethodEdgeAdd:           bind((*Controller).edgeAdd),
		rpc.MethodEdgeRemove:        bind((*Controller).edgeRemove),
		rpc.MethodLabelAdd:          bind((*Controller).labelAdd),
		rpc.MethodLabelRemove:       bind((*Controller).labelRemove),
		rpc.MethodClaim:             bind((*Controller).claim),
		rpc.MethodHeartbeat:         bind((*Controller).heartbeat),
		rpc.MethodRelease:           bind((*Controller).release),
		rpc.MethodHandoff:           bind((*Controller).handoff),
		rpc.MethodSessionNew:        bind((*Controller).sessionNew),
		rpc.MethodSessionList:       bind((*Controller).sessionList),
		rpc.MethodSessionReap:       bind((*Controller).sessionReap),
		rpc.MethodSessionRestart:    bind((*Controller).sessionRestart),
		rpc.MethodSessionPrompt:     bind((*Controller).sessionPrompt),
		rpc.MethodSessionSandbox:    bind((*Controller).sessionSandbox),
		rpc.MethodSessionSend:       bind((*Controller).sessionSend),
		rpc.MethodSessionSubscribe:  bind((*Controller).sessionSubscribe),
		rpc.MethodSessionJournal:    bind((*Controller).sessionJournal),
		rpc.MethodSessionCancel:     bind((*Controller).sessionCancel),
		rpc.MethodJournalSearch:     bind((*Controller).journalSearch),
		rpc.MethodJournalBackfill:   bind((*Controller).journalBackfill),
		rpc.MethodEventsSubscribe:   bind((*Controller).eventsSubscribe),
		rpc.MethodShutdown:          bind((*Controller).shutdown),
		rpc.MethodDrain:             bind((*Controller).drainPools),
		rpc.MethodTriggerList:       bind((*Controller).triggerList),
		rpc.MethodReload:            bind((*Controller).reloadConfig),
		rpc.MethodConfigShow:        bind((*Controller).configShow),
		rpc.MethodRunStart:          bind((*Controller).runStart),
		rpc.MethodRunShow:           bind((*Controller).runShow),
		rpc.MethodRunList:           bind((*Controller).runList),
		rpc.MethodRunCancel:         bind((*Controller).runCancel),
		rpc.MethodRunResume:         bind((*Controller).runResume),
		rpc.MethodBatchStart:        bind((*Controller).batchStart),
		rpc.MethodBatchShow:         bind((*Controller).batchShow),
		rpc.MethodBatchList:         bind((*Controller).batchList),
		rpc.MethodPoolList:          bind((*Controller).poolList),
		rpc.MethodPoolScale:         bind((*Controller).poolScaleRPC),
		rpc.MethodPoolShow:          bind((*Controller).poolShow),
		rpc.MethodSessionCordon:     bind((*Controller).sessionCordonRPC),
		rpc.MethodSessionConfig:     bind((*Controller).sessionConfigRPC),
		rpc.MethodSessionRename:     bind((*Controller).sessionRenameRPC),
		rpc.MethodAttention:         bind((*Controller).attention),
		rpc.MethodAttentionChat:     bind((*Controller).attentionChat),
		rpc.MethodPoolUnquarantine:  bind((*Controller).poolUnquarantine),
		rpc.MethodReportIncidents:   bind((*Controller).reportIncidents),
		rpc.MethodGuard:             bind((*Controller).guard),
		rpc.MethodSecretScan:        bind((*Controller).secretScan),
		rpc.MethodSubscribe:         bind((*Controller).subscribe),
		rpc.MethodReport:            bind((*Controller).report),
		rpc.MethodQuestionAnswer:    bind((*Controller).answerQuestion),
		rpc.MethodArtifactDeliver:   bind((*Controller).artifactDeliver),
		rpc.MethodArtifactList:      bind((*Controller).artifactList),
		rpc.MethodArtifactShow:      bind((*Controller).artifactShow),
		rpc.MethodArtifactAck:       bind((*Controller).artifactAck),
		rpc.MethodActivityList:      bind((*Controller).activityList),
		rpc.MethodActivityRecord:    bind((*Controller).activityRecord),
		rpc.MethodLearnAdd:          bind((*Controller).learnAdd),
		rpc.MethodLearnList:         bind((*Controller).learnList),
		rpc.MethodLearnUpdate:       bind((*Controller).learnUpdate),
		rpc.MethodLearnEvaluate:     bind((*Controller).learnEvaluate),
		rpc.MethodLearnEdits:        bind((*Controller).learnEdits),
		rpc.MethodMemoryWrite:       bind((*Controller).memoryWrite),
		rpc.MethodMemoryShow:        bind((*Controller).memoryShow),
		rpc.MethodResourceAdd:       bind((*Controller).resourceAdd),
		rpc.MethodResourceRemove:    bind((*Controller).resourceRemove),
		rpc.MethodResourceUnhealthy: bind((*Controller).resourceUnhealthy),
		rpc.MethodResourceList:      bind((*Controller).resourceList),
		rpc.MethodWriteKeyClaim:     bind((*Controller).writeKeyClaim),
		rpc.MethodWriteKeyFinish:    bind((*Controller).writeKeyFinish),
		rpc.MethodWaitAdd:           bind((*Controller).waitAdd),
		rpc.MethodWaitList:          bind((*Controller).waitList),
		rpc.MethodWaitClear:         bind((*Controller).waitClear),
		rpc.MethodDecisionResolve:   bind((*Controller).decisionResolve),
		rpc.MethodTicketUndefer:     bind((*Controller).ticketUndefer),
		rpc.MethodDeskExec:          bind((*Controller).deskExec),
	}
}

// storeOptions are the options both stores take.
type storeOptions struct {
	Now      func() time.Time
	Prefix   func(string) string
	IDs      func(string) idgen.Spec
	OnCommit func()
}

// openStore opens the configured ticket store (#493): SQLite at
// root.DB() unless store.driver is postgres.
func openStore(ctx context.Context, root config.Root, sc config.Store, opt storeOptions) (store.Store, error) {
	if sc.DriverName() != config.StorePostgres {
		return sqlite.Open(ctx, root.DB(), sqlite.Options{Now: opt.Now, Prefix: opt.Prefix, IDs: opt.IDs, OnCommit: opt.OnCommit})
	}
	dsn, err := sc.DSN(ctx, root.Dir)
	if err != nil {
		return nil, err
	}
	return postgres.Open(ctx, dsn, postgres.Options{Now: opt.Now, Prefix: opt.Prefix, IDs: opt.IDs, OnCommit: opt.OnCommit})
}
