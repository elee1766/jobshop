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
	"sync"
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

// finding is one detected finding: a signal on a subject (ticket, run,
// session or pool id, or a check's first word).
type finding struct {
	signal, subject, detail string
}

// id is the finding's identity, <signal>:<subject>, the same whatever key
// prefix a trigger uses (JOB-95). Runs carry it in meta.trigger.finding.
func (s finding) id() string { return s.signal + ":" + s.subject }

// keyWith is the run key <prefix>:<signal>:<subject>.
func (s finding) keyWith(prefix string) string { return prefix + ":" + s.id() }

// findingMetaKey is meta.trigger.finding, <signal>:<subject>, set on the runs
// a scan trigger starts with a scan key (no run.key).
const findingMetaKey = "finding"

// runScanTrigger starts one run per fireable finding of a when.scan
// trigger (#143). ev is the delivery that woke it (JOB-72), nil on a poll.
func (c *Controller) runScanTrigger(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, ev triggerEvent) {
	fire, _, err := c.evalScan(ctx, cfg, name, tr, ev)
	if errors.Is(err, errCheckPending) {
		return // its checks run outside the tick (JOB-109)
	}
	if err != nil {
		c.log.ErrorContext(ctx, "scan trigger", "trigger", name, "err", err)
		return
	}
	prefix := tr.ScanKeyPrefix()
	for _, s := range fire {
		key := s.keyWith(prefix)
		if s.signal == config.SignalLinearIssueNoRun {
			busy, err := c.linearIssueHasOpenRun(ctx, s.subject)
			if err != nil {
				c.log.ErrorContext(ctx, "finding recheck", "trigger", name, "key", key, "err", err)
				continue
			}
			if busy {
				c.log.InfoContext(ctx, "finding skipped: issue got a run during the scan", "trigger", name, "key", key)
				continue
			}
		}
		if tr.Handler != "" {
			c.enqueueFinding(ctx, cfg, name, tr, s)
			continue
		}
		t, existing, err := c.fireFinding(ctx, cfg, name, tr, s)
		if errors.Is(err, errCheckPending) {
			continue // its intake check runs outside the tick (JOB-103)
		}
		if err != nil {
			c.log.ErrorContext(ctx, "finding fire", "trigger", name, "key", key, "err", err)
			continue
		}
		if existing {
			// Already handled by an open run with the same key: not a
			// new finding to learn from.
			c.log.DebugContext(ctx, "finding already has an open run", "trigger", name, "key", key, "run", t.ID)
			continue
		}
		c.log.InfoContext(ctx, "scan trigger fired", "trigger", name, "key", key, "run", t.ID, "delivery", ev["delivery"])
		c.recordFinding(ctx, s, t.ID)
	}
}

// enqueueFinding calls a scan trigger's handler for one fireable finding. The
// invocation key is the finding, its run count and a cooldown slot, so one
// finding is offered once per slot; max_runs, the cooldown and the #704 hold
// were checked by evalScan. The handler's start_run gets
// meta.trigger.finding, so the next evaluation sees its run.
func (c *Controller) enqueueFinding(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, s finding) {
	digest := cfg.HookDigests[tr.Handler]
	if digest == "" {
		return
	}
	runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, Meta: []store.MetaMatch{{Pack: triggerPack, Key: "name", Value: name}, {Pack: triggerPack, Key: findingMetaKey, Value: s.id()}}, Limit: store.NoLimit})
	if err != nil {
		c.log.ErrorContext(ctx, "scan handler", "trigger", name, "err", err)
		return
	}
	event := &hooks.Event{
		Source: hooks.ScanSource, Event: s.signal, Delivery: s.id(), Trigger: name,
		Payload: map[string]any{"signal": s.signal, "subject": s.subject, "detail": s.detail, "key": s.keyWith(tr.ScanKeyPrefix()), "runs": len(runs)},
		Fields:  map[string]any{},
	}
	// A handler that declines a finding is offered it again a cooldown later
	// (30m without one), while the signal persists.
	bucket := c.now().Truncate(config.Duration(tr.Cooldown, 30*time.Minute)).UTC().Format(time.RFC3339)
	key := "scan/" + name + "/" + s.id() + "/" + strconv.Itoa(len(runs)) + "/" + bucket
	if err := c.hooks().Put(ctx, key, hooks.Invocation{Trigger: name, Entry: tr.Handler, Digest: digest, Event: event, Args: tr.Args}); err != nil {
		c.log.ErrorContext(ctx, "enqueue scan hook", "trigger", name, "err", err)
		return
	}
	c.kickHooks()
}

// evalScan detects findings and returns those that may fire now: no open
// run for the key, under max_runs, past the cooldown, and under
// max_open open runs in total (JOB-93), so a backlog of old findings (a
// fresh deploy, a long outage) is worked a few at a time. why explains
// an empty result. ev is passed to the checks, nil on a poll.
func (c *Controller) evalScan(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, ev triggerEvent) (fire []finding, why string, err error) {
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
	maxOpen := tr.ScanMaxOpen()
	if open >= maxOpen {
		return nil, fmt.Sprintf("%d scan runs open (max_open %d)", open, maxOpen), nil
	}
	found, err := c.detectFindings(ctx, cfg, tr, ev)
	if err != nil {
		return nil, "", err
	}
	if len(found) == 0 {
		return nil, "no finding", nil
	}
	now := c.now()
	prefix := tr.ScanKeyPrefix()
	var reasons []string
	for _, s := range found {
		key := s.keyWith(prefix)
		f := store.Filter{Kind: store.KindRun, MetaPack: triggerPack, MetaKey: "key", MetaValue: name + "/" + key, Order: store.OrderNewest}
		if tr.Handler != "" {
			// A handler picks its own run keys; its runs carry the finding id.
			f = store.Filter{Kind: store.KindRun, Meta: []store.MetaMatch{{Pack: triggerPack, Key: "name", Value: name}, {Pack: triggerPack, Key: findingMetaKey, Value: s.id()}}, Order: store.OrderNewest}
		}
		runs, err := c.store.List(ctx, f)
		if err != nil {
			return nil, "", err
		}
		r := findingHeld(runs, tr, now)
		if r == "" {
			if r, err = c.subjectHeld(ctx, s); err != nil {
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

// findingHeld says why a finding with these earlier runs (newest first) may
// not fire now, or "".
func findingHeld(runs []store.Ticket, tr config.Trigger, now time.Time) string {
	if len(runs) == 0 {
		return ""
	}
	if runs[0].Status != store.StatusClosed {
		return "run " + runs[0].ID + " is open"
	}
	if len(runs) >= tr.ScanMaxRuns() {
		return fmt.Sprintf("max_runs %d reached", tr.ScanMaxRuns())
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

// subjectHeld says why a finding may not fire from any trigger (#704):
// another trigger's run for the same finding is open, or one of them ended
// in an escalation that is still open. Keyed on the finding's signal and
// subject (meta.trigger.finding), not the trigger or its key prefix
// (JOB-95), so two packs watching one subject start one check.
func (c *Controller) subjectHeld(ctx context.Context, s finding) (string, error) {
	runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, MetaPack: triggerPack, MetaKey: findingMetaKey, MetaValue: s.id()})
	if err != nil {
		return "", err
	}
	for _, r := range runs {
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

// isScanRun: a scan trigger started the run for a finding, so it carries
// meta.trigger.finding.
func isScanRun(r store.Ticket) bool {
	return metaString(r.Meta, triggerPack, findingMetaKey) != ""
}

// fireFinding starts the trigger's workflow for one finding with key
// <key_prefix>:<signal>:<subject>, or the trigger's run key filled with
// the scan vars. existing is true when an open run already had that key
// (one started by hand or by another run, JOB-33).
func (c *Controller) fireFinding(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, s finding) (run store.Ticket, existing bool, err error) {
	tv := map[string]any{config.ScanSignalVar: s.signal, config.ScanSubjectVar: s.subject, config.ScanDetailVar: s.detail}
	key := s.keyWith(tr.ScanKeyPrefix())
	p := rpc.RunStartParams{Workflow: tr.Run.Workflow, Vars: map[string]string{}, Key: key}
	if tr.Run.Key != "" {
		if p.Key, err = tr.Run.Fill(tr.Run.Key, tv); err != nil {
			return store.Ticket{}, false, fmt.Errorf("run key: %w", err)
		}
	}
	// A finding about an issue titles its run with the issue's (JOB-133).
	if p.Title, err = c.findingTitle(ctx, cfg, tr, s, p.Workflow, p.Key); err != nil {
		return store.Ticket{}, false, err
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
		// The #704 hold covers runs keyed on the finding. A run with its
		// own key (pr-<n>, JOB-33) never took part: leave it out.
		m[findingMetaKey] = s.id()
	}
	res, err := c.startRun(ctx, cfg, scope{admin: true, actor: "controller"}, p, map[string]any{triggerPack: m})
	if err != nil {
		return store.Ticket{}, false, err
	}
	rr := res.(rpc.RunResult)
	return rr.Run, rr.Existing, nil
}

// detectFindings evaluates every signal of the trigger, skipping subjects
// that belong to the trigger's own workflow. An event re-evaluates all
// of them; only a check uses its fields.
func (c *Controller) detectFindings(ctx context.Context, cfg *config.Config, tr config.Trigger, ev triggerEvent) ([]finding, error) {
	own, err := c.workflowTickets(ctx, tr.ScanWorkflow())
	if err != nil {
		return nil, err
	}
	now := c.now()
	var out []finding
	pending := false
	for _, sig := range tr.When.Scan {
		after := config.Duration(sig.After, 0)
		var found []finding
		var err error
		switch sig.Signal {
		case config.SignalReadyUnclaimed:
			found, err = c.signalReadyUnclaimed(ctx, cfg, now, after)
		case config.SignalPoolQuarantined:
			found = c.signalPoolQuarantined(cfg)
		case config.SignalStepIdle:
			found, err = c.signalSteps(ctx, now, after)
		case config.SignalSessionSilent:
			found, err = c.signalSessions(ctx, now, after)
		case config.SignalNeedsHuman:
			found, err = c.signalNeedsHuman(ctx, cfg, now, after, tr.ScanWorkflow())
		case config.SignalBlocked:
			found, err = c.signalBlocked(ctx, now, after)
		case config.SignalResourceWait:
			found, err = c.signalResourceWait(ctx, now, after)
		case config.SignalIngestWorker, config.SignalIngestBacklog:
			found, err = c.signalIngest(ctx, cfg, sig.Signal)
		case config.SignalLinearIssueNoRun:
			found, err = c.signalLinearIssues(ctx, cfg, sig)
		default:
			if sig.Script != "" {
				found, err = c.signalScript(ctx, cfg, sig)
			} else {
				found, err = c.signalCheck(ctx, tr, sig, ev)
			}
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
// needs_human signal covers it and the step or session is not idle.
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

// signalReadyUnclaimed: ready tickets in a configured pool nobody claimed
// for after. Chat pools are left out: their tickets wait for a person.
func (c *Controller) signalReadyUnclaimed(ctx context.Context, cfg *config.Config, now time.Time, after time.Duration) ([]finding, error) {
	var out []finding
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
				out = append(out, finding{config.SignalReadyUnclaimed, t.ID, fmt.Sprintf("ready in pool %s, unclaimed for %s", t.Pool, ago(now, last))})
			}
		}
	}
	return out, nil
}

// signalPoolQuarantined: pools paused for crash-looping sessions. The
// subject carries the pause end, so each quarantine is its own finding and
// max_runs does not carry over from the last one.
func (c *Controller) signalPoolQuarantined(cfg *config.Config) []finding {
	var out []finding
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
			out = append(out, finding{config.SignalPoolQuarantined, p + "@" + until.UTC().Format("20060102T150405Z"), "pool " + p + ": sessions crash-looping, paused until " + until.UTC().Format(time.RFC3339) + ": " + why})
		}
	}
	return out
}

// signalSteps: the current step of a running run is held but has had no
// history or comment for after. Open steps are ready_unclaimed's (or
// deferred or blocked on purpose).
func (c *Controller) signalSteps(ctx context.Context, now time.Time, after time.Duration) ([]finding, error) {
	runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, State: store.RunRunning, Order: store.OrderNewest})
	if err != nil {
		return nil, err
	}
	var out []finding
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
					out = append(out, finding{config.SignalStepIdle, s.ID, fmt.Sprintf("step of run %s held by %s, no progress for %s", r.ID, s.Holder, ago(now, last))})
				}
			}
			break // only the first unfinished step
		}
	}
	return out, nil
}

// signalSessions: open sessions holding a ticket whose journal has been
// quiet for after (counted from the claim when the journal is older).
func (c *Controller) signalSessions(ctx context.Context, now time.Time, after time.Duration) ([]finding, error) {
	sessions, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen})
	if err != nil {
		return nil, err
	}
	var out []finding
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
			out = append(out, finding{config.SignalSessionSilent, s.ID, fmt.Sprintf("holds %s, no journal output for %s", held[0].ID, ago(now, last))})
		}
	}
	return out, nil
}

// signalNeedsHuman: needs_human runs, and open needs-human tickets that are
// not a run's escalation (the run covers those), older than after. A run
// of another pack instance is left to that instance's trigger (JOB-17).
func (c *Controller) signalNeedsHuman(ctx context.Context, cfg *config.Config, now time.Time, after time.Duration, workflow string) ([]finding, error) {
	runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, State: store.RunNeedsHuman, Order: store.OrderNewest})
	if err != nil {
		return nil, err
	}
	var out []finding
	for _, r := range runs {
		if r.Status == store.StatusClosed || isScanRun(r) {
			// A signal check that needs a human already escalated: a check
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
			out = append(out, finding{config.SignalNeedsHuman, r.ID, fmt.Sprintf("run %s needs_human for %s", r.Title, ago(now, last))})
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
			out = append(out, finding{config.SignalNeedsHuman, t.ID, fmt.Sprintf("ticket labeled needs-human, unchanged for %s", ago(now, last))})
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

// signalBlocked: open tickets labeled blocked, unchanged for after. This is
// the claimed blocker blocker-check used to re-check.
func (c *Controller) signalBlocked(ctx context.Context, now time.Time, after time.Duration) ([]finding, error) {
	ts, err := c.labeledOpen(ctx, blockedLabel)
	if err != nil {
		return nil, err
	}
	var out []finding
	for _, t := range ts {
		last, err := c.lastActivity(ctx, t)
		if err != nil {
			return nil, err
		}
		if now.Sub(last) >= after {
			out = append(out, finding{config.SignalBlocked, t.ID, fmt.Sprintf("claims a blocker, unchanged for %s", ago(now, last))})
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

// signalCheck runs a check signal: each stdout line is "<subject>
// [detail]". A non-zero exit means no finding. On a trigger with on:, the
// check is a template over {{.event.<field>}}, empty on a poll (JOB-72).
func (c *Controller) signalCheck(ctx context.Context, tr config.Trigger, sig config.ScanSignal, ev triggerEvent) ([]finding, error) {
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
	found := map[string]string{}
	for line := range strings.Lines(out) {
		subject, detail, _ := strings.Cut(strings.TrimSpace(line), " ")
		found[subject] = strings.TrimSpace(detail)
	}
	return c.findingsFrom(ctx, sig, found)
}

// findingsFrom turns subject -> detail into findings in subject order. An
// issue_no_run issue with an open run in sig.Repo is not a finding.
func (c *Controller) findingsFrom(ctx context.Context, sig config.ScanSignal, found map[string]string) ([]finding, error) {
	var busy map[string]bool
	if sig.Signal == config.SignalIssueNoRun {
		var err error
		if busy, err = c.issuesWithOpenRun(ctx, sig.Repo); err != nil {
			return nil, err
		}
	}
	var out []finding
	for _, raw := range slices.Sorted(maps.Keys(found)) {
		subject := strings.TrimPrefix(raw, "#")
		if subject == "" || strings.ContainsAny(subject, " \t\n") || busy[subject] {
			continue
		}
		out = append(out, finding{sig.Signal, subject, found[raw]})
	}
	return out, nil
}

// signalScriptTimeout bounds one signal script call.
const signalScriptTimeout = 2 * time.Minute

// signalScript runs a script signal: fn(args) returns a dict of subject to
// detail. It is called outside the tick like a check (JOB-109), at most
// once per poll (default DefaultLinearPoll); between calls the last
// answer stands.
func (c *Controller) signalScript(ctx context.Context, cfg *config.Config, sig config.ScanSignal) ([]finding, error) {
	key := "script\x00" + sig.Script + "\x00" + sig.Signal
	poll := config.Duration(sig.Poll, config.DefaultLinearPoll)
	now := c.now()
	c.signalScripts.mu.Lock()
	r, ok := c.signalScripts.cache[key]
	c.signalScripts.mu.Unlock()
	if !ok || !now.Before(r.at.Add(poll)) {
		res, err := offTick(ctx, checkNeed{
			key: key, poller: "signal_script", timeout: signalScriptTimeout,
			run: func(ctx context.Context) checkResult {
				result, err := c.callSync(ctx, cfg, "scan."+sig.Signal, sig.Script, signalScriptTimeout, sig.Args)
				return checkResult{val: result.Output, err: err}
			},
		})
		if err == nil {
			err = res.err
		}
		if err != nil {
			return nil, err
		}
		found := map[string]string{}
		out, _ := res.val.(map[string]any)
		for subject, detail := range out {
			if s, ok := detail.(string); ok {
				found[subject] = s
			} else {
				found[subject] = fmt.Sprint(detail)
			}
		}
		r = signalScriptRead{at: now, found: found}
		c.signalScripts.mu.Lock()
		if c.signalScripts.cache == nil {
			c.signalScripts.cache = map[string]signalScriptRead{}
		}
		c.signalScripts.cache[key] = r
		c.signalScripts.mu.Unlock()
	}
	return c.findingsFrom(ctx, sig, r.found)
}

// signalScripts caches script signal answers between polls.
type signalScripts struct {
	mu    sync.Mutex
	cache map[string]signalScriptRead
}

type signalScriptRead struct {
	at    time.Time
	found map[string]string
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
