package triggers

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

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/hooks"
	"github.com/elee1766/jobshop/pkg/store"
)

// Labels the signals read. The controller uses them too.
const (
	// NeedsHumanLabel marks tickets waiting on the operator.
	NeedsHumanLabel = "needs-human"
	// BlockedLabel marks tickets that claim a blocker.
	BlockedLabel = "blocked"
)

// runPack is meta.run on a run's escalation ticket: {"id": <run>}.
const runPack = "run"

// FindingMetaKey is meta.trigger.finding, <signal>:<subject>, set on the
// runs a scan trigger starts with a scan key (no run.key).
const FindingMetaKey = "finding"

// Finding is one detected finding: a signal on a subject (ticket, run,
// session or pool id, or a check's first word).
type Finding struct {
	Signal, Subject, Detail string
}

// ID is the finding's identity, <signal>:<subject>, the same whatever key
// prefix a trigger uses (JOB-95). Runs carry it in meta.trigger.finding.
func (s Finding) ID() string { return s.Signal + ":" + s.Subject }

// KeyWith is the run key <prefix>:<signal>:<subject>.
func (s Finding) KeyWith(prefix string) string { return prefix + ":" + s.ID() }

// Signals are the scan signals the controller reads from state other
// subsystems own: pool pauses, the ingest queue, Linear, resources and
// signal scripts.
type Signals interface {
	// QuarantinedPools are the pool_quarantined findings.
	QuarantinedPools(cfg *config.Config) []Finding
	// JournalAt is when the session's journal last grew, zero if never.
	JournalAt(session string) time.Time
	// WaitsOnQuestion: the ticket's last report is waiting on a question
	// that is still open. The question itself is a needs-human ticket, so
	// the step or session is not idle.
	WaitsOnQuestion(ctx context.Context, t store.Ticket) bool
	// ResourceWaits are the resource_wait findings.
	ResourceWaits(ctx context.Context, now time.Time, after time.Duration) ([]Finding, error)
	// Ingest are the ingest_worker and ingest_backlog findings.
	Ingest(ctx context.Context, cfg *config.Config, signal string) ([]Finding, error)
	// LinearIssues are the linear_issue_no_run findings; LinearIssueBusy
	// rechecks one issue just before its run starts.
	LinearIssues(ctx context.Context, cfg *config.Config, sig config.ScanSignal) ([]Finding, error)
	LinearIssueBusy(ctx context.Context, key string) (bool, error)
	// Script calls a signal script (name is for logs and spans) and returns
	// its dict.
	Script(ctx context.Context, cfg *config.Config, name, entry string, timeout time.Duration, args map[string]any) (map[string]any, error)
}

// runScan starts one run per fireable finding of a when.scan trigger
// (#143). ev is the delivery that woke it (JOB-72), nil on a poll.
func (e *Engine) runScan(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, ev Event) {
	fire, _, err := e.evalScan(ctx, cfg, name, tr, ev)
	if errors.Is(err, ErrCheckPending) {
		return // its checks run outside the tick (JOB-109)
	}
	if err != nil {
		e.log().ErrorContext(ctx, "scan trigger", "trigger", name, "err", err)
		return
	}
	prefix := tr.ScanKeyPrefix()
	for _, s := range fire {
		key := s.KeyWith(prefix)
		if s.Signal == config.SignalLinearIssueNoRun {
			busy, err := e.host.LinearIssueBusy(ctx, s.Subject)
			if err != nil {
				e.log().ErrorContext(ctx, "finding recheck", "trigger", name, "key", key, "err", err)
				continue
			}
			if busy {
				e.log().InfoContext(ctx, "finding skipped: issue got a run during the scan", "trigger", name, "key", key)
				continue
			}
		}
		if tr.Handler != "" {
			e.enqueueFinding(ctx, cfg, name, tr, s)
			continue
		}
		t, existing, err := e.host.FireFinding(ctx, cfg, name, tr, s)
		if errors.Is(err, ErrCheckPending) {
			continue // its intake check runs outside the tick (JOB-103)
		}
		if err != nil {
			e.log().ErrorContext(ctx, "finding fire", "trigger", name, "key", key, "err", err)
			continue
		}
		if existing {
			// Already handled by an open run with the same key: not a
			// new finding to learn from.
			e.log().DebugContext(ctx, "finding already has an open run", "trigger", name, "key", key, "run", t.ID)
			continue
		}
		e.log().InfoContext(ctx, "scan trigger fired", "trigger", name, "key", key, "run", t.ID, "delivery", ev["delivery"])
		e.host.RecordFinding(ctx, s, t.ID)
	}
}

// enqueueFinding calls a scan trigger's handler for one fireable finding. The
// invocation key is the finding, its run count and a cooldown slot, so one
// finding is offered once per slot; max_runs, the cooldown and the #704 hold
// were checked by evalScan. The handler's start_run gets
// meta.trigger.finding, so the next evaluation sees its run.
func (e *Engine) enqueueFinding(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, s Finding) {
	digest := cfg.HookDigests[tr.Handler]
	if digest == "" {
		return
	}
	runs, err := e.host.Store().List(ctx, store.Filter{Kind: store.KindRun, Meta: []store.MetaMatch{{Pack: Pack, Key: "name", Value: name}, {Pack: Pack, Key: FindingMetaKey, Value: s.ID()}}, Limit: store.NoLimit})
	if err != nil {
		e.log().ErrorContext(ctx, "scan handler", "trigger", name, "err", err)
		return
	}
	event := &hooks.Event{
		Source: hooks.ScanSource, Event: s.Signal, Delivery: s.ID(), Trigger: name,
		Payload: map[string]any{"signal": s.Signal, "subject": s.Subject, "detail": s.Detail, "key": s.KeyWith(tr.ScanKeyPrefix()), "runs": len(runs)},
		Fields:  map[string]any{},
	}
	// A handler that declines a finding is offered it again a cooldown later
	// (30m without one), while the signal persists.
	bucket := e.host.Now().Truncate(config.Duration(tr.Cooldown, 30*time.Minute)).UTC().Format(time.RFC3339)
	key := "scan/" + name + "/" + s.ID() + "/" + strconv.Itoa(len(runs)) + "/" + bucket
	if err := e.host.Hooks().Put(ctx, key, hooks.Invocation{Trigger: name, Entry: tr.Handler, Digest: digest, Event: event, Args: tr.Args}); err != nil {
		e.log().ErrorContext(ctx, "enqueue scan hook", "trigger", name, "err", err)
		return
	}
	e.host.KickHooks()
}

// evalScan detects findings and returns those that may fire now: no open
// run for the key, under max_runs, past the cooldown, and under
// max_open open runs in total (JOB-93), so a backlog of old findings (a
// fresh deploy, a long outage) is worked a few at a time. why explains
// an empty result. ev is passed to the checks, nil on a poll.
func (e *Engine) evalScan(ctx context.Context, cfg *config.Config, name string, tr config.Trigger, ev Event) (fire []Finding, why string, err error) {
	s := e.host.Store()
	mine, err := s.List(ctx, store.Filter{Kind: store.KindRun, MetaPack: Pack, MetaKey: "name", MetaValue: name, Order: store.OrderNewest})
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
	found, err := e.Detect(ctx, cfg, tr, ev)
	if err != nil {
		return nil, "", err
	}
	if len(found) == 0 {
		return nil, "no finding", nil
	}
	now := e.host.Now()
	prefix := tr.ScanKeyPrefix()
	var reasons []string
	for _, f := range found {
		key := f.KeyWith(prefix)
		flt := store.Filter{Kind: store.KindRun, MetaPack: Pack, MetaKey: "key", MetaValue: name + "/" + key, Order: store.OrderNewest}
		if tr.Handler != "" {
			// A handler picks its own run keys; its runs carry the finding id.
			flt = store.Filter{Kind: store.KindRun, Meta: []store.MetaMatch{{Pack: Pack, Key: "name", Value: name}, {Pack: Pack, Key: FindingMetaKey, Value: f.ID()}}, Order: store.OrderNewest}
		}
		runs, err := s.List(ctx, flt)
		if err != nil {
			return nil, "", err
		}
		r := findingHeld(runs, tr, now)
		if r == "" {
			if r, err = e.subjectHeld(ctx, f); err != nil {
				return nil, "", err
			}
		}
		if r != "" {
			reasons = append(reasons, key+": "+r)
			continue
		}
		fire = append(fire, f)
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
func (e *Engine) subjectHeld(ctx context.Context, f Finding) (string, error) {
	s := e.host.Store()
	runs, err := s.List(ctx, store.Filter{Kind: store.KindRun, MetaPack: Pack, MetaKey: FindingMetaKey, MetaValue: f.ID()})
	if err != nil {
		return "", err
	}
	for _, r := range runs {
		if r.Status != store.StatusClosed {
			return "run " + r.ID + " is open", nil
		}
		esc, err := s.List(ctx, store.Filter{Label: NeedsHumanLabel, MetaPack: runPack, MetaKey: "id", MetaValue: r.ID})
		if err != nil {
			return "", err
		}
		for _, x := range esc {
			if x.Status != store.StatusClosed {
				return "escalation " + x.ID + " of run " + r.ID + " is open", nil
			}
		}
	}
	return "", nil
}

// IsScanRun reports whether a scan trigger started the run for a finding,
// so it carries meta.trigger.finding.
func IsScanRun(r store.Ticket) bool {
	return MetaString(r.Meta, Pack, FindingMetaKey) != ""
}

// Detect evaluates every signal of the trigger, skipping subjects that
// belong to the trigger's own workflow. An event re-evaluates all of them;
// only a check uses its fields.
func (e *Engine) Detect(ctx context.Context, cfg *config.Config, tr config.Trigger, ev Event) ([]Finding, error) {
	own, err := e.workflowTickets(ctx, tr.ScanWorkflow())
	if err != nil {
		return nil, err
	}
	now := e.host.Now()
	var out []Finding
	pending := false
	for _, sig := range tr.When.Scan {
		after := config.Duration(sig.After, 0)
		var found []Finding
		var err error
		switch sig.Signal {
		case config.SignalReadyUnclaimed:
			found, err = e.signalReadyUnclaimed(ctx, cfg, now, after)
		case config.SignalPoolQuarantined:
			found = e.host.QuarantinedPools(cfg)
		case config.SignalStepIdle:
			found, err = e.signalSteps(ctx, now, after)
		case config.SignalSessionSilent:
			found, err = e.signalSessions(ctx, now, after)
		case config.SignalNeedsHuman:
			found, err = e.signalNeedsHuman(ctx, cfg, now, after, tr.ScanWorkflow())
		case config.SignalBlocked:
			found, err = e.signalBlocked(ctx, now, after)
		case config.SignalResourceWait:
			found, err = e.host.ResourceWaits(ctx, now, after)
		case config.SignalIngestWorker, config.SignalIngestBacklog:
			found, err = e.host.Ingest(ctx, cfg, sig.Signal)
		case config.SignalLinearIssueNoRun:
			found, err = e.host.LinearIssues(ctx, cfg, sig)
		default:
			if sig.Script != "" {
				found, err = e.signalScript(ctx, cfg, sig)
			} else {
				found, err = e.signalCheck(ctx, tr, sig, ev)
			}
		}
		if errors.Is(err, ErrCheckPending) {
			// Go on: every check of the trigger is collected in one pass,
			// so they run at the same time (JOB-109).
			pending = true
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("signal %s: %w", sig.Signal, err)
		}
		for _, s := range found {
			if !own.has(s.Subject) {
				out = append(out, s)
			}
		}
	}
	if pending {
		return nil, ErrCheckPending
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

func (e *Engine) workflowTickets(ctx context.Context, workflow string) (ownTickets, error) {
	s := e.host.Store()
	o := ownTickets{runs: map[string]bool{}, ids: map[string]bool{}}
	runs, err := s.List(ctx, store.Filter{Kind: store.KindRun})
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
		steps, err := store.Children(ctx, s, r.ID)
		if err != nil {
			return o, err
		}
		for _, st := range steps {
			o.ids[st.ID] = true
		}
	}
	esc, err := s.List(ctx, store.Filter{Label: NeedsHumanLabel})
	if err != nil {
		return o, err
	}
	for _, t := range esc {
		if o.runs[MetaString(t.Meta, runPack, "id")] {
			o.ids[t.ID] = true
		}
	}
	return o, nil
}

// MetaString is meta[pack][key] when it is a string.
func MetaString(raw json.RawMessage, pack, key string) string {
	var m map[string]map[string]any
	if json.Unmarshal(raw, &m) != nil {
		return ""
	}
	s, _ := m[pack][key].(string)
	return s
}

// LastActivity is the newest of a ticket's creation, history and
// comments. Heartbeats bump updated_at but write no history, so a held
// ticket nobody works on still ages.
func LastActivity(ctx context.Context, s store.Store, t store.Ticket) (time.Time, error) {
	last := t.CreatedAt
	hist, err := store.History(ctx, s, t.ID)
	if err != nil {
		return last, err
	}
	for _, h := range hist {
		if h.At.After(last) {
			last = h.At
		}
	}
	cs, err := store.Comments(ctx, s, t.ID)
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

// Ago is the time from t to now, rounded to seconds, for finding details.
func Ago(now, t time.Time) string { return now.Sub(t).Round(time.Second).String() }

// LastUnblocked moves last to the newest close of a ticket blocking t: a
// ready ticket waits from when its last blocker closed, and nothing is
// written on t itself when that happens (#235).
func LastUnblocked(ctx context.Context, s store.Store, t store.Ticket, last time.Time) (time.Time, error) {
	edges, err := store.Edges(ctx, s, t.ID)
	if err != nil {
		return last, err
	}
	for _, e := range edges {
		if e.From != t.ID || !e.Kind.Blocking() {
			continue
		}
		b, err := s.Get(ctx, e.To)
		if err != nil {
			return last, err
		}
		if b.ClosedAt.After(last) {
			last = b.ClosedAt
		}
	}
	return last, nil
}

// LabeledOpen lists the open and in-progress tickets with label.
func LabeledOpen(ctx context.Context, s store.Store, label string) ([]store.Ticket, error) {
	var out []store.Ticket
	for _, st := range []store.Status{store.StatusOpen, store.StatusInProgress} {
		ts, err := s.List(ctx, store.Filter{Label: label, Status: st})
		if err != nil {
			return nil, err
		}
		out = append(out, ts...)
	}
	return out, nil
}

// signalReadyUnclaimed: ready tickets in a configured pool nobody claimed
// for after. Chat pools are left out: their tickets wait for a person.
func (e *Engine) signalReadyUnclaimed(ctx context.Context, cfg *config.Config, now time.Time, after time.Duration) ([]Finding, error) {
	s := e.host.Store()
	var out []Finding
	for _, proj := range slices.Sorted(maps.Keys(cfg.Projects)) {
		ts, err := s.List(ctx, store.Filter{Project: proj, Ready: true, Limit: store.DefaultLimit})
		if err != nil {
			return nil, err
		}
		for _, t := range ts {
			p, ok := cfg.Pools[t.Pool]
			if !ok || p.Chat() {
				continue // orphaned tickets are reported by reportOrphans
			}
			if slices.Contains(t.Labels, NeedsHumanLabel) || slices.Contains(t.Labels, BlockedLabel) {
				continue // their own signals cover them
			}
			last, err := LastActivity(ctx, s, t)
			if err != nil {
				return nil, err
			}
			if last, err = LastUnblocked(ctx, s, t, last); err != nil {
				return nil, err
			}
			if now.Sub(last) >= after {
				out = append(out, Finding{config.SignalReadyUnclaimed, t.ID, fmt.Sprintf("ready in pool %s, unclaimed for %s", t.Pool, Ago(now, last))})
			}
		}
	}
	return out, nil
}

// signalSteps: the current step of a running run is held but has had no
// history or comment for after. Open steps are ready_unclaimed's (or
// deferred or blocked on purpose).
func (e *Engine) signalSteps(ctx context.Context, now time.Time, after time.Duration) ([]Finding, error) {
	s := e.host.Store()
	runs, err := s.List(ctx, store.Filter{Kind: store.KindRun, State: store.RunRunning, Order: store.OrderNewest})
	if err != nil {
		return nil, err
	}
	var out []Finding
	for _, r := range runs {
		if r.Status == store.StatusClosed {
			continue
		}
		steps, err := store.Children(ctx, s, r.ID)
		if err != nil {
			return nil, err
		}
		for _, st := range steps {
			if st.Status == store.StatusClosed {
				continue
			}
			if st.Status == store.StatusInProgress && !e.host.WaitsOnQuestion(ctx, st) {
				last, err := LastActivity(ctx, s, st)
				if err != nil {
					return nil, err
				}
				if now.Sub(last) >= after {
					out = append(out, Finding{config.SignalStepIdle, st.ID, fmt.Sprintf("step of run %s held by %s, no progress for %s", r.ID, st.Holder, Ago(now, last))})
				}
			}
			break // only the first unfinished step
		}
	}
	return out, nil
}

// signalSessions: open sessions holding a ticket whose journal has been
// quiet for after (counted from the claim when the journal is older).
func (e *Engine) signalSessions(ctx context.Context, now time.Time, after time.Duration) ([]Finding, error) {
	s := e.host.Store()
	sessions, err := s.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen})
	if err != nil {
		return nil, err
	}
	var out []Finding
	for _, sess := range sessions {
		held, err := s.List(ctx, store.Filter{Holder: sess.ID, Status: store.StatusInProgress, Limit: 1})
		if err != nil {
			return nil, err
		}
		if len(held) == 0 {
			continue // idle sessions are quiet by design
		}
		if e.host.WaitsOnQuestion(ctx, held[0]) {
			continue // waiting for an answer is quiet by design
		}
		last := sess.CreatedAt
		if at := e.host.JournalAt(sess.ID); at.After(last) {
			last = at
		}
		hl, err := LastActivity(ctx, s, held[0])
		if err != nil {
			return nil, err
		}
		if hl.After(last) {
			last = hl
		}
		if now.Sub(last) >= after {
			out = append(out, Finding{config.SignalSessionSilent, sess.ID, fmt.Sprintf("holds %s, no journal output for %s", held[0].ID, Ago(now, last))})
		}
	}
	return out, nil
}

// signalNeedsHuman: needs_human runs, and open needs-human tickets that are
// not a run's escalation (the run covers those), older than after. A run
// of another pack instance is left to that instance's trigger (JOB-17).
func (e *Engine) signalNeedsHuman(ctx context.Context, cfg *config.Config, now time.Time, after time.Duration, workflow string) ([]Finding, error) {
	s := e.host.Store()
	runs, err := s.List(ctx, store.Filter{Kind: store.KindRun, State: store.RunNeedsHuman, Order: store.OrderNewest})
	if err != nil {
		return nil, err
	}
	var out []Finding
	for _, r := range runs {
		if r.Status == store.StatusClosed || IsScanRun(r) {
			// A signal check that needs a human already escalated: a check
			// on it would only add another escalation (#704).
			continue
		}
		if f, err := store.DecodeFields[store.RunFields](r); err == nil && !sameInstance(cfg, f.Workflow, workflow) {
			continue
		}
		last, err := LastActivity(ctx, s, r)
		if err != nil {
			return nil, err
		}
		if now.Sub(last) >= after {
			out = append(out, Finding{config.SignalNeedsHuman, r.ID, fmt.Sprintf("run %s needs_human for %s", r.Title, Ago(now, last))})
		}
	}
	ts, err := LabeledOpen(ctx, s, NeedsHumanLabel)
	if err != nil {
		return nil, err
	}
	for _, t := range ts {
		if MetaString(t.Meta, runPack, "id") != "" {
			continue
		}
		last, err := LastActivity(ctx, s, t)
		if err != nil {
			return nil, err
		}
		if now.Sub(last) >= after {
			out = append(out, Finding{config.SignalNeedsHuman, t.ID, fmt.Sprintf("ticket labeled needs-human, unchanged for %s", Ago(now, last))})
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
func (e *Engine) signalBlocked(ctx context.Context, now time.Time, after time.Duration) ([]Finding, error) {
	s := e.host.Store()
	ts, err := LabeledOpen(ctx, s, BlockedLabel)
	if err != nil {
		return nil, err
	}
	var out []Finding
	for _, t := range ts {
		last, err := LastActivity(ctx, s, t)
		if err != nil {
			return nil, err
		}
		if now.Sub(last) >= after {
			out = append(out, Finding{config.SignalBlocked, t.ID, fmt.Sprintf("claims a blocker, unchanged for %s", Ago(now, last))})
		}
	}
	return out, nil
}

// signalCheck runs a check signal: each stdout line is "<subject>
// [detail]". A non-zero exit means no finding. On a trigger with on:, the
// check is a template over {{.event.<field>}}, empty on a poll (JOB-72).
func (e *Engine) signalCheck(ctx context.Context, tr config.Trigger, sig config.ScanSignal, ev Event) ([]Finding, error) {
	check := sig.Check
	if len(tr.On) > 0 {
		var err error
		if check, err = config.Fill(check, map[string]any{config.EventVar: ev.ShellVars()}); err != nil {
			return nil, fmt.Errorf("check: %w", err)
		}
	}
	out, why, err := e.runCheck(ctx, config.TriggerWhen{Check: check, CheckTimeout: tr.When.CheckTimeout})
	if err != nil || why != "" {
		return nil, err
	}
	found := map[string]string{}
	for line := range strings.Lines(out) {
		subject, detail, _ := strings.Cut(strings.TrimSpace(line), " ")
		found[subject] = strings.TrimSpace(detail)
	}
	return e.findingsFrom(ctx, sig, found)
}

// findingsFrom turns subject -> detail into findings in subject order. An
// issue_no_run issue with an open run in sig.Repo is not a finding.
func (e *Engine) findingsFrom(ctx context.Context, sig config.ScanSignal, found map[string]string) ([]Finding, error) {
	var busy map[string]bool
	if sig.Signal == config.SignalIssueNoRun {
		var err error
		if busy, err = e.issuesWithOpenRun(ctx, sig.Repo); err != nil {
			return nil, err
		}
	}
	var out []Finding
	for _, raw := range slices.Sorted(maps.Keys(found)) {
		subject := strings.TrimPrefix(raw, "#")
		if subject == "" || strings.ContainsAny(subject, " \t\n") || busy[subject] {
			continue
		}
		out = append(out, Finding{sig.Signal, subject, found[raw]})
	}
	return out, nil
}

// signalScriptTimeout bounds one signal script call.
const signalScriptTimeout = 2 * time.Minute

// signalScript runs a script signal: fn(args) returns a dict of subject to
// detail. It is called outside the tick like a check (JOB-109), at most
// once per poll (default DefaultLinearPoll); between calls the last
// answer stands.
func (e *Engine) signalScript(ctx context.Context, cfg *config.Config, sig config.ScanSignal) ([]Finding, error) {
	key := scriptSignalKey(sig)
	poll := config.Duration(sig.Poll, config.DefaultLinearPoll)
	now := e.host.Now()
	e.scripts.mu.Lock()
	r, ok := e.scripts.cache[key]
	e.scripts.mu.Unlock()
	if !ok || !now.Before(r.at.Add(poll)) {
		res, err := OffTick(ctx, Need{
			Key: key, Poller: "signal_script", Timeout: signalScriptTimeout,
			Run: func(ctx context.Context) Result {
				out, err := e.host.Script(ctx, cfg, "scan."+sig.Signal, sig.Script, signalScriptTimeout, sig.Args)
				return Result{Val: out, Err: err}
			},
		})
		if err == nil {
			err = res.Err
		}
		if err != nil {
			return nil, err
		}
		found := map[string]string{}
		out, _ := res.Val.(map[string]any)
		for subject, detail := range out {
			if s, ok := detail.(string); ok {
				found[subject] = s
			} else {
				found[subject] = fmt.Sprint(detail)
			}
		}
		r = scriptRead{at: now, found: found}
		e.scripts.mu.Lock()
		if e.scripts.cache == nil {
			e.scripts.cache = map[string]scriptRead{}
		}
		e.scripts.cache[key] = r
		e.scripts.mu.Unlock()
	}
	return e.findingsFrom(ctx, sig, r.found)
}

// scriptSignalKey names one script signal call: the script and signal with
// the repo and args it runs on. Triggers that share a script (the self-dev
// pack runs the github pack's issue_no_run for another repo) make separate
// calls: keyed on the script alone, one trigger took the other's cached
// answer and fired findings that named issues of the other repo.
func scriptSignalKey(sig config.ScanSignal) string {
	args, _ := json.Marshal(sig.Args) // map keys are sorted
	return "script\x00" + sig.Script + "\x00" + sig.Signal + "\x00" + sig.Repo + "\x00" + string(args)
}

// scriptCache caches script signal answers between polls.
type scriptCache struct {
	mu    sync.Mutex
	cache map[string]scriptRead
}

type scriptRead struct {
	at    time.Time
	found map[string]string
}

// issuesWithOpenRun is the meta.github.issue, as text, of every open run
// in repo. A run of another repo with the same number does not count.
func (e *Engine) issuesWithOpenRun(ctx context.Context, repo string) (map[string]bool, error) {
	out := map[string]bool{}
	for _, st := range []store.Status{store.StatusOpen, store.StatusInProgress} {
		runs, err := e.host.Store().List(ctx, store.Filter{Kind: store.KindRun, Status: st})
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
