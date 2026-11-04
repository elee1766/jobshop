package control

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	host "github.com/elee1766/acphost"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// stallMaxOpen bounds the open runs of one stall trigger, so a backlog of
// old stalls (a fresh deploy, a long outage) is worked a few at a time.
const stallMaxOpen = 3

// needsHumanLabel marks tickets waiting on the operator.
const needsHumanLabel = "needs-human"

// blockedLabel marks tickets that claim a blocker.
const blockedLabel = "blocked"

// stall is one detected stall: a signal on a subject (ticket, run,
// session or pool id, or a check's first word).
type stall struct {
	signal, subject, detail string
}

func (s stall) key() string { return "stall:" + s.signal + ":" + s.subject }

// runStallTrigger starts one run per fireable stall of a when.stall
// trigger (#143).
func (c *Controller) runStallTrigger(ctx context.Context, name string, tr config.Trigger) {
	fire, _, err := c.evalStall(ctx, name, tr)
	if err != nil {
		c.log.Error("stall trigger", "trigger", name, "err", err)
		return
	}
	for _, s := range fire {
		t, err := c.fireStall(ctx, name, tr, s)
		if err != nil {
			c.log.Error("stall fire", "trigger", name, "key", s.key(), "err", err)
			continue
		}
		c.log.Info("stall trigger fired", "trigger", name, "key", s.key(), "run", t.ID)
		c.recordStall(ctx, s, t.ID)
	}
}

// evalStall detects stalls and returns those that may fire now: no open
// run for the key, under max_runs, past the cooldown, and under
// stallMaxOpen open runs in total. why explains an empty result.
func (c *Controller) evalStall(ctx context.Context, name string, tr config.Trigger) (fire []stall, why string, err error) {
	mine, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, MetaPack: triggerPack, MetaKey: "name", MetaValue: name, Newest: true})
	if err != nil {
		return nil, "", err
	}
	open := 0
	for _, t := range mine {
		if t.Status != store.StatusClosed {
			open++
		}
	}
	if open >= stallMaxOpen {
		return nil, fmt.Sprintf("%d stall runs open", open), nil
	}
	found, err := c.detectStalls(ctx, tr)
	if err != nil {
		return nil, "", err
	}
	if len(found) == 0 {
		return nil, "no stall", nil
	}
	now := c.now()
	var reasons []string
	for _, s := range found {
		runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun, MetaPack: triggerPack, MetaKey: "key", MetaValue: name + "/" + s.key(), Newest: true})
		if err != nil {
			return nil, "", err
		}
		if r := stallHeld(runs, tr, now); r != "" {
			reasons = append(reasons, s.key()+": "+r)
			continue
		}
		fire = append(fire, s)
		if open+len(fire) >= stallMaxOpen {
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

// fireStall starts the trigger's workflow for one stall with key
// stall:<signal>:<subject>.
func (c *Controller) fireStall(ctx context.Context, name string, tr config.Trigger, s stall) (store.Ticket, error) {
	tv := map[string]any{config.StallSignalVar: s.signal, config.StallSubjectVar: s.subject, config.StallDetailVar: s.detail}
	p := rpc.RunStartParams{Workflow: tr.Run.Workflow, Vars: map[string]string{}, Key: s.key()}
	for k, v := range tr.Run.Vars {
		out, err := tr.Run.Fill(v, tv)
		if err != nil {
			return store.Ticket{}, fmt.Errorf("run var %s: %w", k, err)
		}
		p.Vars[k] = out
	}
	m := map[string]string{"name": name, "key": name + "/" + s.key()}
	res, err := c.startRun(ctx, scope{admin: true, actor: "controller"}, p, map[string]any{triggerPack: m})
	if err != nil {
		return store.Ticket{}, err
	}
	return res.(rpc.RunResult).Run, nil
}

// detectStalls evaluates every signal of the trigger, skipping subjects
// that belong to the trigger's own workflow.
func (c *Controller) detectStalls(ctx context.Context, tr config.Trigger) ([]stall, error) {
	own, err := c.workflowTickets(ctx, tr.Run.Workflow)
	if err != nil {
		return nil, err
	}
	now := c.now()
	var out []stall
	for _, sig := range tr.When.Stall {
		after := config.Duration(sig.After, 0)
		var found []stall
		var err error
		switch sig.Signal {
		case config.StallReadyUnclaimed:
			found, err = c.stallReadyUnclaimed(ctx, now, after)
		case config.StallPoolQuarantined:
			found = c.stallPoolQuarantined()
		case config.StallStepStalled:
			found, err = c.stallSteps(ctx, now, after)
		case config.StallSessionSilent:
			found, err = c.stallSessions(ctx, now, after)
		case config.StallNeedsHuman:
			found, err = c.stallNeedsHuman(ctx, now, after)
		case config.StallBlocked:
			found, err = c.stallBlocked(ctx, now, after)
		case config.StallLinearIssueNoRun:
			found, err = c.stallLinearIssues(ctx, sig)
		default:
			found, err = c.stallCheck(ctx, tr.When, sig)
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
		steps, err := c.store.ListChildren(ctx, r.ID)
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
	hist, err := c.store.History(ctx, t.ID)
	if err != nil {
		return last, err
	}
	for _, h := range hist {
		if h.At.After(last) {
			last = h.At
		}
	}
	cs, err := c.store.Comments(ctx, t.ID)
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

func ago(now, t time.Time) string { return now.Sub(t).Truncate(time.Second).String() }

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
	edges, err := c.store.Edges(ctx, t.ID)
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
func (c *Controller) stallReadyUnclaimed(ctx context.Context, now time.Time, after time.Duration) ([]stall, error) {
	cfg := c.config()
	var out []stall
	for _, proj := range slices.Sorted(maps.Keys(cfg.Projects)) {
		ts, err := c.store.Ready(ctx, proj, "", store.DefaultLimit)
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
func (c *Controller) stallPoolQuarantined() []stall {
	var out []stall
	for _, p := range slices.Sorted(maps.Keys(c.config().Pools)) {
		if c.poolPaused(p) {
			c.mu.Lock()
			until := c.poolPausedUntil[p]
			_, permanent := c.poolQuarantined[p]
			c.mu.Unlock()
			if permanent {
				continue // already escalated as "Pool <p> quarantined" (#146)
			}
			out = append(out, stall{config.StallPoolQuarantined, p + "@" + until.UTC().Format("20060102T150405Z"), "pool " + p + ": sessions crash-looping, paused until " + until.UTC().Format(time.RFC3339)})
		}
	}
	return out
}

// stallSteps: the current step of a running run is held but has had no
// history or comment for after. Open steps are ready_unclaimed's (or
// deferred or blocked on purpose).
func (c *Controller) stallSteps(ctx context.Context, now time.Time, after time.Duration) ([]stall, error) {
	runs, err := c.store.ListRuns(ctx, "", store.RunRunning, 0)
	if err != nil {
		return nil, err
	}
	var out []stall
	for _, r := range runs {
		if r.Status == store.StatusClosed {
			continue
		}
		steps, err := c.store.ListChildren(ctx, r.ID)
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
// not a run's escalation (the run covers those), older than after.
func (c *Controller) stallNeedsHuman(ctx context.Context, now time.Time, after time.Duration) ([]stall, error) {
	runs, err := c.store.ListRuns(ctx, "", store.RunNeedsHuman, 0)
	if err != nil {
		return nil, err
	}
	var out []stall
	for _, r := range runs {
		if r.Status == store.StatusClosed {
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
// [detail]". A non-zero exit means no stall.
func (c *Controller) stallCheck(ctx context.Context, w config.TriggerWhen, sig config.StallSignal) ([]stall, error) {
	out, why, err := c.runCheck(ctx, config.TriggerWhen{Check: sig.Check, CheckTimeout: w.CheckTimeout})
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
