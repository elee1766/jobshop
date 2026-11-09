package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/execution/sandbox"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// Artifacts (#508): deliverables a human reads and acks in Desk. An
// artifact is a message ticket to "human" with label artifact, like a
// question, so it is never claimed and is not an orphan. Open means
// unread; ack closes it. Attachments are copied by the controller into
// <root>/artifacts/<id>/, a dir sessions cannot write from the sandbox.

const (
	artifactLabel    = "artifact"
	artifactMetaPack = "artifact"
	// resolutionAcknowledged closes an artifact a human read.
	resolutionAcknowledged = "acknowledged"
)

// artifactDir is where an artifact's attachments live.
func (c *Controller) artifactDir(id string) string {
	return filepath.Join(c.root.Dir, "artifacts", id)
}

func (c *Controller) artifactDeliver(ctx context.Context, sc scope, p rpc.ArtifactDeliverParams) (any, error) {
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: artifact title is required", domain.ErrInvalid)
	}
	am := rpc.ArtifactMeta{Session: sc.session, Pool: sc.pool}
	project := sc.project
	var src *domain.Ticket
	if p.Ticket != "" || !sc.admin {
		id, err := c.current(ctx, sc, p.Ticket)
		if err != nil {
			return nil, err
		}
		t, err := c.store.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		if !sc.admin && t.Project != sc.project {
			return nil, forbidden("session is scoped to project %s", sc.project)
		}
		src, project, am.Ticket = &t, t.Project, t.ID
		if t.Pool != "" {
			am.Pool = t.Pool
		}
		am.Run, am.Trigger = c.artifactRun(ctx, t)
	}
	if project == "" {
		project = c.config().DefaultProject
	}
	if project == "" {
		return nil, fmt.Errorf("%w: no project; pass a source ticket with -t", domain.ErrInvalid)
	}
	files, err := c.artifactFiles(ctx, sc, p.Files)
	if err != nil {
		return nil, err
	}
	id, err := c.store.NewID(project)
	if err != nil {
		return nil, err
	}
	if len(files) > 0 {
		if am.Files, err = c.copyArtifactFiles(id, files); err != nil {
			return nil, err
		}
	}
	masked := false
	for i, f := range am.Files {
		p := filepath.Join(c.artifactDir(id), f.Name)
		m, err := redactFile(c.redactor(nil), p)
		if err != nil {
			_ = os.RemoveAll(c.artifactDir(id))
			return nil, err
		}
		if m {
			if st, err := os.Stat(p); err == nil {
				am.Files[i].Size = st.Size()
			}
		}
		masked = masked || m
	}
	if masked {
		c.recordRedacted(sc, "artifact", id)
	}
	c.redactWrite(sc, "artifact", id, &title, &p.Body)
	meta, err := json.Marshal(map[string]rpc.ArtifactMeta{artifactMetaPack: am})
	if err != nil {
		return nil, err
	}
	fields, err := json.Marshal(domain.MessageFields{To: questionTo, From: sc.actor})
	if err != nil {
		return nil, err
	}
	t, err := store.Create(ctx, c.store, sc.actor, store.NewTicket{
		ID: id, Project: project, Kind: domain.KindMessage, Title: title, Body: p.Body,
		Labels: []string{artifactLabel}, Meta: meta, Fields: fields,
	})
	if err != nil {
		_ = os.RemoveAll(c.artifactDir(id))
		return nil, err
	}
	if src != nil {
		if err := store.AddEdge(ctx, c.store, sc.actor, domain.Edge{From: t.ID, To: src.ID, Kind: domain.EdgeTracks}); err != nil {
			return nil, err
		}
		if _, err := store.AddComment(ctx, c.store, sc.actor, src.ID, "artifact delivered: "+t.ID+" "+title); err != nil {
			return nil, err
		}
	}
	return rpc.Artifact{Ticket: t, Meta: am}, nil
}

// artifactRun finds the run of a source ticket (itself, or a step's run)
// and the trigger that started it.
func (c *Controller) artifactRun(ctx context.Context, t domain.Ticket) (run, trigger string) {
	switch t.Kind {
	case domain.KindRun:
		run = t.ID
	case domain.KindStep:
		if f, err := domain.DecodeFields[domain.StepFields](t); err == nil && f.Run != "" {
			run = f.Run
		} else {
			run = t.ParentID
		}
	}
	if run == "" {
		return "", metaString(t.Meta, triggerPack, "name")
	}
	r, err := c.store.Get(ctx, run)
	if err != nil {
		return run, ""
	}
	return run, metaString(r.Meta, triggerPack, "name")
}

// artifactFile is one attachment to copy.
type artifactFile struct {
	path string
	name string
	size int64
}

// artifactFiles checks the paths and the size cap before anything is
// written. A session may only pass files inside its session dir or
// worktree, after symlinks are resolved.
func (c *Controller) artifactFiles(ctx context.Context, sc scope, paths []string) ([]artifactFile, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	var allowed []string
	if !sc.admin {
		st, err := c.store.Get(ctx, sc.session)
		if err != nil {
			return nil, err
		}
		// The session dir holds the worktree (work/). A pool dir without a
		// repo is shared by sessions, so it is not the session's own.
		if r, err := filepath.EvalSymlinks(c.root.SessionDir(st.ID)); err == nil {
			allowed = append(allowed, r)
		}
	}
	var out []artifactFile
	var total int64
	for _, p := range paths {
		if !filepath.IsAbs(p) {
			return nil, fmt.Errorf("%w: attachment %q must be an absolute path", domain.ErrInvalid, p)
		}
		resolved, err := filepath.EvalSymlinks(p)
		if err != nil {
			return nil, fmt.Errorf("%w: attachment %s: %w", domain.ErrInvalid, p, err)
		}
		if !sc.admin && !slices.ContainsFunc(allowed, func(d string) bool { return sandbox.Within(resolved, d) }) {
			return nil, forbidden("attachment %s is outside the session dir", p)
		}
		fi, err := os.Stat(resolved)
		if err != nil {
			return nil, fmt.Errorf("%w: attachment %s: %w", domain.ErrInvalid, p, err)
		}
		if !fi.Mode().IsRegular() {
			return nil, fmt.Errorf("%w: attachment %s is not a regular file", domain.ErrInvalid, p)
		}
		name := filepath.Base(p)
		if !validArtifactName(name) {
			return nil, fmt.Errorf("%w: attachment name %q is not allowed", domain.ErrInvalid, name)
		}
		if slices.ContainsFunc(out, func(f artifactFile) bool { return f.name == name }) {
			return nil, fmt.Errorf("%w: two attachments are named %s", domain.ErrInvalid, name)
		}
		total += fi.Size()
		out = append(out, artifactFile{path: resolved, name: name, size: fi.Size()})
	}
	if limit := c.config().Artifacts.MaxBytesOrDefault(); total > limit {
		return nil, fmt.Errorf("%w: attachments total %d bytes, over artifacts.max_bytes %d; nothing was delivered", domain.ErrInvalid, total, limit)
	}
	return out, nil
}

// validArtifactName is a plain file name Desk can serve by name.
func validArtifactName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.HasPrefix(name, ".") &&
		!strings.ContainsAny(name, "/\\\x00") && len(name) <= 255
}

// copyArtifactFiles copies into a temp dir and renames it into place, so
// a failed copy leaves nothing behind.
func (c *Controller) copyArtifactFiles(id string, files []artifactFile) ([]rpc.ArtifactFile, error) {
	base := filepath.Join(c.root.Dir, "artifacts")
	if err := os.MkdirAll(base, 0o750); err != nil {
		return nil, err
	}
	tmp, err := os.MkdirTemp(base, ".tmp-"+id+"-")
	if err != nil {
		return nil, err
	}
	out := make([]rpc.ArtifactFile, 0, len(files))
	for _, f := range files {
		n, err := copyFile(f.path, filepath.Join(tmp, f.name), f.size)
		if err != nil {
			_ = os.RemoveAll(tmp)
			return nil, err
		}
		typ := mime.TypeByExtension(filepath.Ext(f.name))
		out = append(out, rpc.ArtifactFile{Name: f.name, Size: n, Type: typ})
	}
	if err := os.Rename(tmp, c.artifactDir(id)); err != nil {
		_ = os.RemoveAll(tmp)
		return nil, err
	}
	return out, nil
}

// copyFile copies at most limit bytes: a file that grew after the size
// check fails instead of passing the cap.
func copyFile(src, dst string, limit int64) (int64, error) {
	in, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return 0, err
	}
	n, err := io.Copy(out, io.LimitReader(in, limit+1))
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err == nil && n > limit {
		err = fmt.Errorf("%w: attachment %s changed size while copying", domain.ErrInvalid, filepath.Base(src))
	}
	return n, err
}

// artifactOf parses meta.artifact; ok is false for other tickets.
func artifactOf(t domain.Ticket) (rpc.ArtifactMeta, bool) {
	var am rpc.ArtifactMeta
	if !slices.Contains(t.Labels, artifactLabel) || t.Kind != domain.KindMessage {
		return am, false
	}
	var m map[string]json.RawMessage
	if json.Unmarshal(t.Meta, &m) == nil {
		_ = json.Unmarshal(m[artifactMetaPack], &am)
	}
	return am, true
}

func toArtifact(t domain.Ticket, am rpc.ArtifactMeta) rpc.Artifact {
	return rpc.Artifact{Ticket: t, Meta: am, Read: t.Status == domain.StatusClosed}
}

// artifactList lists artifacts unread first, then newest first.
func (c *Controller) artifactList(ctx context.Context, sc scope, p rpc.ArtifactListParams) (any, error) {
	f := store.Filter{Kind: domain.KindMessage, Label: artifactLabel, Order: store.OrderNewest}
	if !sc.admin {
		f.Project = sc.project
	}
	if p.Unread {
		f.Status = domain.StatusOpen
	}
	ts, err := c.store.List(ctx, f)
	if err != nil {
		return nil, err
	}
	out := []rpc.Artifact{}
	for _, t := range ts {
		am, ok := artifactOf(t)
		if !ok || (p.Trigger != "" && am.Trigger != p.Trigger) || (p.Pool != "" && am.Pool != p.Pool) {
			continue
		}
		out = append(out, toArtifact(t, am))
	}
	slices.SortStableFunc(out, func(a, b rpc.Artifact) int {
		if a.Read != b.Read {
			if a.Read {
				return 1
			}
			return -1
		}
		return b.Ticket.CreatedAt.Compare(a.Ticket.CreatedAt)
	})
	if p.Limit > 0 && len(out) > p.Limit {
		out = out[:p.Limit]
	}
	return out, nil
}

func (c *Controller) getArtifact(ctx context.Context, sc scope, id string) (domain.Ticket, rpc.ArtifactMeta, error) {
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return t, rpc.ArtifactMeta{}, err
	}
	if !sc.admin && t.Project != sc.project {
		return t, rpc.ArtifactMeta{}, domain.ErrNotFound
	}
	am, ok := artifactOf(t)
	if !ok {
		return t, am, fmt.Errorf("%w: %s is not an artifact", domain.ErrInvalid, id)
	}
	return t, am, nil
}

func (c *Controller) artifactShow(ctx context.Context, sc scope, p rpc.IDParams) (any, error) {
	t, am, err := c.getArtifact(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	r := rpc.ArtifactShowResult{Artifact: toArtifact(t, am)}
	th, err := c.store.Thread(ctx, t.ID, store.Parts{Comments: true, History: true})
	if err != nil {
		return nil, err
	}
	r.Comments, r.History = th.Comments, th.History
	if sc.admin {
		r.Dir = c.artifactDir(t.ID)
	}
	return r, nil
}

// artifactAck closes an artifact acknowledged; history records who and when.
func (c *Controller) artifactAck(ctx context.Context, sc scope, p rpc.ArtifactAckParams) (any, error) {
	if err := requireAdmin(sc, "artifact ack"); err != nil {
		return nil, err
	}
	t, am, err := c.getArtifact(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Status == domain.StatusClosed {
		return nil, fmt.Errorf("%w: artifact %s is already acknowledged", domain.ErrClosed, t.ID)
	}
	if note := strings.TrimSpace(p.Note); note != "" {
		if _, err := store.AddComment(ctx, c.store, sc.actor, t.ID, note); err != nil {
			return nil, err
		}
	}
	if t, err = store.Close(ctx, c.store, sc.actor, t.ID, nil, resolutionAcknowledged); err != nil {
		return nil, err
	}
	return toArtifact(t, am), nil
}

// artifactPruneEvery spaces attachment prunes; retention is days.
const artifactPruneEvery = time.Hour

// pruneArtifacts deletes attachments of artifacts acked longer than
// artifacts.retention ago and marks their files pruned. The ticket stays.
func (c *Controller) pruneArtifacts(ctx context.Context, cfg *config.Config) {
	now := c.now()
	if now.Sub(c.artifactsPrunedAt) < artifactPruneEvery {
		return
	}
	c.artifactsPrunedAt = now
	n, err := c.pruneArtifactsBefore(ctx, now.Add(-cfg.Artifacts.RetentionDuration()))
	if err != nil {
		if ctx.Err() == nil {
			c.log.ErrorContext(ctx, "artifact prune", "err", err)
		}
		return
	}
	if n > 0 {
		c.log.InfoContext(ctx, "artifacts pruned", "artifacts", n)
	}
}

func (c *Controller) pruneArtifactsBefore(ctx context.Context, cutoff time.Time) (int, error) {
	ts, err := c.store.List(ctx, store.Filter{Kind: domain.KindMessage, Label: artifactLabel, Status: domain.StatusClosed})
	if err != nil {
		return 0, err
	}
	n := 0
	var errs []error
	for _, t := range ts {
		am, ok := artifactOf(t)
		if !ok || t.ClosedAt.IsZero() || !t.ClosedAt.Before(cutoff) ||
			!slices.ContainsFunc(am.Files, func(f rpc.ArtifactFile) bool { return !f.Pruned }) {
			continue
		}
		if err := os.RemoveAll(c.artifactDir(t.ID)); err != nil {
			errs = append(errs, err)
			continue
		}
		for i := range am.Files {
			am.Files[i].Pruned = true
		}
		if err := c.setArtifactMeta(ctx, t.ID, am); err != nil {
			errs = append(errs, err)
			continue
		}
		n++
	}
	return n, errors.Join(errs...)
}

// setArtifactMeta replaces meta.artifact, keeping other packs.
func (c *Controller) setArtifactMeta(ctx context.Context, id string, am rpc.ArtifactMeta) error {
	for attempt := 0; ; attempt++ {
		t, err := c.store.Get(ctx, id)
		if err != nil {
			return err
		}
		m := map[string]json.RawMessage{}
		if len(t.Meta) > 0 {
			_ = json.Unmarshal(t.Meta, &m)
		}
		m[artifactMetaPack], _ = json.Marshal(am)
		mb, err := json.Marshal(m)
		if err != nil {
			return err
		}
		_, err = c.store.Update(ctx, "controller", id, store.Update{Meta: mb, IfRevision: t.Revision})
		if errors.Is(err, domain.ErrConflict) && attempt < 5 {
			continue
		}
		return err
	}
}
