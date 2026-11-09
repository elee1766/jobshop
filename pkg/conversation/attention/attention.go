// Package attention projects what waits on a human into rpc.AttentionItem
// rows and describes the choices each row offers (#185, split out of
// pkg/control by JOB-147).
//
// The projection is a read. The controller supplies the tickets and the
// subsystem rows through Source, and keeps every write: grants, the
// stale-option check, answering questions, closing and resuming work and
// the ticket and session mutations an answer causes all stay in
// pkg/control. Actions only describes choices, so a supervisor read can
// never change work.
package attention

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/elee1766/jobshop/pkg/conversation/question"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

const (
	// NeedsHumanLabel marks work on hold for a human.
	NeedsHumanLabel = "needs-human"
	// QuestionLabel marks a ticket where a worker asks a human (#185, #186).
	QuestionLabel = "question"
	// ProposedLabel marks the planner's proposed next target, which waits
	// on a human to approve (#488).
	ProposedLabel = "proposed"

	// KeyPrefix starts the session key of an attention chat.
	KeyPrefix = "attention:"
)

// TicketKinds are the attention kinds whose ID is a ticket.
var TicketKinds = []string{
	rpc.AttentionRun, rpc.AttentionTicket, rpc.AttentionQuestion, rpc.AttentionDecision,
	rpc.AttentionStale, rpc.AttentionWait,
}

// Key is the default session key of the chat that resolves an item.
func Key(kind, id string) string { return KeyPrefix + kind + ":" + id }

// ParseKey splits an attention chat key, ok false for any other key.
func ParseKey(key string) (kind, id string, ok bool) {
	rest, ok := strings.CutPrefix(key, KeyPrefix)
	if !ok {
		return "", "", false
	}
	kind, id, _ = strings.Cut(rest, ":")
	return kind, id, true
}

// Waiting is a ticket whose wait passed its timeout.
type Waiting struct {
	Ticket domain.Ticket
	Wait   *domain.Wait
}

// Pause is a pool paused for crash-looping (JOB-27) until Until.
type Pause struct {
	Pool  string
	Until time.Time
	Why   string
}

// Health is the controller's own state that can wait on a human.
type Health struct {
	Started time.Time
	// ConfigErr and WatchErr are empty when the config loaded cleanly.
	ConfigErr, WatchErr     string
	ConfigErrAt, WatchErrAt time.Time
	// Quarantined maps a pool to why it stopped for good.
	Quarantined map[string]string
	Paused      []Pause
	// PausedFor is how long a pause lasts, to date its start.
	PausedFor time.Duration
}

// ActionSource reads what describing choices needs.
type ActionSource interface {
	// Ticket is the ticket with id, or an error.
	Ticket(ctx context.Context, id string) (domain.Ticket, error)
	// Steps are the steps of a run, in order.
	Steps(ctx context.Context, run string) ([]domain.Ticket, error)
}

// Source is what the controller supplies to Build.
type Source interface {
	ActionSource
	Now() time.Time
	// LabeledOpen lists open and in-progress tickets with label.
	LabeledOpen(ctx context.Context, label string) ([]domain.Ticket, error)
	// NeedsHumanRuns lists runs in needs_human, newest first.
	NeedsHumanRuns(ctx context.Context) ([]domain.Ticket, error)
	// TimedOut lists tickets whose wait passed its timeout.
	TimedOut(ctx context.Context) ([]Waiting, error)
	// Subsystems are rows owned by other subsystems: resource waits,
	// ingest, journal lag.
	Subsystems(ctx context.Context) ([]rpc.AttentionItem, error)
	// Stale are orphaned tickets older than attention.orphan_age.
	Stale(ctx context.Context) ([]rpc.AttentionItem, error)
	Health() Health
}

// Meta keys of an item on its ticket's meta.desk (the pack is written by
// the chat lifecycle).
const (
	MetaSummary   = "attention_summary"
	MetaContext   = "attention_context"
	MetaRecommend = "attention_recommend"
	MetaChat      = "attention_chat"
)

// Build lists only what waits on a human, across all projects, oldest
// first. Empty when healthy.
func Build(ctx context.Context, src Source) ([]rpc.AttentionItem, error) {
	items := []rpc.AttentionItem{}
	// Runs and pools an open escalation already stands for are listed once,
	// as the escalation ticket.
	escalatedRuns, escalatedPools := map[string]bool{}, map[string]bool{}
	seen := map[string]bool{}
	for _, label := range []string{NeedsHumanLabel, QuestionLabel, ProposedLabel} {
		ts, err := src.LabeledOpen(ctx, label)
		if err != nil {
			return nil, err
		}
		for _, t := range ts {
			if id := domain.MetaString(t.Meta, domain.MetaRun, "id"); id != "" {
				escalatedRuns[id] = true
			}
			if p := domain.MetaString(t.Meta, domain.MetaPool, "name"); p != "" {
				escalatedPools[p] = true
			}
			if seen[t.ID] {
				continue
			}
			seen[t.ID] = true
			kind := rpc.AttentionTicket
			switch {
			case label == ProposedLabel && !slices.Contains(t.Labels, NeedsHumanLabel):
				kind = rpc.AttentionDecision
			case (label == QuestionLabel && !slices.Contains(t.Labels, NeedsHumanLabel)) || domain.MetaString(t.Meta, domain.MetaQuestion, "ticket") != "":
				kind = rpc.AttentionQuestion
			}
			items = append(items, TicketItem(kind, t))
		}
	}
	runs, err := src.NeedsHumanRuns(ctx)
	if err != nil {
		return nil, err
	}
	// Waits past their timeout (#381), with reason and owner. Listed even
	// though the timeout labeled them needs-human, so the title says why.
	waits, err := src.TimedOut(ctx)
	if err != nil {
		return nil, err
	}
	for _, w := range waits {
		t := w.Ticket
		it := TicketItem(rpc.AttentionWait, t)
		it.Title = fmt.Sprintf("%s: wait past timeout: %s", t.Title, w.Wait.Reason)
		it.Owner, it.Since = w.Wait.Owner, w.Wait.Timeout
		if domain.MetaString(t.Meta, domain.MetaDesk, MetaSummary) == "" {
			it.Summary = SummaryOf(w.Wait.Reason)
		}
		if i := slices.IndexFunc(items, func(x rpc.AttentionItem) bool { return x.ID == t.ID }); i >= 0 {
			items[i] = it
		} else {
			items = append(items, it)
		}
	}
	for _, r := range runs {
		if !escalatedRuns[r.ID] && r.Status != domain.StatusClosed {
			items = append(items, TicketItem(rpc.AttentionRun, r))
		}
	}
	more, err := src.Subsystems(ctx)
	if err != nil {
		return nil, err
	}
	items = append(items, more...)
	stale, err := src.Stale(ctx)
	if err != nil {
		return nil, err
	}
	for _, it := range stale {
		// A needs-human or question ticket with no pool is listed once,
		// as itself.
		if !slices.ContainsFunc(items, func(x rpc.AttentionItem) bool { return x.ID == it.ID }) {
			items = append(items, it)
		}
	}
	items = append(items, healthItems(src.Health(), escalatedPools)...)
	for i := range items {
		if items[i].Kind == rpc.AttentionQuestion && len(items[i].Options) > 0 {
			continue
		}
		for _, a := range Actions(ctx, src, items[i]) {
			items[i].Options = append(items[i].Options, a.Option)
		}
	}
	slices.SortStableFunc(items, func(a, b rpc.AttentionItem) int {
		if n := a.Since.Compare(b.Since); n != 0 {
			return n
		}
		return strings.Compare(a.ID, b.ID)
	})
	return items, nil
}

// healthItems are the config and pool rows. A pool an open escalation
// stands for is listed once, as that ticket.
func healthItems(h Health, escalatedPools map[string]bool) []rpc.AttentionItem {
	var items []rpc.AttentionItem
	if h.ConfigErr != "" {
		items = append(items, rpc.AttentionItem{Kind: rpc.AttentionConfig, ID: "config", Title: "serving last good config, a pack failed to load: " + h.ConfigErr, Since: h.ConfigErrAt})
	}
	if h.WatchErr != "" {
		items = append(items, rpc.AttentionItem{Kind: rpc.AttentionConfigInvalid, ID: "config-invalid", Title: "a watched config change failed to load, running config kept: " + h.WatchErr, Since: h.WatchErrAt})
	}
	pools := make([]string, 0, len(h.Quarantined))
	for p := range h.Quarantined {
		pools = append(pools, p)
	}
	sort.Strings(pools)
	for _, p := range pools {
		if !escalatedPools[p] {
			items = append(items, rpc.AttentionItem{Kind: rpc.AttentionPool, ID: p, Title: "pool " + p + " quarantined: " + h.Quarantined[p], Since: h.Started})
		}
	}
	// Crash-loop pauses (JOB-27) name the last session error.
	for _, p := range h.Paused {
		if _, permanent := h.Quarantined[p.Pool]; permanent || escalatedPools[p.Pool] {
			continue
		}
		items = append(items, rpc.AttentionItem{
			Kind: rpc.AttentionPool, ID: p.Pool, Since: p.Until.Add(-h.PausedFor),
			Title:   ClipTitle("pool " + p.Pool + " paused until " + p.Until.UTC().Format(time.RFC3339) + ", sessions crash-looping: " + p.Why),
			Summary: SummaryOf(p.Why),
		})
	}
	return items
}

// ClipTitle keeps the first line of s, at most 200 bytes.
func ClipTitle(s string) string {
	s, _, _ = strings.Cut(s, "\n")
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}

// maxSummary bounds an attention summary, about two lines of a row.
const maxSummary = 240

// MaxSummary is the byte bound of a summary, before its ellipsis.
const MaxSummary = maxSummary

// SummaryOf is s's first non-empty lines, at most maxSummary bytes, on one line.
func SummaryOf(s string) string {
	var parts []string
	n := 0
	for l := range strings.Lines(s) {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		parts = append(parts, l)
		if n += len(l); n >= maxSummary || len(parts) == 2 {
			break
		}
	}
	out := strings.Join(parts, " ")
	if len(out) > maxSummary {
		cut := maxSummary
		for cut > 0 && !utf8.RuneStart(out[cut]) {
			cut--
		}
		out = out[:cut] + "..."
	}
	return out
}

// TicketItem lists t. A run waits since it last changed (when it went
// needs_human), a ticket since it was filed.
func TicketItem(kind string, t domain.Ticket) rpc.AttentionItem {
	it := rpc.AttentionItem{Kind: kind, ID: t.ID, Project: t.Project, Title: t.Title, Since: t.CreatedAt}
	if kind == rpc.AttentionRun {
		it.Since = t.UpdatedAt
	}
	it.Summary = SummaryOf(domain.MetaString(t.Meta, domain.MetaDesk, MetaSummary))
	if it.Summary == "" {
		it.Summary = SummaryOf(t.Body)
	}
	if qm, ok := question.Of(t); ok {
		it.Context, it.Options = qm.Context, qm.Options
	}
	if full := domain.MetaString(t.Meta, domain.MetaDesk, MetaContext); full != "" {
		it.Context = full
	} else if it.Context == "" {
		it.Context = t.Body
	}
	var m struct {
		GitHub struct {
			Repo  string `json:"repo"`
			Issue any    `json:"issue"`
			PR    any    `json:"pr"`
		} `json:"github"`
		Linear struct {
			Issue     string `json:"issue"`
			Workspace string `json:"workspace"`
		} `json:"linear"`
	}
	if len(t.Meta) > 0 && json.Unmarshal(t.Meta, &m) == nil {
		it.Repo = m.GitHub.Repo
		switch {
		case m.Linear.Issue != "":
			it.Ref = m.Linear.Issue
			if m.Linear.Workspace != "" {
				it.URL = "https://linear.app/" + url.PathEscape(m.Linear.Workspace) + "/issue/" + url.PathEscape(m.Linear.Issue)
			}
		case m.GitHub.Issue != nil:
			it.Ref = fmt.Sprintf("#%v", m.GitHub.Issue)
			if m.GitHub.Repo != "" {
				it.URL = fmt.Sprintf("https://github.com/%s/issues/%v", m.GitHub.Repo, m.GitHub.Issue)
			}
		case m.GitHub.PR != nil:
			it.Ref = fmt.Sprintf("PR #%v", m.GitHub.PR)
			if m.GitHub.Repo != "" {
				it.URL = fmt.Sprintf("https://github.com/%s/pull/%v", m.GitHub.Repo, m.GitHub.PR)
			}
		}
	}
	return it
}

// Verbs of an Action, which the controller executes.
const (
	VerbQuestion     = "question"
	VerbDecision     = "decision"
	VerbResume       = "resume"
	VerbCancel       = "cancel"
	VerbUnquarantine = "unquarantine"
	VerbClear        = "clear"
	VerbGuard        = "guard"
	VerbExtend       = "extend"
	VerbKeep         = "keep"
	VerbDiscuss      = "discuss"
)

// Action is one choice of an item: what Desk shows (Option) and what the
// controller does when it is picked (Verb on Target, with Step).
type Action struct {
	Option       domain.QuestionOption
	Verb, Target string
	Step         string
}

// Actions describes the choices of it. It only reads, so supervisor
// reads cannot mutate work.
func Actions(ctx context.Context, src ActionSource, it rpc.AttentionItem) []Action {
	var actions []Action
	add := func(title, description, verb, target, step string) {
		actions = append(actions, Action{domain.QuestionOption{ID: len(actions) + 1, Title: title, Description: description}, verb, target, step})
	}
	t, _ := src.Ticket(ctx, it.ID)
	runID, pool := domain.MetaString(t.Meta, domain.MetaRun, "id"), domain.MetaString(t.Meta, domain.MetaPool, "name")
	if it.Kind == rpc.AttentionRun {
		runID = it.ID
	}
	if it.Kind == rpc.AttentionPool {
		pool = it.ID
	}
	switch {
	case it.Kind == rpc.AttentionQuestion && len(it.Options) > 0:
		for _, opt := range it.Options {
			actions = append(actions, Action{Option: opt, Verb: VerbQuestion, Target: it.ID})
		}
	case it.Kind == rpc.AttentionDecision:
		for _, action := range []string{rpc.DecisionApprove, rpc.DecisionReject, rpc.DecisionDismiss} {
			add(strings.ToUpper(action[:1])+action[1:], "Resolve the proposal as "+action, VerbDecision, it.ID, action)
		}
	case it.Kind == rpc.AttentionWait:
		add("Clear wait", "Make the ticket eligible again", VerbClear, it.ID, "")
		add("Extend one hour", "Keep waiting with a new timeout", VerbExtend, it.ID, "")
		if runID != "" {
			add("Cancel run", "Stop this workflow run", VerbCancel, runID, "")
		}
	case domain.MetaString(t.Meta, domain.MetaHook, "key") != "":
		add("Allow", "Resolve the guardrail approval", VerbGuard, it.ID, "done")
		add("Deny", "Resolve the guardrail approval", VerbGuard, it.ID, "denied")
	case runID != "":
		add("Resume", "Retry the first unfinished step", VerbResume, runID, "")
		if steps, err := src.Steps(ctx, runID); err == nil {
			for _, st := range steps {
				sf, _ := domain.DecodeFields[domain.StepFields](st)
				if !domain.StepFinished(st) || len(actions) >= 4 {
					break
				}
				add("Resume at "+sf.Step, "Reopen this earlier step", VerbResume, runID, sf.Step)
			}
		}
		add("Cancel run", "Stop this workflow run", VerbCancel, runID, "")
	case pool != "":
		add("Unquarantine after fix", "Let the pool accept work again", VerbUnquarantine, pool, "")
		add("Keep quarantined", "Leave the pool stopped", VerbKeep, pool, "")
	default:
		add("Discuss", "Open the resolving chat with context", VerbDiscuss, it.ID, "")
		add("Leave waiting", "Keep this item open for a later decision", VerbKeep, it.ID, "")
	}
	if recommended := domain.MetaString(t.Meta, domain.MetaDesk, MetaRecommend); recommended != "" {
		for i := range actions {
			actions[i].Option.Recommended = recommended == strconv.Itoa(actions[i].Option.ID)
		}
	}
	return actions
}
