package control

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/redact"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// A GitHub token shape the default gitleaks rules detect, split so the
// literal is not a token in this file.
const fakeGHToken = "ghp_" + "Zx9Qw3Er5Ty7Ui9Op1As3Df5Gh7Jk9Lz1Xc3V"

// TestRedactOnWrite (#574): comments, reports and learnings holding a
// token are stored masked.
func TestRedactOnWrite(t *testing.T) {
	t.Parallel()
	c, _, _ := scanController(t)
	ctx := t.Context()
	step, _, sc := startWork(t, c, "r1")

	_, err := c.comment(ctx, sc, rpc.CommentParams{Body: "token " + fakeGHToken})
	require.NoError(t, err)
	cms, err := store.Comments(ctx, c.store, step.ID)
	require.NoError(t, err)
	last := cms[len(cms)-1].Body
	assert.NotContains(t, last, fakeGHToken)
	assert.Contains(t, last, redact.Mask)

	_, err = c.report(ctx, sc, rpc.ReportParams{State: rpc.ReportProgress, Note: "used " + fakeGHToken})
	require.NoError(t, err)
	got, err := c.store.Get(ctx, step.ID)
	require.NoError(t, err)
	assert.NotContains(t, string(got.Meta), fakeGHToken)

	out, err := c.learnAdd(ctx, sc, rpc.LearnParams{Kind: domain.LearningWorkaround, Summary: "set " + fakeGHToken, Fix: "unset " + fakeGHToken})
	require.NoError(t, err)
	l := out.(domain.Learning)
	assert.NotContains(t, l.Summary+l.Fix, fakeGHToken)
	ls := learnings(t, c, store.LearningFilter{})
	for _, x := range ls {
		assert.NotContains(t, x.Summary+x.Fix, fakeGHToken)
	}
}

// TestRedactArtifact: an artifact's body and text attachment are masked;
// a binary attachment is copied as is.
func TestRedactArtifact(t *testing.T) {
	t.Parallel()
	c, _, _ := scanController(t)
	sc, _ := artifactSession(t, c)
	txt := filepath.Join(c.root.SessionDir(sc.session), "work", "log.txt")
	require.NoError(t, os.WriteFile(txt, []byte("line\nkey "+fakeGHToken+"\n"), 0o600))
	bin := filepath.Join(c.root.SessionDir(sc.session), "work", "blob.bin")
	require.NoError(t, os.WriteFile(bin, []byte("\x00"+fakeGHToken), 0o600))

	out, err := c.artifactDeliver(t.Context(), sc, rpc.ArtifactDeliverParams{Title: "r", Body: "b " + fakeGHToken, Files: []string{txt, bin}})
	require.NoError(t, err)
	a := out.(rpc.Artifact)
	assert.NotContains(t, a.Ticket.Body, fakeGHToken)
	b, err := os.ReadFile(filepath.Join(c.artifactDir(a.Ticket.ID), "log.txt"))
	require.NoError(t, err)
	assert.NotContains(t, string(b), fakeGHToken)
	assert.Equal(t, int64(len(b)), a.Meta.Files[0].Size, "size is of the masked file")
	b, err = os.ReadFile(filepath.Join(c.artifactDir(a.Ticket.ID), "blob.bin"))
	require.NoError(t, err)
	assert.Contains(t, string(b), fakeGHToken, "binary attachments are not rewritten")
}

// TestRedactActivity: a masked write is recorded as redacted, without the
// secret.
func TestRedactActivity(t *testing.T) {
	t.Parallel()
	c, _, _ := activityController(t)
	c.redactWrite(scope{session: "js-s1"}, "comment on", "js-1", new("plain"))
	assert.Empty(t, listActivity(t, c, rpc.ActivityListParams{Kind: "redact"}).Items, "clean text records nothing")
	s := "x " + fakeGHToken
	c.redactWrite(scope{session: "js-s1"}, "comment on", "js-1", &s)
	res := listActivity(t, c, rpc.ActivityListParams{Kind: "redact"})
	require.Len(t, res.Items, 1)
	assert.True(t, res.Items[0].Redacted)
	b, _ := json.Marshal(res.Items[0])
	assert.NotContains(t, string(b), fakeGHToken)
}

// TestRedactEvent: a session event with a token in its text or tool title
// is sent masked.
func TestRedactEvent(t *testing.T) {
	t.Parallel()
	red := redact.New([]string{"MY_API_TOKEN=env-secret-value-1234"})
	ev := rpc.SessionEvent{
		Text: "out " + fakeGHToken, Context: "env-secret-value-1234", Error: fakeGHToken,
		Tool: &rpc.ToolSummary{ID: "t1", Title: "curl -H " + fakeGHToken, Status: "completed"},
	}
	redactEvent(red, &ev)
	b, _ := json.Marshal(ev)
	assert.NotContains(t, string(b), fakeGHToken)
	assert.NotContains(t, string(b), "env-secret-value-1234")
	assert.Contains(t, ev.Text, redact.Mask)
	assert.Equal(t, "t1", ev.Tool.ID)
}
