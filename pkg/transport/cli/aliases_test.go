package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control/controltest"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Names agents use for things that exist under another name (JOB-201).
// Each form parses to what the existing form parses to.

func parseErr(args ...string) (*root, error) {
	var r root
	p, err := kong.New(&r, kong.Exit(func(int) {}))
	if err != nil {
		return nil, err
	}
	_, err = p.Parse(args)
	return &r, err
}

func TestAliasCloseMessage(t *testing.T) {
	want := parse(t, "close", "js-1", "--comment", "why").Close
	for _, flag := range []string{"-m", "--message"} {
		assert.Equal(t, want, parse(t, "close", "js-1", flag, "why").Close, flag)
	}
	assert.Equal(t, "why", parse(t, "ticket", "close", "-m", "why").Ticket.Close.Comment)
}

func TestAliasCommentTicketPositional(t *testing.T) {
	id, body, err := parse(t, "comment", "-t", "js-1", "hello").Comment.target()
	require.NoError(t, err)
	assert.Equal(t, [2]string{"js-1", "hello"}, [2]string{id, body})

	id, body, err = parse(t, "comment", "js-1", "hello").Comment.target()
	require.NoError(t, err)
	assert.Equal(t, [2]string{"js-1", "hello"}, [2]string{id, body}, "ticket then body")

	id, body, err = parse(t, "comment", "hello").Comment.target()
	require.NoError(t, err)
	assert.Equal(t, [2]string{"", "hello"}, [2]string{id, body})

	id, body, err = parse(t, "comment", "-t", "js-1", "-").Comment.target()
	require.NoError(t, err)
	assert.Equal(t, [2]string{"js-1", "-"}, [2]string{id, body}, "stdin marker")

	_, err = parseErr("comment", "-t", "a", "b", "c")
	require.Error(t, err, "-t with a positional ticket")
	_, err = parseErr("comment", "a", "b", "c")
	require.Error(t, err, "three positionals")
	_, err = parseErr("comment")
	require.Error(t, err, "no body")
}

func TestAliasLinearShow(t *testing.T) {
	r := parse(t, "linear", "show", "JOB-7")
	assert.Equal(t, "JOB-7", r.Linear.Issue.Key)
}

func TestAliasJSONNoop(t *testing.T) {
	assert.Equal(t, parse(t, "ticket", "show", "js-1").Ticket.Show.ID, parse(t, "ticket", "show", "js-1", "--json").Ticket.Show.ID)
	assert.True(t, parse(t, "ticket", "show", "js-1", "--json").Ticket.Show.JSON)
	assert.True(t, parse(t, "run", "list", "--json").Run.List.JSON)
	assert.Equal(t, "running", parse(t, "run", "list", "--json", "-s", "running").Run.List.State)
}

func TestAliasHandoffReason(t *testing.T) {
	want := parse(t, "handoff", "--to", "plan", "--note", "why").Handoff
	for _, flag := range []string{"-r", "--reason"} {
		assert.Equal(t, want, parse(t, "handoff", "--to", "plan", flag, "why").Handoff, flag)
	}
}

func TestAliasRunCancelReason(t *testing.T) {
	r := parse(t, "run", "cancel", "js-r", "-r", "superseded")
	assert.Equal(t, "superseded", r.Run.Cancel.Reason)
	assert.Equal(t, "js-r", r.Run.Cancel.ID)
	assert.Empty(t, parse(t, "run", "cancel", "js-r").Run.Cancel.Reason)
}

func TestAliasWaitRm(t *testing.T) {
	r := parse(t, "wait", "rm", "js-1", "-r", "done")
	assert.Equal(t, parse(t, "wait", "clear", "js-1", "-r", "done").Wait.Clear, r.Wait.Clear)
	assert.Equal(t, "js-1", r.Wait.Clear.ID)
}

// --version prints the line `version` does and stops before any command.
func TestAliasVersionFlag(t *testing.T) {
	var want, got bytes.Buffer
	old := stdout
	t.Cleanup(func() { stdout = old })

	stdout = &want
	require.Equal(t, 0, Main([]string{"version"}, nil))
	stdout = &got
	require.Equal(t, 0, Main([]string{"--version"}, nil))
	assert.NotEmpty(t, want.String())
	assert.Equal(t, want.String(), got.String())
}

// searchLinear is a fake Linear that records the team filter of each search.
type searchLinear struct {
	mu    sync.Mutex
	teams []string
}

func (f *searchLinear) serve(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req linearReq
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &req)
		f.mu.Lock()
		defer f.mu.Unlock()
		filter, _ := req.Variables["filter"].(map[string]any)
		team, _ := filter["team"].(map[string]any)
		key, _ := team["key"].(map[string]any)
		eq, _ := key["eq"].(string)
		f.teams = append(f.teams, eq)
		_, _ = io.WriteString(w, `{"data":{"searchIssues":{"nodes":[{"identifier":"`+eq+`-1","title":"found"}],"pageInfo":{"hasNextPage":false}},"issues":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}`)
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

// TestLinearSearchDefaultTeam: without --team the host's one Linear team is
// used; two teams, or none, are an error naming them.
func TestLinearSearchDefaultTeam(t *testing.T) {
	run := func(t *testing.T, teams []string, args ...string) (string, error) {
		t.Helper()
		f := &fakeController{decision: domain.GuardAllow, keys: map[string]string{}, teams: teams}
		g := &Globals{Socket: f.serve(t)}
		lin := &searchLinear{}
		l := &linearCmd{APIURL: lin.serve(t)}
		r := parse(t, append([]string{"linear"}, args...)...)
		var out bytes.Buffer
		old := stdout
		stdout = &out
		defer func() { stdout = old }()
		var err error
		switch args[0] {
		case "search":
			err = r.Linear.Search.Run(g, l)
		default:
			err = r.Linear.List.Run(g, l)
		}
		if err == nil {
			assert.Len(t, lin.teams, 1)
		}
		return out.String(), err
	}

	for _, args := range [][]string{{"search", "found"}, {"list"}} {
		t.Run(args[0]+" one team", func(t *testing.T) {
			out, err := run(t, []string{"ENG"}, args...)
			require.NoError(t, err)
			if args[0] == "search" {
				assert.Contains(t, out, `"ENG-1"`)
			}
		})
		t.Run(args[0]+" explicit team wins", func(t *testing.T) {
			out, err := run(t, []string{"ENG", "OPS"}, append(append([]string(nil), args...), "--team", "OPS")...)
			require.NoError(t, err)
			if args[0] == "search" {
				assert.Contains(t, out, `"OPS-1"`)
			}
		})
		t.Run(args[0]+" two teams", func(t *testing.T) {
			_, err := run(t, []string{"ENG", "OPS"}, args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "ENG, OPS")
		})
		t.Run(args[0]+" no teams", func(t *testing.T) {
			_, err := run(t, nil, args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "no Linear team")
		})
	}
}

// TestRunCancelReason: run cancel -r cancels the run and leaves the reason
// as a comment on it.
func TestRunCancelReason(t *testing.T) {
	root := controltest.NewRoot(t, `projects: {js: {}}
agents: {worker: {command: echo}}
pools: {dev: {agent: worker, max_sessions: 0, dir: $ROOT}}
workflows:
  example:
    steps:
      - {id: inspect, pool: dev, prompt: Inspect, title: Inspect}
`)
	controller := controltest.StartWith(t, root)
	var run rpc.RunResult
	controller.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "example"}, &run)
	id := run.Run.ID

	old := stdout
	t.Cleanup(func() { stdout = old })
	var out bytes.Buffer
	stdout = &out
	require.Equal(t, 0, Main([]string{"--socket", root.AdminSocket(), "run", "cancel", id, "-r", "superseded by js-9", "--json"}, nil))

	var shown rpc.ShowResult
	controller.Call(rpc.MethodTicketShow, rpc.IDParams{ID: id}, &shown)
	assert.Equal(t, domain.StatusClosed, shown.Ticket.Status)
	assert.Equal(t, domain.RunCancelled, shown.Ticket.Resolution)
	require.Len(t, shown.Comments, 1)
	assert.True(t, strings.Contains(shown.Comments[0].Body, "superseded by js-9"), shown.Comments[0].Body)
}
