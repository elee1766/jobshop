package control

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/execution/automation"
	"github.com/gfx-labs/jobshop/pkg/orchestration/hooks"
)

// checkHookResources checks the resource keys of a start_step against the
// configured resources, as ResolveResources does for a workflow step, and
// drops duplicates.
func checkHookResources(cfg *config.Config, keys []string) ([]string, error) {
	var out []string
	for _, key := range keys {
		if err := cfg.CheckResourceKey(key); err != nil {
			return nil, fmt.Errorf("needs_resources: %w", err)
		}
		if !slices.Contains(out, key) {
			out = append(out, key)
		}
	}
	return out, nil
}

// hookPlacement resolves a start_step placement like resolvePlacement does
// for a workflow step: pools through the callback's reference scope,
// required keys added to each option, options ordered by weight (highest
// first, ties in call order). Every pool must be in the run's project, so
// the ticket can move between them.
func hookPlacement(cfg *config.Config, in hooks.Invocation, a *domain.RunAutomation, project string, p *automation.ActionPlacement) (*domain.StepPlacement, error) {
	required, err := checkHookResources(cfg, p.Required)
	if err != nil {
		return nil, fmt.Errorf("placement required: %w", err)
	}
	sp := &domain.StepPlacement{}
	if p.PreferFor != "" {
		d, err := time.ParseDuration(p.PreferFor)
		if err != nil || d < 0 {
			return nil, fmt.Errorf("placement prefer_for %q: want a Go duration >= 0", p.PreferFor)
		}
		sp.PreferMS = d.Milliseconds()
	}
	for i, o := range p.Options {
		pool, err := hookReference(cfg, in, a, "pool", o.Pool)
		if err != nil {
			return nil, fmt.Errorf("placement option %d: %w", i, err)
		}
		if cfg.Pools[pool].Project != project {
			return nil, fmt.Errorf("placement option %d: pool %s is outside run project %s", i, pool, project)
		}
		resources, err := checkHookResources(cfg, o.Resources)
		if err != nil {
			return nil, fmt.Errorf("placement option %d: %w", i, err)
		}
		all := slices.Clone(required)
		for _, r := range resources {
			if !slices.Contains(all, r) {
				all = append(all, r)
			}
		}
		sp.Options = append(sp.Options, domain.PlacementOption{Pool: pool, Resources: all, Weight: o.Weight})
	}
	if len(sp.Options) == 0 {
		return nil, errors.New("placement: at least one option is required")
	}
	slices.SortStableFunc(sp.Options, func(a, b domain.PlacementOption) int { return b.Weight - a.Weight })
	return sp, nil
}
