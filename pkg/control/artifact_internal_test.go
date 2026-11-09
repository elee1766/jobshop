package control

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// artifactSession makes a session holding a work step and returns its scope.
func artifactSession(t *testing.T, c *Controller) (scope, domain.Ticket) {
	t.Helper()
	ctx := t.Context()
	res, err := c.startRun(ctx, c.config(), scope{admin: true, actor: "admin"}, rpc.RunStartParams{Workflow: "work", Vars: map[string]string{"n": "1"}, Key: "w1"}, nil)
	require.NoError(t, err)
	step := res.(rpc.RunResult).Steps[0]
	sess := mkTicket(t, c, store.NewTicket{Kind: domain.KindSession, Pool: "dev", Fields: json.RawMessage(`{"agent":"fake"}`)})
	_, _, err = c.store.Claim(ctx, step.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(c.root.SessionDir(sess.ID), "work"), 0o700))
	return scope{actor: sess.ID, session: sess.ID, project: "js", pool: "dev"}, step
}

func writeSessionFile(t *testing.T, c *Controller, sc scope, name string, size int) string {
	t.Helper()
	p := filepath.Join(c.root.SessionDir(sc.session), "work", name)
	require.NoError(t, os.WriteFile(p, []byte(strings.Repeat("x", size)), 0o600))
	return p
}

func TestArtifactDeliverAck(t *testing.T) {
	t.Parallel()
	c, clk, _ := scanController(t)
	ctx := t.Context()
	sc, step := artifactSession(t, c)
	admin := scope{admin: true, actor: "admin"}

	// Without attachments.
	out, err := c.artifactDeliver(ctx, sc, rpc.ArtifactDeliverParams{Title: "Weekly report", Body: "# hi"})
	require.NoError(t, err)
	a := out.(rpc.Artifact)
	assert.Equal(t, domain.KindMessage, a.Ticket.Kind)
	assert.Contains(t, a.Ticket.Labels, artifactLabel)
	assert.Equal(t, step.ID, a.Meta.Ticket)
	assert.Equal(t, step.ParentID, a.Meta.Run)
	assert.Equal(t, "dev", a.Meta.Pool)
	assert.Equal(t, sc.session, a.Meta.Session)
	assert.Empty(t, a.Meta.Files)
	assert.NoDirExists(t, c.artifactDir(a.Ticket.ID))
	edges, err := store.Edges(ctx, c.store, a.Ticket.ID)
	require.NoError(t, err)
	assert.Contains(t, edges, domain.Edge{From: a.Ticket.ID, To: step.ID, Kind: domain.EdgeTracks, CreatedAt: edges[0].CreatedAt})
	cms, err := store.Comments(ctx, c.store, step.ID)
	require.NoError(t, err)
	assert.Equal(t, "artifact delivered: "+a.Ticket.ID+" Weekly report", cms[len(cms)-1].Body)
	assert.Empty(t, runnerInbox(t, c, sc.session), "the artifact is not delivered to any session")

	// With attachments.
	p1 := writeSessionFile(t, c, sc, "out.csv", 10)
	p2 := writeSessionFile(t, c, sc, "chart.png", 20)
	out, err = c.artifactDeliver(ctx, sc, rpc.ArtifactDeliverParams{Title: "With files", Body: "b", Files: []string{p1, p2}})
	require.NoError(t, err)
	b := out.(rpc.Artifact)
	require.Len(t, b.Meta.Files, 2)
	assert.Equal(t, "out.csv", b.Meta.Files[0].Name)
	assert.Equal(t, int64(10), b.Meta.Files[0].Size)
	assert.Equal(t, "image/png", b.Meta.Files[1].Type)
	assert.FileExists(t, filepath.Join(c.artifactDir(b.Ticket.ID), "chart.png"))

	// List: unread, newest first; read ones last.
	ls, err := c.artifactList(ctx, admin, rpc.ArtifactListParams{})
	require.NoError(t, err)
	require.Len(t, ls.([]rpc.Artifact), 2)

	// A session cannot ack.
	_, err = c.artifactAck(ctx, sc, rpc.ArtifactAckParams{ID: a.Ticket.ID})
	require.ErrorIs(t, err, rpc.ErrForbidden)

	// Ack closes it with the note and history.
	out, err = c.artifactAck(ctx, admin, rpc.ArtifactAckParams{ID: b.Ticket.ID, Note: "looks good"})
	require.NoError(t, err)
	acked := out.(rpc.Artifact)
	assert.True(t, acked.Read)
	assert.Equal(t, resolutionAcknowledged, acked.Ticket.Resolution)
	show, err := c.artifactShow(ctx, admin, rpc.IDParams{ID: b.Ticket.ID})
	require.NoError(t, err)
	sr := show.(rpc.ArtifactShowResult)
	assert.Equal(t, "looks good", sr.Comments[len(sr.Comments)-1].Body)
	last := sr.History[len(sr.History)-1]
	assert.Equal(t, "close", last.Op)
	assert.Equal(t, "admin", last.Actor)
	_, err = c.artifactAck(ctx, admin, rpc.ArtifactAckParams{ID: b.Ticket.ID})
	require.ErrorIs(t, err, domain.ErrClosed)

	ls, err = c.artifactList(ctx, admin, rpc.ArtifactListParams{})
	require.NoError(t, err)
	got := ls.([]rpc.Artifact)
	assert.Equal(t, a.Ticket.ID, got[0].Ticket.ID, "unread first")
	ls, err = c.artifactList(ctx, admin, rpc.ArtifactListParams{Unread: true})
	require.NoError(t, err)
	assert.Len(t, ls.([]rpc.Artifact), 1)

	// Retention: not yet; after 90d the files go and the ticket stays.
	n, err := c.pruneArtifactsBefore(ctx, c.now().Add(-c.config().Artifacts.RetentionDuration()))
	require.NoError(t, err)
	assert.Zero(t, n)
	clk.advance(91 * 24 * time.Hour)
	c.pruneArtifacts(t.Context(), c.config())
	assert.NoDirExists(t, c.artifactDir(b.Ticket.ID))
	pt, err := c.store.Get(ctx, b.Ticket.ID)
	require.NoError(t, err)
	am, ok := artifactOf(pt)
	require.True(t, ok)
	for _, f := range am.Files {
		assert.True(t, f.Pruned, f.Name)
	}
	n, err = c.pruneArtifactsBefore(ctx, c.now())
	require.NoError(t, err)
	assert.Zero(t, n, "pruned once")
}

func TestArtifactDeliverRefused(t *testing.T) {
	t.Parallel()
	c, _, dir := scanControllerYAML(t, scanYAML+"artifacts: {max_bytes: 100}\n")
	ctx := t.Context()
	sc, _ := artifactSession(t, c)

	// Outside the session dir, directly and through a symlink.
	outside := filepath.Join(dir, "secret.txt")
	require.NoError(t, os.WriteFile(outside, []byte("s"), 0o600))
	_, err := c.artifactDeliver(ctx, sc, rpc.ArtifactDeliverParams{Title: "x", Files: []string{outside}})
	require.ErrorIs(t, err, rpc.ErrForbidden)
	link := filepath.Join(c.root.SessionDir(sc.session), "work", "link.txt")
	require.NoError(t, os.Symlink(outside, link))
	_, err = c.artifactDeliver(ctx, sc, rpc.ArtifactDeliverParams{Title: "x", Files: []string{link}})
	require.ErrorIs(t, err, rpc.ErrForbidden)

	// Over the cap: refused with nothing written.
	p1 := writeSessionFile(t, c, sc, "a.txt", 60)
	p2 := writeSessionFile(t, c, sc, "b.txt", 60)
	_, err = c.artifactDeliver(ctx, sc, rpc.ArtifactDeliverParams{Title: "big", Files: []string{p1, p2}})
	require.ErrorIs(t, err, domain.ErrInvalid)
	assert.Contains(t, err.Error(), "max_bytes")
	assert.NoDirExists(t, filepath.Join(dir, "artifacts"))
	ls, err := c.artifactList(ctx, scope{admin: true}, rpc.ArtifactListParams{})
	require.NoError(t, err)
	assert.Empty(t, ls.([]rpc.Artifact))
}

func TestArtifactTrigger(t *testing.T) {
	t.Parallel()
	c, _, _ := scanController(t)
	ctx := t.Context()
	res, err := c.startRun(ctx, c.config(), scope{admin: true, actor: "controller"}, rpc.RunStartParams{Workflow: "work", Vars: map[string]string{"n": "2"}}, map[string]any{triggerPack: map[string]string{"name": "nightly"}})
	require.NoError(t, err)
	step := res.(rpc.RunResult).Steps[0]
	out, err := c.artifactDeliver(ctx, scope{admin: true, actor: "admin"}, rpc.ArtifactDeliverParams{Title: "Nightly", Ticket: step.ID})
	require.NoError(t, err)
	a := out.(rpc.Artifact)
	assert.Equal(t, "nightly", a.Meta.Trigger)
	ls, err := c.artifactList(ctx, scope{admin: true}, rpc.ArtifactListParams{Trigger: "nightly"})
	require.NoError(t, err)
	assert.Len(t, ls.([]rpc.Artifact), 1)
	ls, err = c.artifactList(ctx, scope{admin: true}, rpc.ArtifactListParams{Trigger: "other"})
	require.NoError(t, err)
	assert.Empty(t, ls.([]rpc.Artifact))
}
