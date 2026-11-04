package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os/exec"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/cron"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
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
func (c *Controller) runTriggers() {
	for _, name := range slices.Sorted(maps.Keys(c.config().Triggers)) {
		tr := c.config().Triggers[name]
		if len(tr.When.Stall) > 0 {
			c.runStallTrigger(c.ctx, name, tr)
			continue
		}
		st, err := c.evalTrigger(c.ctx, name, tr)
		if err != nil {
			c.log.Error("trigger", "trigger", name, "err", err)
			continue
		}
		if !st.holds {
			continue
		}
		t, err := c.fireTrigger(c.ctx, name, tr, st.key)
		if err != nil {
			c.log.Error("trigger fire", "trigger", name, "err", err)
			continue
		}
		c.log.Info("trigger fired", "trigger", name, "ticket", t.ID)
	}
}

// evalTrigger checks dedup, cooldown, cron and counts, in that order.
func (c *Controller) evalTrigger(ctx context.Context, name string, tr config.Trigger) (triggerState, error) {
	var st triggerState
	now := c.now()
	mine := store.Filter{MetaPack: triggerPack, MetaKey: "name", MetaValue: name}

	newest := mine
	newest.Newest, newest.Limit = true, 1
	last, err := c.store.List(ctx, newest)
	if err != nil {
		return st, err
	}
	var lastT *store.Ticket
	if len(last) > 0 {
		lastT = &last[0]
		st.last = lastT.CreatedAt
	}

	if tr.When.Cron != "" {
		// The first slot after the last fire (or, for a trigger that never
		// fired, after this controller started) is due once now reaches
		// it. Slots missed while the controller was down collapse into one.
		base := c.started
		if lastT != nil {
			base = lastT.CreatedAt
		}
		st.nextCron = cronNext(tr.When, base)
	}

	for _, s := range []store.Status{store.StatusOpen, store.StatusInProgress} {
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

	if lastT != nil && tr.Cooldown != "" {
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

	if why, err := c.countsHold(ctx, tr.When); err != nil || why != "" {
		st.why = why
		return st, err
	}

	if l := tr.When.Learnings; l != nil {
		var since time.Time
		if lastT != nil {
			since = lastT.CreatedAt
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
	n, err := c.store.CountLearnings(ctx, from, store.LearningOpen)
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
func (c *Controller) countsHold(ctx context.Context, w config.TriggerWhen) (string, error) {
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
		{"ready", w.ReadyBelow, w.ReadyAbove, func() (int, error) { return c.readyCount(ctx, w, kind) }},
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
func (c *Controller) readyCount(ctx context.Context, w config.TriggerWhen, kind store.Kind) (int, error) {
	if !slices.Contains(store.ClaimableKinds, kind) {
		return 0, nil
	}
	projects := slices.Sorted(maps.Keys(c.config().Projects))
	if p, ok := c.config().Pools[w.Pool]; ok {
		projects = []string{p.Project}
	}
	n := 0
	for _, proj := range projects {
		ts, err := c.store.Ready(ctx, proj, w.Pool, store.DefaultLimit)
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

// runCheck runs w.Check with sh in the root dir. Exit 0: holds, and the
// trimmed stdout is the dedup key. Non-zero exit: does not hold (why says
// so); the command failing to start or timing out is an error.
func (c *Controller) runCheck(ctx context.Context, w config.TriggerWhen) (key, why string, err error) {
	ctx, cancel := context.WithTimeout(ctx, config.Duration(w.CheckTimeout, time.Minute))
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", w.Check)
	cmd.Dir = c.root.Dir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return "", "", fmt.Errorf("check: %w", ctx.Err())
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return "", fmt.Sprintf("check exited %d", exit.ExitCode()), nil
	}
	if err != nil {
		return "", "", fmt.Errorf("check: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(string(out)), "", nil
}

func (c *Controller) fireTrigger(ctx context.Context, name string, tr config.Trigger, key string) (store.Ticket, error) {
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
	if tr.Run != nil {
		return c.fireRun(ctx, tr.Run, m, key)
	}
	tt := *tr.Ticket
	if key != "" {
		tt.Body = strings.TrimSpace(tt.Body + "\n\nCheck output: " + key)
	}
	return c.store.Create(ctx, "controller", store.NewTicket{
		Project: tt.Project, Kind: store.KindTask, Title: tt.Title, Body: tt.Body,
		Priority: tt.Priority, Pool: tt.Pool, Labels: tt.Labels, Meta: meta,
	})
}

// fireRun starts the trigger's workflow run. meta.trigger goes on the run
// ticket, so the open run holds the trigger and its key dedups like a
// ticket's. A run already open under the same run key is not an error;
// the trigger stays held by that run only if it carries the trigger meta.
func (c *Controller) fireRun(ctx context.Context, r *config.TriggerRun, m map[string]string, key string) (store.Ticket, error) {
	out, _, _ := strings.Cut(key, "\n")
	tv := map[string]any{config.CheckOutputVar: strings.TrimSpace(out)}
	p := rpc.RunStartParams{Workflow: r.Workflow, Vars: map[string]string{}}
	for k, v := range r.Vars {
		s, err := config.FillTemplate(v, tv)
		if err != nil {
			return store.Ticket{}, fmt.Errorf("run var %s: %w", k, err)
		}
		p.Vars[k] = s
	}
	k, err := config.FillTemplate(r.Key, tv)
	if err != nil {
		return store.Ticket{}, fmt.Errorf("run key: %w", err)
	}
	p.Key = k
	res, err := c.startRun(ctx, scope{admin: true, actor: "controller"}, p, map[string]any{triggerPack: m})
	if err != nil {
		return store.Ticket{}, err
	}
	return res.(rpc.RunResult).Run, nil
}

func (c *Controller) triggerList(ctx context.Context, sc scope, _ struct{}) (any, error) {
	if err := c.requireGrant(sc, "trigger-list", "trigger list"); err != nil {
		return nil, err
	}
	out := []rpc.TriggerInfo{}
	for name, tr := range c.config().Triggers {
		if len(tr.When.Stall) > 0 {
			fire, why, err := c.evalStall(ctx, name, tr)
			if err != nil {
				return nil, err
			}
			info := rpc.TriggerInfo{Name: name, Holds: len(fire) > 0, Why: why}
			for _, s := range fire {
				info.Stalls = append(info.Stalls, s.key())
			}
			out = append(out, info)
			continue
		}
		st, err := c.evalTrigger(ctx, name, tr)
		if err != nil {
			return nil, err
		}
		info := rpc.TriggerInfo{Name: name, Holds: st.holds, Why: st.why, Open: st.open}
		if !st.last.IsZero() {
			info.LastFired = &st.last
		}
		if !st.nextCron.IsZero() {
			info.NextCron = &st.nextCron
		}
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
