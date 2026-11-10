package workflow

import (
	"context"
	"fmt"

	"github.com/gfx-labs/jobshop/pkg/integrations/github"
	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
)

// Pass caches provider reads within one evaluation pass, so waits on one
// PR share a fetch. A Pass is used by one goroutine, under the
// controller's pass lock.
//
// The controller lists the waits first and starts the Pass after, so every
// cached observation is newer than the snapshot it is applied to. A
// Decision made from them is still only advice: the controller saves it
// at the snapshot's revision (see Decision).
type Pass struct {
	src     Sources
	prs     map[string]github.PRInfo
	reviews map[string][]github.Review
	checks  map[string][]github.Check
	runs    map[string][]github.WorkflowRun
	issues  map[string]linear.Context
}

// NewPass starts a pass over src.
func NewPass(src Sources) *Pass {
	return &Pass{
		src: src,
		prs: map[string]github.PRInfo{}, reviews: map[string][]github.Review{}, checks: map[string][]github.Check{},
		runs: map[string][]github.WorkflowRun{}, issues: map[string]linear.Context{},
	}
}

// PRKey is "owner/name#n".
func PRKey(repo string, n int) string { return fmt.Sprintf("%s#%d", repo, n) }

func (p *Pass) pr(ctx context.Context, repo string, n int) (github.PRInfo, error) {
	k := PRKey(repo, n)
	if v, ok := p.prs[k]; ok {
		return v, nil
	}
	v, err := p.src.GitHub.PR(ctx, repo, n)
	if err == nil {
		p.prs[k] = v
	}
	return v, err
}

func (p *Pass) reviewsOf(ctx context.Context, repo string, n int) ([]github.Review, error) {
	k := PRKey(repo, n)
	if v, ok := p.reviews[k]; ok {
		return v, nil
	}
	v, err := p.src.GitHub.Reviews(ctx, repo, n)
	if err == nil {
		p.reviews[k] = v
	}
	return v, err
}

func (p *Pass) checksOf(ctx context.Context, repo string, n int, sha string, required bool) ([]github.Check, error) {
	k := fmt.Sprintf("%s#%d@%s/%t", repo, n, sha, required)
	if v, ok := p.checks[k]; ok {
		return v, nil
	}
	v, err := p.src.GitHub.Checks(ctx, repo, n, sha, required)
	if err == nil {
		p.checks[k] = v
	}
	return v, err
}

func (p *Pass) runsOf(ctx context.Context, repo, sha string) ([]github.WorkflowRun, error) {
	k := repo + "@" + sha
	if v, ok := p.runs[k]; ok {
		return v, nil
	}
	v, err := p.src.GitHub.WorkflowRuns(ctx, repo, sha)
	if err == nil {
		p.runs[k] = v
	}
	return v, err
}

func (p *Pass) linearIssue(ctx context.Context, key string) (linear.Context, error) {
	if v, ok := p.issues[key]; ok {
		return v, nil
	}
	v, err := p.src.Linear.IssueContext(ctx, key)
	if err == nil {
		p.issues[key] = v
	}
	return v, err
}
