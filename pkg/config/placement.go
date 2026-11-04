package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Placement lets a step run on one of several pools (#409). Options are
// tried by weight: until PreferFor has passed since the step could start,
// only the top option whose When holds is used; after that the step moves
// to the highest-weight option whose resources are free.
type Placement struct {
	// Required are resource keys every option needs, like needs_resources.
	Required []string `yaml:"required"`
	// Options are the pools the step may run on.
	Options []PlacementOption `yaml:"options"`
	// PreferFor (Go duration) holds the step on its top option before it
	// may fall back. Empty or 0 falls back as soon as the top is full.
	PreferFor string `yaml:"prefer_for"`
}

// PlacementOption is one pool a step may run on.
type PlacementOption struct {
	Pool      string   `yaml:"pool"`
	Resources []string `yaml:"resources"`
	Weight    int      `yaml:"weight"`
	// When is "A == B" or "A != B" over run vars, decided at run start.
	// An option whose When is false is dropped for that run.
	When string `yaml:"when"`
}

// PreferDuration is PreferFor parsed; 0 when unset.
func (p *Placement) PreferDuration() time.Duration {
	d, _ := time.ParseDuration(p.PreferFor)
	return d
}

// ResolvedOption is a placement option filled from run vars.
type ResolvedOption struct {
	Pool      string
	Resources []string
	Weight    int
}

// ResolvePlacement fills the options from vars, drops those whose when
// is false, and orders the rest by weight (highest first, ties in
// declaration order). Required keys are added to each option's
// resources. It fails when no option is left.
func (c *Config) ResolvePlacement(w *Workflow, p *Placement, vars map[string]any) ([]ResolvedOption, error) {
	req, err := c.ResolveResources(w, p.Required, vars)
	if err != nil {
		return nil, fmt.Errorf("placement required: %w", err)
	}
	var out []ResolvedOption
	proj := ""
	for i, o := range p.Options {
		if o.When != "" {
			s, err := w.Render(o.When, vars)
			if err != nil {
				return nil, fmt.Errorf("placement option %d when: %w", i, err)
			}
			ok, err := EvalWhen(s)
			if err != nil {
				return nil, fmt.Errorf("placement option %d: %w", i, err)
			}
			if !ok {
				continue
			}
		}
		pool, err := c.ResolvePool(w, o.Pool, vars)
		if err != nil {
			return nil, fmt.Errorf("placement option %d: %w", i, err)
		}
		// The step ticket moves between option pools, so they must share
		// a project.
		if pp := c.Pools[pool].Project; proj == "" {
			proj = pp
		} else if pp != proj {
			return nil, fmt.Errorf("placement option %d: pool %s is in project %s, not %s", i, pool, pp, proj)
		}
		rs, err := c.ResolveResources(w, o.Resources, vars)
		if err != nil {
			return nil, fmt.Errorf("placement option %d: %w", i, err)
		}
		all := slices.Clone(req)
		for _, r := range rs {
			if !slices.Contains(all, r) {
				all = append(all, r)
			}
		}
		out = append(out, ResolvedOption{Pool: pool, Resources: all, Weight: o.Weight})
	}
	if len(out) == 0 {
		return nil, errors.New("placement: no option's when holds")
	}
	slices.SortStableFunc(out, func(a, b ResolvedOption) int { return b.Weight - a.Weight })
	return out, nil
}

// EvalWhen evaluates a filled "A == B" or "A != B". Each side is trimmed
// and may be quoted with ' or ".
func EvalWhen(s string) (bool, error) {
	op := "=="
	a, b, ok := strings.Cut(s, "!=")
	if ok {
		op = "!="
	} else if a, b, ok = strings.Cut(s, "=="); !ok {
		return false, fmt.Errorf("when %q: want A == B or A != B", s)
	}
	unq := func(x string) string {
		x = strings.TrimSpace(x)
		if len(x) >= 2 && (x[0] == '\'' || x[0] == '"') && x[len(x)-1] == x[0] {
			x = x[1 : len(x)-1]
		}
		return x
	}
	eq := unq(a) == unq(b)
	return eq == (op == "=="), nil
}

// checkPlacement checks a step's placement at load.
func (c *Config) checkPlacement(w *Workflow, s *Step) error {
	p := s.Placement
	switch {
	case s.Pool != "" || len(s.NeedsResources) > 0:
		return errors.New("placement replaces pool and needs_resources: set one or the other")
	case s.Auto():
		return errors.New("placement is for agent steps")
	case len(p.Options) == 0:
		return errors.New("placement: at least one option is required")
	}
	if p.PreferFor != "" {
		if d, err := time.ParseDuration(p.PreferFor); err != nil || d < 0 {
			return fmt.Errorf("placement prefer_for %q: want a Go duration >= 0", p.PreferFor)
		}
	}
	if err := c.checkStepResources(w, p.Required); err != nil {
		return fmt.Errorf("placement required: %w", err)
	}
	for i, o := range p.Options {
		if o.Pool == "" {
			return fmt.Errorf("placement option %d: pool is required", i)
		}
		if err := c.checkStepPool(w, o.Pool); err != nil {
			return fmt.Errorf("placement option %d: %w", i, err)
		}
		if err := c.checkStepResources(w, o.Resources); err != nil {
			return fmt.Errorf("placement option %d: %w", i, err)
		}
		if o.When == "" {
			continue
		}
		if !strings.Contains(o.When, "==") && !strings.Contains(o.When, "!=") {
			return fmt.Errorf("placement option %d: when %q: want A == B or A != B", i, o.When)
		}
		vs, err := refs(o.When, w.Template)
		if err != nil {
			return fmt.Errorf("placement option %d when: %w", i, err)
		}
		for _, v := range vs {
			if _, ok := w.PackVars[v]; !ok && !slices.Contains(w.Vars, v) {
				return fmt.Errorf("placement option %d: when uses %s, not in workflow vars or pack vars", i, placeholder(v, w.Template))
			}
		}
	}
	return nil
}
