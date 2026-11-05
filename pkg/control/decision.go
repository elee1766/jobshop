package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
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

// decisionRef is a decision ticket's proposal issue.
type decisionRef struct {
	Repo  string
	Issue int
}

func decisionOf(t store.Ticket) (decisionRef, bool) {
	var m struct {
		GitHub struct {
			Repo  string `json:"repo"`
			Issue any    `json:"issue"`
		} `json:"github"`
	}
	if len(t.Meta) == 0 || json.Unmarshal(t.Meta, &m) != nil {
		return decisionRef{}, false
	}
	n := refNumber(m.GitHub.Issue)
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
		return nil, fmt.Errorf("%w: %s is not a decision (no %s label)", store.ErrInvalid, t.ID, proposedLabel)
	}
	if t.Status == store.StatusClosed {
		return nil, fmt.Errorf("%w: %s is already closed (%s)", store.ErrClosed, t.ID, t.Resolution)
	}
	ref, ok := decisionOf(t)
	if !ok && p.Action != rpc.DecisionDismiss {
		return nil, fmt.Errorf("%w: %s has no meta.github repo and issue", store.ErrInvalid, t.ID)
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
		comment = fmt.Sprintf("Approved: %s#%d relabeled %s -> %s.", ref.Repo, ref.Issue, proposedLabel, targetLabel)
	case rpc.DecisionReject:
		resolution = "rejected"
		body := fmt.Sprintf("Rejected by %s.", who)
		if note != "" {
			body += "\n\n" + note
		}
		if err := c.rejectIssue(ctx, ref, body); err != nil {
			return nil, err
		}
		comment = fmt.Sprintf("Rejected: %s#%d closed as not planned.", ref.Repo, ref.Issue)
	case rpc.DecisionDismiss:
		resolution = "dismissed"
		comment = "Dismissed: not now. The proposal issue is left as is."
	default:
		return nil, fmt.Errorf("%w: action must be approve, reject or dismiss", store.ErrInvalid)
	}
	if note != "" {
		comment += "\n\n" + note
	}
	if _, err := c.store.AddComment(ctx, who, t.ID, comment); err != nil {
		return nil, err
	}
	// The ticket may be held (the planner parked it while holding, or a
	// session works it): close it as admin, its holder loses the claim.
	closed, err := c.store.CloseTicket(ctx, who, t.ID, nil, resolution)
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
func (c *Controller) undeferWaiting(ctx context.Context, d store.Ticket, ref decisionRef, hasRef bool, who string) []string {
	ts, err := c.store.List(ctx, store.Filter{Project: d.Project, Status: store.StatusOpen, Limit: math.MaxInt32})
	if err != nil {
		c.log.Warn("decision: list waiting", "ticket", d.ID, "err", err)
		return nil
	}
	now := c.now()
	var out []string
	for _, t := range ts {
		if t.ID == d.ID || !t.DeferUntil.After(now) {
			continue
		}
		if t.Kind != store.KindTask && t.Kind != store.KindStep {
			continue
		}
		if !hasRef || !waitsOn(t, ref) {
			continue
		}
		if _, err := c.clearDefer(ctx, t, who, "decision "+d.ID+" resolved"); err != nil {
			c.log.Warn("decision: undefer", "ticket", t.ID, "err", err)
			continue
		}
		out = append(out, t.ID)
	}
	return out
}

// waitsOn reports whether t is about the proposal issue.
func waitsOn(t store.Ticket, ref decisionRef) bool {
	if r, ok := decisionOf(t); ok && strings.EqualFold(r.Repo, ref.Repo) && r.Issue == ref.Issue {
		return true
	}
	text := t.Title + "\n" + t.Body
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
func (c *Controller) clearDefer(ctx context.Context, t store.Ticket, who, why string) (store.Ticket, error) {
	zero := time.Time{}
	u, err := c.store.Update(ctx, who, t.ID, store.Update{DeferUntil: &zero})
	if err != nil {
		return u, err
	}
	if _, err := c.store.AddComment(ctx, who, t.ID, "Deferral cleared: "+why); err != nil {
		c.log.Warn("undefer comment", "ticket", t.ID, "err", err)
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
		return nil, fmt.Errorf("%w: ticket id is required", store.ErrInvalid)
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

// approveIssue adds target, removes proposed and comments.
func (c *Controller) approveIssue(ctx context.Context, ref decisionRef, body string) error {
	base := fmt.Sprintf("/repos/%s/issues/%d", ref.Repo, ref.Issue)
	if err := c.githubWrite(ctx, ref.Repo, http.MethodPost, base+"/labels", map[string]any{"labels": []string{targetLabel}}); err != nil {
		return err
	}
	if err := c.githubWrite(ctx, ref.Repo, http.MethodDelete, base+"/labels/"+url.PathEscape(proposedLabel), nil); err != nil && !isNotFound(err) {
		return err
	}
	return c.githubWrite(ctx, ref.Repo, http.MethodPost, base+"/comments", map[string]any{"body": body})
}

// rejectIssue comments and closes the issue as not planned.
func (c *Controller) rejectIssue(ctx context.Context, ref decisionRef, body string) error {
	base := fmt.Sprintf("/repos/%s/issues/%d", ref.Repo, ref.Issue)
	if err := c.githubWrite(ctx, ref.Repo, http.MethodPost, base+"/comments", map[string]any{"body": body}); err != nil {
		return err
	}
	return c.githubWrite(ctx, ref.Repo, http.MethodPatch, base, map[string]any{"state": "closed", "state_reason": "not_planned"})
}

type githubStatusError struct {
	status int
	msg    string
}

func (e githubStatusError) Error() string { return e.msg }

func isNotFound(err error) bool {
	var ge githubStatusError
	return errors.As(err, &ge) && ge.status == http.StatusNotFound
}

// githubWrite calls the GitHub REST API as the identity covering repo,
// else the host's gh token (pollerAuth).
func (c *Controller) githubWrite(ctx context.Context, repo, method, path string, body any) error {
	base, tok, err := c.pollerAuth(ctx, c.config(), "", repo)
	if err != nil {
		return fmt.Errorf("github auth for %s: %w", repo, err)
	}
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rd = bytes.NewReader(b)
	}
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, base+path, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+tok)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode/100 != 2 {
		return githubStatusError{resp.StatusCode, fmt.Sprintf("github %s %s: %s: %s", method, path, resp.Status, strings.TrimSpace(string(b)))}
	}
	return nil
}
