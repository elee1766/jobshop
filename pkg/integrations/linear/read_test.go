package linear_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/integrations/apicall"
	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
)

// recorded serves testdata files chosen by pick and records each call.
func recorded(t *testing.T, pick func(c call) string) (*linear.Client, *[]call) {
	t.Helper()
	var calls []call
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var c call
		require.NoError(t, json.Unmarshal(b, &c))
		calls = append(calls, c)
		name := pick(c)
		if name == "" {
			_, _ = io.WriteString(w, `{"data":{"team":null}}`)
			return
		}
		data, err := os.ReadFile(filepath.Join("testdata", name))
		require.NoError(t, err)
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)
	return linear.New(srv.URL, apicall.Static("lin_api_x")), &calls
}

func TestContext(t *testing.T) {
	c, calls := recorded(t, func(call) string { return "context.json" })
	is, err := c.GetContext(t.Context(), "ENG-5", false)
	require.NoError(t, err)
	assert.Equal(t, float64(linear.RecentComments), (*calls)[0].Variables["comments"])
	assert.Equal(t, "Ann", is.Assignee)
	assert.Equal(t, []string{"backend", "Feature"}, is.Labels)
	require.NotNil(t, is.Parent)
	assert.Equal(t, "ENG-1", is.Parent.Identifier)
	require.Len(t, is.Children, 1)
	assert.Equal(t, "ENG-6", is.Children[0].Identifier)
	require.Len(t, is.Blockers, 2)
	open := is.OpenBlockers()
	require.Len(t, open, 1, "a done blocker does not block")
	assert.Equal(t, "ENG-3", open[0].Identifier)
	types := map[string]string{}
	for _, r := range is.Relations {
		types[r.Issue.Identifier] = r.Type
	}
	assert.Equal(t, map[string]string{"ENG-9": "related", "ENG-3": "blocked_by", "ENG-2": "blocked_by", "ENG-8": "duplicate"}, types)
	require.Len(t, is.Comments, 2)
	assert.Equal(t, "older", is.Comments[0].Body, "oldest first")
	assert.True(t, is.Comments[0].Bot)
	assert.True(t, is.Comments[1].Self)

	_, err = c.GetContext(t.Context(), "ENG-5", true)
	require.NoError(t, err)
	assert.Equal(t, float64(250), (*calls)[1].Variables["comments"], "--all-comments")
}

func TestListPaging(t *testing.T) {
	c, calls := recorded(t, func(cl call) string {
		if cl.Variables["after"] == "cur-2" {
			return "list_page2.json"
		}
		return "list_page1.json"
	})
	f := linear.Filter{Team: "ENG", State: "Todo", Label: "bug", Assignee: "me", Limit: 2}
	p, err := c.List(t.Context(), f)
	require.NoError(t, err)
	require.Len(t, p.Issues, 2)
	assert.True(t, p.HasNext)
	assert.Equal(t, "cur-2", p.EndCursor)
	assert.Equal(t, "Ann", p.Issues[0].Assignee)
	assert.Equal(t, []string{"bug"}, p.Issues[0].Labels)
	assert.Equal(t, []string{}, p.Issues[1].Labels)
	v := (*calls)[0].Variables
	assert.Equal(t, float64(2), v["first"])
	assert.NotContains(t, v, "after")
	filter, _ := json.Marshal(v["filter"])
	assert.JSONEq(t, `{"team":{"key":{"eq":"ENG"}},"state":{"name":{"eqIgnoreCase":"Todo"}},"labels":{"name":{"eqIgnoreCase":"bug"}},"assignee":{"isMe":{"eq":true}}}`, string(filter))

	f.After = p.EndCursor
	p, err = c.List(t.Context(), f)
	require.NoError(t, err)
	require.Len(t, p.Issues, 1)
	assert.False(t, p.HasNext)
	assert.Empty(t, p.EndCursor, "no cursor on the last page")
	assert.Equal(t, "cur-2", (*calls)[1].Variables["after"])

	_, err = c.List(t.Context(), linear.Filter{Team: "ENG", Limit: 1000, Assignee: "Bo"})
	require.NoError(t, err)
	v = (*calls)[2].Variables
	assert.Equal(t, float64(250), v["first"], "page size capped")
	filter, _ = json.Marshal(v["filter"])
	assert.Contains(t, string(filter), `"assignee":{"name":{"eqIgnoreCase":"Bo"}}`)
}

func TestSearch(t *testing.T) {
	c, calls := recorded(t, func(call) string { return "search.json" })
	p, err := c.Search(t.Context(), "export", linear.Filter{Team: "ENG", After: "x"})
	require.NoError(t, err)
	require.Len(t, p.Issues, 1)
	assert.Equal(t, "ENG-5", p.Issues[0].Identifier)
	assert.False(t, p.HasNext)
	assert.Equal(t, "export", (*calls)[0].Variables["term"])
	assert.Equal(t, "x", (*calls)[0].Variables["after"])
	assert.True(t, strings.Contains((*calls)[0].Query, "searchIssues"))
	_, err = c.Search(t.Context(), " ", linear.Filter{})
	require.Error(t, err)
}

// TestCatalog: one connection per query, following the cursor (#653).
// A query naming more than one connection gets Linear's "Query too
// complex" answer, as the combined query did on a real workspace.
func TestCatalog(t *testing.T) {
	c, calls := recorded(t, func(cl call) string {
		if cl.Variables["team"] == "NOPE" {
			return ""
		}
		var conns []string
		for _, f := range []string{"projects(", "labels(", "states(", "members("} {
			if strings.Contains(cl.Query, f) {
				conns = append(conns, strings.TrimSuffix(f, "("))
			}
		}
		if len(conns) != 1 {
			return "catalog.json"
		}
		name := map[string]string{"members": "users"}[conns[0]]
		if name == "" {
			name = conns[0]
		}
		if name == "labels" {
			if cl.Variables["after"] == "c-l1" {
				return "catalog_labels_2.json"
			}
			return "catalog_labels_1.json"
		}
		return "catalog_" + name + ".json"
	})
	want := map[string][]linear.Entry{
		"projects": {{ID: "p-1", Name: "Reports", Type: "started"}},
		"labels":   {{ID: "l-1", Name: "bug"}, {ID: "l-2", Name: "backend"}},
		"states":   {{ID: "st-todo", Name: "Todo", Type: "unstarted"}, {ID: "st-done", Name: "Done", Type: "completed"}},
		"users":    {{ID: "u-1", Name: "Ann", Type: "ann@example.com"}},
	}
	for kind, w := range want {
		es, err := c.Catalog(t.Context(), "ENG", kind)
		require.NoError(t, err, kind)
		assert.Equal(t, w, es, kind)
	}
	assert.Len(t, *calls, 5, "one query per kind, two pages of labels")
	for _, cl := range *calls {
		assert.InDelta(t, 100, cl.Variables["first"], 0)
	}
	_, err := c.Catalog(t.Context(), "ENG", "cycles")
	require.ErrorContains(t, err, "not one of")
	_, err = c.Catalog(t.Context(), "NOPE", "labels")
	require.ErrorContains(t, err, "team NOPE not found")
}

func TestEligible(t *testing.T) {
	c, _ := recorded(t, func(call) string { return "context.json" })
	is, err := c.GetContext(t.Context(), "ENG-5", false)
	require.NoError(t, err)

	e := linear.Eligible(is, nil, "")
	assert.False(t, e.Eligible)
	assert.Equal(t, "blocked by ENG-3", e.Reason)

	free := is
	free.Blockers = nil
	assert.True(t, linear.Eligible(free, nil, "").Eligible)

	cases := map[string]struct {
		mod  func(*linear.Context)
		skip []string
		run  string
		want string
	}{
		"done":     {mod: func(c *linear.Context) { c.State = linear.State{Name: "Done", Type: "completed"} }, want: "state Done is done"},
		"canceled": {mod: func(c *linear.Context) { c.State = linear.State{Name: "Won't do", Type: "canceled"} }, want: "state Won't do is canceled"},
		"archived": {mod: func(c *linear.Context) { at := time.Now(); c.ArchivedAt = &at }, want: "archived"},
		"open run": {run: "js-run1", want: "open run js-run1"},
		"label":    {skip: []string{"wontfix", " feature "}, want: "label Feature is skipped"},
	}
	for name, tc := range cases {
		x := free
		if tc.mod != nil {
			tc.mod(&x)
		}
		e := linear.Eligible(x, tc.skip, tc.run)
		assert.False(t, e.Eligible, name)
		assert.Equal(t, tc.want, e.Reason, name)
		assert.Equal(t, "ENG-5", e.Issue)
	}
	assert.True(t, linear.Eligible(free, []string{"", "wontfix"}, "").Eligible, "other labels pass")

	// JOB-68: an open stated GitHub dependency blocks; CheckDeps keeps
	// only the open ones, so a closed one does not.
	dep := free
	dep.Description = "Depends on acphost#155 ([link](https://github.com/gfx-labs/acphost/issues/155)): x\nDepends on gfx-labs/jobshop#9"
	open := func(_ context.Context, d linear.GitHubDep) (bool, error) { return d.Number == 155, nil }
	require.NoError(t, dep.CheckDeps(t.Context(), "gfx-labs", open))
	e = linear.Eligible(dep, nil, "")
	assert.False(t, e.Eligible)
	assert.Equal(t, "depends on open gfx-labs/acphost#155", e.Reason)
	require.NoError(t, dep.CheckDeps(t.Context(), "gfx-labs", func(context.Context, linear.GitHubDep) (bool, error) { return false, nil }))
	assert.True(t, linear.Eligible(dep, nil, "").Eligible, "closed dependencies do not block")
}

func TestDependencies(t *testing.T) {
	desc := "Depends on acphost#155 (x)\n- depends on: o/r#2\nDepends on https://github.com/a/b/issues/3\nThis depends on z#4 inline\nDepends on acphost#155"
	var got []string
	for _, d := range linear.Dependencies(desc, "me") {
		got = append(got, d.String())
	}
	assert.Equal(t, []string{"me/acphost#155", "o/r#2", "a/b#3"}, got, "line starts only, deduped")
	assert.Empty(t, linear.Dependencies("Depends on acphost#155", ""), "bare repo needs an owner")
}
