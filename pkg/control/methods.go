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
	"github.com/elee1766/jobshop/pkg/clock"
	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
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
		return "", fmt.Errorf("%w: no project given and no default_project configured", store.ErrInvalid)
	}
	if _, ok := c.config().Projects[asked]; !ok {
		return "", fmt.Errorf("%w: unknown project %q", store.ErrInvalid, asked)
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
		return "", fmt.Errorf("%w: ticket id is required", store.ErrInvalid)
	}
	held, err := c.store.List(ctx, store.Filter{Holder: sc.session, Status: store.StatusInProgress})
	if err != nil {
		return "", err
	}
	switch len(held) {
	case 0:
		return "", fmt.Errorf("%w: session holds no ticket; pass an id", store.ErrInvalid)
	case 1:
		return held[0].ID, nil
	}
	return "", fmt.Errorf("%w: session holds %d tickets; pass an id", store.ErrInvalid, len(held))
}

// heldClaim returns the claim for id. Sessions must hold it; admin acts on
// whatever claim is current.
func (c *Controller) heldClaim(ctx context.Context, sc scope, id string) (store.Claim, error) {
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return store.Claim{}, err
	}
	if t.Status != store.StatusInProgress {
		return store.Claim{}, fmt.Errorf("%w: %s is %s", store.ErrNotHolder, id, t.Status)
	}
	if !sc.admin && t.Holder != sc.session {
		return store.Claim{}, fmt.Errorf("%w: %s is held by %s", store.ErrNotHolder, id, t.Holder)
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
	if err != nil || t.Status != store.StatusInProgress || t.Project != sc.project {
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
		p.Kind = store.KindTask
	}
	if p.Kind == store.KindSession {
		return nil, fmt.Errorf("%w: use session new to create sessions", store.ErrInvalid)
	}
	if p.Kind == store.KindRun || p.Kind == store.KindStep {
		return nil, fmt.Errorf("%w: use run start to create runs and steps", store.ErrInvalid)
	}
	if p.Kind == store.KindBatch {
		return nil, fmt.Errorf("%w: use batch start to create batches", store.ErrInvalid)
	}
	if p.Holder != "" && !sc.admin {
		return nil, forbidden("assigning needs the admin socket")
	}
	if err := c.checkAssign(ctx, p.Holder, p.Force); err != nil {
		return nil, err
	}
	// A session's follow-up task with no pool would never be claimed (#244);
	// it goes to the creator's pool, which can reroute it.
	if p.Kind == store.KindTask && p.Pool == "" && !sc.admin {
		p.Pool = sc.pool
	}
	if err := c.checkMetaPacks(p.Meta); err != nil {
		return nil, err
	}
	if p.Kind == store.KindMessage {
		// A message's sender is the caller; sessions cannot speak for others.
		var f store.MessageFields
		if len(p.Fields) > 0 {
			if err := json.Unmarshal(p.Fields, &f); err != nil {
				return nil, fmt.Errorf("%w: message fields: %w", store.ErrInvalid, err)
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
		return c.orphaned(ctx, p.Project, p.Limit)
	}
	f := store.Filter{
		Project: p.Project, Kind: p.Kind, Status: p.Status, Pool: p.Pool,
		Holder: p.Holder, ParentID: p.ParentID, Label: p.Label, Limit: p.Limit,
		Workflow: p.Workflow, State: p.State, Since: p.Since, After: p.After,
	}
	if p.Run != "" {
		if p.ParentID != "" && p.ParentID != p.Run {
			return nil, fmt.Errorf("%w: run and parent_id differ", store.ErrInvalid)
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
		return nil, fmt.Errorf("%w: order %q: want priority or newest", store.ErrInvalid, p.Order)
	}
	keys := slices.Sorted(maps.Keys(p.Meta))
	for _, k := range keys {
		pack, key, ok := strings.Cut(k, ".")
		if !ok || pack == "" || key == "" {
			return nil, fmt.Errorf("%w: meta filter %q: want pack.key", store.ErrInvalid, k)
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
	return c.store.Update(ctx, sc.actor, id, store.Update{
		Title: p.Title, Body: p.Body, Priority: p.Priority, Pool: p.Pool, Holder: p.Holder,
		DeferUntil: p.DeferUntil, Fields: p.Fields, Meta: p.Meta, IfRevision: p.IfRevision,
	})
}

func (c *Controller) closeTicket(ctx context.Context, sc scope, p rpc.CloseParams) (any, error) {
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
	if p.Resolution == store.ResolutionRequestChanges {
		t, err := c.closeRequestChanges(ctx, sc, id, cl, p.Comment)
		if err != nil {
			return nil, err
		}
		// Within the cap the step is open again: count the close it made.
		counted := t
		counted.Resolution = store.ResolutionRequestChanges
		c.countClose(ctx, counted)
		c.checkRun(ctx, t.ParentID)
		return t, nil
	}
	if strings.TrimSpace(p.Comment) != "" {
		if _, err := store.AddComment(ctx, c.store, sc.actor, id, p.Comment); err != nil {
			return nil, err
		}
	}
	if p.Outcome != "" {
		if err := c.setStepOutcome(ctx, sc, id, p.Outcome); err != nil {
			return nil, err
		}
	}
	if p.Resolution == store.ResolutionRebase {
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
	t, err := store.Close(ctx, c.store, sc.actor, id, cl, p.Resolution)
	if err != nil {
		return nil, err
	}
	c.countClose(ctx, t)
	if p.Resolution == "failed" {
		c.recordStepFailure(ctx, t, "closed failed", failedReason(t))
	}
	if t.Kind == store.KindStep && t.ParentID != "" {
		c.checkRun(ctx, t.ParentID)
	}
	if t.Kind == store.KindSession {
		c.stopRunner(t.ID)
		c.endSession(ctx, t.ID, "session closed")
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
	case "done", "failed", store.ResolutionRebase, store.ResolutionRequestChanges, store.RunCancelled:
		return nil
	case store.StepSkipped:
		if sc.admin {
			return nil
		}
	}
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return err
	}
	if t.Kind != store.KindStep {
		return nil
	}
	return fmt.Errorf("%w: step %s takes resolution done, failed, rebase, request_changes or cancelled, not %q; put a summary in a comment (jobshop comment) and a verdict in --outcome",
		store.ErrInvalid, id, res)
}

// endSession stops a session's socket, frees everything it held and removes
// its worktree.
func (c *Controller) endSession(ctx context.Context, id, reason string) {
	c.closeSession(id)
	c.setReadOnly(id, false)
	// Every caller stopped the host first, so exit.json is written.
	c.saveEnded(ctx, id)
	if t, err := c.store.Get(ctx, id); err == nil {
		c.removeWorktree(t, "endSession: "+reason)
		c.removeSlotHome(t)
	}
	released, err := c.store.Expire(ctx, "controller", id, reason)
	if err != nil {
		c.log.ErrorContext(ctx, "release session claims", "session", id, "err", err)
		return
	}
	c.stepEvents(ctx, released)
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
	return struct{}{}, store.AddEdge(ctx, c.store, sc.actor, store.Edge{From: p.From, To: p.To, Kind: p.Kind})
}

func (c *Controller) edgeRemove(ctx context.Context, sc scope, p rpc.EdgeParams) (any, error) {
	return struct{}{}, store.RemoveEdge(ctx, c.store, sc.actor, store.Edge{From: p.From, To: p.To, Kind: p.Kind})
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
		return nil, fmt.Errorf("%w: admin claims need --holder", store.ErrInvalid)
	}
	var (
		r   rpc.ClaimResult
		err error
	)
	if p.ID != "" {
		if t, gerr := c.store.Get(ctx, p.ID); gerr == nil && (t.Status != store.StatusInProgress || t.Holder != holder) {
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
		held, _ := c.store.List(ctx, store.Filter{Holder: holder, Status: store.StatusInProgress, Limit: 1})
		if len(held) == 0 && c.sessionCordon(ctx, holder) != nil {
			// A cordoned session takes no new work (#281); it may still
			// claim a ticket pre-assigned to it by id.
			return nil, fmt.Errorf("%w: session %s is cordoned", store.ErrNoWork, holder)
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
	if r.Ticket.Kind == store.KindStep && r.Ticket.ParentID != "" {
		c.runBatchEvent(ctx, r.Ticket.ParentID) // a batch is running once a member step is claimed
	}
	return r, nil
}

func (c *Controller) countClose(ctx context.Context, t store.Ticket) {
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
	if errors.Is(err, store.ErrNotHolder) {
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
			return nil, fmt.Errorf("%w: until must be in the future", store.ErrInvalid)
		}
		if !p.TimeWait {
			if err := eventWaitHint(p.Reason); err != nil {
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
	c.stepEvent(ctx, t)
	return t, nil
}

// handoff re-pools, releases and notes a ticket in one store write (#492).
// A session hands off only the ticket it holds; admin, or a pool granted
// "handoff" (stall-check), any open ticket.
func (c *Controller) handoff(ctx context.Context, sc scope, p rpc.HandoffParams) (any, error) {
	if p.Pool == "" {
		return nil, fmt.Errorf("%w: handoff needs --to <pool>", store.ErrInvalid)
	}
	if _, ok := c.config().Pools[p.Pool]; !ok {
		return nil, fmt.Errorf("%w: unknown pool %q", store.ErrInvalid, p.Pool)
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
	c.stepEvent(ctx, t)
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
	if p.Pool != "" {
		if pool, ok := c.config().Pools[p.Pool]; ok && p.Agent == "" {
			p.Agent = pool.Agent
		}
		if err := c.humanSlot(ctx, p.Pool); err != nil {
			return nil, err
		}
	}
	t, err := c.newSession(ctx, sc.actor, proj, p.Pool, p.Agent, p.Key, p.Title, p.DisplayName)
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
		return nil, fmt.Errorf("%w: since must not be negative", store.ErrInvalid)
	}
	ended := p.State != "" || p.Since > 0
	statuses := []store.Status{store.StatusOpen, store.StatusInProgress}
	switch {
	case ended:
		statuses = []store.Status{store.StatusClosed}
	case p.All:
		statuses = []store.Status{""}
	}
	var cutoff time.Time
	if p.Since > 0 {
		cutoff = c.now().Add(-p.Since)
	}
	var tickets []store.Ticket
	for _, s := range statuses {
		ts, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: s})
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ts...)
	}

	// One query for what every session holds, not one per session.
	inProgress, err := c.store.List(ctx, store.Filter{Status: store.StatusInProgress, Limit: math.MaxInt32})
	if err != nil {
		return nil, err
	}
	held := map[string][]store.Ticket{}
	for _, t := range inProgress {
		if t.Holder != "" {
			held[t.Holder] = append(held[t.Holder], t)
		}
	}

	var open []store.Ticket
	if !ended {
		for _, t := range tickets {
			if t.Status != store.StatusClosed {
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
		if t.Status == store.StatusClosed && !orphans[t.ID] {
			// Its host is stopped: the state saved at the end is current.
			info = closedSessionInfo(t)
		} else {
			info = c.sessionInfoOf(t.ID, &t, held[t.ID])
		}
		info.Ticket = &t
		// A closed ticket whose host still runs is a leak too.
		info.Orphan = t.Status == store.StatusClosed && info.Alive
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
		if seen[id] || !hostAlive(c.root.SessionDir(id)) {
			continue
		}
		info, err := c.sessionInfo(ctx, id)
		if err != nil {
			return nil, err
		}
		if t, err := c.store.Get(ctx, id); err == nil {
			info.Ticket = &t
			if t.Kind == store.KindSession && t.Status != store.StatusClosed {
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
	var sess *store.Ticket
	if t, err := c.store.Get(ctx, id); err == nil {
		sess = &t
	}
	held, err := c.store.List(ctx, store.Filter{Holder: id, Status: store.StatusInProgress})
	if err != nil {
		return rpc.SessionInfo{ID: id}, err
	}
	return c.sessionInfoOf(id, sess, held), nil
}

// sessionInfoOf is sessionInfo given the session's ticket (nil if it has
// none) and the in-progress tickets it holds.
func (c *Controller) sessionInfoOf(id string, sess *store.Ticket, held []store.Ticket) rpc.SessionInfo {
	dir := c.root.SessionDir(id)
	info := rpc.SessionInfo{ID: id, Alive: hostAlive(dir), HostPid: hostPid(dir)}
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
		if r.inTurn.Load() {
			info.Busy = true
			t := r.runningTool()
			info.Tool, info.ToolTarget, info.ToolIntent = t.title, t.target, t.intent
		}
	}
	c.mu.Unlock()
	if info.ToolTarget != "" || info.ToolIntent != "" {
		// A command line can carry a secret: redact as the journal is.
		red := c.redactor(journalEnv(dir))
		info.ToolTarget, info.ToolIntent = red.Text(info.ToolTarget), red.Text(info.ToolIntent)
	}
	info.Drained = r != nil && c.poolDrained(r.pool)
	var sessTicket store.Ticket
	if sess != nil {
		sessTicket = *sess
		if f, err := store.DecodeFields[store.SessionFields](sessTicket); err == nil {
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
func defaultDisplayName(sess store.Ticket, held []store.Ticket) string {
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
		return nil, fmt.Errorf("%w: session id is required", store.ErrInvalid)
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Kind != store.KindSession {
		return nil, fmt.Errorf("%w: %s is a %s, not a session", store.ErrInvalid, p.ID, t.Kind)
	}
	if t.Status == store.StatusClosed {
		return nil, fmt.Errorf("%w: session %s", store.ErrClosed, p.ID)
	}
	c.mu.Lock()
	r := c.runners[p.ID]
	c.mu.Unlock()
	if r == nil {
		return nil, fmt.Errorf("%w: session %s has no controller-driven host", store.ErrInvalid, p.ID)
	}
	c.log.InfoContext(ctx, "session restart requested", "session", p.ID, "now", p.Now)
	r.requestRestart(p.Now)
	return rpc.SessionRestartResult{ID: p.ID, HostPid: hostPid(c.root.SessionDir(p.ID)), Now: p.Now}, nil
}

// sessionCancel cancels the running turn of a controller-driven session
// (#535). Nothing is sent; messages already queued are delivered next.
func (c *Controller) sessionCancel(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	if err := requireAdmin(sc, "session cancel"); err != nil {
		return nil, err
	}
	if p.ID == "" {
		return nil, fmt.Errorf("%w: session id is required", store.ErrInvalid)
	}
	c.mu.Lock()
	r := c.runners[p.ID]
	c.mu.Unlock()
	if r == nil {
		return nil, fmt.Errorf("%w: session %s has no controller-driven host", store.ErrInvalid, p.ID)
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
	open, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusOpen})
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
	closed, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusClosed})
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
		if !p.DryRun && hostAlive(dir) {
			c.log.WarnContext(ctx, "reap: host still alive, keeping dir", "session", t.ID)
			continue
		}
		if !p.DryRun {
			c.removeWorktree(t, "session reap")
			c.log.WarnContext(ctx, "reap: removing session dir", "session", t.ID, "dir", dir)
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
			return fmt.Errorf("%w: meta: unknown pack %q (add it to meta_packs in %s)", store.ErrInvalid, name, config.FileName)
		}
	}
	return nil
}
