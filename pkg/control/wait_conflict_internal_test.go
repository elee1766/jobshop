package control

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Recorded GET /pulls/{n} responses (trimmed) by mergeable state (#609).
const (
	recPRConflict = `{"number":5,"state":"open","merged":false,"mergeable":false,"mergeable_state":"dirty","head":{"sha":"aaaa1111bbbb"},"base":{"ref":"master"}}`
	recPRUnknown  = `{"number":5,"state":"open","merged":false,"mergeable":null,"mergeable_state":"unknown","head":{"sha":"aaaa1111bbbb"},"base":{"ref":"master"}}`
	recPRClean    = `{"number":5,"state":"open","merged":false,"mergeable":true,"mergeable_state":"blocked","head":{"sha":"aaaa1111bbbb"},"base":{"ref":"master"}}`
)

// ciWait parks a step held by a session on success-or-failure CI for PR 5.
func ciWait(t *testing.T, c *Controller) (store.Ticket, store.Ticket) {
	t.Helper()
	ctx := t.Context()
	sess := mkSession(t, c)
	run := mkTicket(t, c, store.NewTicket{Kind: store.KindRun, Title: "r", Fields: json.RawMessage(`{"workflow":"w","state":"running"}`)})
	step := mkTicket(t, c, store.NewTicket{
		Kind: store.KindStep, Title: "implement", Pool: "dev", ParentID: run.ID,
		Fields: json.RawMessage(`{"run":"` + run.ID + `","step":"implement","index":0,"retries":3,"failures":0}`),
	})
	_, _, err := c.store.Claim(ctx, step.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	_, err = c.waitAdd(ctx, sessScope(sess), rpc.WaitAddParams{
		Conds: []store.WaitCond{
			{Type: store.WaitChecks, Repo: "acme/widget", PR: 5, State: "success"},
			{Type: store.WaitChecks, Repo: "acme/widget", PR: 5, State: "failure"},
		},
		Any: true, Reason: "CI on PR #5", Resume: true,
	})
	require.NoError(t, err)
	return sess, step
}

// TestWaitChecksConflict: a conflicting PR clears a CI wait in one pass
// with the reason and resumes the session; an uncomputed mergeable and
// a clean PR with pending checks keep waiting.
func TestWaitChecksConflict(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	sess, step := ciWait(t, c)

	prs.set(recPRUnknown, `[[]]`, `[]`)
	c.pollWaits(ctx, prs, threads)
	assert.NotNil(t, waitOf(mustGet(t, c, step.ID)), "UNKNOWN is read again next pass")

	prs.set(recPRClean, `[[]]`, `[{"name":"build","bucket":"pending"}]`)
	c.pollWaits(ctx, prs, threads)
	assert.NotNil(t, waitOf(mustGet(t, c, step.ID)), "mergeable with pending checks waits")

	prs.set(recPRConflict, `[[]]`, `[]`)
	c.pollWaits(ctx, prs, threads)
	got := mustGet(t, c, step.ID)
	assert.Nil(t, waitOf(got))
	assert.Equal(t, sess.ID, got.Holder, "resume pre-assigns the parking session")
	cs, err := c.store.Comments(ctx, step.ID)
	require.NoError(t, err)
	assert.Contains(t, cs[len(cs)-1].Body, "invalidated: PR #5 conflicts with master; CI will not run")
}

// TestWaitChecksNoCI: a sha checks wait with no check run at all is
// listed in attention after noCIAfter, before its timeout.
func TestWaitChecksNoCI(t *testing.T) {
	t.Parallel()
	c, clk, prs, threads := waitController(t)
	ctx := t.Context()
	a := mkTicket(t, c, store.NewTicket{Title: "post-merge", Pool: "dev"})
	timeout := clk.now().Add(time.Hour)
	_, err := c.waitAdd(ctx, scope{admin: true, actor: "admin"}, rpc.WaitAddParams{
		ID: a.ID, Reason: "master CI", Timeout: &timeout,
		Conds: []store.WaitCond{{Type: store.WaitChecks, Repo: "acme/widget", SHA: "dddd4444eeee", State: "success"}},
	})
	require.NoError(t, err)
	prs.checks[prKey("acme/widget", 0)] = `[]`

	c.pollWaits(ctx, prs, threads)
	items, err := c.attentionItems(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)

	clk.advance(noCIAfter)
	c.pollWaits(ctx, prs, threads)
	items, err = c.attentionItems(ctx)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, rpc.AttentionWait, items[0].Kind)
	assert.True(t, strings.HasSuffix(items[0].Title, "CI did not start for dddd444"), items[0].Title)

	prs.checks[prKey("acme/widget", 0)] = `[{"name":"build","bucket":"pending"}]`
	clk.advance(time.Minute)
	c.pollWaits(ctx, prs, threads)
	items, err = c.attentionItems(ctx)
	require.NoError(t, err)
	assert.Empty(t, items, "a started run clears the item")
}
