package automation

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/execution/automation/api"
	"go.starlark.net/starlark"
)

// ActionPlacement is start_step(placement=...): weighted pool options for an
// agent step, the Starlark form of a workflow step's placement (#409). There
// is no per-option when: the script decides before it builds the dict.
type ActionPlacement struct {
	// Required are resource keys every option needs.
	Required []string `json:"required,omitempty"`
	// Options are the pools the step may run on.
	Options []ActionPlacementOption `json:"options"`
	// PreferFor (Go duration) holds the step on its top option before it may
	// fall back.
	PreferFor string `json:"prefer_for,omitempty"`
}

// ActionPlacementOption is one pool a step may run on.
type ActionPlacementOption struct {
	Pool      string   `json:"pool"`
	Resources []string `json:"resources,omitempty"`
	Weight    int      `json:"weight,omitempty"`
}

func clonePlacement(p *ActionPlacement) *ActionPlacement {
	if p == nil {
		return nil
	}
	out := *p
	out.Required = slices.Clone(p.Required)
	out.Options = slices.Clone(p.Options)
	for i := range out.Options {
		out.Options[i].Resources = slices.Clone(out.Options[i].Resources)
	}
	return &out
}

// stringList reads a Starlark list of nonempty strings.
func stringList(kind, name string, list *starlark.List) ([]string, error) {
	var out []string
	if list == nil {
		return out, nil
	}
	for i := 0; i < list.Len(); i++ {
		s, ok := starlark.AsString(list.Index(i))
		if !ok || strings.TrimSpace(s) == "" {
			return nil, fmt.Errorf("%s: %s must be a list of nonempty strings", kind, name)
		}
		out = append(out, s)
	}
	return out, nil
}

// parsePlacement checks the placement dict the way config checks a
// workflow step's placement. Resource names are checked against the
// configuration by the controller.
func parsePlacement(d *starlark.Dict, limit int) (*ActionPlacement, error) {
	data, err := api.FromValue(d, limit)
	if err != nil {
		return nil, fmt.Errorf("start_step: placement: %w", err)
	}
	m := data.(map[string]any)
	for k := range m {
		if k != "required" && k != "options" && k != "prefer_for" {
			return nil, fmt.Errorf("start_step: placement: unknown key %q (want required, options, prefer_for)", k)
		}
	}
	p := &ActionPlacement{}
	if p.Required, err = anyStrings(m["required"], "placement required"); err != nil {
		return nil, err
	}
	if v, ok := m["prefer_for"]; ok {
		s, isString := v.(string)
		if !isString {
			return nil, errors.New("start_step: placement prefer_for must be a duration string")
		}
		if d, err := time.ParseDuration(s); err != nil || d < 0 {
			return nil, fmt.Errorf("start_step: placement prefer_for %q: want a Go duration >= 0", s)
		}
		p.PreferFor = s
	}
	options, _ := m["options"].([]any)
	if len(options) == 0 {
		return nil, errors.New("start_step: placement: at least one option is required")
	}
	for i, raw := range options {
		o, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("start_step: placement option %d must be a dict", i)
		}
		for k := range o {
			if k != "pool" && k != "resources" && k != "weight" {
				return nil, fmt.Errorf("start_step: placement option %d: unknown key %q (want pool, resources, weight)", i, k)
			}
		}
		opt := ActionPlacementOption{}
		if opt.Pool, _ = o["pool"].(string); strings.TrimSpace(opt.Pool) == "" {
			return nil, fmt.Errorf("start_step: placement option %d: pool is required", i)
		}
		if opt.Resources, err = anyStrings(o["resources"], fmt.Sprintf("placement option %d resources", i)); err != nil {
			return nil, err
		}
		if w, ok := o["weight"]; ok {
			n, isInt := w.(int64)
			if !isInt {
				return nil, fmt.Errorf("start_step: placement option %d: weight must be an integer", i)
			}
			opt.Weight = int(n)
		}
		p.Options = append(p.Options, opt)
	}
	return p, nil
}

func anyStrings(v any, what string) ([]string, error) {
	if v == nil {
		return nil, nil
	}
	items, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("start_step: %s must be a list of nonempty strings", what)
	}
	var out []string
	for _, item := range items {
		s, ok := item.(string)
		if !ok || strings.TrimSpace(s) == "" {
			return nil, fmt.Errorf("start_step: %s must be a list of nonempty strings", what)
		}
		out = append(out, s)
	}
	return out, nil
}

// checkStepFields validates the start_step kwargs that mirror workflow step
// fields, with the rules config applies to their YAML counterparts. It runs
// after the type checks, so a.Type, a.Script and a.RetryPool are final.
func (a *Action) checkStepFields(placement *starlark.Dict, needsResources *starlark.List, maxReviewRounds starlark.Value, limit int) error {
	script := a.Script != ""
	command := a.Type == "command"
	var err error
	if a.NeedsResources, err = stringList("start_step", "needs_resources", needsResources); err != nil {
		return err
	}
	if placement != nil {
		switch {
		case script || command:
			return errors.New("start_step: placement is for agent steps")
		case a.Pool != "" || len(a.NeedsResources) > 0:
			return errors.New("start_step: placement replaces pool and needs_resources: set one or the other")
		case a.RetryPool != "":
			return errors.New("start_step: retry_pool is for agent steps with a pool")
		}
		if a.Placement, err = parsePlacement(placement, limit); err != nil {
			return err
		}
	}
	if a.RetryAfter != "" {
		if !script && !command {
			return errors.New("start_step: retry_after needs script or type=\"command\"")
		}
		if d, err := time.ParseDuration(a.RetryAfter); err != nil || d <= 0 {
			return errors.New("start_step: retry_after must be a positive duration")
		}
	}
	if maxReviewRounds != nil && maxReviewRounds != starlark.None {
		n, ok := maxReviewRounds.(starlark.Int)
		v, fits := n.Int64()
		if !ok || !fits || v < 0 || v > math.MaxInt32 {
			return errors.New("start_step: max_review_rounds must be an integer >= 0")
		}
		if script || command {
			return errors.New("start_step: max_review_rounds is for agent steps")
		}
		r := int(v)
		a.MaxReviewRounds = &r
	}
	if a.OnRequestChanges != "" && (script || command) {
		return errors.New("start_step: on_request_changes is for agent steps")
	}
	return nil
}
