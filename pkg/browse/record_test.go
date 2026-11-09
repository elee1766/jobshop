package browse

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/integrations/apicall"
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/integrations/linear"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// recorder saves each response body the real providers send under
// testdata/<name>.json, where name is the step the test is on. The replay
// tests read those files, so they run against what the providers really
// answer. It is skipped unless BROWSE_RECORD is set; testdata/README.md says
// how the files are made and trimmed.
type recorder struct {
	t    *testing.T
	dir  string
	name *string
	rt   http.RoundTripper
}

func (r recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := r.rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.NoError(r.t, err)
	require.NoError(r.t, os.WriteFile(filepath.Join(r.dir, *r.name+".json"), body, 0o644))
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return resp, nil
}

type liveHost struct {
	cfg *config.Config
	lin *linear.Client
	gh  *github.Client
}

func (h liveHost) Config() *config.Config                { return h.cfg }
func (h liveHost) Now() time.Time                        { return time.Now() }
func (h liveHost) Linear(string) (*linear.Client, error) { return h.lin, nil }
func (h liveHost) GitHub(string) (*github.Client, error) { return h.gh, nil }

func TestRecord(t *testing.T) {
	if os.Getenv("BROWSE_RECORD") == "" {
		t.Skip("set BROWSE_RECORD=1 with LINEAR_TOKEN and GH_TOKEN to record")
	}
	team, repo := os.Getenv("BROWSE_TEAM"), os.Getenv("BROWSE_REPO")
	issue, pr := os.Getenv("BROWSE_ISSUE"), os.Getenv("BROWSE_PR")
	dir := "testdata"
	require.NoError(t, os.MkdirAll(dir, 0o755))
	name := new(string)
	lin := linear.New("", apicall.Static(os.Getenv("LINEAR_TOKEN")))
	lin.HTTP.Transport = recorder{t, dir, name, http.DefaultTransport}
	gh := github.New("", apicall.Static(os.Getenv("GH_TOKEN")))
	gh.HTTP.Transport = recorder{t, dir, name, http.DefaultTransport}
	cfg := &config.Config{
		Identities: map[string]config.Identity{"lin": {Kind: config.IdentityLinear, Teams: []string{team}, Workspace: "w"}},
		Activity:   &config.Activity{Repos: []string{repo}},
	}
	s := New(liveHost{cfg, lin, gh}, slog.Default())
	ctx := context.Background()
	step := func(n string) { *name = n }

	step("linear_list")
	_, err := s.List(ctx, rpc.BrowseListParams{Provider: rpc.BrowseLinear, Scope: team, Limit: 8})
	require.NoError(t, err)
	step("linear_show")
	_, err = s.Show(ctx, rpc.BrowseShowParams{Provider: rpc.BrowseLinear, Ref: issue})
	require.NoError(t, err)
	step("github_list")
	_, err = s.List(ctx, rpc.BrowseListParams{Provider: rpc.BrowseGitHub, Scope: repo, View: "all", Limit: 8})
	require.NoError(t, err)
	step("github_show")
	_, err = s.Show(ctx, rpc.BrowseShowParams{Provider: rpc.BrowseGitHub, Ref: pr})
	require.NoError(t, err)
}
