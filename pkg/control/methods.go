package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/host"
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

func (c *Controller) whoami(_ context.Context, sc scope, _ struct{}) (any, error) {
	return rpc.Whoami{Actor: sc.actor, Admin: sc.admin, Session: sc.session, Project: sc.project, Pool: sc.pool}, nil
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
	if p.Holder != "" && !sc.admin {
		return nil, forbidden("assigning needs the admin socket")
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
	return c.store.Create(ctx, sc.actor, store.NewTicket{
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
	if r.Edges, err = c.store.Edges(ctx, id); err != nil {
		return nil, err
	}
	if r.Comments, err = c.store.Comments(ctx, id); err != nil {
		return nil, err
	}
	if r.History, err = c.store.History(ctx, id); err != nil {
		return nil, err
	}
	return r, nil
}

func (c *Controller) list(ctx context.Context, sc scope, p rpc.ListParams) (any, error) {
	if !sc.admin {
		p.Project = sc.project
	}
	if p.Orphaned {
		return c.orphaned(ctx, p.Project, p.Limit)
	}
	return c.store.List(ctx, store.Filter{
		Project: p.Project, Kind: p.Kind, Status: p.Status, Pool: p.Pool,
		Holder: p.Holder, ParentID: p.ParentID, Label: p.Label, Limit: p.Limit,
	})
}

func (c *Controller) ready(ctx context.Context, sc scope, p rpc.ReadyParams) (any, error) {
	proj, err := c.project(sc, p.Project)
	if err != nil {
		return nil, err
	}
	return c.store.Ready(ctx, proj, p.Pool, p.Limit)
}

func (c *Controller) update(ctx context.Context, sc scope, p rpc.UpdateParams) (any, error) {
	if p.Holder != nil && !sc.admin {
		return nil, forbidden("assigning needs the admin socket")
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
	var cl *store.Claim
	if !sc.admin {
		h, err := c.heldClaim(ctx, sc, id)
		if err != nil {
			return nil, err
		}
		cl = &h
	}
	t, err := c.store.CloseTicket(ctx, sc.actor, id, cl, p.Resolution)
	if err != nil {
		return nil, err
	}
	c.countClose(ctx, t)
	if t.Kind == store.KindStep && t.ParentID != "" {
		c.checkRun(ctx, t.ParentID)
	}
	if t.Kind == store.KindSession {
		c.stopRunner(t.ID)
		c.endSession(ctx, t.ID, "session closed")
	}
	return t, nil
}

// endSession stops a session's socket, frees everything it held and removes
// its worktree.
func (c *Controller) endSession(ctx context.Context, id, reason string) {
	c.closeSession(id)
	if t, err := c.store.Get(ctx, id); err == nil {
		c.removeWorktree(t)
	}
	released, err := c.store.ReleaseHolder(ctx, "controller", id, reason)
	if err != nil {
		c.log.Error("release session claims", "session", id, "err", err)
		return
	}
	c.stepEvents(ctx, released)
	if len(released) > 0 {
		c.log.Info("released session claims", "session", id, "tickets", released)
	}
}

func (c *Controller) reopen(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	if err := requireAdmin(sc, "reopen"); err != nil {
		return nil, err
	}
	return c.store.Reopen(ctx, sc.actor, p.ID)
}

func (c *Controller) comment(ctx context.Context, sc scope, p rpc.CommentParams) (any, error) {
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	return c.store.AddComment(ctx, sc.actor, id, p.Body)
}

func (c *Controller) edgeAdd(ctx context.Context, sc scope, p rpc.EdgeParams) (any, error) {
	return struct{}{}, c.store.AddEdge(ctx, sc.actor, store.Edge{From: p.From, To: p.To, Kind: p.Kind})
}

func (c *Controller) edgeRemove(ctx context.Context, sc scope, p rpc.EdgeParams) (any, error) {
	return struct{}{}, c.store.RemoveEdge(ctx, sc.actor, store.Edge{From: p.From, To: p.To, Kind: p.Kind})
}

func (c *Controller) labelAdd(ctx context.Context, sc scope, p rpc.LabelParams) (any, error) {
	return struct{}{}, c.store.AddLabel(ctx, sc.actor, p.ID, p.Label)
}

func (c *Controller) labelRemove(ctx context.Context, sc scope, p rpc.LabelParams) (any, error) {
	return struct{}{}, c.store.RemoveLabel(ctx, sc.actor, p.ID, p.Label)
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
		r.Ticket, r.Claim, err = c.store.Claim(ctx, p.ID, holder, c.leaseDur())
	} else {
		proj, perr := c.project(sc, p.Project)
		if perr != nil {
			return nil, perr
		}
		r.Ticket, r.Claim, err = c.store.ClaimNext(ctx, store.ClaimNext{
			Project: proj, Pool: pool, Holder: holder, Lease: c.leaseDur(),
		})
	}
	if err != nil {
		return nil, err
	}
	c.metrics.Claims.Add(ctx, 1, metric.WithAttributes(telemetry.AttrPool.String(r.Ticket.Pool)))
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
	if err != nil {
		return nil, err
	}
	var until time.Time
	if p.Until != nil {
		if !p.Until.After(time.Now()) {
			return nil, fmt.Errorf("%w: until must be in the future", store.ErrInvalid)
		}
		until = *p.Until
	}
	t, err := c.store.Release(ctx, cl, p.Reason, until)
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
	}
	t, err := c.newSession(ctx, sc.actor, proj, p.Pool, p.Agent, p.Key, p.Title)
	if err != nil {
		return nil, err
	}
	return rpc.SessionNewResult{Ticket: t, Socket: c.socketPath(t.ID)}, nil
}

// sessionList joins session tickets with the live state of their session
// directories, so store and processes can be compared (plan 8.4 item 5).
func (c *Controller) sessionList(ctx context.Context, sc scope, p rpc.SessionListParams) (any, error) {
	if err := requireAdmin(sc, "session list"); err != nil {
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
		cutoff = time.Now().Add(-p.Since)
	}
	var tickets []store.Ticket
	for _, s := range statuses {
		ts, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: s})
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ts...)
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
		info, err := c.sessionInfo(ctx, t.ID)
		if err != nil {
			return nil, err
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

	entries, err := os.ReadDir(c.root.SessionDir(""))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() || seen[e.Name()] || !host.Alive(c.root.SessionDir(e.Name())) {
			continue
		}
		info, err := c.sessionInfo(ctx, e.Name())
		if err != nil {
			return nil, err
		}
		if t, err := c.store.Get(ctx, e.Name()); err == nil {
			info.Ticket = &t
			if t.Kind == store.KindSession && t.Status != store.StatusClosed {
				continue // listed above unless the store changed under us
			}
		}
		info.Orphan = true
		out = append(out, info)
	}
	return out, nil
}

// sessionInfo reads the live state of one session directory.
func (c *Controller) sessionInfo(ctx context.Context, id string) (rpc.SessionInfo, error) {
	dir := c.root.SessionDir(id)
	info := rpc.SessionInfo{ID: id, Alive: host.Alive(dir), HostPid: hostPid(dir)}
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
	}
	c.mu.Unlock()
	info.Drained = r != nil && c.poolDrained(r.pool)
	if !info.Alive {
		info.AgentPid = 0
		if ex, ok := host.ReadExit(dir); ok {
			info.Exit = ex.Reason
		}
	}
	if seq, at := host.JournalTail(dir); seq > 0 {
		info.Seq, info.SeqTime = seq, &at
	}
	held, err := c.store.List(ctx, store.Filter{Holder: id, Status: store.StatusInProgress})
	if err != nil {
		return info, err
	}
	if len(held) > 0 {
		info.Current = held[0].ID
	}
	return info, nil
}

// sessionRestart stops a session's host so its driver respawns it with a
// freshly built spec and resumes the conversation (plan 8.4). The
// session's claims are kept and it is not counted as a crash.
func (c *Controller) sessionRestart(ctx context.Context, sc scope, p rpc.SessionRestartParams) (any, error) {
	if err := requireAdmin(sc, "session restart"); err != nil {
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
	c.log.Info("session restart requested", "session", p.ID, "now", p.Now)
	r.requestRestart(p.Now)
	return rpc.SessionRestartResult{ID: p.ID, HostPid: hostPid(c.root.SessionDir(p.ID)), Now: p.Now}, nil
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
			c.log.Warn("stopping orphaned host", "session", id)
			stopHost(c.root.SessionDir(id))
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
		if !p.DryRun && host.Alive(dir) {
			c.log.Warn("reap: host still alive, keeping dir", "session", t.ID)
			continue
		}
		if !p.DryRun {
			c.removeWorktree(t)
			if err := os.RemoveAll(dir); err != nil {
				return res, fmt.Errorf("remove %s: %w", dir, err)
			}
		}
		res.Removed = append(res.Removed, t.ID)
	}
	return res, nil
}

func (c *Controller) now() time.Time {
	if c.opt.Now != nil {
		return c.opt.Now()
	}
	return time.Now()
}

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
