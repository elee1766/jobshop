package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"os"
	"slices"
	"strings"
	"time"

	"go.opentelemetry.io/otel/metric"

	host "github.com/elee1766/acphost/pkg/acphost"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/execution/session"
	"github.com/elee1766/jobshop/pkg/orchestration/hooks"
	"github.com/elee1766/jobshop/pkg/orchestration/workflow"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/support/clock"
	"github.com/elee1766/jobshop/pkg/support/telemetry"
	"github.com/elee1766/jobshop/pkg/transport/rpc"

	"github.com/elee1766/jobshop/pkg/domain"
)

func bytesReader(b []byte) io.Reader { return bytes.NewReader(b) }

func forbidden(format string, args ...any) error {
	return fmt.Errorf("%w: %s", rpc.ErrForbidden, fmt.Sprintf(format, args...))
}

func requireAdmin(sc scope, what string) error {
	if !sc.admin {
		return forbidden("%s needs the admin socket", what)
	}
	return nil
}

// requireGrant is requireAdmin for methods a pool can grant its sessions
// with pools.<name>.admin. The live config is read per call, so a reload
// that removes the grant takes effect at once.
func (c *Controller) requireGrant(sc scope, method, what string) error {
	if sc.admin {
		return nil
	}
	if p, ok := c.config().Pools[sc.pool]; ok && p.Allows(method) {
		return nil
	}
	return forbidden("%s needs the admin socket or pool admin grant %q", what, method)
}

// project picks the request's project: a session is pinned to its own.
func (c *Controller) project(sc scope, asked string) (string, error) {
	if !sc.admin {
		if asked != "" && asked != sc.project {
			return "", forbidden("session is scoped to project %s", sc.project)
		}
		return sc.project, nil
	}
	if asked == "" {
		asked = c.config().DefaultProject
	}
	if asked == "" {
		return "", fmt.Errorf("%w: no project given and no default_project configured", domain.ErrInvalid)
	}
	if _, ok := c.config().Projects[asked]; !ok {
		return "", fmt.Errorf("%w: unknown project %q", domain.ErrInvalid, asked)
	}
	return asked, nil
}

// current resolves an omitted ticket id to the session's one in-progress
// claim. Current work is derived from claims rather than stored, so it can
// never disagree with them.
func (c *Controller) current(ctx context.Context, sc scope, id string) (string, error) {
	if id != "" {
		return id, nil
	}
	if sc.admin {
		return "", fmt.Errorf("%w: ticket id is required", domain.ErrInvalid)
	}
	held, err := c.store.List(ctx, store.Filter{Holder: sc.session, Status: domain.StatusInProgress})
	if err != nil {
		return "", err
	}
	switch len(held) {
	case 0:
		return "", fmt.Errorf("%w: session holds no ticket; pass an id", domain.ErrInvalid)
	case 1:
		return held[0].ID, nil
	}
	return "", fmt.Errorf("%w: session holds %d tickets; pass an id", domain.ErrInvalid, len(held))
}

// heldClaim returns the claim for id. Sessions must hold it; admin acts on
// whatever claim is current.
func (c *Controller) heldClaim(ctx context.Context, sc scope, id string) (store.Claim, error) {
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return store.Claim{}, err
	}
	if t.Status != domain.StatusInProgress {
		return store.Claim{}, fmt.Errorf("%w: %s is %s", domain.ErrNotHolder, id, t.Status)
	}
	if !sc.admin && t.Holder != sc.session {
		return store.Claim{}, fmt.Errorf("%w: %s is held by %s", domain.ErrNotHolder, id, t.Holder)
	}
	return store.Claim{TicketID: t.ID, Holder: t.Holder, Fence: t.Fence, LeaseUntil: t.LeaseUntil}, nil
}

// grantedClaim is the claim on id for a session whose pool holds the
// ticket-release grant, when the ticket is held by another session of its
// project. It returns notHolder otherwise.
func (c *Controller) grantedClaim(ctx context.Context, sc scope, id string, notHolder error) (store.Claim, error) {
	if sc.admin || c.requireGrant(sc, "ticket-release", "release") != nil {
		return store.Claim{}, notHolder
	}
	t, err := c.store.Get(ctx, id)
	if err != nil || t.Status != domain.StatusInProgress || t.Project != sc.project {
		return store.Claim{}, notHolder
	}
	c.log.InfoContext(ctx, "release by grant", "ticket", id, "holder", t.Holder, "by", sc.actor)
	return store.Claim{TicketID: t.ID, Holder: t.Holder, Fence: t.Fence, LeaseUntil: t.LeaseUntil}, nil
}

func (c *Controller) whoami(_ context.Context, sc scope, _ struct{}) (any, error) {
	return rpc.Whoami{Actor: sc.actor, Admin: sc.admin, Session: sc.session, Project: sc.project, Pool: sc.pool}, nil
}

// version is open to every caller: post-merge sessions check the deploy
// with it.
func (c *Controller) version(_ context.Context, _ scope, _ struct{}) (any, error) {
	return rpc.VersionResult{Version: c.opt.Version, Commit: c.opt.Commit}, nil
}

func (c *Controller) create(ctx context.Context, sc scope, p rpc.CreateParams) (any, error) {
	proj, err := c.project(sc, p.Project)
	if err != nil {
		return nil, err
	}
	if p.Kind == "" {
		p.Kind = domain.KindTask
	}
	if p.Kind == domain.KindSession {
		return nil, fmt.Errorf("%w: use session new to create sessions", domain.ErrInvalid)
	}
	if p.Kind == domain.KindRun || p.Kind == domain.KindStep {
		return nil, fmt.Errorf("%w: use run start to create runs and steps", domain.ErrInvalid)
	}
	if p.Kind == domain.KindBatch {
		return nil, fmt.Errorf("%w: use batch start to create batches", domain.ErrInvalid)
	}
	if p.Holder != "" && !sc.admin {
		return nil, forbidden("assigning needs the admin socket")
	}
	if err := c.checkAssign(ctx, p.Holder, p.Force); err != nil {
		return nil, err
	}
	// A session's follow-up task with no pool would never be claimed (#244);
	// it goes to the creator's pool, which can reroute it.
	if p.Kind == domain.KindTask && p.Pool == "" && !sc.admin {
		p.Pool = sc.pool
	}
	if err := c.checkMetaPacks(p.Meta); err != nil {
		return nil, err
	}
	if p.Kind == domain.KindMessage {
		// A message's sender is the caller; sessions cannot speak for others.
		var f domain.MessageFields
		if len(p.Fields) > 0 {
			if err := json.Unmarshal(p.Fields, &f); err != nil {
				return nil, fmt.Errorf("%w: message fields: %w", domain.ErrInvalid, err)
			}
		}
		if !sc.admin || f.From == "" {
			f.From = sc.actor
		}
		if p.Fields, err = json.Marshal(f); err != nil {
			return nil, err
		}
	}
	return store.Create(ctx, c.store, sc.actor, store.NewTicket{
		Project: proj, Kind: p.Kind, Title: p.Title, Body: p.Body, Priority: p.Priority,
		Pool: p.Pool, Holder: p.Holder, ParentID: p.ParentID, Labels: p.Labels,
		Fields: p.Fields, Meta: p.Meta,
	})
}

func (c *Controller) show(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	var r rpc.ShowResult
	if r.Ticket, err = c.store.Get(ctx, id); err != nil {
		return nil, err
	}
	th, err := c.store.Thread(ctx, id, store.AllParts)
	if err != nil {
		return nil, err
	}
	r.Edges, r.Comments, r.History = th.Edges, th.Comments, th.History
	return r, nil
}

func (c *Controller) list(ctx context.Context, sc scope, p rpc.ListParams) (any, error) {
	if !sc.admin {
		p.Project = sc.project
	}
	if p.Orphaned {
		return c.orphaned(ctx, c.config(), p.Project, p.Limit)
	}
	f := store.Filter{
		Query:   p.Query,
		Project: p.Project, Kind: p.Kind, Status: p.Status, Pool: p.Pool,
		Holder: p.Holder, ParentID: p.ParentID, Label: p.Label, Limit: p.Limit,
		Workflow: p.Workflow, State: p.State, Since: p.Since, After: p.After,
	}
	if p.Run != "" {
		if p.ParentID != "" && p.ParentID != p.Run {
			return nil, fmt.Errorf("%w: run and parent_id differ", domain.ErrInvalid)
		}
		f.ParentID = p.Run
	}
	if p.Newest {
		f.Order = store.OrderNewest
	}
	switch p.Order {
	case "", "priority":
	case "newest":
		f.Order = store.OrderNewest
	default:
		return nil, fmt.Errorf("%w: order %q: want priority or newest", domain.ErrInvalid, p.Order)
	}
	keys := slices.Sorted(maps.Keys(p.Meta))
	for _, k := range keys {
		pack, key, ok := strings.Cut(k, ".")
		if !ok || pack == "" || key == "" {
			return nil, fmt.Errorf("%w: meta filter %q: want pack.key", domain.ErrInvalid, k)
		}
		f.Meta = append(f.Meta, store.MetaMatch{Pack: pack, Key: key, Value: p.Meta[k]})
	}
	if p.Report != "" {
		f.MetaPack, f.MetaKey, f.MetaValue = reportMetaPack, "state", p.Report
	}
	return c.store.List(ctx, f)
}

func (c *Controller) ready(ctx context.Context, sc scope, p rpc.ReadyParams) (any, error) {
	proj, err := c.project(sc, p.Project)
	if err != nil {
		return nil, err
	}
	return c.store.List(ctx, store.Filter{Project: proj, Pool: p.Pool, Ready: true, Limit: p.Limit})
}

func (c *Controller) update(ctx context.Context, sc scope, p rpc.UpdateParams) (any, error) {
	if p.Holder != nil && !sc.admin {
		return nil, forbidden("assigning needs the admin socket")
	}
	if p.Holder != nil {
		if err := c.checkAssign(ctx, *p.Holder, p.Force); err != nil {
			return nil, err
		}
	}
	if err := c.checkMetaPacks(p.Meta); err != nil {
		return nil, err
	}
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	if !sc.admin && len(p.Fields) > 0 {
		if p.IfRevision, err = c.checkCommandSpec(ctx, id, p.Fields, p.IfRevision); err != nil {
			return nil, err
		}
	}
	return c.store.Update(ctx, sc.actor, id, store.Update{
		Title: p.Title, Body: p.Body, Priority: p.Priority, Pool: p.Pool, Holder: p.Holder,
		DeferUntil: p.DeferUntil, Fields: p.Fields, Meta: p.Meta, IfRevision: p.IfRevision,
	})
}

// checkCommandSpec refuses a session's fields update that adds, replaces
// or removes a step's command definition: what a command step runs was
// accepted from a callback, and a session may edit data, not that. It
// returns the revision the update must apply to, so the check and the
// write see the same ticket.
func (c *Controller) checkCommandSpec(ctx context.Context, id string, fields json.RawMessage, ifRevision int64) (int64, error) {
	ticket, err := c.store.Get(ctx, id)
	if err != nil || ticket.Kind != domain.KindStep {
		return ifRevision, err
	}
	before, err := domain.DecodeFields[domain.StepFields](ticket)
	if err != nil {
		return ifRevision, err
	}
	var after domain.StepFields
	if err := json.Unmarshal(fields, &after); err != nil {
		return ifRevision, fmt.Errorf("%w: step fields: %w", domain.ErrInvalid, err)
	}
	oldSpec, _ := json.Marshal(before.CommandSpec)
	newSpec, _ := json.Marshal(after.CommandSpec)
	if !bytes.Equal(oldSpec, newSpec) || before.CommandSpec != nil && before.Type != after.Type {
		return ifRevision, forbidden("changing command definitions needs the admin socket")
	}
	if ifRevision == 0 {
		ifRevision = ticket.Revision
	}
	return ifRevision, nil
}

func (c *Controller) closeTicket(ctx context.Context, sc scope, p rpc.CloseParams) (any, error) {
	return c.closeTicketIn(ctx, c.config(), sc, p)
}

// closeTicketIn is closeTicket under cfg: the pass's when the supervisor
// closes a chat (JOB-104).
func (c *Controller) closeTicketIn(ctx context.Context, cfg *config.Config, sc scope, p rpc.CloseParams) (any, error) {
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	if p.Resolution == "" {
		p.Resolution = "done"
	}
	if err := closeOutcome(&p); err != nil {
		return nil, err
	}
	var cl *store.Claim
	if !sc.admin {
		h, err := c.heldClaim(ctx, sc, id)
		if err != nil {
			return nil, err
		}
		cl = &h
	}
	if err := c.checkStepResolution(ctx, sc, id, p.Resolution); err != nil {
		return nil, err
	}
	if p.Resolution == domain.ResolutionRequestChanges {
		if p.Output != nil || p.Outcome != "" {
			return nil, fmt.Errorf("%w: request_changes accepts findings, not a successful result", domain.ErrInvalid)
		}
		t, err := c.closeRequestChanges(ctx, cfg, sc, id, cl, p.Comment)
		if err != nil {
			return nil, err
		}
		// Within the cap the step is open again: count the close it made.
		counted := t
		counted.Resolution = domain.ResolutionRequestChanges
		c.countClose(ctx, counted)
		c.checkRun(ctx, cfg, t.ParentID)
		return t, nil
	}
	if strings.TrimSpace(p.Comment) != "" {
		if _, err := store.AddComment(ctx, c.store, sc.actor, id, p.Comment); err != nil {
			return nil, err
		}
	}
	var result *store.StepResult
	cur, err := c.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.Outcome != "" || p.Output != nil || cur.Kind == domain.KindStep {
		if cur.Kind != domain.KindStep {
			return nil, fmt.Errorf("%w: outcome and output require a workflow step", domain.ErrInvalid)
		}
		if p.Outcome != "" && !config.OutcomePattern.MatchString(p.Outcome) {
			return nil, fmt.Errorf("%w: invalid outcome %q", domain.ErrInvalid, p.Outcome)
		}
		f, err := domain.DecodeFields[domain.StepFields](cur)
		if err != nil {
			return nil, err
		}
		if p.Resolution == "done" {
			if err := hooks.CheckStepOutput(cfg.Automation.OutputBytes, f.OutputSchema, p.Output); err != nil {
				p.Resolution, p.Outcome, p.Output = "failed", "", nil
				if _, err := store.AddComment(ctx, c.store, "controller", id, "Invalid structured output: "+err.Error()); err != nil {
					return nil, err
				}
			}
		}
		result = &store.StepResult{Revision: cur.Revision, Outcome: p.Outcome, Output: p.Output}
	}
	if p.Resolution == domain.ResolutionRebase {
		if cur, err := c.store.Get(ctx, id); err == nil {
			res, why := c.rebaseResolution(ctx, cur)
			if why != "" {
				if _, err := store.AddComment(ctx, c.store, "controller", id, why); err != nil {
					c.log.WarnContext(ctx, "rebase check comment", "step", id, "err", err)
				}
			}
			p.Resolution = res
		}
	}
	t, err := c.store.Release(ctx, sc.actor, id, cl, store.Outcome{Close: p.Resolution, StepResult: result})
	if err != nil {
		return nil, err
	}
	c.countClose(ctx, t)
	if p.Resolution == "failed" {
		c.recordStepFailure(ctx, t, "closed failed", failedReason(t))
	}
	if t.Kind == domain.KindStep && t.ParentID != "" {
		c.checkRun(ctx, cfg, t.ParentID)
	}
	if t.Kind == domain.KindSession {
		c.stopRunner(t.ID)
		c.endSession(ctx, cfg, t.ID, "session closed")
	}
	c.escalationClosed(ctx, t)
	return t, nil
}

// checkStepResolution refuses a step close whose resolution the run
// engine does not act on (JOB-92). A free-text resolution closed the step
// but never finished it (store.StepFinished), so the run stayed open.
// skipped is the controller's: only an admin caller may set it.
func (c *Controller) checkStepResolution(ctx context.Context, sc scope, id, res string) error {
	switch res {
	case "done", "failed", domain.ResolutionRebase, domain.ResolutionRequestChanges, domain.RunCancelled:
		return nil
	case domain.StepSkipped:
		if sc.admin {
			return nil
		}
	}
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return err
	}
	if t.Kind != domain.KindStep {
		return nil
	}
	return fmt.Errorf("%w: step %s takes resolution done, failed, rebase, request_changes or cancelled, not %q; put a summary in a comment (jobshop comment) and a verdict in --outcome",
		domain.ErrInvalid, id, res)
}

// endSession stops a session's socket, frees everything it held and removes
// its worktree.
func (c *Controller) endSession(ctx context.Context, cfg *config.Config, id, reason string) {
	c.closeSession(id)
	c.setReadOnly(id, false)
	// Every caller stopped the host first, so exit.json is written.
	c.saveEnded(ctx, id)
	if t, err := c.store.Get(ctx, id); err == nil {
		c.removeWorktree(cfg, t, "endSession: "+reason)
		c.removeSlotHome(cfg, t)
	}
	released, err := c.store.Expire(ctx, "controller", id, reason)
	if err != nil {
		c.log.ErrorContext(ctx, "release session claims", "session", id, "err", err)
		return
	}
	c.stepEvents(ctx, cfg, released)
	if len(released) > 0 {
		c.log.InfoContext(ctx, "released session claims", "session", id, "tickets", released)
	}
}

func (c *Controller) reopen(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	if err := requireAdmin(sc, "reopen"); err != nil {
		return nil, err
	}
	return store.Reopen(ctx, c.store, sc.actor, p.ID)
}

func (c *Controller) comment(ctx context.Context, sc scope, p rpc.CommentParams) (any, error) {
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	c.redactWrite(sc, "comment on", id, &p.Body)
	return store.AddComment(ctx, c.store, sc.actor, id, p.Body)
}

func (c *Controller) edgeAdd(ctx context.Context, sc scope, p rpc.EdgeParams) (any, error) {
	return struct{}{}, store.AddEdge(ctx, c.store, sc.actor, domain.Edge{From: p.From, To: p.To, Kind: p.Kind})
}

func (c *Controller) edgeRemove(ctx context.Context, sc scope, p rpc.EdgeParams) (any, error) {
	return struct{}{}, store.RemoveEdge(ctx, c.store, sc.actor, domain.Edge{From: p.From, To: p.To, Kind: p.Kind})
}

func (c *Controller) labelAdd(ctx context.Context, sc scope, p rpc.LabelParams) (any, error) {
	if err := store.AddLabel(ctx, c.store, sc.actor, p.ID, p.Label); err != nil {
		return nil, err
	}
	if p.Label == needsHumanLabel {
		c.holdChat(ctx, p.ID)
	}
	return struct{}{}, nil
}

func (c *Controller) labelRemove(ctx context.Context, sc scope, p rpc.LabelParams) (any, error) {
	return struct{}{}, store.RemoveLabel(ctx, c.store, sc.actor, p.ID, p.Label)
}

func (c *Controller) claim(ctx context.Context, sc scope, p rpc.ClaimParams) (any, error) {
	holder, pool := p.Holder, p.Pool
	if !sc.admin {
		if holder != "" && holder != sc.session {
			return nil, forbidden("a session claims only for itself")
		}
		holder = sc.session
		if pool == "" {
			pool = sc.pool
		}
	} else if holder == "" {
		return nil, fmt.Errorf("%w: admin claims need --holder", domain.ErrInvalid)
	}
	var (
		r   rpc.ClaimResult
		err error
	)
	if p.ID != "" {
		if t, gerr := c.store.Get(ctx, p.ID); gerr == nil && (t.Status != domain.StatusInProgress || t.Holder != holder) {
			if err := c.claimHook(ctx, sc, t); err != nil {
				return nil, err
			}
		}
		r.Ticket, r.Claim, err = c.store.Claim(ctx, p.ID, holder, c.leaseDur())
	} else {
		proj, perr := c.project(sc, p.Project)
		if perr != nil {
			return nil, perr
		}
		// Resuming held work is not a new claim, so it skips claim hooks.
		held, _ := c.store.List(ctx, store.Filter{Holder: holder, Status: domain.StatusInProgress, Limit: 1})
		if len(held) == 0 && c.sessionCordon(ctx, holder) != nil {
			// A cordoned session takes no new work (#281); it may still
			// claim a ticket pre-assigned to it by id.
			return nil, fmt.Errorf("%w: session %s is cordoned", domain.ErrNoWork, holder)
		}
		r.Ticket, r.Claim, err = c.store.ClaimNext(ctx, store.ClaimNext{
			Project: proj, Pool: pool, Holder: holder, Lease: c.leaseDur(),
		})
		if err == nil && len(held) == 0 {
			if herr := c.claimHook(ctx, sc, r.Ticket); herr != nil {
				// ClaimNext picks atomically, so the hook sees the ticket
				// after the pick. A refused pick is deferred, not failed,
				// so the next claim moves on to other work.
				until := c.now().Add(claimDenyDefer)
				if _, rerr := c.store.Release(ctx, "", r.Claim.TicketID, &r.Claim, store.Outcome{Reason: herr.Error(), Until: until}); rerr != nil {
					c.log.ErrorContext(ctx, "release after claim hook", "ticket", r.Ticket.ID, "err", rerr)
				}
				return nil, herr
			}
		}
	}
	if err != nil {
		return nil, err
	}
	c.metrics.Claims.Add(ctx, 1, metric.WithAttributes(telemetry.AttrPool.String(r.Ticket.Pool)))
	if r.Ticket.Kind == domain.KindStep && r.Ticket.ParentID != "" {
		c.runBatchEvent(ctx, r.Ticket.ParentID) // a batch is running once a member step is claimed
	}
	return r, nil
}

func (c *Controller) countClose(ctx context.Context, t domain.Ticket) {
	c.metrics.Closes.Add(ctx, 1, metric.WithAttributes(
		telemetry.AttrPool.String(t.Pool), telemetry.AttrKind.String(string(t.Kind)),
		telemetry.AttrResolution.String(t.Resolution)))
}

func (c *Controller) heartbeat(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	cl, err := c.heldClaim(ctx, sc, id)
	if err != nil {
		return nil, err
	}
	return c.store.Heartbeat(ctx, cl, c.leaseDur())
}

func (c *Controller) release(ctx context.Context, sc scope, p rpc.ReleaseParams) (any, error) {
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	cl, err := c.heldClaim(ctx, sc, id)
	if errors.Is(err, domain.ErrNotHolder) {
		// An escalation chat may hand back a blocked worker's ticket
		// (pool grant ticket-release), in its own project only.
		cl, err = c.grantedClaim(ctx, sc, id, err)
	}
	if err != nil {
		return nil, err
	}
	var until time.Time
	if p.Until != nil {
		if !p.Until.After(c.now()) {
			return nil, fmt.Errorf("%w: until must be in the future", domain.ErrInvalid)
		}
		if !p.TimeWait {
			if err := workflow.TimerHint(p.Reason); err != nil {
				return nil, err
			}
		}
		until = *p.Until
	}
	t, err := c.store.Release(ctx, "", cl.TicketID, &cl, store.Outcome{Reason: p.Reason, Until: until})
	if err != nil {
		return nil, err
	}
	if until.IsZero() {
		// A plain release of a step counts as a failed attempt.
		c.recordStepFailure(ctx, t, "released", p.Reason)
	}
	c.stepEvent(ctx, c.config(), t)
	return t, nil
}

// handoff re-pools, releases and notes a ticket in one store write (#492).
// A session hands off only the ticket it holds; admin, or a pool granted
// "handoff" (triage), any open ticket.
func (c *Controller) handoff(ctx context.Context, sc scope, p rpc.HandoffParams) (any, error) {
	if p.Pool == "" {
		return nil, fmt.Errorf("%w: handoff needs --to <pool>", domain.ErrInvalid)
	}
	if _, ok := c.config().Pools[p.Pool]; !ok {
		return nil, fmt.Errorf("%w: unknown pool %q", domain.ErrInvalid, p.Pool)
	}
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	var cl *store.Claim
	if c.requireGrant(sc, "handoff", "handoff") != nil {
		h, err := c.heldClaim(ctx, sc, id)
		if err != nil {
			return nil, forbidden("handoff of a ticket this session does not hold: %v", err)
		}
		cl = &h
	}
	t, err := c.store.Release(ctx, sc.actor, id, cl, store.Outcome{Pool: p.Pool, Note: p.Note})
	if err != nil {
		return nil, err
	}
	c.stepEvent(ctx, c.config(), t)
	return t, nil
}

func (c *Controller) sessionNew(ctx context.Context, sc scope, p rpc.SessionNewParams) (any, error) {
	if err := requireAdmin(sc, "session new"); err != nil {
		return nil, err
	}
	proj, err := c.project(sc, p.Project)
	if err != nil {
		return nil, err
	}
	cfg := c.config()
	if p.Pool != "" {
		if pool, ok := cfg.Pools[p.Pool]; ok && p.Agent == "" {
			p.Agent = pool.Agent
		}
		if err := c.humanSlot(ctx, cfg, p.Pool); err != nil {
			return nil, err
		}
	}
	t, err := c.newSession(ctx, cfg, sc.actor, proj, p.Pool, p.Agent, p.Key, p.Title, p.DisplayName)
	if err != nil {
		return nil, err
	}
	return rpc.SessionNewResult{Ticket: t, Socket: c.socketPath(t.ID)}, nil
}

// sessionList joins session tickets with the live state of their session
// directories, so store and processes can be compared (plan 8.4 item 5).
func (c *Controller) sessionList(ctx context.Context, sc scope, p rpc.SessionListParams) (any, error) {
	if err := c.requireGrant(sc, "session-list", "session list"); err != nil {
		return nil, err
	}
	if p.Since < 0 {
		return nil, fmt.Errorf("%w: since must not be negative", domain.ErrInvalid)
	}
	ended := p.State != "" || p.Since > 0
	statuses := []domain.Status{domain.StatusOpen, domain.StatusInProgress}
	switch {
	case ended:
		statuses = []domain.Status{domain.StatusClosed}
	case p.All:
		statuses = []domain.Status{""}
	}
	var cutoff time.Time
	if p.Since > 0 {
		cutoff = c.now().Add(-p.Since)
	}
	var tickets []domain.Ticket
	for _, s := range statuses {
		ts, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: s, Limit: store.NoLimit})
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ts...)
	}

	// One query for what every session holds, not one per session.
	inProgress, err := c.store.List(ctx, store.Filter{Status: domain.StatusInProgress, Limit: math.MaxInt32})
	if err != nil {
		return nil, err
	}
	held := map[string][]domain.Ticket{}
	for _, t := range inProgress {
		if t.Holder != "" {
			held[t.Holder] = append(held[t.Holder], t)
		}
	}

	var open []domain.Ticket
	if !ended {
		for _, t := range tickets {
			if t.Status != domain.StatusClosed {
				open = append(open, t)
			}
		}
	}
	// The supervisor pass walks the session directories (JOB-32); a
	// request only rechecks the hosts it found.
	orphans := map[string]bool{}
	for _, id := range c.scannedOrphans(open) {
		orphans[id] = true
	}

	out := []rpc.SessionInfo{}
	seen := map[string]bool{}
	for i := range tickets {
		t := tickets[i]
		if p.State != "" && t.Resolution != p.State {
			continue
		}
		if !cutoff.IsZero() && t.ClosedAt.Before(cutoff) {
			continue
		}
		var info rpc.SessionInfo
		if t.Status == domain.StatusClosed && !orphans[t.ID] {
			// Its host is stopped: the state saved at the end is current.
			info = closedSessionInfo(t)
		} else {
			info = c.sessionInfoOf(t.ID, &t, held[t.ID])
		}
		info.Ticket = &t
		// A closed ticket whose host still runs is a leak too.
		info.Orphan = t.Status == domain.StatusClosed && info.Alive
		seen[t.ID] = true
		out = append(out, info)
	}
	if ended {
		return out, nil
	}

	ids := make([]string, 0, len(orphans))
	for id := range orphans {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		if seen[id] || !session.HostAlive(c.root.SessionDir(id)) {
			continue
		}
		info, err := c.sessionInfo(ctx, id)
		if err != nil {
			return nil, err
		}
		if t, err := c.store.Get(ctx, id); err == nil {
			info.Ticket = &t
			if t.Kind == domain.KindSession && t.Status != domain.StatusClosed {
				continue // opened since the pass
			}
		}
		info.Orphan = true
		out = append(out, info)
	}
	return out, nil
}

// sessionInfo reads the live state of one session directory.
func (c *Controller) sessionInfo(ctx context.Context, id string) (rpc.SessionInfo, error) {
	var sess *domain.Ticket
	if t, err := c.store.Get(ctx, id); err == nil {
		sess = &t
	}
	held, err := c.store.List(ctx, store.Filter{Holder: id, Status: domain.StatusInProgress})
	if err != nil {
		return rpc.SessionInfo{ID: id}, err
	}
	return c.sessionInfoOf(id, sess, held), nil
}

// sessionInfoOf is sessionInfo given the session's ticket (nil if it has
// none) and the in-progress tickets it holds.
func (c *Controller) sessionInfoOf(id string, sess *domain.Ticket, held []domain.Ticket) rpc.SessionInfo {
	dir := c.root.SessionDir(id)
	info := rpc.SessionInfo{ID: id, Alive: session.HostAlive(dir), HostPid: session.HostPid(dir)}
	c.mu.Lock()
	r := c.runners[id]
	if r != nil {
		r.mu.Lock()
		info.AgentPid = r.agentPid
		if !r.backoffUntil.IsZero() {
			bu := r.backoffUntil
			info.BackoffUntil = &bu
		}
		r.mu.Unlock()
		info.Crashes = r.crashCount()
		if r.inTurn.Load() || r.turning.Load() {
			// turning also covers the unclosed reminder, a turn inTurn
			// does not (JOB-184): the timeline reads the same flag.
			info.Busy = true
			t := r.traffic.Running()
			info.Tool, info.ToolTarget, info.ToolIntent = t.Title, t.Target, t.Intent
		}
	}
	c.mu.Unlock()
	if info.ToolTarget != "" || info.ToolIntent != "" {
		// A command line can carry a secret: redact as the journal is.
		red := c.redactor(sessionEnv(dir))
		info.ToolTarget, info.ToolIntent = red.Text(info.ToolTarget), red.Text(info.ToolIntent)
	}
	info.Drained = r != nil && c.poolDrained(r.pool)
	var sessTicket domain.Ticket
	if sess != nil {
		sessTicket = *sess
		info.Ticket = sess
		if f, err := domain.DecodeFields[domain.SessionFields](sessTicket); err == nil {
			info.DisplayName = f.DisplayName
			if f.Cordoned != nil {
				info.Cordon = f.Cordoned.Reason
			}
			info.Model, info.Effort, info.Override = f.Model, f.Effort, f.Override
			if f.AgentName != "" {
				info.Agent = &rpc.AgentInfo{Name: f.AgentName, Version: f.AgentVersion}
			}
		}
	}
	if !info.Alive {
		info.AgentPid = 0
		if ex, ok := host.ReadExit(dir); ok {
			info.Exit = ex.Reason
		}
	}
	if seq, at := host.JournalTail(dir); seq > 0 {
		info.Seq, info.SeqTime = seq, &at
	}
	if len(held) > 0 {
		info.Current = held[0].ID
	}
	if info.DisplayName == "" {
		info.DisplayName = defaultDisplayName(sessTicket, held)
	}
	return info
}

// defaultDisplayName labels a session with no name of its own (#626):
// its pool and the title of the ticket it works, else its ticket title.
func defaultDisplayName(sess domain.Ticket, held []domain.Ticket) string {
	if len(held) > 0 && held[0].Title != "" {
		if sess.Pool != "" {
			return sess.Pool + ": " + held[0].Title
		}
		return held[0].Title
	}
	return sess.Title
}

// sessionRestart stops a session's host so its driver respawns it with a
// freshly built spec and resumes the conversation (plan 8.4). The
// session's claims are kept and it is not counted as a crash.
func (c *Controller) sessionRestart(ctx context.Context, sc scope, p rpc.SessionRestartParams) (any, error) {
	if err := c.requireGrant(sc, "session-restart", "session restart"); err != nil {
		return nil, err
	}
	if p.ID == "" {
		return nil, fmt.Errorf("%w: session id is required", domain.ErrInvalid)
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Kind != domain.KindSession {
		return nil, fmt.Errorf("%w: %s is a %s, not a session", domain.ErrInvalid, p.ID, t.Kind)
	}
	if t.Status == domain.StatusClosed {
		return nil, fmt.Errorf("%w: session %s", domain.ErrClosed, p.ID)
	}
	c.mu.Lock()
	r := c.runners[p.ID]
	c.mu.Unlock()
	if r == nil {
		return nil, fmt.Errorf("%w: session %s has no controller-driven host", domain.ErrInvalid, p.ID)
	}
	c.log.InfoContext(ctx, "session restart requested", "session", p.ID, "now", p.Now)
	r.requestRestart(p.Now)
	return rpc.SessionRestartResult{ID: p.ID, HostPid: session.HostPid(c.root.SessionDir(p.ID)), Now: p.Now}, nil
}

// sessionCancel cancels the running turn of a controller-driven session
// (#535). Nothing is sent; messages already queued are delivered next.
func (c *Controller) sessionCancel(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	if err := requireAdmin(sc, "session cancel"); err != nil {
		return nil, err
	}
	if p.ID == "" {
		return nil, fmt.Errorf("%w: session id is required", domain.ErrInvalid)
	}
	c.mu.Lock()
	r := c.runners[p.ID]
	c.mu.Unlock()
	if r == nil {
		return nil, fmt.Errorf("%w: session %s has no controller-driven host", domain.ErrInvalid, p.ID)
	}
	if !r.inTurn.Load() {
		return rpc.SessionCancelResult{ID: p.ID}, nil
	}
	select {
	case r.cancelNow <- struct{}{}:
	default:
	}
	c.log.InfoContext(ctx, "session cancel requested", "session", p.ID)
	return rpc.SessionCancelResult{ID: p.ID, Cancelled: true}, nil
}

// sessionReap stops orphaned hosts (the startup reapOrphans rule) and, with
// Closed, deletes old closed sessions' directories (plan 8.4 item 5). An
// open session is never touched: only closed tickets' directories are
// removed, and only after their host is dead.
func (c *Controller) sessionReap(ctx context.Context, sc scope, p rpc.SessionReapParams) (any, error) {
	if err := requireAdmin(sc, "session reap"); err != nil {
		return nil, err
	}
	if p.OlderThan <= 0 {
		p.OlderThan = 24 * time.Hour
	}
	res := rpc.SessionReapResult{Stopped: []string{}, Removed: []string{}, DryRun: p.DryRun}
	open, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusOpen})
	if err != nil {
		return nil, err
	}
	for _, id := range c.orphans(open) {
		if !p.DryRun {
			c.log.WarnContext(ctx, "stopping orphaned host", "session", id)
			c.stopHost(c.root.SessionDir(id))
		}
		res.Stopped = append(res.Stopped, id)
	}
	if !p.Closed {
		return res, nil
	}
	closed, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusClosed})
	if err != nil {
		return nil, err
	}
	cutoff := c.now().Add(-p.OlderThan)
	for _, t := range closed {
		dir := c.root.SessionDir(t.ID)
		if t.ClosedAt.IsZero() || t.ClosedAt.After(cutoff) {
			continue
		}
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		// Still alive after the stop above means stopHost could not end it;
		// leave it for a later run rather than delete under a live host.
		if !p.DryRun && session.HostAlive(dir) {
			c.log.WarnContext(ctx, "reap: host still alive, keeping dir", "session", t.ID)
			continue
		}
		if !p.DryRun {
			c.removeWorktree(c.config(), t, "session reap")
			c.log.WarnContext(ctx, "reap: removing session dir", "session", t.ID, "dir", dir)
			if err := os.RemoveAll(c.attachmentDir(t.ID)); err != nil {
				return res, err
			}
			if err := os.RemoveAll(dir); err != nil {
				return res, fmt.Errorf("remove %s: %w", dir, err)
			}
		}
		res.Removed = append(res.Removed, t.ID)
	}
	return res, nil
}

// clk is the controller's clock; the wall clock for a Controller built
// without Run (internal tests).
func (c *Controller) clk() clock.Clock {
	if c.opt.Clock != nil {
		return c.opt.Clock
	}
	return clock.Real{}
}

func (c *Controller) now() time.Time { return c.clk().Now() }

// after is time.After on the controller's clock.
func (c *Controller) after(d time.Duration) <-chan time.Time { return c.clk().After(d) }

// since is time.Since on the controller's clock.
func (c *Controller) since(t time.Time) time.Duration { return c.now().Sub(t) }

func (c *Controller) shutdown(_ context.Context, sc scope, p rpc.ShutdownParams) (any, error) {
	if err := requireAdmin(sc, "shutdown"); err != nil {
		return nil, err
	}
	if p.All {
		// Ending sessions takes a while; the caller waits for the socket to
		// close rather than for this response.
		go c.stopAll()
		return struct{}{}, nil
	}
	// Let the response flush before sockets close.
	time.AfterFunc(50*time.Millisecond, c.cancel)
	return struct{}{}, nil
}

// checkMetaPacks rejects top-level meta keys that are not known packs, so
// a typo such as "gihtub" is not stored where no pack filter finds it.
func (c *Controller) checkMetaPacks(raw json.RawMessage) error {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil // shape errors come from store.NormalizeMeta
	}
	cfg := c.live.Load().cfg
	for name := range m {
		if !cfg.KnownMetaPack(name) {
			return fmt.Errorf("%w: meta: unknown pack %q (add it to meta_packs in %s)", domain.ErrInvalid, name, config.FileName)
		}
	}
	return nil
}
