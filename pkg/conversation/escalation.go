package conversation

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/conversation/attention"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// Escalation chat (#379). When work is on hold for a human (a run or pool
// escalation ticket, a wait past its timeout, a guard ask_human, any task
// or step labeled needs-human), the controller opens one session in the
// configured chat pool, keyed escalation:<ticket>, and sends the pack's
// prompt for the cause as context for the model, not as a visible message.
// meta.escalation records the key and session with a conditional update,
// so a restart or a repeated sweep never opens a second chat. Closing the
// ticket sends one final message. Questions keep their own answer path and
// get no chat. No escalation chat config: the ticket only. Gateways and
// Desk attach to the key.
//
// This file decides whether a hold gets a chat and what its prompt says.
// The controller opens the session and writes meta.escalation.

// Escalation is meta.escalation on an escalation ticket.
type Escalation struct {
	Key     string `json:"key"`
	Pool    string `json:"pool"`
	Session string `json:"session,omitempty"`
	// Cause is why the chat exists (Cause* constants) and Held the ticket
	// on hold, for Desk's list.
	Cause string `json:"cause,omitempty"`
	Held  string `json:"held,omitempty"`
	Final bool   `json:"final,omitempty"`
}

// EscalationOf is meta.escalation of t, false when it has no chat key.
func EscalationOf(t domain.Ticket) (Escalation, bool) {
	var e Escalation
	raw, ok := domain.MetaMap(t.Meta)[domain.MetaEscalation]
	if !ok || json.Unmarshal(raw, &e) != nil || e.Key == "" {
		return e, false
	}
	return e, true
}

// Escalation causes (meta.escalation.cause, the cause template var).
const (
	CauseRun   = "run"   // a run step failed at its retry limit
	CausePool  = "pool"  // a pool is quarantined
	CauseWait  = "wait"  // a wait passed its timeout
	CauseGuard = "guard" // a guardrail hook asked a human
	CauseLabel = "label" // anything else labeled needs-human
)

// Cause is what put work on hold for a human, as the prompt template
// sees it (config.EscalationVars). Strings are empty when the cause has
// no such part.
type Cause struct {
	Kind, Reason, Held, HeldTitle, Worker      string
	Run, Workflow, Step, IssueRef, Pool, Quarn string
}

// maxReason bounds the reason text a prompt carries.
const maxReason = 1000

// Hold reports whether t is a hold that wants an escalation chat: a task
// or step labeled needs-human that is not a question. Questions (report
// --question, admission and intake approvals) keep their own answer path
// and get no chat.
func Hold(t domain.Ticket) bool {
	return (t.Kind == domain.KindTask || t.Kind == domain.KindStep) && slices.Contains(t.Labels, attention.NeedsHumanLabel) &&
		!slices.Contains(t.Labels, attention.QuestionLabel) && domain.MetaString(t.Meta, domain.MetaQuestion, "ticket") == ""
}

// Reason is why the run's steps need a human.
func Reason(steps []domain.Ticket) string {
	for _, s := range steps {
		if domain.ReviewCapped(s) {
			return "a review step requested changes past its review round limit"
		}
	}
	return "a step failed at its retry limit"
}

// ClassifyHold classifies the hold ticket t and gathers its cause. A run
// escalation names the failed step and the session that last held it; a
// timed-out wait, a guard ask and a plain label name the held ticket.
func ClassifyHold(ctx context.Context, s store.Store, t domain.Ticket) Cause {
	switch {
	case domain.MetaString(t.Meta, domain.MetaRun, "escalation") != "":
		x := Cause{
			Kind: CauseRun, Run: domain.MetaString(t.Meta, domain.MetaRun, "id"), Workflow: domain.MetaString(t.Meta, domain.MetaRun, "workflow"),
			IssueRef: IssueRef(t), Reason: "a step failed at its retry limit",
		}
		if steps, err := store.Children(ctx, s, x.Run); err == nil {
			for i := len(steps) - 1; i >= 0; i-- {
				if !domain.StepExhausted(steps[i]) {
					continue
				}
				st := steps[i]
				x.Held, x.HeldTitle, x.Worker = st.ID, st.Title, store.LastClaimant(ctx, s, st.ID)
				if sf, err := domain.DecodeFields[domain.StepFields](st); err == nil {
					x.Step = sf.Step
					x.Reason = fmt.Sprintf("step %s failed %d/%d times", sf.Step, sf.Failures, sf.Retries)
					if domain.ReviewCapped(st) {
						x.Reason = fmt.Sprintf("step %s requested changes in review round %d, over its limit of %d", sf.Step, sf.ReviewRounds, sf.ReviewRoundLimit())
					}
				}
				if r := failedReason(st); r != "" {
					x.Reason += ": " + clip(r, maxReason)
				} else if cs, err := store.Comments(ctx, s, st.ID); err == nil && len(cs) > 0 {
					x.Reason += ". Last comment by " + cs[len(cs)-1].Author + ": " + clip(cs[len(cs)-1].Body, maxReason)
				}
				break
			}
		}
		return x
	case domain.MetaString(t.Meta, domain.MetaPool, "name") != "":
		q := domain.MetaString(t.Meta, domain.MetaPool, "quarantine")
		return Cause{Kind: CausePool, Pool: domain.MetaString(t.Meta, domain.MetaPool, "name"), Quarn: q, Reason: q}
	case domain.MetaString(t.Meta, domain.MetaHook, "session") != "":
		x := Cause{
			Kind: CauseGuard, Worker: domain.MetaString(t.Meta, domain.MetaHook, "session"), Held: domain.MetaString(t.Meta, domain.MetaHook, "ticket"),
			Reason: domain.MetaString(t.Meta, domain.MetaHook, "reason"), IssueRef: IssueRef(t),
		}
		if x.Held != "" {
			if h, err := s.Get(ctx, x.Held); err == nil {
				x.HeldTitle = h.Title
			}
		}
		return x
	}
	x := Cause{Kind: CauseLabel, Held: t.ID, HeldTitle: t.Title, Worker: t.Holder, IssueRef: IssueRef(t), Reason: clip(t.Body, maxReason)}
	if w := domain.WaitOf(t); w != nil && w.TimedOut {
		x.Kind = CauseWait
		x.Reason = fmt.Sprintf("wait (%s) passed its timeout %s; owner %s", w.Reason, w.Timeout.Format(time.RFC3339), cmp.Or(w.Owner, "unset"))
		x.Worker = cmp.Or(w.Session, x.Worker)
	} else if hist, err := store.History(ctx, s, t.ID); err == nil {
		for i := len(hist) - 1; i >= 0; i-- {
			if hist[i].Op == "label_add" && strings.Contains(string(hist[i].Diff), attention.NeedsHumanLabel) {
				if hist[i].Actor != "controller" && hist[i].Actor != "admin" {
					x.Worker = hist[i].Actor
				}
				break
			}
		}
		if cs, err := store.Comments(ctx, s, t.ID); err == nil && len(cs) > 0 {
			x.Reason = clip(cs[len(cs)-1].Body, maxReason)
		}
	}
	if x.Worker != "" {
		if w, err := s.Get(ctx, x.Worker); err != nil || w.Kind != domain.KindSession {
			x.Worker = "" // an actor like admin, not a session to message
		}
	}
	return x
}

// failedReason is the note of the worker's last failed or dropped report.
func failedReason(t domain.Ticket) string {
	if st := domain.MetaString(t.Meta, "report", "state"); st == rpc.ReportFailed || st == rpc.ReportDropped {
		return domain.MetaString(t.Meta, "report", "note")
	}
	return ""
}

// clip trims s and cuts it to n bytes with an ellipsis.
func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

// Policy is the chat pool and rendered prompt for hold ticket t with
// cause x, ok false when t gets no chat: no policy for its cause, or a
// pool quarantine of the chat pool itself. A prompt that fails to render
// is returned as err with ok false.
func Policy(cfg *config.Config, t domain.Ticket, x Cause) (pool, prompt string, ok bool, err error) {
	var e *config.EscalationChat
	var tmpl string
	switch x.Kind {
	case CauseRun:
		e = cfg.Escalation
		if w, has := cfg.Workflows[x.Workflow]; has && w.Escalate != nil {
			e = w.Escalate
		}
		if e != nil {
			tmpl = e.Prompt
		}
	case CausePool:
		if e = poolPolicy(cfg, t.Pool); e != nil && e.ChatPool != x.Pool {
			tmpl = e.PoolPrompt
		}
	default:
		e = holdPolicy(cfg, t.Project)
		if e != nil {
			tmpl = e.HoldPrompt
		}
	}
	if e == nil || e.ChatPool == "" || tmpl == "" {
		return "", "", false, nil
	}
	out, err := config.Fill(tmpl, map[string]any{
		"ticket": t.ID, "title": t.Title, "cause": x.Kind, "reason": x.Reason, "body": t.Body,
		"held": x.Held, "held_title": x.HeldTitle, "worker": x.Worker,
		"run": x.Run, "workflow": x.Workflow, "step": x.Step, "issue_ref": x.IssueRef,
		"pool": x.Pool, "quarantine": x.Quarn, "error": x.Quarn,
	})
	if err != nil {
		return "", "", false, err
	}
	return e.ChatPool, out, true, nil
}

// Configured reports whether any escalation chat policy is set, so the
// sweep can skip the store when none is.
func Configured(cfg *config.Config) bool {
	return cfg.Escalation != nil || slices.ContainsFunc(slices.Sorted(maps.Keys(cfg.Workflows)), func(n string) bool { return cfg.Workflows[n].Escalate != nil })
}

// holdPolicy is the policy for a hold that is neither a run nor a pool
// escalation: no run names the workflow, so the first one (by name) with
// a hold prompt whose chat pool is in the ticket's project owns it, else
// the top-level block.
func holdPolicy(cfg *config.Config, project string) *config.EscalationChat {
	ok := func(e *config.EscalationChat) bool {
		if e == nil || e.HoldPrompt == "" {
			return false
		}
		p, has := cfg.Pools[e.ChatPool]
		return has && p.Project == project
	}
	for _, name := range slices.Sorted(maps.Keys(cfg.Workflows)) {
		if e := cfg.Workflows[name].Escalate; ok(e) {
			return e
		}
	}
	if ok(cfg.Escalation) {
		return cfg.Escalation
	}
	return nil
}

// poolPolicy is the policy for a pool quarantine escalation filed in
// pool dest: the escalate block of a workflow escalating to dest that has
// a pool prompt, else the top-level block.
func poolPolicy(cfg *config.Config, dest string) *config.EscalationChat {
	for _, name := range slices.Sorted(maps.Keys(cfg.Workflows)) {
		w := cfg.Workflows[name]
		if w.Escalate != nil && w.Escalate.PoolPrompt != "" && w.EscalatePool == dest {
			return w.Escalate
		}
	}
	return cfg.Escalation
}

// IssueRef tells the chat session which issue to read, with its repo.
func IssueRef(t domain.Ticket) string {
	repo := ""
	if r := domain.MetaString(t.Meta, "github", "repo"); r != "" {
		repo = " --repo " + r
	}
	var m map[string]map[string]any
	_ = json.Unmarshal(t.Meta, &m)
	if k := domain.MetaString(t.Meta, "linear", "issue"); k != "" {
		return fmt.Sprintf("its Linear issue (`jobshop linear issue %s`)", k)
	}
	g := m["github"]
	switch {
	case g["issue"] != nil:
		return fmt.Sprintf("its GitHub issue (`gh issue view %v%s --comments`)", g["issue"], repo)
	case g["pr"] != nil:
		return fmt.Sprintf("its GitHub PR (`gh pr view %v%s --comments`)", g["pr"], repo)
	}
	return "any GitHub issue it names (the run's vars)"
}

// FinalMessage is what an escalation chat is told once its ticket closed.
func FinalMessage(ticket, resolution string) string {
	if resolution == "" {
		resolution = "done"
	}
	return fmt.Sprintf("jobshop: escalation %s closed (%s). Nothing more is needed unless the operator writes again.", ticket, resolution)
}

// ResolveComment is the close comment of a ticket resolved from its
// escalation chat.
const ResolveComment = "Resolved by the operator from the escalation chat."
