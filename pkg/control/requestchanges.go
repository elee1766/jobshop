package control

import (
	"context"
	"fmt"
	"strings"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/memedit"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Request changes (JOB-98). A reviewing step closes request_changes with
// its findings; the store closes it and, in the same transaction, counts
// a review round, reopens it behind the step it reviews and reopens that
// step with the findings in fields.changes. Past the reviewer's round cap
// the reviewer stays closed and checkRun moves the run to needs_human.

// maxFindings bounds the findings kept in fields.changes; the full text is
// also a comment on the reviewer's ticket.
const maxFindings = 16000

// closeOutcome maps `close --outcome request_changes` to the resolution,
// so either spelling works. Any other resolution with that outcome is a
// contradiction.
func closeOutcome(p *rpc.CloseParams) error {
	if p.Outcome != store.ResolutionRequestChanges {
		return nil
	}
	switch p.Resolution {
	case "", "done", store.ResolutionRequestChanges:
		// The resolution is the record; fields.outcome stays for when:
		// gates, which never open on a request.
		p.Resolution, p.Outcome = store.ResolutionRequestChanges, ""
		return nil
	}
	return fmt.Errorf("%w: --outcome request_changes closes with resolution request_changes, not %q", store.ErrInvalid, p.Resolution)
}

// requestChangesTarget is the closed step of t's run that a
// request_changes close of t reopens.
func (c *Controller) requestChangesTarget(ctx context.Context, cfg *config.Config, t store.Ticket, findings string) (store.Ticket, error) {
	if t.Kind != store.KindStep || t.ParentID == "" {
		return store.Ticket{}, fmt.Errorf("%w: request_changes is for workflow steps; %s is a %s", store.ErrInvalid, t.ID, t.Kind)
	}
	if t.Status == store.StatusClosed {
		return store.Ticket{}, fmt.Errorf("%w: %s", store.ErrClosed, t.ID)
	}
	if strings.TrimSpace(findings) == "" {
		return store.Ticket{}, fmt.Errorf("%w: request_changes needs the findings: --comment \"<findings>\"", store.ErrInvalid)
	}
	sf, err := store.DecodeFields[store.StepFields](t)
	if err != nil {
		return store.Ticket{}, err
	}
	name := sf.OnRequestChanges
	if name == "" {
		name = c.configTarget(ctx, cfg, t.ParentID, sf)
	}
	if name == "" {
		return store.Ticket{}, fmt.Errorf("%w: step %s has no step to send changes to (set on_request_changes in the workflow)", store.ErrInvalid, sf.Step)
	}
	steps, err := store.Children(ctx, c.store, t.ParentID)
	if err != nil {
		return store.Ticket{}, err
	}
	for _, s := range steps {
		f, err := store.DecodeFields[store.StepFields](s)
		if err != nil || f.Step != name {
			continue
		}
		if s.Status != store.StatusClosed {
			return store.Ticket{}, fmt.Errorf("%w: step %s is %s; request_changes reopens a finished step", store.ErrInvalid, name, s.Status)
		}
		return s, nil
	}
	return store.Ticket{}, fmt.Errorf("%w: run %s has no step %q", store.ErrInvalid, t.ParentID, name)
}

// configTarget reads the target from cfg for a run
// started before steps recorded on_request_changes.
func (c *Controller) configTarget(ctx context.Context, cfg *config.Config, runID string, sf store.StepFields) string {
	run, err := c.store.Get(ctx, runID)
	if err != nil {
		return ""
	}
	rf, err := store.DecodeFields[store.RunFields](run)
	if err != nil {
		return ""
	}
	w, ok := cfg.Workflows[rf.Workflow]
	if !ok || sf.Index >= len(w.Steps) || w.Steps[sf.Index].ID != sf.Step {
		return ""
	}
	return w.RequestChangesTarget(sf.Index)
}

// closeRequestChanges is close for resolution request_changes: the close
// and the reopen of the reviewed step in one store write, then the
// retry-pool move and the review-finding learning.
func (c *Controller) closeRequestChanges(ctx context.Context, cfg *config.Config, sc scope, id string, cl *store.Claim, findings string) (store.Ticket, error) {
	cur, err := c.store.Get(ctx, id)
	if err != nil {
		return store.Ticket{}, err
	}
	target, err := c.requestChangesTarget(ctx, cfg, cur, findings)
	if err != nil {
		return store.Ticket{}, err
	}
	findings = strings.TrimSpace(findings)
	kept := findings
	if len(kept) > maxFindings {
		kept = strings.ToValidUTF8(kept[:maxFindings], "") + " [truncated; full text on " + id + "]"
	}
	if _, err := store.AddComment(ctx, c.store, sc.actor, id, "Changes requested:\n\n"+findings); err != nil {
		return store.Ticket{}, err
	}
	t, err := c.store.Release(ctx, sc.actor, id, cl, store.Outcome{Close: store.ResolutionRequestChanges, Reopen: target.ID, Note: kept})
	if err != nil {
		return store.Ticket{}, err
	}
	if t.Status == store.StatusClosed {
		// Over the round cap: nothing reopened; checkRun escalates.
		c.log.WarnContext(ctx, "review round limit reached", "step", id, "run", t.ParentID)
		return t, nil
	}
	if target, err = c.store.Get(ctx, target.ID); err == nil {
		c.escalatePool(ctx, cfg, &target, "changes requested by "+id)
		c.record(ctx, store.LearningCorrection, target, memedit.ReviewFindingPrefix+c.stepName(ctx, target)+": "+findings, "")
	}
	c.log.InfoContext(ctx, "changes requested", "step", id, "reopened", target.ID, "run", t.ParentID)
	return t, nil
}

// changesPrompt is the ticket prompt section for an open change request.
func changesPrompt(t store.Ticket) string {
	if t.Kind != store.KindStep {
		return ""
	}
	f, err := store.DecodeFields[store.StepFields](t)
	if err != nil || f.Changes == nil {
		return ""
	}
	ch := f.Changes
	return fmt.Sprintf("# Changes requested\n\nStep %s (%s) reviewed your earlier result and requested changes (review round %d of %d). "+
		"Address every finding below on the same branch or PR, then close this step as usual: %s reviews it again.\n\n%s\n\n",
		ch.Step, ch.Ticket, ch.Round, ch.Max, ch.Step, ch.Findings)
}
