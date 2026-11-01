package control

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
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
		asked = c.cfg.DefaultProject
	}
	if asked == "" {
		return "", fmt.Errorf("%w: no project given and no default_project configured", store.ErrInvalid)
	}
	if _, ok := c.cfg.Projects[asked]; !ok {
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
	if p.Holder != "" && !sc.admin {
		return nil, forbidden("assigning needs the admin socket")
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
		r.Ticket, r.Claim, err = c.store.Claim(ctx, p.ID, holder, c.lease)
	} else {
		proj, perr := c.project(sc, p.Project)
		if perr != nil {
			return nil, perr
		}
		r.Ticket, r.Claim, err = c.store.ClaimNext(ctx, store.ClaimNext{
			Project: proj, Pool: pool, Holder: holder, Lease: c.lease,
		})
	}
	if err != nil {
		return nil, err
	}
	return r, nil
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
	return c.store.Heartbeat(ctx, cl, c.lease)
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
	return c.store.Release(ctx, cl, p.Reason)
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
		if pool, ok := c.cfg.Pools[p.Pool]; ok && p.Agent == "" {
			p.Agent = pool.Agent
		}
	}
	t, err := c.newSession(ctx, sc.actor, proj, p.Pool, p.Agent, p.Key, p.Title)
	if err != nil {
		return nil, err
	}
	return rpc.SessionNewResult{Ticket: t, Socket: c.socketPath(t.ID)}, nil
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
