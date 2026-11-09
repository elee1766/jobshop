// Package triggers evaluates triggers and scans (JOB-145), split out of
// pkg/control. It decides when a trigger's conditions hold, which
// findings a scan trigger may fire, and runs the slow parts (check
// commands, provider reads) outside the supervisor pass.
//
// The controller keeps everything that writes: starting runs, filing
// tickets, running trigger commands and recording learnings stay behind
// Host (Fire, FireFinding, RecordFinding), so the controller is still the
// only owner of the run and ticket graph. Evaluation reads the store and
// keeps a little runtime state of its own: the last poll of each trigger,
// the background check jobs and the script signal cache. Durable state
// (what a trigger filed, its keys, cooldowns, finding history) lives in
// the store as meta.trigger on the tickets and runs a trigger created, so
// a restart or a replay finds it again and files nothing twice.
package triggers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/cron"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/knowledge"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/support/telemetry"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// Pack is the meta pack on tickets a trigger created:
// {"trigger": {"name": "<trigger>"}}. Dedup and cooldown read it back, so
// trigger state lives in the store and survives restarts.
const Pack = "trigger"

// CommandRun is what a check command did.
type CommandRun struct {
	Stdout, Stderr string
	// Exit is "code N" text of a non-zero exit or a signal, "" when the
	// command exited 0.
	Exit string
}

// Host is what the engine needs from the controller. Methods that write
// (Fire, FireFinding, RecordFinding) and the signals other subsystems own
// (pool pauses, ingest, Linear, resources, scripts) stay in the controller.
type Host interface {
	// Config is the active config. Pass code takes the pass's config
	// instead (JOB-104): only the entry points that start a new pass (an
	// event job, a check replay) call it.
	Config() *config.Config
	Now() time.Time
	// Started is when the controller started, the base of a cron slot for a
	// trigger that never fired.
	Started() time.Time
	Log() *slog.Logger
	Store() store.Store
	// Background reports whether checks run outside the tick (set by Run;
	// a controller built by hand keeps them inline).
	Background() bool

	// Context is the controller's, cancelled on stop. Go runs fn in a
	// goroutine the controller joins on shutdown.
	Context() context.Context
	Go(fn func())
	Draining() bool
	// LockTick takes the tick lock for who ("event", "check") working for
	// trigger, and returns the unlock func.
	LockTick(ctx context.Context, who, trigger string) func()
	// RootSpan starts a root span for work in its own goroutine.
	RootSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span)
	// TimedOut logs and counts an off-pass call that ran out its timeout.
	TimedOut(poller string, timeout time.Duration, args ...any)

	// RunCheck runs command with sh in the root dir, isolated like any
	// controller command. A non-zero exit is not an error.
	RunCheck(ctx context.Context, command string) (CommandRun, error)

	// Hooks is the callback engine a trigger with a handler enqueues to
	// (pkg/hooks owns execution); KickHooks wakes its worker.
	Hooks() Callbacks
	KickHooks()
	// EditsHold is the edits condition (memory edits due for evaluation).
	EditsHold(ctx context.Context, w config.EditsWhen, now time.Time) (string, error)

	// Fire files what tr says for a fire: a ticket, a run or a command. ev
	// is the event that fired it, nil for a poll.
	Fire(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, key string, ev Event) (domain.Ticket, error)
	// FireFinding starts the workflow of a scan trigger for one finding.
	// existing is true when an open run already had its key.
	FireFinding(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, s Finding) (run domain.Ticket, existing bool, err error)
	// RecordFinding records the learning of a fired finding.
	RecordFinding(ctx context.Context, s Finding, run string)

	Signals
}

// Callbacks is the part of pkg/hooks the engine uses: the durable queue a
// handler trigger enqueues to, and the last accepted fire for cooldowns.
type Callbacks interface {
	Put(ctx context.Context, key string, in hooks.Invocation) error
	LastFire(ctx context.Context, name string) (time.Time, error)
}

// Engine evaluates triggers for one controller. Its poll, check job and
// script cache state is tick-only (under the tick lock) except where
// noted.
type Engine struct {
	host Host

	// polled is each trigger's last poll, for fallback: (JOB-62);
	// tick-only state.
	polled map[string]time.Time

	// jobs is the background check state, under the tick lock.
	jobs checkJobs
	// slots bounds the check commands running at once.
	slots     chan struct{}
	slotsOnce sync.Once

	scripts scriptCache
}

// New returns an engine with nothing polled.
func New(h Host) *Engine { return &Engine{host: h} }

func (e *Engine) log() *slog.Logger { return e.host.Log() }

// State is what one evaluation found.
type State struct {
	Holds    bool
	Why      string    // first condition that does not hold
	Open     string    // open ticket the trigger created, if any
	Last     time.Time // created_at of its newest ticket
	NextCron time.Time
	Key      string // check output, stored as meta.trigger.key
}

// Run fires every trigger whose conditions hold. Called each tick.
func (e *Engine) Run(ctx context.Context, cfg *config.Config) {
	for _, name := range slices.Sorted(maps.Keys(cfg.Triggers)) {
		tr := cfg.Triggers[name]
		if e.host.Background() && e.jobs.running[name] {
			continue // its check is still running; poll again next tick
		}
		if !e.pollDue(name, tr) {
			continue
		}
		e.RunChecked(ctx, cfg, name, false, func(ctx context.Context, cfg *config.Config) {
			// Looked up again: a replay after the check runs on the then
			// current config, where the trigger may be gone.
			if tr, ok := cfg.Triggers[name]; ok {
				e.RunTrigger(ctx, cfg, name, tr)
			}
		})
	}
}

// RunTrigger evaluates one trigger that is due, under a span naming it and
// the result: fired, held (conditions do not hold), finding or error
// (JOB-105).
func (e *Engine) RunTrigger(ctx context.Context, cfg *config.Config, name string, tr config.Trigger) {
	ctx, span := telemetry.StartSpan(ctx, "trigger "+name, attribute.String("jobshop.trigger", name))
	result := "held"
	var runErr error
	defer func() {
		span.SetAttributes(attribute.String("jobshop.trigger.result", result))
		telemetry.End(span, runErr)
	}()
	if len(tr.When.Scan) > 0 {
		e.runScan(ctx, cfg, name, tr, nil)
		result = "scan"
		return
	}
	st, err := e.Eval(ctx, cfg, name, tr)
	if errors.Is(err, ErrCheckPending) {
		result = "check pending"
		return
	}
	if err != nil {
		result, runErr = "error", err
		e.log().ErrorContext(ctx, "trigger", "trigger", name, "err", err)
		return
	}
	if !st.Holds {
		return
	}
	if tr.Handler != "" {
		digest := cfg.HookDigests[tr.Handler]
		if digest == "" {
			return
		}
		key := st.Key
		if key == "" {
			key = e.host.Now().UTC().Truncate(time.Minute).Format(time.RFC3339)
		}
		event := &hooks.Event{Source: "timer", Event: "timer", Delivery: key, Trigger: name, Payload: map[string]any{}, Fields: map[string]any{}}
		if err := e.host.Hooks().Put(ctx, "timer/"+name+"/"+key, hooks.Invocation{Trigger: name, Entry: tr.Handler, Digest: digest, Event: event, Args: tr.Args}); err != nil {
			e.log().ErrorContext(ctx, "enqueue timer hook", "trigger", name, "err", err)
		}
		e.host.KickHooks()
		return
	}
	t, err := e.host.Fire(ctx, cfg, name, tr, st.Key, nil)
	if errors.Is(err, ErrCheckPending) {
		result = "check pending" // the run's intake check (JOB-103)
		return
	}
	if err != nil {
		result, runErr = "error", err
		e.log().ErrorContext(ctx, "trigger fire", "trigger", name, "err", err)
		return
	}
	result = "fired"
	e.log().InfoContext(ctx, "trigger fired", "trigger", name, "ticket", t.ID)
}

// Eval checks dedup, cooldown, cron and counts, in that order.
func (e *Engine) Eval(ctx context.Context, cfg *config.Config, name string, tr config.Trigger) (State, error) {
	var st State
	s := e.host.Store()
	now := e.host.Now()
	mine := store.Filter{MetaPack: Pack, MetaKey: "name", MetaValue: name}

	newest := mine
	newest.Order, newest.Limit = store.OrderNewest, 1
	last, err := s.List(ctx, newest)
	if err != nil {
		return st, err
	}
	var lastT *domain.Ticket
	if len(last) > 0 {
		lastT = &last[0]
		st.Last = lastT.CreatedAt
	}
	if tr.Handler != "" {
		lastHook, err := e.host.Hooks().LastFire(ctx, name)
		if err != nil {
			return st, err
		}
		if lastHook.After(st.Last) {
			st.Last = lastHook
		}
	}

	if tr.When.Cron != "" {
		// The first slot after the last fire (or, for a trigger that never
		// fired, after this controller started) is due once now reaches
		// it. Slots missed while the controller was down collapse into one.
		base := e.host.Started()
		if !st.Last.IsZero() {
			base = st.Last
		}
		st.NextCron = cronNext(tr.When, base)
	}

	openCount := 0
	if tr.Handler != "" && tr.MaxOpen > 0 {
		if openCount, err = e.OpenCount(ctx, name); err != nil {
			return st, err
		}
	}
	for _, status := range []domain.Status{domain.StatusOpen, domain.StatusInProgress} {
		if tr.Handler != "" {
			break
		}
		f := mine
		f.Status, f.Limit = status, 1
		open, err := s.List(ctx, f)
		if err != nil {
			return st, err
		}
		if len(open) > 0 {
			st.Open = open[0].ID
			st.Why = "ticket " + open[0].ID + " is still open"
			return st, nil
		}
	}
	if tr.Handler != "" && tr.MaxOpen > 0 && openCount >= tr.MaxOpen {
		st.Why = "max_open reached"
		return st, nil
	}

	if tr.Handler != "" && tr.Cooldown != "" {
		// From the last fire or the close of the newest ticket it filed,
		// whichever is later: a max_open hold ends at the close, and the
		// cooldown runs from there.
		base := st.Last
		if lastT != nil && lastT.ClosedAt.After(base) {
			base = lastT.ClosedAt
		}
		if until := base.Add(config.Duration(tr.Cooldown, 0)); !base.IsZero() && now.Before(until) {
			st.Why = "cooldown until " + until.UTC().Format(time.RFC3339)
			return st, nil
		}
	}
	if tr.Handler == "" && lastT != nil && tr.Cooldown != "" {
		// lastT was read before the open-status queries. If it was closed
		// in between, its ClosedAt is still zero here and the cooldown
		// would compute from year 1 (#47). Treat that as not yet cooled.
		if lastT.ClosedAt.IsZero() {
			st.Why = "ticket " + lastT.ID + " closed during evaluation"
			return st, nil
		}
		until := lastT.ClosedAt.Add(config.Duration(tr.Cooldown, 0))
		if now.Before(until) {
			st.Why = "cooldown until " + until.UTC().Format(time.RFC3339)
			return st, nil
		}
	}

	if tr.When.Cron != "" {
		if st.NextCron.IsZero() || now.Before(st.NextCron) {
			st.Why = "cron not due"
			return st, nil
		}
		// Due: report the following slot as next.
		st.NextCron = cronNext(tr.When, now)
	}

	if why, err := e.countsHold(ctx, cfg, tr.When); err != nil || why != "" {
		st.Why = why
		return st, err
	}

	if l := tr.When.Learnings; l != nil {
		var since time.Time
		if !st.Last.IsZero() {
			since = st.Last
		}
		why, err := e.learningsHold(ctx, *l, since, now)
		if err != nil || why != "" {
			st.Why = why
			return st, err
		}
	}

	if ed := tr.When.Edits; ed != nil {
		why, err := e.host.EditsHold(ctx, *ed, now)
		if err != nil || why != "" {
			st.Why = why
			return st, err
		}
	}

	// Last, so the command runs only when everything else holds.
	if tr.When.Check != "" {
		key, why, err := e.runCheck(ctx, tr.When)
		if err != nil || why != "" {
			st.Why = why
			return st, err
		}
		if key != "" {
			seen, err := s.List(ctx, store.Filter{MetaPack: Pack, MetaKey: "key", MetaValue: name + "/" + key, Limit: 1})
			if err != nil {
				return st, err
			}
			if len(seen) > 0 {
				st.Why = "key " + key + " already filed as " + seen[0].ID
				return st, nil
			}
		}
		st.Key = key
	}
	st.Holds = true
	return st, nil
}

// OpenCount counts the open runs and tasks a trigger filed. Steps of its
// runs carry no meta.trigger, but are excluded by kind in case one does.
func (e *Engine) OpenCount(ctx context.Context, name string) (int, error) {
	ts, err := e.host.Store().List(ctx, store.Filter{
		MetaPack: Pack, MetaKey: "name", MetaValue: name,
		Kinds:    []domain.Kind{domain.KindRun, domain.KindTask},
		Statuses: []domain.Status{domain.StatusOpen, domain.StatusInProgress}, Limit: store.NoLimit,
	})
	return len(ts), err
}

// cronNext is the first slot of w.Cron after t, in w's timezone (Parse
// already validated both). Zero if it never matches.
func cronNext(w config.TriggerWhen, t time.Time) time.Time {
	sched, err := cron.Parse(w.Cron)
	if err != nil {
		return time.Time{}
	}
	loc, err := time.LoadLocation(w.Timezone) // "" is UTC
	if err != nil {
		return time.Time{}
	}
	return sched.Next(t.In(loc))
}

// learningsHold returns "" when enough open learnings were recorded since
// the last fire (#273): at least l.Min, or at least one once l.Every has
// passed. A trigger that never fired (zero since) counts every open one
// and its every is due.
func (e *Engine) learningsHold(ctx context.Context, l config.LearningsWhen, since, now time.Time) (string, error) {
	from := since
	if !from.IsZero() {
		// Strictly after the last fire: one recorded in the same
		// millisecond belongs to that pass.
		from = from.Add(time.Millisecond)
	}
	n, err := knowledge.CountLearnings(ctx, e.host.Store(), from, domain.LearningCount{
		Status: domain.LearningOpen, Kinds: l.Kinds, Scopes: l.Scope, Source: l.Source,
	})
	if err != nil {
		return "", err
	}
	if n >= l.Min {
		return "", nil
	}
	if l.Every != "" && n > 0 && !now.Before(since.Add(config.Duration(l.Every, 0))) {
		return "", nil
	}
	return fmt.Sprintf("open learnings %d is below %d", n, l.Min), nil
}

// countsHold returns "" when every count threshold in w holds, else the
// first that does not. Counts are capped at store.DefaultLimit (500).
func (e *Engine) countsHold(ctx context.Context, cfg *config.Config, w config.TriggerWhen) (string, error) {
	if !w.HasCounts() {
		return "", nil
	}
	kind := domain.Kind(w.Kind)
	if kind == "" {
		kind = domain.KindTask
	}
	count := func(s domain.Status) (int, error) {
		ts, err := e.host.Store().List(ctx, store.Filter{Kind: kind, Status: s, Pool: w.Pool, Label: w.Label})
		return len(ts), err
	}
	checks := []struct {
		name         string
		below, above *int
		n            func() (int, error)
	}{
		{"ready", w.ReadyBelow, w.ReadyAbove, func() (int, error) { return e.readyCount(ctx, cfg, w, kind) }},
		{"open", w.OpenBelow, w.OpenAbove, func() (int, error) { return count(domain.StatusOpen) }},
		{"in_progress", w.InProgressBelow, w.InProgressAbove, func() (int, error) { return count(domain.StatusInProgress) }},
	}
	for _, ch := range checks {
		if ch.below == nil && ch.above == nil {
			continue
		}
		n, err := ch.n()
		if err != nil {
			return "", err
		}
		if ch.below != nil && n >= *ch.below {
			return fmt.Sprintf("%s %d is not below %d", ch.name, n, *ch.below), nil
		}
		if ch.above != nil && n <= *ch.above {
			return fmt.Sprintf("%s %d is not above %d", ch.name, n, *ch.above), nil
		}
	}
	return "", nil
}

// readyCount counts claimable tickets: open, unblocked and not deferred.
// Only claimable kinds (tasks) can be ready, so other kinds count zero.
func (e *Engine) readyCount(ctx context.Context, cfg *config.Config, w config.TriggerWhen, kind domain.Kind) (int, error) {
	if !slices.Contains(domain.ClaimableKinds, kind) {
		return 0, nil
	}
	projects := slices.Sorted(maps.Keys(cfg.Projects))
	if p, ok := cfg.Pools[w.Pool]; ok {
		projects = []string{p.Project}
	}
	n := 0
	for _, proj := range projects {
		ts, err := e.host.Store().List(ctx, store.Filter{Project: proj, Pool: w.Pool, Ready: true, Limit: store.DefaultLimit})
		if err != nil {
			return 0, err
		}
		for _, t := range ts {
			if w.Label == "" || slices.Contains(t.Labels, w.Label) {
				n++
			}
		}
	}
	return n, nil
}

// runCheck is the check of w: run now, or, under a Pass on ctx (JOB-109),
// the stored result, or ErrCheckPending and a note of the command to run
// outside the tick.
func (e *Engine) runCheck(ctx context.Context, w config.TriggerWhen) (key, why string, err error) {
	cp, ok := ctx.Value(passKey{}).(*Pass)
	if !ok {
		return e.runCheckNow(ctx, w)
	}
	k := checkKey(w)
	if r, ok := cp.Results[k]; ok {
		return r.Out, r.Why, r.Err
	}
	cp.note(Need{Key: k, When: w})
	return "", "", ErrCheckPending
}

// runCheckNow runs w.Check with sh in the root dir. Exit 0: holds, and the
// trimmed stdout is the dedup key. Non-zero exit: does not hold (why says
// so); the command failing to start or timing out is an error.
func (e *Engine) runCheckNow(ctx context.Context, w config.TriggerWhen) (key, why string, err error) {
	ctx, cancel := context.WithTimeout(ctx, config.Duration(w.CheckTimeout, time.Minute))
	defer cancel()
	run, err := e.host.RunCheck(ctx, w.Check)
	if ctx.Err() != nil {
		return "", "", fmt.Errorf("check: %w", ctx.Err())
	}
	if err != nil {
		return "", "", fmt.Errorf("check: %w: %s", err, strings.TrimSpace(run.Stderr))
	}
	if run.Exit != "" {
		return "", "check exited " + run.Exit, nil
	}
	return strings.TrimSpace(run.Stdout), "", nil
}

// Info evaluates one trigger for trigger.list.
func (e *Engine) Info(ctx context.Context, cfg *config.Config, name string, tr config.Trigger) (rpc.TriggerInfo, error) {
	if len(tr.When.Scan) > 0 {
		fire, why, err := e.evalScan(ctx, cfg, name, tr, nil)
		if err != nil {
			return rpc.TriggerInfo{}, err
		}
		info := rpc.TriggerInfo{Name: name, Holds: len(fire) > 0, Why: why}
		for _, s := range fire {
			info.Findings = append(info.Findings, s.KeyWith(tr.ScanKeyPrefix()))
		}
		return info, nil
	}
	st, err := e.Eval(ctx, cfg, name, tr)
	if err != nil {
		return rpc.TriggerInfo{}, err
	}
	info := rpc.TriggerInfo{Name: name, Holds: st.Holds, Why: st.Why, Open: st.Open}
	if !st.Last.IsZero() {
		info.LastFired = &st.Last
	}
	if !st.NextCron.IsZero() {
		info.NextCron = &st.NextCron
	}
	return info, nil
}

// pollDue reports whether a tick should poll tr, and records the poll. A
// trigger with on: polls every fallback (default
// config.DefaultTriggerFallback), and never without when conditions;
// without on:, every tick unless fallback is set. Scan triggers too
// (JOB-72). Tick-only state.
func (e *Engine) pollDue(name string, tr config.Trigger) bool {
	every := time.Duration(0)
	if len(tr.On) > 0 {
		// A handler decides for itself whether to act, so its fallback
		// polls with no when conditions.
		if !tr.When.HasConditions() && (tr.Handler == "" || tr.Fallback == "") {
			return false
		}
		every = config.Duration(tr.Fallback, config.DefaultTriggerFallback)
	} else if tr.Fallback != "" {
		every = config.Duration(tr.Fallback, 0)
	}
	if every == 0 {
		return true
	}
	now := e.host.Now()
	if last, ok := e.polled[name]; ok && now.Before(last.Add(every)) {
		return false
	}
	if e.polled == nil {
		e.polled = map[string]time.Time{}
	}
	e.polled[name] = now
	return true
}
