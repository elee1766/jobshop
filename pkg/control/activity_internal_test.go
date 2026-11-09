package control

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/activity"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

const activityYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  dev: {agent: fake, dir: %q}
  plan: {agent: fake, dir: %q, admin: [activity]}
workflows:
  pr-work:
    vars: [issue]
    issue: issue
    repo: elee1766/jobshop
    steps:
      - {id: implement, pool: dev, title: "Implement #{{.issue}}", prompt: "do it"}
activity:
  window: 24h
  max_items: 3
  repos: [elee1766/acphost]
`

type fakeActivity struct {
	mu    sync.Mutex
	gh    map[string][]rpc.ActivityItem
	calls map[string]int
	err   error
}

func (f *fakeActivity) GitHub(_ context.Context, repo string, st *activity.RepoPoll) ([]rpc.ActivityItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls[repo]++
	if f.err != nil {
		return nil, f.err
	}
	st.EventsETag = "etag-" + repo
	return f.gh[repo], nil
}

func (f *fakeActivity) Linear(context.Context, string, time.Time) ([]rpc.ActivityItem, error) {
	return nil, nil
}

func activityController(t *testing.T) (*Controller, *scanClock, *fakeActivity) {
	t.Helper()
	c, clk, _ := subController(t)
	require.NoError(t, os.WriteFile(filepath.Join(c.root.Dir, config.FileName), fmt.Appendf(nil, activityYAML, c.root.Dir, c.root.Dir), 0o600))
	cfg, err := config.Load(c.root)
	require.NoError(t, err)
	lc, err := newLiveConfig(cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	c.activitySvc().Load()
	return c, clk, &fakeActivity{gh: map[string][]rpc.ActivityItem{}, calls: map[string]int{}}
}

func ghItem(repo, id, kind, actor, ref string, at time.Time) rpc.ActivityItem {
	return rpc.ActivityItem{ID: "gh:" + repo + ":" + id, Source: rpc.ActivityGitHub, Repo: repo, Kind: kind, Actor: actor, Bot: strings.HasSuffix(actor, "[bot]"), At: at, Title: kind, Ref: ref}
}

func listActivity(t *testing.T, c *Controller, p rpc.ActivityListParams) rpc.ActivityListResult {
	t.Helper()
	v, err := c.activityList(t.Context(), scope{admin: true}, p)
	require.NoError(t, err)
	return v.(rpc.ActivityListResult)
}

// TestActivityPollListAndLinks: the controller feeds polled items through
// the inbox and admission into the feed, and listing links them to runs
// and tickets from meta at list time.
func TestActivityPollListAndLinks(t *testing.T) {
	t.Parallel()
	c, clk, src := activityController(t)
	now := clk.now()
	run := mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: "run", Fields: json.RawMessage(`{"workflow":"pr-work","state":"running"}`), Meta: json.RawMessage(`{"github":{"repo":"elee1766/jobshop","issue":255}}`)})
	mkTicket(t, c, store.NewTicket{Kind: domain.KindStep, Title: "step", Fields: json.RawMessage(`{"step":"implement","run":"` + run.ID + `"}`), ParentID: run.ID, Meta: json.RawMessage(`{"github":{"repo":"elee1766/jobshop","pr":"260"}}`)})
	src.gh["elee1766/jobshop"] = []rpc.ActivityItem{
		ghItem("elee1766/jobshop", "1", "comment", "elee1766", "255", now.Add(-time.Hour)),
		ghItem("elee1766/jobshop", "2", "review", "elee-agent[bot]", "260", now.Add(-2*time.Hour)),
		ghItem("elee1766/jobshop", "old", "push", "elee1766", "", now.Add(-48*time.Hour)),
	}
	src.gh["elee1766/acphost"] = []rpc.ActivityItem{ghItem("elee1766/acphost", "3", "ci", "elee1766", "", now.Add(-30*time.Minute))}

	assert.Equal(t, 3, c.pollActivity(t, src))
	assert.Equal(t, 0, c.pollActivity(t, src), "items are deduped by id")

	res := listActivity(t, c, rpc.ActivityListParams{})
	require.Len(t, res.Items, 3, "the 48h-old push is outside the 24h window")
	assert.Equal(t, []string{"github:elee1766/acphost", "github:elee1766/jobshop"}, res.Sources)
	assert.Equal(t, "ci", res.Items[0].Kind, "newest first")
	assert.Equal(t, []string{run.ID}, res.Items[1].Tickets, "issue links to its run")
	assert.Equal(t, []string{run.ID}, res.Items[2].Tickets, "a step's PR links to the step's run")
	assert.Len(t, listActivity(t, c, rpc.ActivityListParams{Ticket: run.ID}).Items, 2)
}

func TestActivityGrantAndOff(t *testing.T) {
	t.Parallel()
	c, _, _ := activityController(t)
	_, err := c.activityList(t.Context(), scope{pool: "dev"}, rpc.ActivityListParams{})
	require.ErrorContains(t, err, "activity")
	_, err = c.activityList(t.Context(), scope{pool: "plan"}, rpc.ActivityListParams{})
	require.NoError(t, err)
	_, err = c.activityList(t.Context(), scope{admin: true}, rpc.ActivityListParams{Bots: "robots"})
	require.ErrorIs(t, err, domain.ErrInvalid)
}

// pollActivity runs one poll pass on src.
func (c *Controller) pollActivity(t *testing.T, src activity.Source) int {
	t.Helper()
	return c.activitySvc().Poll(t.Context(), src)
}

// forgetActivity drops the feed cache file and the loaded service, as a
// lost cache directory would.
func (c *Controller) forgetActivity(t *testing.T) {
	t.Helper()
	require.NoError(t, os.Remove(filepath.Join(c.root.Dir, activity.File)))
	c.activity = activity.New(activityHost{c}, c.log, c.root.Dir)
}
