package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/cron"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/learning"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// triggerPack is the meta pack on tickets a trigger created:
// {"trigger": {"name": "<trigger>"}}. Dedup and cooldown read it back, so
// trigger state lives in the store and survives restarts.
const triggerPack = "trigger"

// triggerState is what one evaluation found.
type triggerState struct {
	holds    bool
	why      string    // first condition that does not hold
	open     string    // open ticket the trigger created, if any
	last     time.Time // created_at of its newest ticket
	nextCron time.Time
	key      string // check output, stored as meta.trigger.key
}

// runTriggers fires every trigger whose conditions hold. Called each tick.
func (c *Controller) runTriggers(ctx context.Context, cfg *config.Config) {
	for _, name := range slices.Sorted(maps.Keys(cfg.Triggers)) {
		tr := cfg.Triggers[name]
		if c.backgroundChecks && c.checkJobs.running[name] {
			continue // its check is still running; poll again next tick
		}
		if !c.triggerPollDue(name, tr) {
			continue
		}
		c.runTriggerChecks(ctx, cfg, name, false, func(ctx context.Context, cfg *config.Config) {
			// Looked up again: a replay after the check runs on the then
			// current config, where the trigger may be gone.
			if tr, ok := cfg.Triggers[name]; ok {
				c.runTrigger(ctx, cfg, name, tr)
			}
		})
	}
}

// runTrigger evaluates one trigger that is due, under a span naming it and
// the result: fired, held (conditions do not hold), stall or error
// (JOB-105).
func (c *Controller) runTrigger(ctx context.Context, cfg *config.Config, name string, tr config.Trigger) {
	ctx, span := telemetry.StartSpan(ctx, "trigger "+name, attribute.String("jobshop.trigger", name))
	result := "held"
	var runErr error
	defer func() {
		span.SetAttributes(attribute.String("jobshop.trigger.result", result))
		telemetry.End(span, runErr)
	}()
	if len(tr.When.Stall) > 0 {
		c.runStallTrigger(ctx, cfg, name, tr, nil)
		result = "stall"
		return
	}
	st, err := c.evalTrigger(ctx, cfg, name, tr)
	if errors.Is(err, errCheckPending) {
		result = "check pending"
		return
	}
	if err != nil {
		result, runErr = "error", err
		c.log.ErrorContext(ctx, "trigger", "trigger", name, "err", err)
		return
	}
	if !st.holds {
		return
	}
	if tr.Handler != "" {
		digest := cfg.HookDigests[tr.Handler]
		if digest == "" {
			return
		}
		key := st.key
		if key == "" {
			key = c.now().UTC().Truncate(time.Minute).Format(time.RFC3339)
		}
		event := &hooks.Event{Source: "timer", Event: "timer", Delivery: key, Trigger: name, Payload: map[string]any{}, Fields: map[string]any{}}
		if err := c.hooks().Put(ctx, "timer/"+name+"/"+key, hooks.Invocation{Trigger: name, Entry: tr.Handler, Digest: digest, Event: event, Args: tr.Args}); err != nil {
			c.log.ErrorContext(ctx, "enqueue timer hook", "trigger", name, "err", err)
		}
		c.kickHooks()
		return
	}
	t, err := c.fireTrigger(ctx, cfg, name, tr, st.key)
	if errors.Is(err, errCheckPending) {
		result = "check pending" // the run's intake check (JOB-103)
		return
	}
	if err != nil {
		result, runErr = "error", err
		c.log.ErrorContext(ctx, "trigger fire", "trigger", name, "err", err)
		return
	}
	result = "fired"
	c.log.InfoContext(ctx, "trigger fired", "trigger", name, "ticket", t.ID)
}

// evalTrigger checks dedup, cooldown, cron and counts, in that order.
func (c *Controller) evalTrigger(ctx context.Context, cfg *config.Config, name string, tr config.Trigger) (triggerState, error) {
	var st triggerState
	now := c.now()
	mine := store.Filter{MetaPack: triggerPack, MetaKey: "name", MetaValue: name}

	newest := mine
	newest.Order, newest.Limit = store.OrderNewest, 1
	last, err := c.store.List(ctx, newest)
	if err != nil {
		return st, err
	}
	var lastT *store.Ticket
	if len(last) > 0 {
		lastT = &last[0]
		st.last = lastT.CreatedAt
	}
	if tr.Handler != "" {
		lastHook, err := c.hooks().LastFire(ctx, name)
		if err != nil {
			return st, err
		}
		if lastHook.After(st.last) {
			st.last = lastHook
		}
	}

	if tr.When.Cron != "" {
		// The first slot after the last fire (or, for a trigger that never
		// fired, after this controller started) is due once now reaches
		// it. Slots missed while the controller was down collapse into one.
		base := c.started
		if !st.last.IsZero() {
			base = st.last
		}
		st.nextCron = cronNext(tr.When, base)
	}

	openCount := 0
	if tr.Handler != "" && tr.MaxOpen > 0 {
		if openCount, err = c.triggerOpen(ctx, name); err != nil {
			return st, err
		}
	}
	for _, s := range []store.Status{store.StatusOpen, store.StatusInProgress} {
		if tr.Handler != "" {
			break
		}
		f := mine
		f.Status, f.Limit = s, 1
		open, err := c.store.List(ctx, f)
		if err != nil {
			return st, err
		}
		if len(open) > 0 {
			st.open = open[0].ID
			st.why = "ticket " + open[0].ID + " is still open"
			return st, nil
		}
	}
	if tr.Handler != "" && tr.MaxOpen > 0 && openCount >= tr.MaxOpen {
		st.why = "max_open reached"
		return st, nil
	}

	if tr.Handler != "" && tr.Cooldown != "" {
		// From the last fire or the close of the newest ticket it filed,
		// whichever is later: a max_open hold ends at the close, and the
		// cooldown runs from there.
		base := st.last
		if lastT != nil && lastT.ClosedAt.After(base) {
			base = lastT.ClosedAt
		}
		if until := base.Add(config.Duration(tr.Cooldown, 0)); !base.IsZero() && now.Before(until) {
			st.why = "cooldown until " + until.UTC().Format(time.RFC3339)
			return st, nil
		}
	}
	if tr.Handler == "" && lastT != nil && tr.Cooldown != "" {
		// lastT was read before the open-status queries. If it was closed
		// in between, its ClosedAt is still zero here and the cooldown
		// would compute from year 1 (#47). Treat that as not yet cooled.
		if lastT.ClosedAt.IsZero() {
			st.why = "ticket " + lastT.ID + " closed during evaluation"
			return st, nil
		}
		until := lastT.ClosedAt.Add(config.Duration(tr.Cooldown, 0))
		if now.Before(until) {
			st.why = "cooldown until " + until.UTC().Format(time.RFC3339)
			return st, nil
		}
	}

	if tr.When.Cron != "" {
		if st.nextCron.IsZero() || now.Before(st.nextCron) {
			st.why = "cron not due"
			return st, nil
		}
		// Due: report the following slot as next.
		st.nextCron = cronNext(tr.When, now)
	}

	if why, err := c.countsHold(ctx, cfg, tr.When); err != nil || why != "" {
		st.why = why
		return st, err
	}

	if l := tr.When.Learnings; l != nil {
		var since time.Time
		if !st.last.IsZero() {
			since = st.last
		}
		why, err := c.learningsHold(ctx, *l, since, now)
		if err != nil || why != "" {
			st.why = why
			return st, err
		}
	}

	if e := tr.When.Edits; e != nil {
		why, err := c.editsHold(ctx, *e, now)
		if err != nil || why != "" {
			st.why = why
			return st, err
		}
	}

	// Last, so the command runs only when everything else holds.
	if tr.When.Check != "" {
		key, why, err := c.runCheck(ctx, tr.When)
		if err != nil || why != "" {
			st.why = why
			return st, err
		}
		if key != "" {
			seen, err := c.store.List(ctx, store.Filter{MetaPack: triggerPack, MetaKey: "key", MetaValue: name + "/" + key, Limit: 1})
			if err != nil {
				return st, err
			}
			if len(seen) > 0 {
				st.why = "key " + key + " already filed as " + seen[0].ID
				return st, nil
			}
		}
		st.key = key
	}
	st.holds = true
	return st, nil
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
func (c *Controller) learningsHold(ctx context.Context, l config.LearningsWhen, since, now time.Time) (string, error) {
	from := since
	if !from.IsZero() {
		// Strictly after the last fire: one recorded in the same
		// millisecond belongs to that pass.
		from = from.Add(time.Millisecond)
	}
	n, err := learning.Count(ctx, c.store, from, store.LearningCount{
		Status: store.LearningOpen, Kinds: l.Kinds, Scopes: l.Scope, Source: l.Source,
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
func (c *Controller) countsHold(ctx context.Context, cfg *config.Config, w config.TriggerWhen) (string, error) {
	if !w.HasCounts() {
		return "", nil
	}
	kind := store.Kind(w.Kind)
	if kind == "" {
		kind = store.KindTask
	}
	count := func(s store.Status) (int, error) {
		ts, err := c.store.List(ctx, store.Filter{Kind: kind, Status: s, Pool: w.Pool, Label: w.Label})
		return len(ts), err
	}
	checks := []struct {
		name         string
		below, above *int
		n            func() (int, error)
	}{
		{"ready", w.ReadyBelow, w.ReadyAbove, func() (int, error) { return c.readyCount(ctx, cfg, w, kind) }},
		{"open", w.OpenBelow, w.OpenAbove, func() (int, error) { return count(store.StatusOpen) }},
		{"in_progress", w.InProgressBelow, w.InProgressAbove, func() (int, error) { return count(store.StatusInProgress) }},
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
func (c *Controller) readyCount(ctx context.Context, cfg *config.Config, w config.TriggerWhen, kind store.Kind) (int, error) {
	if !slices.Contains(store.ClaimableKinds, kind) {
		return 0, nil
	}
	projects := slices.Sorted(maps.Keys(cfg.Projects))
	if p, ok := cfg.Pools[w.Pool]; ok {
		projects = []string{p.Project}
	}
	n := 0
	for _, proj := range projects {
		ts, err := c.store.List(ctx, store.Filter{Project: proj, Pool: w.Pool, Ready: true, Limit: store.DefaultLimit})
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

// runCheck is the check of w: run now, or, under a checkPass on ctx
// (JOB-109), the stored result, or errCheckPending and a note of the
// command to run outside the tick.
func (c *Controller) runCheck(ctx context.Context, w config.TriggerWhen) (key, why string, err error) {
	cp, ok := ctx.Value(checkPassKey{}).(*checkPass)
	if !ok {
		return c.runCheckNow(ctx, w)
	}
	k := checkKey(w)
	if r, ok := cp.results[k]; ok {
		return r.out, r.why, r.err
	}
	cp.note(checkNeed{key: k, when: w})
	return "", "", errCheckPending
}

// runCheckNow runs w.Check with sh in the root dir. Exit 0: holds, and the
// trimmed stdout is the dedup key. Non-zero exit: does not hold (why says
// so); the command failing to start or timing out is an error.
func (c *Controller) runCheckNow(ctx context.Context, w config.TriggerWhen) (key, why string, err error) {
	ctx, cancel := context.WithTimeout(ctx, config.Duration(w.CheckTimeout, time.Minute))
	defer cancel()
	// Triggers have no pool: unisolated in the root dir, through
	// proc.Start for the process-group kill (#810).
	var stdout bytes.Buffer
	var stderr strings.Builder
	ex, err := c.runIsolated(ctx, isolatedCmd{argv: shArgv(w.Check), dir: c.root.Dir, stdout: &stdout, stderr: &stderr, name: "trigger-check"})
	if ctx.Err() != nil {
		return "", "", fmt.Errorf("check: %w", ctx.Err())
	}
	if err != nil {
		return "", "", fmt.Errorf("check: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	if ex.ExitCode != 0 || ex.Signal != "" {
		return "", "check exited " + exitText(ex), nil
	}
	return strings.TrimSpace(stdout.String()), "", nil
}

func (c *Controller) fireTrigger(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, key string) (store.Ticket, error) {
	return c.fireTriggerEvent(ctx, cfg, name, tr, key, nil)
}

// fireTriggerEvent fires tr; ev is the event that fired it, nil for a
// poll. Its fields are the event template var (empty on a poll).
func (c *Controller) fireTriggerEvent(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, key string, ev triggerEvent) (store.Ticket, error) {
	m := map[string]string{"name": name}
	if key != "" {
		// Prefixed with the trigger name so keys of different triggers
		// do not collide in the meta filter.
		m["key"] = name + "/" + key
	}
	meta, err := json.Marshal(map[string]map[string]string{triggerPack: m})
	if err != nil {
		return store.Ticket{}, err
	}
	out, _, _ := strings.Cut(key, "\n")
	tv := map[string]any{config.CheckOutputVar: strings.TrimSpace(out), config.EventVar: ev.vars()}
	if tr.Run != nil {
		return c.fireRun(ctx, cfg, tr.Run, m, tv)
	}
	if tr.Command != nil {
		return c.fireCommand(ctx, cfg, name, tr, meta, tv)
	}
	tt := *tr.Ticket
	if key != "" {
		tt.Body = strings.TrimSpace(tt.Body + "\n\nCheck output: " + key)
	}
	return store.Create(ctx, c.store, "controller", store.NewTicket{
		Project: tt.Project, Kind: store.KindTask, Title: tt.Title, Body: tt.Body,
		Priority: tt.Priority, Pool: tt.Pool, Labels: tt.Labels, Meta: meta,
	})
}

// fireRun starts the trigger's workflow run. meta.trigger goes on the run
// ticket, so the open run holds the trigger and its key dedups like a
// ticket's. A run already open under the same run key is not an error;
// the trigger stays held by that run only if it carries the trigger meta.
func (c *Controller) fireRun(ctx context.Context, cfg *config.Config, r *config.TriggerRun, m map[string]string, tv map[string]any) (store.Ticket, error) {
	p := rpc.RunStartParams{Workflow: r.Workflow, Vars: map[string]string{}}
	for k, v := range r.Vars {
		s, err := r.Fill(v, tv)
		if err != nil {
			return store.Ticket{}, fmt.Errorf("run var %s: %w", k, err)
		}
		p.Vars[k] = s
	}
	k, err := r.Fill(r.Key, tv)
	if err != nil {
		return store.Ticket{}, fmt.Errorf("run key: %w", err)
	}
	p.Key = k
	res, err := c.startRun(ctx, cfg, scope{admin: true, actor: "controller"}, p, map[string]any{triggerPack: m})
	if err != nil {
		return store.Ticket{}, err
	}
	return res.(rpc.RunResult).Run, nil
}

func (c *Controller) triggerList(ctx context.Context, sc scope, _ struct{}) (any, error) {
	if err := c.requireGrant(sc, "trigger-list", "trigger list"); err != nil {
		return nil, err
	}
	// Triggers are evaluated in parallel: most of the time is check
	// commands (often gh), and one after another they made the Desk
	// pools view take ~5s (#522). Evaluation only reads, so order does
	// not matter.
	cfg := c.config()
	trs := cfg.Triggers
	names := slices.Sorted(maps.Keys(trs))
	out := make([]rpc.TriggerInfo, len(names))
	errs := make([]error, len(names))
	var wg sync.WaitGroup
	for i, name := range names {
		wg.Go(func() { out[i], errs[i] = c.triggerInfo(ctx, cfg, name, trs[name]) })
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	labels := map[string][]string{}
	for _, in := range cfg.Instances {
		for _, name := range in.Triggers {
			labels[name] = config.InstanceLabels(in)
		}
	}
	for i := range out {
		out[i].Labels = labels[out[i].Name]
	}
	return out, nil
}

// triggerInfo evaluates one trigger for trigger.list.
func (c *Controller) triggerInfo(ctx context.Context, cfg *config.Config, name string, tr config.Trigger) (rpc.TriggerInfo, error) {
	if len(tr.When.Stall) > 0 {
		fire, why, err := c.evalStall(ctx, cfg, name, tr, nil)
		if err != nil {
			return rpc.TriggerInfo{}, err
		}
		info := rpc.TriggerInfo{Name: name, Holds: len(fire) > 0, Why: why}
		for _, s := range fire {
			info.Stalls = append(info.Stalls, s.keyWith(tr.StallKeyPrefix()))
		}
		return info, nil
	}
	st, err := c.evalTrigger(ctx, cfg, name, tr)
	if err != nil {
		return rpc.TriggerInfo{}, err
	}
	info := rpc.TriggerInfo{Name: name, Holds: st.holds, Why: st.why, Open: st.open}
	if !st.last.IsZero() {
		info.LastFired = &st.last
	}
	if !st.nextCron.IsZero() {
		info.NextCron = &st.nextCron
	}
	return info, nil
}
