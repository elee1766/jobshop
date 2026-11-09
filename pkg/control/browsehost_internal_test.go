package control

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// browse.* is open to Desk and the admin socket, and to a pool only with the
// "browse" grant. A call with no scope configured never reaches a provider.
func TestBrowseGrantAndScope(t *testing.T) {
	c, _, _ := activityController(t)
	ctx := t.Context()

	_, err := c.browseScope(ctx, scope{pool: "dev"}, struct{}{})
	require.ErrorIs(t, err, rpc.ErrForbidden)
	_, err = c.browseList(ctx, scope{pool: "dev"}, rpc.BrowseListParams{Provider: rpc.BrowseGitHub, Scope: "elee1766/acphost"})
	require.ErrorIs(t, err, rpc.ErrForbidden)
	_, err = c.browseShow(ctx, scope{pool: "dev"}, rpc.BrowseShowParams{Provider: rpc.BrowseGitHub, Ref: "elee1766/acphost#1"})
	require.ErrorIs(t, err, rpc.ErrForbidden)
	_, err = c.browseCatalog(ctx, scope{pool: "dev"}, rpc.BrowseCatalogParams{Team: "JOB"})
	require.ErrorIs(t, err, rpc.ErrForbidden)

	v, err := c.browseScope(ctx, scope{admin: true}, struct{}{})
	require.NoError(t, err)
	res := v.(rpc.BrowseScopeResult)
	assert.Equal(t, []rpc.BrowseGitHubScope{{Repo: "elee1766/acphost"}, {Repo: "elee1766/jobshop"}}, res.GitHub, "workflow repos and activity.repos")
	assert.Empty(t, res.Linear)
	assert.False(t, res.Writes.Enabled)

	// A repo outside that is refused before any provider is asked.
	_, err = c.browseList(ctx, scope{admin: true}, rpc.BrowseListParams{Provider: rpc.BrowseGitHub, Scope: "other/repo"})
	require.ErrorIs(t, err, domain.ErrInvalid)
}

// A run links to an item only when its meta names it, for the ticket kinds
// that carry runs, and a step links to its run.
func TestBrowseRunLinks(t *testing.T) {
	c, _, _ := activityController(t)
	ctx := t.Context()
	run := mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: "implement 255", Fields: json.RawMessage(`{"workflow":"pr-work","state":"running"}`), Meta: json.RawMessage(`{"github":{"repo":"elee1766/jobshop","issue":255},"linear":{"issue":"JOB-163"}}`)})
	mkTicket(t, c, store.NewTicket{Kind: domain.KindStep, Title: "step", Fields: json.RawMessage(`{"step":"implement","run":"` + run.ID + `"}`), ParentID: run.ID, Meta: json.RawMessage(`{"github":{"repo":"elee1766/jobshop","pr":"260"}}`)})

	gh := c.browseRuns(ctx, rpc.BrowseItem{Provider: rpc.BrowseGitHub, Ref: "elee1766/jobshop#255", Kind: "issue"})
	require.Len(t, gh, 1)
	assert.Equal(t, rpc.BrowseLinked{Kind: "run", Ref: run.ID, Title: "implement 255", State: "running", Relation: "run"}, gh[0])
	pr := c.browseRuns(ctx, rpc.BrowseItem{Provider: rpc.BrowseGitHub, Ref: "elee1766/jobshop#260", Kind: "pr"})
	require.Len(t, pr, 1)
	assert.Equal(t, run.ID, pr[0].Ref, "a step's link is its run")
	lin := c.browseRuns(ctx, rpc.BrowseItem{Provider: rpc.BrowseLinear, Ref: "JOB-163"})
	require.Len(t, lin, 1)
	assert.Equal(t, run.ID, lin[0].Ref)

	assert.Empty(t, c.browseRuns(ctx, rpc.BrowseItem{Provider: rpc.BrowseGitHub, Ref: "elee1766/jobshop#9"}), "no invented links")
	assert.Empty(t, c.browseRuns(ctx, rpc.BrowseItem{Provider: rpc.BrowseLinear, Ref: "JOB-9"}))
	assert.Empty(t, c.browseRuns(ctx, rpc.BrowseItem{Provider: rpc.BrowseGitHub, Ref: "not-a-ref"}))
}

// New activity moves the scope revision once per scope, which is what tells
// an open Desk view to read again.
func TestBrowseRevisionFollowsActivity(t *testing.T) {
	c, _, _ := activityController(t)
	svc := c.browseService()
	before := svc.Revision()
	c.browseChanged(nil)
	assert.Equal(t, before, svc.Revision())

	at := time.Unix(1700000000, 0)
	c.browseChanged([]rpc.ActivityItem{
		ghItem("elee1766/jobshop", "1", "pr", "a", "elee1766/jobshop#1", at),
		ghItem("elee1766/jobshop", "2", "pr", "a", "elee1766/jobshop#2", at),
		{Source: rpc.ActivityLinear, Repo: "JOB", Kind: "comment", At: at},
	})
	assert.Equal(t, before+2, svc.Revision(), "once per repo or team")
	assert.Equal(t, svc.Revision(), svc.Scope().Revision)
}
