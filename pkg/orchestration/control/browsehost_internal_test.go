package control

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
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
	assert.Equal(t, []rpc.BrowseGitHubScope{{Repo: "elee1766/acphost"}, {Repo: "gfx-labs/jobshop"}}, res.GitHub, "workflow repos and activity.repos")
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
	run := mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: "implement 255", Fields: json.RawMessage(`{"workflow":"pr-work","state":"running"}`), Meta: json.RawMessage(`{"github":{"repo":"gfx-labs/jobshop","issue":255},"linear":{"issue":"JOB-163"}}`)})
	mkTicket(t, c, store.NewTicket{Kind: domain.KindStep, Title: "step", Fields: json.RawMessage(`{"step":"implement","run":"` + run.ID + `"}`), ParentID: run.ID, Meta: json.RawMessage(`{"github":{"repo":"gfx-labs/jobshop","pr":"260"}}`)})

	gh := c.browseRuns(ctx, rpc.BrowseItem{Provider: rpc.BrowseGitHub, Ref: "gfx-labs/jobshop#255", Kind: "issue"}, nil)
	require.Len(t, gh, 1)
	assert.Equal(t, rpc.BrowseLinked{Kind: "run", Ref: run.ID, Title: "implement 255", State: "running", Relation: "run"}, gh[0])
	pr := c.browseRuns(ctx, rpc.BrowseItem{Provider: rpc.BrowseGitHub, Ref: "gfx-labs/jobshop#260", Kind: "pr"}, nil)
	require.Len(t, pr, 1)
	assert.Equal(t, run.ID, pr[0].Ref, "a step's link is its run")
	lin := c.browseRuns(ctx, rpc.BrowseItem{Provider: rpc.BrowseLinear, Ref: "JOB-163"}, nil)
	require.Len(t, lin, 1)
	assert.Equal(t, run.ID, lin[0].Ref)

	assert.Empty(t, c.browseRuns(ctx, rpc.BrowseItem{Provider: rpc.BrowseGitHub, Ref: "gfx-labs/jobshop#9"}, nil), "no invented links")
	assert.Empty(t, c.browseRuns(ctx, rpc.BrowseItem{Provider: rpc.BrowseLinear, Ref: "JOB-9"}, nil))
	assert.Empty(t, c.browseRuns(ctx, rpc.BrowseItem{Provider: rpc.BrowseGitHub, Ref: "not-a-ref"}, nil))
}

// Runs started by triggers and callbacks carry no meta.github: they link by
// their pr var with the workflow's repo, or by the tracker URL in their
// title. An item also lists the runs of the work it is tied to.
func TestBrowseRunLinksWithoutMeta(t *testing.T) {
	c, _, _ := activityController(t)
	ctx := t.Context()
	ship := mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: "Run pr-work 75", Fields: json.RawMessage(`{"workflow":"pr-work","state":"running","vars":{"pr":"75","head":"abc"}}`)})
	review := mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: `Review [jobshop#75](https://github.com/gfx-labs/jobshop/pull/75): fix\(desk\): links (JOB-77)`, Fields: json.RawMessage(`{"workflow":"gone","state":"done","vars":{"subject":"75@abc"}}`)})
	implement := mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: "Implement [JOB-77](https://linear.app/gfx-labs/issue/JOB-77/desk-links)", Fields: json.RawMessage(`{"workflow":"gone","state":"done"}`)})
	mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: "Run pr-work 750", Fields: json.RawMessage(`{"workflow":"pr-work","state":"running","vars":{"pr":"750"}}`)})

	ids := func(links []rpc.BrowseLinked) (out []string) {
		for _, l := range links {
			out = append(out, l.Ref+" "+l.Relation)
		}
		return out
	}
	pr := rpc.BrowseItem{Provider: rpc.BrowseGitHub, Ref: "gfx-labs/jobshop#75", Kind: "pr"}
	assert.ElementsMatch(t, []string{ship.ID + " run", review.ID + " run"}, ids(c.browseRuns(ctx, pr, nil)), "pr var and title link, not #750")

	// The PR mentions JOB-77: that issue's run is listed once, after its own.
	withIssue := c.browseRuns(ctx, pr, []rpc.BrowseLinked{{Kind: "issue", Ref: "JOB-77", Relation: "mentions"}})
	require.Len(t, withIssue, 3)
	assert.Equal(t, rpc.BrowseLinked{Kind: "run", Ref: implement.ID, Title: implement.Title, State: "done", Relation: "via JOB-77"}, withIssue[2])

	// The Linear issue has the PR attached: the PR's runs follow its own.
	issue := c.browseRuns(ctx, rpc.BrowseItem{Provider: rpc.BrowseLinear, Ref: "JOB-77"}, []rpc.BrowseLinked{
		{Kind: "pr", Ref: "gfx-labs/jobshop#75", Relation: "attached"},
		{Kind: "issue", Ref: "JOB-1", Relation: "blocks"},
	})
	assert.Equal(t, implement.ID+" run", ids(issue)[0])
	assert.ElementsMatch(t, []string{ship.ID + " via gfx-labs/jobshop#75", review.ID + " via gfx-labs/jobshop#75"}, ids(issue)[1:])
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
		ghItem("gfx-labs/jobshop", "1", "pr", "a", "gfx-labs/jobshop#1", at),
		ghItem("gfx-labs/jobshop", "2", "pr", "a", "gfx-labs/jobshop#2", at),
		{Source: rpc.ActivityLinear, Repo: "JOB", Kind: "comment", At: at},
	})
	assert.Equal(t, before+2, svc.Revision(), "once per repo or team")
	assert.Equal(t, svc.Revision(), svc.Scope().Revision)
}
