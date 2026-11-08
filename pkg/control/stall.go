package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	host "github.com/elee1766/acphost/pkg/acphost"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// needsHumanLabel marks tickets waiting on the operator.
const needsHumanLabel = "needs-human"

// blockedLabel marks tickets that claim a blocker.
const blockedLabel = "blocked"

// stall is one detected stall: a signal on a subject (ticket, run,
// session or pool id, or a check's first word).
type stall struct {
	signal, subject, detail string
}

// id is the stall's identity, <signal>:<subject>, the same whatever key
// prefix a trigger uses (JOB-95). Runs carry it in meta.trigger.stall.
func (s stall) id() string { return s.signal + ":" + s.subject }

// key is the run key with the default prefix, stall:<signal>:<subject>.
func (s stall) key() string { return s.keyWith(config.DefaultStallKeyPrefix) }

// keyWith is the run key <prefix>:<signal>:<subject>.
func (s stall) keyWith(prefix string) string { return prefix + ":" + s.id() }

// stallMetaKey is meta.trigger.stall, <signal>:<subject>, set on the runs
// a stall trigger starts with a stall key (no run.key).
const stallMetaKey = "stall"

// runStallTrigger starts one run per fireable stall of a when.stall
// trigger (#143). ev is the delivery that woke it (JOB-72), nil on a poll.
func (c *Controller) runStallTrigger(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, ev triggerEvent) {
	fire, _, err := c.evalStall(ctx, cfg, name, tr, ev)
	if errors.Is(err, errCheckPending) {
		return // its checks run outside the tick (JOB-109)
	}
	if err != nil {
		c.log.ErrorContext(ctx, "stall trigger", "trigger", name, "err", err)
		return
	}
	prefix := tr.StallKeyPrefix()
	for _, s := range fire {
		key := s.keyWith(prefix)
		if s.signal == config.StallLinearIssueNoRun {
			busy, err := c.linearIssueHasOpenRun(ctx, s.subject)
			if err != nil {
				c.log.ErrorContext(ctx, "stall recheck", "trigger", name, "key", key, "err", err)
				continue
			}
			if busy {
				c.log.InfoContext(ctx, "stall skipped: issue got a run during the scan", "trigger", name, "key", key)
				continue
			}
		}
		if tr.Handler != "" {
			c.enqueueStall(ctx, cfg, name, tr, s)
			continue
		}
		t, existing, err := c.fireStall(ctx, cfg, name, tr, s)
		if errors.Is(err, errCheckPending) {
			continue // its intake check runs outside the tick (JOB-103)
		}
		if err != nil {
			c.log.ErrorContext(ctx, "stall fire", "trigger", name, "key", key, "err", err)
			continue
		}
		if existing {
			// Already handled by an open run with the same key: not a
			// new stall to learn from.
			c.log.DebugContext(ctx, "stall already has an open run", "trigger", name, "key", key, "run", t.ID)
			continue
		}
		c.log.InfoContext(ctx, "stall trigger fired", "trigger", name, "key", key, "run", t.ID, "delivery", ev["delivery"])
		c.recordStall(ctx, s, t.ID)
	}
}

// enqueueStall calls a stall trigger's handler for one fireable stall. The
// invocation key is the stall, its run count and a cooldown slot, so one
// stall is offered once per slot; max_runs, the cooldown and the #704 hold
// were checked by evalStall. The handler's start_run gets
// meta.trigger.stall, so the next evaluation sees its run.
func (c *Controller) enqueueStall(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, s stall) {
	digest := cfg.HookDigests[tr.Handler]
	if digest == "" {
		return
	}
	runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, Meta: []store.MetaMatch{{Pack: triggerPack, Key: "name", Value: name}, {Pack: triggerPack, Key: stallMetaKey, Value: s.id()}}, Limit: store.NoLimit})
	if err != nil {
		c.log.ErrorContext(ctx, "stall handler", "trigger", name, "err", err)
		return
	}
	event := &hooks.Event{
		Source: hooks.StallSource, Event: s.signal, Delivery: s.id(), Trigger: name,
		Payload: map[string]any{"signal": s.signal, "subject": s.subject, "detail": s.detail, "key": s.keyWith(tr.StallKeyPrefix()), "runs": len(runs)},
		Fields:  map[string]any{},
	}
	// A handler that declines a stall is offered it again a cooldown later
	// (30m without one), while the signal persists.
	bucket := c.now().Truncate(config.Duration(tr.Cooldown, 30*time.Minute)).UTC().Format(time.RFC3339)
	key := "stall/" + name + "/" + s.id() + "/" + strconv.Itoa(len(runs)) + "/" + bucket
	if err := c.hooks().Put(ctx, key, hooks.Invocation{Trigger: name, Entry: tr.Handler, Digest: digest, Event: event, Args: tr.Args}); err != nil {
		c.log.ErrorContext(ctx, "enqueue stall hook", "trigger", name, "err", err)
		return
	}
	c.kickHooks()
}

// evalStall detects stalls and returns those that may fire now: no open
// run for the key, under max_runs, past the cooldown, and under
// max_open open runs in total (JOB-93), so a backlog of old stalls (a
// fresh deploy, a long outage) is worked a few at a time. why explains
// an empty result. ev is passed to the checks, nil on a poll.
func (c *Controller) evalStall(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, ev triggerEvent) (fire []stall, why string, err error) {
	mine, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, MetaPack: triggerPack, MetaKey: "name", MetaValue: name, Order: store.OrderNewest})
	if err != nil {
		return nil, "", err
	}
	open := 0
	for _, t := range mine {
		if t.Status != store.StatusClosed {
			open++
		}
	}
	maxOpen := tr.StallMaxOpen()
	if open >= maxOpen {
		return nil, fmt.Sprintf("%d stall runs open (max_open %d)", open, maxOpen), nil
	}
	found, err := c.detectStalls(ctx, cfg, tr, ev)
	if err != nil {
		return nil, "", err
	}
	if len(found) == 0 {
		return nil, "no stall", nil
	}
	now := c.now()
	prefix := tr.StallKeyPrefix()
	var reasons []string
	for _, s := range found {
		key := s.keyWith(prefix)
		f := store.Filter{Kind: store.KindRun, MetaPack: triggerPack, MetaKey: "key", MetaValue: name + "/" + key, Order: store.OrderNewest}
		if tr.Handler != "" {
			// A handler picks its own run keys; its runs carry the stall id.
			f = store.Filter{Kind: store.KindRun, Meta: []store.MetaMatch{{Pack: triggerPack, Key: "name", Value: name}, {Pack: triggerPack, Key: stallMetaKey, Value: s.id()}}, Order: store.OrderNewest}
		}
		runs, err := c.store.List(ctx, f)
		if err != nil {
			return nil, "", err
		}
		r := stallHeld(runs, tr, now)
		if r == "" {
			if r, err = c.stallSubjectHeld(ctx, s); err != nil {
				return nil, "", err
			}
		}
		if r != "" {
			reasons = append(reasons, key+": "+r)
			continue
		}
		fire = append(fire, s)
		if open+len(fire) >= maxOpen {
			break
		}
	}
	if len(fire) == 0 {
		why = strings.Join(reasons, "; ")
	}
	return fire, why, nil
}

// stallHeld says why a stall with these earlier runs (newest first) may
// not fire now, or "".
func stallHeld(runs []store.Ticket, tr config.Trigger, now time.Time) string {
	if len(runs) == 0 {
		return ""
	}
	if runs[0].Status != store.StatusClosed {
		return "run " + runs[0].ID + " is open"
	}
	if len(runs) >= tr.StallMaxRuns() {
		return fmt.Sprintf("max_runs %d reached", tr.StallMaxRuns())
	}
	if tr.Cooldown != "" {
		if runs[0].ClosedAt.IsZero() {
			return "run " + runs[0].ID + " closed during evaluation"
		}
		until := runs[0].ClosedAt.Add(config.Duration(tr.Cooldown, 0))
		if now.Before(until) {
			return "cooldown until " + until.UTC().Format(time.RFC3339)
		}
	}
	return ""
}

// stallSubjectHeld says why a stall may not fire from any trigger (#704):
// another trigger's run for the same stall is open, or one of them ended
// in an escalation that is still open. Keyed on the stall's signal and
// subject (meta.trigger.stall), not the trigger or its key prefix
// (JOB-95), so two packs watching one subject start one check.
func (c *Controller) stallSubjectHeld(ctx context.Context, s stall) (string, error) {
	runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, MetaPack: triggerPack, MetaKey: stallMetaKey, MetaValue: s.id()})
	if err != nil {
		return "", err
	}
	// Runs started before meta.trigger.stall have only the default key.
	legacy, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, Key: s.key()})
	if err != nil {
		return "", err
	}
	seen := map[string]bool{}
	for _, r := range append(runs, legacy...) {
		if seen[r.ID] || !isStallRun(r) {
			continue
		}
		seen[r.ID] = true
		if r.Status != store.StatusClosed {
			return "run " + r.ID + " is open", nil
		}
		esc, err := c.store.List(ctx, store.Filter{Label: needsHumanLabel, MetaPack: runMetaPack, MetaKey: "id", MetaValue: r.ID})
		if err != nil {
			return "", err
		}
		for _, e := range esc {
			if e.Status != store.StatusClosed {
				return "escalation " + e.ID + " of run " + r.ID + " is open", nil
			}
		}
	}
	return "", nil
}

// isStallRun: the run was started by a stall trigger. fireStall sets
// meta.trigger.stall on a stall-keyed run (any key_prefix, JOB-95), and
// meta.trigger.key is <trigger>/stall:<signal>:<subject> on runs with
// the default prefix, old ones and those with their own run.key.
func isStallRun(r store.Ticket) bool {
	if metaString(r.Meta, triggerPack, stallMetaKey) != "" {
		return true
	}
	_, k, ok := strings.Cut(metaString(r.Meta, triggerPack, "key"), "/")
	return ok && strings.HasPrefix(k, config.DefaultStallKeyPrefix+":")
}

// fireStall starts the trigger's workflow for one stall with key
// <key_prefix>:<signal>:<subject>, or the trigger's run key filled with
// the stall vars. existing is true when an open run already had that key
// (one started by hand or by another run, JOB-33).
func (c *Controller) fireStall(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, s stall) (run store.Ticket, existing bool, err error) {
	tv := map[string]any{config.StallSignalVar: s.signal, config.StallSubjectVar: s.subject, config.StallDetailVar: s.detail}
	key := s.keyWith(tr.StallKeyPrefix())
	p := rpc.RunStartParams{Workflow: tr.Run.Workflow, Vars: map[string]string{}, Key: key}
	if tr.Run.Key != "" {
		if p.Key, err = tr.Run.Fill(tr.Run.Key, tv); err != nil {
			return store.Ticket{}, false, fmt.Errorf("run key: %w", err)
		}
	}
	for k, v := range tr.Run.Vars {
		out, err := tr.Run.Fill(v, tv)
		if err != nil {
			return store.Ticket{}, false, fmt.Errorf("run var %s: %w", k, err)
		}
		p.Vars[k] = out
	}
	m := map[string]string{"name": name, "key": name + "/" + key}
	if tr.Run.Key == "" {
		// The #704 hold covers runs keyed on the stall. A run with its
		// own key (pr-<n>, JOB-33) never took part: leave it out.
		m[stallMetaKey] = s.id()
	}
	res, err := c.startRun(ctx, cfg, scope{admin: true, actor: "controller"}, p, map[string]any{triggerPack: m})
	if err != nil {
		return store.Ticket{}, false, err
	}
	rr := res.(rpc.RunResult)
	return rr.Run, rr.Existing, nil
}

// detectStalls evaluates every signal of the trigger, skipping subjects
// that belong to the trigger's own workflow. An event re-evaluates all
// of them; only a check uses its fields.
func (c *Controller) detectStalls(ctx context.Context, cfg *config.Config, tr config.Trigger, ev triggerEvent) ([]stall, error) {
	own, err := c.workflowTickets(ctx, tr.StallWorkflow())
	if err != nil {
		return nil, err
	}
	now := c.now()
	var out []stall
	pending := false
	for _, sig := range tr.When.Stall {
		after := config.Duration(sig.After, 0)
		var found []stall
		var err error
		switch sig.Signal {
		case config.StallReadyUnclaimed:
			found, err = c.stallReadyUnclaimed(ctx, cfg, now, after)
		case config.StallPoolQuarantined:
			found = c.stallPoolQuarantined(cfg)
		case config.StallStepStalled:
			found, err = c.stallSteps(ctx, now, after)
		case config.StallSessionSilent:
			found, err = c.stallSessions(ctx, now, after)
		case config.StallNeedsHuman:
			found, err = c.stallNeedsHuman(ctx, cfg, now, after, tr.StallWorkflow())
		case config.StallBlocked:
			found, err = c.stallBlocked(ctx, now, after)
		case config.StallResourceWait:
			found, err = c.stallResourceWait(ctx, now, after)
		case config.StallIngestWorker, config.StallIngestBacklog:
			found, err = c.stallIngest(ctx, cfg, sig.Signal)
		case config.StallLinearIssueNoRun:
			found, err = c.stallLinearIssues(ctx, cfg, sig)
		default:
			found, err = c.stallCheck(ctx, tr, sig, ev)
		}
		if errors.Is(err, errCheckPending) {
			// Go on: every check of the trigger is collected in one pass,
			// so they run at the same time (JOB-109).
			pending = true
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("signal %s: %w", sig.Signal, err)
		}
		for _, s := range found {
			if !own.has(s.subject) {
				out = append(out, s)
			}
		}
	}
	if pending {
		return nil, errCheckPending
	}
	return out, nil
}

// ownTickets is the runs of one workflow; a ticket belongs to them when it
// is one, is a step of one, or is an escalation of one.
type ownTickets struct {
	runs map[string]bool
	ids  map[string]bool // runs, steps and escalations
}

func (o ownTickets) has(id string) bool { return o.ids[id] }

func (c *Controller) workflowTickets(ctx context.Context, workflow string) (ownTickets, error) {
	o := ownTickets{runs: map[string]bool{}, ids: map[string]bool{}}
	runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun})
	if err != nil {
		return o, err
	}
	for _, r := range runs {
		f, err := store.DecodeFields[store.RunFields](r)
		if err != nil || f.Workflow != workflow {
			continue
		}
		o.runs[r.ID], o.ids[r.ID] = true, true
		if r.Status == store.StatusClosed {
			continue
		}
		steps, err := store.Children(ctx, c.store, r.ID)
		if err != nil {
			return o, err
		}
		for _, s := range steps {
			o.ids[s.ID] = true
		}
	}
	esc, err := c.store.List(ctx, store.Filter{Label: needsHumanLabel})
	if err != nil {
		return o, err
	}
	for _, t := range esc {
		if o.runs[metaString(t.Meta, runMetaPack, "id")] {
			o.ids[t.ID] = true
		}
	}
	return o, nil
}

// metaString is meta[pack][key] when it is a string.
func metaString(raw json.RawMessage, pack, key string) string {
	var m map[string]map[string]any
	if json.Unmarshal(raw, &m) != nil {
		return ""
	}
	s, _ := m[pack][key].(string)
	return s
}

// lastActivity is the newest of a ticket's creation, history and
// comments. Heartbeats bump updated_at but write no history, so a held
// ticket nobody works on still ages.
func (c *Controller) lastActivity(ctx context.Context, t store.Ticket) (time.Time, error) {
	last := t.CreatedAt
	hist, err := store.History(ctx, c.store, t.ID)
	if err != nil {
		return last, err
	}
	for _, h := range hist {
		if h.At.After(last) {
			last = h.At
		}
	}
	cs, err := store.Comments(ctx, c.store, t.ID)
	if err != nil {
		return last, err
	}
	for _, cm := range cs {
		if cm.CreatedAt.After(last) {
			last = cm.CreatedAt
		}
	}
	// A deferred ticket waits from its defer time, not before.
	if t.DeferUntil.After(last) {
		last = t.DeferUntil
	}
	return last, nil
}

func ago(now, t time.Time) string { return now.Sub(t).Round(time.Second).String() }

// waitsOnQuestion: the ticket's last report is waiting on a question that
// is still open. The question itself is a needs-human ticket, so the
// needs_human signal covers it and the step or session is not stalled.
func (c *Controller) waitsOnQuestion(ctx context.Context, t store.Ticket) bool {
	rm, ok := reportOf(t)
	if !ok || rm.State != rpc.ReportWaiting || rm.Question == "" {
		return false
	}
	q, err := c.store.Get(ctx, rm.Question)
	return err == nil && q.Status != store.StatusClosed
}

// lastUnblocked moves last to the newest close of a ticket blocking t: a
// ready ticket waits from when its last blocker closed, and nothing is
// written on t itself when that happens (#235).
func (c *Controller) lastUnblocked(ctx context.Context, t store.Ticket, last time.Time) (time.Time, error) {
	edges, err := store.Edges(ctx, c.store, t.ID)
	if err != nil {
		return last, err
	}
	for _, e := range edges {
		if e.From != t.ID || !e.Kind.Blocking() {
			continue
		}
		b, err := c.store.Get(ctx, e.To)
		if err != nil {
			return last, err
		}
		if b.ClosedAt.After(last) {
			last = b.ClosedAt
		}
	}
	return last, nil
}

// stallReadyUnclaimed: ready tickets in a configured pool nobody claimed
// for after. Chat pools are left out: their tickets wait for a person.
func (c *Controller) stallReadyUnclaimed(ctx context.Context, cfg *config.Config, now time.Time, after time.Duration) ([]stall, error) {
	var out []stall
	for _, proj := range slices.Sorted(maps.Keys(cfg.Projects)) {
		ts, err := c.store.List(ctx, store.Filter{Project: proj, Ready: true, Limit: store.DefaultLimit})
		if err != nil {
			return nil, err
		}
		for _, t := range ts {
			p, ok := cfg.Pools[t.Pool]
			if !ok || p.Chat() {
				continue // orphaned tickets are reported by reportOrphans
			}
			if slices.Contains(t.Labels, needsHumanLabel) || slices.Contains(t.Labels, blockedLabel) {
				continue // their own signals cover them
			}
			last, err := c.lastActivity(ctx, t)
			if err != nil {
				return nil, err
			}
			if last, err = c.lastUnblocked(ctx, t, last); err != nil {
				return nil, err
			}
			if now.Sub(last) >= after {
				out = append(out, stall{config.StallReadyUnclaimed, t.ID, fmt.Sprintf("ready in pool %s, unclaimed for %s", t.Pool, ago(now, last))})
			}
		}
	}
	return out, nil
}

// stallPoolQuarantined: pools paused for crash-looping sessions. The
// subject carries the pause end, so each quarantine is its own stall and
// max_runs does not carry over from the last one.
func (c *Controller) stallPoolQuarantined(cfg *config.Config) []stall {
	var out []stall
	for _, p := range slices.Sorted(maps.Keys(cfg.Pools)) {
		if c.poolPaused(p) {
			c.mu.Lock()
			until := c.poolPausedUntil[p]
			_, permanent := c.poolQuarantined[p]
			why := c.poolPauseWhy(p)
			c.mu.Unlock()
			if permanent {
				continue // already escalated as "Pool <p> quarantined" (#146)
			}
			out = append(out, stall{config.StallPoolQuarantined, p + "@" + until.UTC().Format("20060102T150405Z"), "pool " + p + ": sessions crash-looping, paused until " + until.UTC().Format(time.RFC3339) + ": " + why})
		}
	}
	return out
}

// stallSteps: the current step of a running run is held but has had no
// history or comment for after. Open steps are ready_unclaimed's (or
// deferred or blocked on purpose).
func (c *Controller) stallSteps(ctx context.Context, now time.Time, after time.Duration) ([]stall, error) {
	runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, State: store.RunRunning, Order: store.OrderNewest})
	if err != nil {
		return nil, err
	}
	var out []stall
	for _, r := range runs {
		if r.Status == store.StatusClosed {
			continue
		}
		steps, err := store.Children(ctx, c.store, r.ID)
		if err != nil {
			return nil, err
		}
		for _, s := range steps {
			if s.Status == store.StatusClosed {
				continue
			}
			if s.Status == store.StatusInProgress && !c.waitsOnQuestion(ctx, s) {
				last, err := c.lastActivity(ctx, s)
				if err != nil {
					return nil, err
				}
				if now.Sub(last) >= after {
					out = append(out, stall{config.StallStepStalled, s.ID, fmt.Sprintf("step of run %s held by %s, no progress for %s", r.ID, s.Holder, ago(now, last))})
				}
			}
			break // only the first unfinished step
		}
	}
	return out, nil
}

// stallSessions: open sessions holding a ticket whose journal has been
// quiet for after (counted from the claim when the journal is older).
func (c *Controller) stallSessions(ctx context.Context, now time.Time, after time.Duration) ([]stall, error) {
	sessions, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen})
	if err != nil {
		return nil, err
	}
	var out []stall
	for _, s := range sessions {
		held, err := c.store.List(ctx, store.Filter{Holder: s.ID, Status: store.StatusInProgress, Limit: 1})
		if err != nil {
			return nil, err
		}
		if len(held) == 0 {
			continue // idle sessions are quiet by design
		}
		if c.waitsOnQuestion(ctx, held[0]) {
			continue // waiting for an answer is quiet by design
		}
		last := s.CreatedAt
		if _, at := host.JournalTail(c.root.SessionDir(s.ID)); at.After(last) {
			last = at
		}
		hl, err := c.lastActivity(ctx, held[0])
		if err != nil {
			return nil, err
		}
		if hl.After(last) {
			last = hl
		}
		if now.Sub(last) >= after {
			out = append(out, stall{config.StallSessionSilent, s.ID, fmt.Sprintf("holds %s, no journal output for %s", held[0].ID, ago(now, last))})
		}
	}
	return out, nil
}

// stallNeedsHuman: needs_human runs, and open needs-human tickets that are
// not a run's escalation (the run covers those), older than after. A run
// of another pack instance is left to that instance's trigger (JOB-17).
func (c *Controller) stallNeedsHuman(ctx context.Context, cfg *config.Config, now time.Time, after time.Duration, workflow string) ([]stall, error) {
	runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, State: store.RunNeedsHuman, Order: store.OrderNewest})
	if err != nil {
		return nil, err
	}
	var out []stall
	for _, r := range runs {
		if r.Status == store.StatusClosed || isStallRun(r) {
			// A stall check that needs a human already escalated: a check
			// on it would only add another escalation (#704).
			continue
		}
		if f, err := store.DecodeFields[store.RunFields](r); err == nil && !sameInstance(cfg, f.Workflow, workflow) {
			continue
		}
		last, err := c.lastActivity(ctx, r)
		if err != nil {
			return nil, err
		}
		if now.Sub(last) >= after {
			out = append(out, stall{config.StallNeedsHuman, r.ID, fmt.Sprintf("run %s needs_human for %s", r.Title, ago(now, last))})
		}
	}
	ts, err := c.labeledOpen(ctx, needsHumanLabel)
	if err != nil {
		return nil, err
	}
	for _, t := range ts {
		if metaString(t.Meta, runMetaPack, "id") != "" {
			continue
		}
		last, err := c.lastActivity(ctx, t)
		if err != nil {
			return nil, err
		}
		if now.Sub(last) >= after {
			out = append(out, stall{config.StallNeedsHuman, t.ID, fmt.Sprintf("ticket labeled needs-human, unchanged for %s", ago(now, last))})
		}
	}
	return out, nil
}

// sameInstance: workflows a and b come from the same packs entry, or one
// of them comes from none (declared in jobshop.yaml itself), so two
// instances of one pack do not both check a run.
func sameInstance(cfg *config.Config, a, b string) bool {
	ia, ib := -1, -1
	for i, in := range cfg.Instances {
		if slices.Contains(in.Workflows, a) {
			ia = i
		}
		if slices.Contains(in.Workflows, b) {
			ib = i
		}
	}
	return ia < 0 || ib < 0 || ia == ib
}

// stallBlocked: open tickets labeled blocked, unchanged for after. This is
// the claimed blocker blocker-check used to re-check.
func (c *Controller) stallBlocked(ctx context.Context, now time.Time, after time.Duration) ([]stall, error) {
	ts, err := c.labeledOpen(ctx, blockedLabel)
	if err != nil {
		return nil, err
	}
	var out []stall
	for _, t := range ts {
		last, err := c.lastActivity(ctx, t)
		if err != nil {
			return nil, err
		}
		if now.Sub(last) >= after {
			out = append(out, stall{config.StallBlocked, t.ID, fmt.Sprintf("claims a blocker, unchanged for %s", ago(now, last))})
		}
	}
	return out, nil
}

func (c *Controller) labeledOpen(ctx context.Context, label string) ([]store.Ticket, error) {
	var out []store.Ticket
	for _, st := range []store.Status{store.StatusOpen, store.StatusInProgress} {
		ts, err := c.store.List(ctx, store.Filter{Label: label, Status: st})
		if err != nil {
			return nil, err
		}
		out = append(out, ts...)
	}
	return out, nil
}

// stallCheck runs a check signal: each stdout line is "<subject>
// [detail]". A non-zero exit means no stall. On a trigger with on:, the
// check is a template over {{.event.<field>}}, empty on a poll (JOB-72).
func (c *Controller) stallCheck(ctx context.Context, tr config.Trigger, sig config.StallSignal, ev triggerEvent) ([]stall, error) {
	check := sig.Check
	if len(tr.On) > 0 {
		var err error
		if check, err = config.Fill(check, map[string]any{config.EventVar: ev.shellVars()}); err != nil {
			return nil, fmt.Errorf("check: %w", err)
		}
	}
	out, why, err := c.runCheck(ctx, config.TriggerWhen{Check: check, CheckTimeout: tr.When.CheckTimeout})
	if err != nil || why != "" {
		return nil, err
	}
	var found []stall
	var busy map[string]bool
	if sig.Signal == config.StallIssueNoRun {
		if busy, err = c.issuesWithOpenRun(ctx, sig.Repo); err != nil {
			return nil, err
		}
	}
	for line := range strings.Lines(out) {
		subject, detail, _ := strings.Cut(strings.TrimSpace(line), " ")
		subject = strings.TrimPrefix(subject, "#")
		if subject == "" || busy[subject] {
			continue
		}
		found = append(found, stall{sig.Signal, subject, strings.TrimSpace(detail)})
	}
	return found, nil
}

// issuesWithOpenRun is the meta.github.issue, as text, of every open run
// in repo. A run of another repo with the same number does not count.
func (c *Controller) issuesWithOpenRun(ctx context.Context, repo string) (map[string]bool, error) {
	out := map[string]bool{}
	for _, st := range []store.Status{store.StatusOpen, store.StatusInProgress} {
		runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, Status: st})
		if err != nil {
			return nil, err
		}
		for _, r := range runs {
			var m struct {
				GitHub struct {
					Repo  string `json:"repo"`
					Issue any    `json:"issue"`
				} `json:"github"`
			}
			if json.Unmarshal(r.Meta, &m) == nil && m.GitHub.Issue != nil && strings.EqualFold(m.GitHub.Repo, repo) {
				out[fmt.Sprint(m.GitHub.Issue)] = true
			}
		}
	}
	return out, nil
}
