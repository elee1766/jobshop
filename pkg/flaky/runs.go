package flaky

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// Run is one failed CI run on master and the failed-tests.json it uploaded.
type Run struct {
	ID        int64     `json:"id"`
	SHA       string    `json:"sha"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	// Tests is nil when the run has no failed-tests artifact (a build
	// error, or the run failed in another job).
	Tests []Test `json:"-"`
}

// Group is one test across the runs it failed in, for the planner's
// incident step (JOB-13).
type Group struct {
	Test    string `json:"test"`
	Package string `json:"package"`
	Flaky   int    `json:"flaky"`
	Broken  int    `json:"broken"`
	// Verdict is the newest run's verdict.
	Verdict string    `json:"verdict"`
	Runs    []int64   `json:"runs"`
	Last    time.Time `json:"last"`
}

// Report is the output of `jobshop report ci-failures`.
type Report struct {
	Repo string `json:"repo"`
	// Runs lists the failed runs read; NoArtifact the ones without a
	// failed-tests artifact.
	Runs       []Run   `json:"runs"`
	NoArtifact []int64 `json:"no_artifact,omitempty"`
	Tests      []Group `json:"tests"`
}

// GroupRuns groups the failed tests of runs by test, most failures first.
func GroupRuns(runs []Run) []Group {
	type key struct{ pkg, test string }
	byKey := map[key]*Group{}
	// Newest first, so the first verdict seen is the latest.
	sorted := append([]Run(nil), runs...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].CreatedAt.After(sorted[j].CreatedAt) })
	for _, r := range sorted {
		for _, t := range r.Tests {
			k := key{t.Package, t.Test}
			g := byKey[k]
			if g == nil {
				g = &Group{Test: t.Test, Package: t.Package, Verdict: t.Verdict, Last: r.CreatedAt}
				byKey[k] = g
			}
			switch t.Verdict {
			case Flaky:
				g.Flaky++
			case Broken:
				g.Broken++
			}
			g.Runs = append(g.Runs, r.ID)
		}
	}
	out := make([]Group, 0, len(byKey))
	for _, g := range byKey {
		out = append(out, *g)
	}
	sort.Slice(out, func(i, j int) bool {
		if len(out[i].Runs) != len(out[j].Runs) {
			return len(out[i].Runs) > len(out[j].Runs)
		}
		if out[i].Package != out[j].Package {
			return out[i].Package < out[j].Package
		}
		return out[i].Test < out[j].Test
	})
	return out
}

// GH reads failed master runs and their failed-tests artifacts with gh.
type GH struct {
	Repo     string
	Branch   string
	Workflow string
	// Gh runs gh with args and returns stdout. Default: exec gh.
	Gh func(ctx context.Context, args ...string) ([]byte, error)
}

func (g GH) gh(ctx context.Context, args ...string) ([]byte, error) {
	if g.Gh != nil {
		return g.Gh(ctx, args...)
	}
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

// Report lists the last limit failed runs and groups their failed tests.
func (g GH) Report(ctx context.Context, limit int) (Report, error) {
	args := []string{"run", "list", "-R", g.Repo, "--branch", g.Branch, "--status", "failure", "--limit", strconv.Itoa(limit), "--json", "databaseId,headSha,url,createdAt"}
	if g.Workflow != "" {
		args = append(args, "--workflow", g.Workflow)
	}
	out, err := g.gh(ctx, args...)
	if err != nil {
		return Report{}, fmt.Errorf("gh run list: %w", err)
	}
	var list []struct {
		DatabaseID int64     `json:"databaseId"`
		HeadSha    string    `json:"headSha"`
		URL        string    `json:"url"`
		CreatedAt  time.Time `json:"createdAt"`
	}
	if err := json.Unmarshal(out, &list); err != nil {
		return Report{}, fmt.Errorf("gh run list: %w", err)
	}
	rep := Report{Repo: g.Repo, Runs: []Run{}}
	for _, l := range list {
		r := Run{ID: l.DatabaseID, SHA: l.HeadSha, URL: l.URL, CreatedAt: l.CreatedAt}
		r.Tests, err = g.download(ctx, r.ID)
		if err != nil {
			return Report{}, err
		}
		if r.Tests == nil {
			rep.NoArtifact = append(rep.NoArtifact, r.ID)
		}
		rep.Runs = append(rep.Runs, r)
	}
	rep.Tests = GroupRuns(rep.Runs)
	return rep, nil
}

// download fetches failed-tests.json of a run. A run without the artifact
// returns nil, nil: gh fails with "no valid artifacts found".
func (g GH) download(ctx context.Context, id int64) ([]Test, error) {
	dir, err := os.MkdirTemp("", "failed-tests-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	if _, err := g.gh(ctx, "run", "download", strconv.FormatInt(id, 10), "-R", g.Repo, "-n", "failed-tests", "-D", dir); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, nil
	}
	b, err := os.ReadFile(filepath.Join(dir, "failed-tests.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	tests := []Test{}
	if err := json.Unmarshal(b, &tests); err != nil {
		return nil, fmt.Errorf("run %d failed-tests.json: %w", id, err)
	}
	return tests, nil
}
