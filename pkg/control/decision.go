package control

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/integrations/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// Decisions (#603, docs/notes/603-decision-resolve.md). A planner
// proposal is a ticket labeled proposed whose meta.github names the
// proposal issue. decision.resolve acts on the issue, not only on the
// ticket: approve relabels it proposed -> target, reject closes it as not
// planned, dismiss leaves it. Each closes the ticket and clears the
// deferral of tickets waiting on the proposal, so the planner resumes
// now instead of after its release --until.

const (
	targetLabel          = "target"
	grantDecisionResolve = "decision-resolve"
	grantTicketUndefer   = "ticket-undefer"
)

// decisionRef is a decision ticket's proposal issue: a GitHub issue, or a
// Linear issue key (meta.linear.issue, JOB-17), which wins when set.
type decisionRef struct {
	Repo   string
	Issue  int
	Linear string
}

func (r decisionRef) String() string {
	if r.Linear != "" {
		return r.Linear
	}
	return fmt.Sprintf("%s#%d", r.Repo, r.Issue)
}

func decisionOf(t domain.Ticket) (decisionRef, bool) {
	var m struct {
		GitHub struct {
			Repo  string `json:"repo"`
			Issue any    `json:"issue"`
		} `json:"github"`
		Linear struct {
			Issue string `json:"issue"`
		} `json:"linear"`
	}
	if len(t.Meta) == 0 || json.Unmarshal(t.Meta, &m) != nil {
		return decisionRef{}, false
	}
	if m.Linear.Issue != "" {
		return decisionRef{Linear: m.Linear.Issue}, true
	}
	n := github.RefNumber(m.GitHub.Issue)
	return decisionRef{Repo: m.GitHub.Repo, Issue: n}, m.GitHub.Repo != "" && n > 0
}

func (c *Controller) decisionResolve(ctx context.Context, sc scope, p rpc.DecisionResolveParams) (any, error) {
	if err := c.requireGrant(sc, grantDecisionResolve, "decision resolve"); err != nil {
		return nil, err
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if !sc.admin && t.Project != sc.project {
		return nil, forbidden("%s is in another project", p.ID)
	}
	if !slices.Contains(t.Labels, proposedLabel) {
		return nil, fmt.Errorf("%w: %s is not a decision (no %s label)", domain.ErrInvalid, t.ID, proposedLabel)
	}
	if t.Status == domain.StatusClosed {
		return nil, fmt.Errorf("%w: %s is already closed (%s)", domain.ErrClosed, t.ID, t.Resolution)
	}
	ref, ok := decisionOf(t)
	if !ok && p.Action != rpc.DecisionDismiss {
		return nil, fmt.Errorf("%w: %s has no meta.github repo and issue or meta.linear issue", domain.ErrInvalid, t.ID)
	}
	who := sc.actor
	note := strings.TrimSpace(p.Note)
	var resolution, comment string
	switch p.Action {
	case rpc.DecisionApprove:
		resolution = "approved"
		body := fmt.Sprintf("Approved as target by %s.", who)
		if note != "" {
			body += "\n\n" + note
		}
		if err := c.approveIssue(ctx, ref, body); err != nil {
			return nil, err
		}
		comment = fmt.Sprintf("Approved: %s relabeled %s -> %s.", ref, proposedLabel, targetLabel)
		if ref.Linear != "" {
			comment += " Open sub-issues moved to " + c.linearIntakeState(ref.Linear) + "."
		}
	case rpc.DecisionReject:
		resolution = "rejected"
		body := fmt.Sprintf("Rejected by %s.", who)
		if note != "" {
			body += "\n\n" + note
		}
		if err := c.rejectIssue(ctx, ref, body); err != nil {
			return nil, err
		}
		comment = fmt.Sprintf("Rejected: %s closed as not planned.", ref)
	case rpc.DecisionDismiss:
		resolution = "dismissed"
		comment = "Dismissed: not now. The proposal issue is left as is."
	default:
		return nil, fmt.Errorf("%w: action must be approve, reject or dismiss", domain.ErrInvalid)
	}
	if note != "" {
		comment += "\n\n" + note
	}
	if _, err := store.AddComment(ctx, c.store, who, t.ID, comment); err != nil {
		return nil, err
	}
	// The ticket may be held (the planner parked it while holding, or a
	// session works it): close it as admin, its holder loses the claim.
	closed, err := store.Close(ctx, c.store, who, t.ID, nil, resolution)
	if err != nil {
		return nil, err
	}
	c.countClose(ctx, closed)
	undeferred := c.undeferWaiting(ctx, t, ref, ok, who)
	c.wakeWaits()
	return rpc.DecisionResolveResult{Ticket: closed, Undeferred: undeferred}, nil
}

// undeferWaiting clears defer_until on open tickets waiting on the
// proposal: in the decision's project, with the same meta.github issue,
// or naming the issue (owner/repo#n or #n) in their title or body.
func (c *Controller) undeferWaiting(ctx context.Context, d domain.Ticket, ref decisionRef, hasRef bool, who string) []string {
	ts, err := c.store.List(ctx, store.Filter{Project: d.Project, Status: domain.StatusOpen, Limit: math.MaxInt32})
	if err != nil {
		c.log.WarnContext(ctx, "decision: list waiting", "ticket", d.ID, "err", err)
		return nil
	}
	now := c.now()
	var out []string
	for _, t := range ts {
		if t.ID == d.ID || !t.DeferUntil.After(now) {
			continue
		}
		if t.Kind != domain.KindTask && t.Kind != domain.KindStep {
			continue
		}
		if !hasRef || !waitsOn(t, ref) {
			continue
		}
		if _, err := c.clearDefer(ctx, t, who, "decision "+d.ID+" resolved"); err != nil {
			c.log.WarnContext(ctx, "decision: undefer", "ticket", t.ID, "err", err)
			continue
		}
		out = append(out, t.ID)
	}
	return out
}

// waitsOn reports whether t is about the proposal issue.
func waitsOn(t domain.Ticket, ref decisionRef) bool {
	if r, ok := decisionOf(t); ok && strings.EqualFold(r.Repo, ref.Repo) && r.Issue == ref.Issue && strings.EqualFold(r.Linear, ref.Linear) {
		return true
	}
	text := t.Title + "\n" + t.Body
	if ref.Linear != "" {
		return containsKey(text, ref.Linear)
	}
	full := fmt.Sprintf("%s#%d", ref.Repo, ref.Issue)
	short := fmt.Sprintf("#%d", ref.Issue)
	if strings.Contains(strings.ToLower(text), strings.ToLower(full)) {
		return true
	}
	// "#56" but not "#560".
	for i := 0; ; {
		j := strings.Index(text[i:], short)
		if j < 0 {
			return false
		}
		end := i + j + len(short)
		if end == len(text) || text[end] < '0' || text[end] > '9' {
			return true
		}
		i = end
	}
}

// clearDefer sets defer_until to zero and notes who did it.
func (c *Controller) clearDefer(ctx context.Context, t domain.Ticket, who, why string) (domain.Ticket, error) {
	zero := time.Time{}
	u, err := c.store.Update(ctx, who, t.ID, store.Update{DeferUntil: &zero})
	if err != nil {
		return u, err
	}
	if _, err := store.AddComment(ctx, c.store, who, t.ID, "Deferral cleared: "+why); err != nil {
		c.log.WarnContext(ctx, "undefer comment", "ticket", t.ID, "err", err)
	}
	return u, nil
}

// ticketUndefer clears a ticket's deferral. Sessions need the
// ticket-undefer grant and stay in their own project (#603).
func (c *Controller) ticketUndefer(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	if err := c.requireGrant(sc, grantTicketUndefer, "ticket undefer"); err != nil {
		return nil, err
	}
	if p.ID == "" {
		return nil, fmt.Errorf("%w: ticket id is required", domain.ErrInvalid)
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if !sc.admin && t.Project != sc.project {
		return nil, forbidden("%s is in another project", p.ID)
	}
	if t.DeferUntil.IsZero() {
		return t, nil
	}
	return c.clearDefer(ctx, t, sc.actor, "by "+sc.actor)
}

// containsKey: text names the Linear key, "JOB-1" but not "JOB-12".
func containsKey(text, key string) bool {
	lt, lk := strings.ToLower(text), strings.ToLower(key)
	for i := 0; ; {
		j := strings.Index(lt[i:], lk)
		if j < 0 {
			return false
		}
		end := i + j + len(lk)
		if end == len(lt) || lt[end] < '0' || lt[end] > '9' {
			return true
		}
		i = end
	}
}

// approveIssue adds target, removes proposed and comments. On Linear it
// also moves the open sub-issues to the intake state, since the planner
// files them in Backlog (JOB-17).
func (c *Controller) approveIssue(ctx context.Context, ref decisionRef, body string) error {
	if ref.Linear != "" {
		return c.approveLinear(ctx, ref.Linear, body)
	}
	base := fmt.Sprintf("/repos/%s/issues/%d", ref.Repo, ref.Issue)
	if err := c.githubWrite(ctx, ref.Repo, http.MethodPost, base+"/labels", map[string]any{"labels": []string{targetLabel}}); err != nil {
		return err
	}
	if err := c.githubWrite(ctx, ref.Repo, http.MethodDelete, base+"/labels/"+url.PathEscape(proposedLabel), nil); err != nil && !github.IsNotFound(err) {
		return err
	}
	return c.githubWrite(ctx, ref.Repo, http.MethodPost, base+"/comments", map[string]any{"body": body})
}

// rejectIssue comments and closes the issue as not planned.
func (c *Controller) rejectIssue(ctx context.Context, ref decisionRef, body string) error {
	if ref.Linear != "" {
		return c.rejectLinear(ctx, ref.Linear, body)
	}
	base := fmt.Sprintf("/repos/%s/issues/%d", ref.Repo, ref.Issue)
	if err := c.githubWrite(ctx, ref.Repo, http.MethodPost, base+"/comments", map[string]any{"body": body}); err != nil {
		return err
	}
	return c.githubWrite(ctx, ref.Repo, http.MethodPatch, base, map[string]any{"state": "closed", "state_reason": "not_planned"})
}

func (c *Controller) approveLinear(ctx context.Context, key, body string) error {
	cl, err := c.linearClient(key)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := cl.AddLabel(ctx, key, targetLabel); err != nil {
		return err
	}
	if err := cl.RemoveLabel(ctx, key, proposedLabel); err != nil {
		return err
	}
	lc, err := cl.GetContext(ctx, key, false)
	if err != nil {
		return err
	}
	state := c.linearIntakeState(key)
	for _, ch := range lc.Children {
		if !ch.Open() || strings.EqualFold(ch.State.Name, state) {
			continue
		}
		if _, err := cl.SetState(ctx, ch.Identifier, state); err != nil {
			return fmt.Errorf("move %s to %s: %w", ch.Identifier, state, err)
		}
	}
	return cl.AddComment(ctx, key, body)
}

func (c *Controller) rejectLinear(ctx context.Context, key, body string) error {
	cl, err := c.linearClient(key)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := cl.AddComment(ctx, key, body); err != nil {
		return err
	}
	state, err := cl.CanceledState(ctx, key)
	if err != nil {
		return err
	}
	_, err = cl.SetState(ctx, key, state)
	return err
}

// linearIntakeState is the first of linear_intake_states for the key's
// team, read from the linear_issue_no_run scan signal; Todo without one.
func (c *Controller) linearIntakeState(key string) string {
	team, _ := linear.Team(key)
	cfg := c.config()
	for _, name := range slices.Sorted(maps.Keys(cfg.Triggers)) {
		for _, sig := range cfg.Triggers[name].When.Scan {
			if sig.Signal != config.SignalLinearIssueNoRun || !strings.EqualFold(sig.Team, team) {
				continue
			}
			if s := config.SplitList(sig.States); len(s) > 0 {
				return s[0]
			}
		}
	}
	return "Todo"
}

// githubWrite calls the GitHub REST API as the identity covering repo,
// else the host's gh token.
func (c *Controller) githubWrite(ctx context.Context, repo, method, path string, body any) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	_, err := c.githubOrHost(c.config(), "", repo).Do(ctx, github.Request{Method: method, Path: path, Body: body, Route: "/repos/{repo}"})
	return err
}
