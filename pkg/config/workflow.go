package config

import (
	"errors"
	"fmt"
	"slices"
)

// DefaultStepRetries is a step's retry limit when retries is unset.
const DefaultStepRetries = 2

// Workflow trackers (#201).
const (
	TrackerGitHub = "github"
	TrackerLinear = "linear"
)

// Workflow is an ordered list of steps started as one run (docs/plan.md
// section 6, "Runs and workflows"). Step titles and prompts are templates
// filled at run start from PackVars plus the run's vars.
type Workflow struct {
	// Vars are the var names a run must supply. Pack vars give defaults.
	Vars []string `yaml:"vars"`
	// Issue names the var holding a GitHub issue number, copied into
	// every ticket's meta.github.issue.
	Issue string `yaml:"issue"`
	// PR names the var holding a GitHub PR number (meta.github.pr).
	PR string `yaml:"pr"`
	// Repo is the GitHub repo (owner/name) the issue or PR lives in,
	// copied into meta.github.repo. Required with issue or pr (#183).
	Repo string `yaml:"repo"`
	// Tracker is where the issue var's issue lives: "github" (default)
	// or "linear" (#201). With linear, the issue var holds a key such as
	// ENG-123, runs get meta.linear = {workspace, issue}, and Repo is the
	// code repo for PRs (meta.github.repo).
	Tracker string `yaml:"tracker"`
	// LinearWorkspace names the Linear workspace; required with linear.
	LinearWorkspace string `yaml:"linear_workspace"`
	// EscalatePool gets the escalation ticket when a run needs a human.
	EscalatePool string `yaml:"escalate_pool"`
	Steps        []Step `yaml:"steps"`
	// PackVars are the vars of the pack that declared the workflow, set by
	// the loader, not by users.
	PackVars map[string]any `yaml:"pack_vars,omitempty"`
	// Template is the declaring pack's syntax (SyntaxGo or ""), set by
	// the loader.
	Template string `yaml:"template,omitempty"`
}

// Step is one ticket of a run, blocked on the previous step.
type Step struct {
	ID     string `yaml:"id"`
	Pool   string `yaml:"pool"`
	Title  string `yaml:"title"`
	Prompt string `yaml:"prompt"`
	// PromptFile is read into Prompt at load time, relative to the file
	// that declares it.
	PromptFile string `yaml:"prompt_file"`
	// Retries is how many failures the step takes before the run needs a
	// human. Nil means DefaultStepRetries.
	Retries *int `yaml:"retries"`
}

// RetryLimit is Retries or the default.
func (s Step) RetryLimit() int {
	if s.Retries == nil {
		return DefaultStepRetries
	}
	return *s.Retries
}

// templateKeys are the step fields kept as templates: pack substitution
// skips them so run vars survive until run start.
// pool is one too, so a run var can route a step to a pool (#384).
var templateKeys = []string{"title", "prompt", "pool"}

// checkTriggerRun validates a run: trigger. The workflow must exist and
// every workflow var must be set (by the trigger or by pack vars); vars
// and key may only reference ${check_output}, and only with a check.
func (c *Config) checkTriggerRun(tr *Trigger) error {
	r := tr.Run
	w, ok := c.Workflows[r.Workflow]
	if !ok {
		return fmt.Errorf("run: unknown workflow %q", r.Workflow)
	}
	for k := range r.Vars {
		if !slices.Contains(w.Vars, k) {
			return fmt.Errorf("run: workflow %s has no var %q", r.Workflow, k)
		}
	}
	for _, k := range w.Vars {
		if _, ok := r.Vars[k]; ok {
			continue
		}
		if _, ok := w.PackVars[k]; !ok {
			return fmt.Errorf("run: workflow %s needs var %s", r.Workflow, k)
		}
	}
	templates := map[string]string{"key": r.Key}
	for k, v := range r.Vars {
		templates["vars."+k] = v
	}
	for where, s := range templates {
		vs, err := refs(s, r.Template)
		if err != nil {
			return fmt.Errorf("run: %s: %w", where, err)
		}
		for _, v := range vs {
			if v != CheckOutputVar {
				return fmt.Errorf("run: %s uses %s; only %s is available", where, placeholder(v, r.Template), placeholder(CheckOutputVar, r.Template))
			}
			if tr.When.Check == "" {
				return fmt.Errorf("run: %s uses %s, which needs when.check", where, placeholder(CheckOutputVar, r.Template))
			}
		}
	}
	return nil
}

// holdTemplates removes step templates from the raw workflows section so
// pack var substitution leaves them alone. The returned func puts them
// back.
func holdTemplates(m map[string]any) func() {
	type held struct {
		step map[string]any
		key  string
		v    any
	}
	var hs []held
	wfs, _ := m["workflows"].(map[string]any)
	for _, w := range wfs {
		wm, _ := w.(map[string]any)
		steps, _ := wm["steps"].([]any)
		for _, s := range steps {
			sm, _ := s.(map[string]any)
			for _, k := range templateKeys {
				if v, ok := sm[k]; ok {
					hs = append(hs, held{sm, k, v})
					delete(sm, k)
				}
			}
		}
	}
	return func() {
		for _, h := range hs {
			h.step[h.key] = h.v
		}
	}
}

// setPackVars records the declaring pack's vars on each workflow it
// declares, and rejects a user-set pack_vars.
func setPackVars(m map[string]any, vars map[string]any, file string) error {
	wfs, _ := m["workflows"].(map[string]any)
	for name, w := range wfs {
		wm, _ := w.(map[string]any)
		if wm == nil {
			continue
		}
		if _, ok := wm["pack_vars"]; ok {
			return fmt.Errorf("%s: workflows.%s: pack_vars is set by the loader", file, name)
		}
		if vars != nil {
			wm["pack_vars"] = vars
		}
	}
	return nil
}

// FillTemplate replaces ${name} in s from vars ($$ is a literal $). A
// missing var is an error.
func FillTemplate(s string, vars map[string]any) (string, error) {
	out, err := substitute(s, vars, "", SyntaxDollar)
	if err != nil {
		return "", err
	}
	return fmt.Sprint(out), nil
}

// checkStepPool checks a step's pool. A pool naming a run var is checked
// at run start (ResolvePool); one filled from pack vars alone is checked
// here.
func (c *Config) checkStepPool(w *Workflow, pool string) error {
	runVar := false
	vs, err := refs(pool, w.Template)
	if err != nil {
		return fmt.Errorf("pool: %w", err)
	}
	for _, v := range vs {
		_, packed := w.PackVars[v]
		switch {
		case slices.Contains(w.Vars, v):
			runVar = true
		case !packed:
			return fmt.Errorf("pool uses %s, not in workflow vars or pack vars", placeholder(v, w.Template))
		}
	}
	if runVar {
		return nil
	}
	_, err = c.ResolvePool(w, pool, w.PackVars)
	return err
}

// ResolvePool fills a step's pool template of w from vars and checks the
// pool exists.
func (c *Config) ResolvePool(w *Workflow, pool string, vars map[string]any) (string, error) {
	name, err := w.Render(pool, vars)
	if err != nil {
		return "", fmt.Errorf("pool: %w", err)
	}
	if _, ok := c.Pools[name]; !ok {
		return "", fmt.Errorf("unknown pool %q", name)
	}
	return name, nil
}

// Render fills a step title, prompt or pool template of w from vars, in
// the declaring pack's syntax.
func (w *Workflow) Render(s string, vars map[string]any) (string, error) {
	return Fill(s, w.Template, vars)
}

// Fill fills a trigger run var or key template from vars.
func (r *TriggerRun) Fill(s string, vars map[string]any) (string, error) {
	return Fill(s, r.Template, vars)
}

// templateVars returns the ${name} references in s ($$ is a literal $).
func templateVars(s string) []string {
	var out []string
	for _, m := range varPattern.FindAllStringSubmatch(s, -1) {
		if m[1] != "" {
			out = append(out, m[1])
		}
	}
	return out
}

func (c *Config) checkWorkflow(name string, w *Workflow) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("name must match %s", namePattern)
	}
	for _, v := range w.Vars {
		if !varName.MatchString(v) {
			return fmt.Errorf("var %q is not a valid name", v)
		}
	}
	if w.Issue != "" && !slices.Contains(w.Vars, w.Issue) {
		return fmt.Errorf("issue %q is not in vars", w.Issue)
	}
	if w.PR != "" && !slices.Contains(w.Vars, w.PR) {
		return fmt.Errorf("pr %q is not in vars", w.PR)
	}
	if w.Issue != "" && w.PR != "" {
		return errors.New("set issue or pr, not both")
	}
	if (w.Issue != "" || w.PR != "") && w.Repo == "" {
		return errors.New("repo (owner/name) is required with issue or pr")
	}
	if w.Repo != "" && !RepoPattern.MatchString(w.Repo) {
		return fmt.Errorf("repo %q must be owner/name", w.Repo)
	}
	switch w.Tracker {
	case "", TrackerGitHub:
		if w.LinearWorkspace != "" {
			return errors.New("linear_workspace needs tracker: linear")
		}
	case TrackerLinear:
		if w.Issue == "" || w.LinearWorkspace == "" {
			return errors.New("tracker linear needs issue and linear_workspace")
		}
	default:
		return fmt.Errorf("tracker %q must be %q or %q", w.Tracker, TrackerGitHub, TrackerLinear)
	}
	if w.EscalatePool != "" {
		if _, ok := c.Pools[w.EscalatePool]; !ok {
			return fmt.Errorf("escalate_pool: unknown pool %q", w.EscalatePool)
		}
	}
	if len(w.Steps) == 0 {
		return errors.New("steps: at least one step is required")
	}
	ids := map[string]bool{}
	for i := range w.Steps {
		s := &w.Steps[i]
		where := fmt.Sprintf("steps[%d]", i)
		if s.ID != "" {
			where = "step " + s.ID
		}
		if !namePattern.MatchString(s.ID) {
			return fmt.Errorf("%s: id %q must match %s", where, s.ID, namePattern)
		}
		if ids[s.ID] {
			return fmt.Errorf("%s: duplicate step id", where)
		}
		ids[s.ID] = true
		if err := c.checkStepPool(w, s.Pool); err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		if s.Title == "" {
			return fmt.Errorf("%s: title is required", where)
		}
		if s.PromptFile != "" {
			return fmt.Errorf("%s: prompt_file needs a config file location", where)
		}
		if s.Prompt == "" {
			return fmt.Errorf("%s: set prompt or prompt_file", where)
		}
		if s.Retries != nil && *s.Retries < 0 {
			return fmt.Errorf("%s: retries must be >= 0", where)
		}
		for _, f := range []struct{ key, v string }{{"title", s.Title}, {"prompt", s.Prompt}} {
			vs, err := refs(f.v, w.Template)
			if err != nil {
				return fmt.Errorf("%s: %s: %w", where, f.key, err)
			}
			for _, v := range vs {
				if _, ok := w.PackVars[v]; !ok && !slices.Contains(w.Vars, v) {
					return fmt.Errorf("%s: %s uses %s, not in workflow vars or pack vars", where, f.key, placeholder(v, w.Template))
				}
			}
		}
	}
	return nil
}
