package config

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"
)

// DefaultStepRetries is a step's retry limit when retries is unset.
const DefaultStepRetries = 2

// OutcomePattern is a step outcome word (`jobshop close --outcome`).
var OutcomePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

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
	// Escalate opens an escalation chat for the run's escalation tickets
	// (#379). Nil: the top-level escalation block, if any.
	Escalate *EscalationChat `yaml:"escalate"`
	Steps    []Step          `yaml:"steps"`
	// PackVars are the vars of the pack that declared the workflow, set by
	// the loader, not by users.
	PackVars map[string]any `yaml:"pack_vars,omitempty"`
	// Template is the declaring pack's syntax (SyntaxGo or ""), set by
	// the loader.
	Template string `yaml:"template,omitempty"`
}

// Step types (#396). Command and http steps are run by the controller
// itself, with no session.
const (
	StepAgent   = "agent"
	StepCommand = "command"
	StepHTTP    = "http"
)

// Step on_restart values: what the controller does with an attempt it
// was running when it stopped.
const (
	OnRestartRetry = "retry"
	OnRestartFail  = "fail"
)

// Defaults for command and http steps.
const (
	DefaultStepTimeout     = 10 * time.Minute
	DefaultStepRetryAfter  = time.Minute
	DefaultStepConcurrency = 4
)

// Step is one ticket of a run, blocked on the previous step.
type Step struct {
	ID string `yaml:"id"`
	// Type is agent (default), command or http.
	Type string `yaml:"type,omitempty"`
	// Pool runs an agent step. For command and http steps it is optional
	// and only picks the project and the working directory (the pool's
	// dir, else its repo).
	Pool   string `yaml:"pool"`
	Title  string `yaml:"title"`
	Prompt string `yaml:"prompt"`
	// Run is a command step's command: a list (argv) or a string run with
	// sh -c. Templates are filled at execution time.
	Run any `yaml:"run,omitempty"`
	// URL, Method (default POST), Headers, HeaderFiles (header name to a
	// file whose trimmed contents are the value) and Body make an http
	// step's request.
	URL         string            `yaml:"url,omitempty"`
	Method      string            `yaml:"method,omitempty"`
	Headers     map[string]string `yaml:"headers,omitempty"`
	HeaderFiles map[string]string `yaml:"header_files,omitempty"`
	Body        string            `yaml:"body,omitempty"`
	// Timeout (Go duration) bounds one attempt. Default 10m.
	Timeout string `yaml:"timeout,omitempty"`
	// RetryAfter is the wait after a "not yet" result with no
	// Retry-After header. Default 1m.
	RetryAfter string `yaml:"retry_after,omitempty"`
	// Concurrency caps attempts of this step running at once across runs.
	Concurrency int `yaml:"concurrency,omitempty"`
	// OnRestart is retry (default) or fail.
	OnRestart string `yaml:"on_restart,omitempty"`
	// PromptFile is read into Prompt at load time, relative to the file
	// that declares it.
	PromptFile string `yaml:"prompt_file"`
	// Retries is how many failures the step takes before the run needs a
	// human. Nil means DefaultStepRetries.
	Retries *int `yaml:"retries"`
	// NeedsResources are resource keys the step holds while in progress
	// (#400): a resource name, or "name:<key>" for a keyed resource.
	// Run vars fill them at run start.
	NeedsResources []string `yaml:"needs_resources"`
	// Needs are the step ids this step waits on (#385). Empty means the
	// previous step; the first step waits on nothing. Two steps with the
	// same needs become ready together.
	Needs []string `yaml:"needs"`
	// When gates the step on an earlier step's recorded outcome; when it
	// does not match, the step closes skipped instead of running.
	When *StepWhen `yaml:"when"`
	// Placement replaces pool and needs_resources with weighted options
	// (#409).
	Placement *Placement `yaml:"placement,omitempty"`
	// Worktree runs a command step in a git worktree of its pool's repo
	// at the pool ref, one per run, shared by the run's command steps and
	// removed when the run closes (#419).
	Worktree bool `yaml:"worktree,omitempty"`
}

// StepWhen is a step condition: run only if step closed with outcome.
type StepWhen struct {
	Step    string `yaml:"step"`
	Outcome string `yaml:"outcome"`
}

// StepNeeds returns the step ids step i waits on: its needs, or the
// previous step.
func (w Workflow) StepNeeds(i int) []string {
	if len(w.Steps[i].Needs) > 0 {
		return w.Steps[i].Needs
	}
	if i == 0 {
		return nil
	}
	return []string{w.Steps[i-1].ID}
}

// Auto reports whether the controller runs the step (command or http).
func (s Step) Auto() bool { return s.Type == StepCommand || s.Type == StepHTTP }

// Argv is a command step's run as argv; a string runs with sh -c.
func (s Step) Argv() ([]string, error) {
	switch r := s.Run.(type) {
	case string:
		if strings.TrimSpace(r) == "" {
			return nil, errors.New("run is empty")
		}
		return []string{"sh", "-c", r}, nil
	case []string:
		if len(r) == 0 {
			return nil, errors.New("run is empty")
		}
		return r, nil
	case []any:
		if len(r) == 0 {
			return nil, errors.New("run is empty")
		}
		out := make([]string, len(r))
		for i, a := range r {
			out[i] = fmt.Sprint(a)
		}
		return out, nil
	case nil:
		return nil, errors.New("run is required")
	}
	return nil, fmt.Errorf("run must be a string or a list, not %T", s.Run)
}

// TimeoutDur is Timeout or the default.
func (s Step) TimeoutDur() time.Duration {
	if d, err := time.ParseDuration(s.Timeout); err == nil && d > 0 {
		return d
	}
	return DefaultStepTimeout
}

// RetryAfterDur is RetryAfter or the default.
func (s Step) RetryAfterDur() time.Duration {
	if d, err := time.ParseDuration(s.RetryAfter); err == nil && d > 0 {
		return d
	}
	return DefaultStepRetryAfter
}

// ConcurrencyLimit is Concurrency or the default.
func (s Step) ConcurrencyLimit() int {
	if s.Concurrency > 0 {
		return s.Concurrency
	}
	return DefaultStepConcurrency
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
// run, url, body and headers are filled when a command or http step runs
// (#396). placement holds pools, keys and when conditions filled at run
// start (#409).
var templateKeys = []string{"title", "prompt", "pool", "needs_resources", "placement", "run", "url", "body", "headers"}

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
	if w.Escalate != nil {
		if err := c.checkEscalation(w.Escalate); err != nil {
			return fmt.Errorf("escalate: %w", err)
		}
	}
	if len(w.Steps) == 0 {
		return errors.New("steps: at least one step is required")
	}
	ids := map[string]bool{}
	// ancestors[id] is every step id that id transitively waits on.
	ancestors := map[string]map[string]bool{}
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
		if s.Type == "" {
			s.Type = StepAgent
		}
		if s.Auto() {
			if err := c.checkAutoStep(w, s); err != nil {
				return fmt.Errorf("%s: %w", where, err)
			}
		} else if s.Type != StepAgent {
			return fmt.Errorf("%s: type %q must be agent, command or http", where, s.Type)
		}
		if s.Worktree && (s.Type != StepCommand || s.Pool == "") {
			return fmt.Errorf("%s: worktree needs a command step with a pool", where)
		}
		if s.Placement != nil {
			if err := c.checkPlacement(w, s); err != nil {
				return fmt.Errorf("%s: %w", where, err)
			}
		} else if s.Pool != "" || !s.Auto() {
			if err := c.checkStepPool(w, s.Pool); err != nil {
				return fmt.Errorf("%s: %w", where, err)
			}
		}
		if s.Title == "" && s.Auto() {
			s.Title = s.Type + " " + s.ID
		}
		if err := c.checkStepResources(w, s.NeedsResources); err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		if s.Title == "" {
			return fmt.Errorf("%s: title is required", where)
		}
		if s.PromptFile != "" {
			return fmt.Errorf("%s: prompt_file needs a config file location", where)
		}
		if s.Prompt == "" && !s.Auto() {
			return fmt.Errorf("%s: set prompt or prompt_file", where)
		}
		if s.Retries != nil && *s.Retries < 0 {
			return fmt.Errorf("%s: retries must be >= 0", where)
		}
		// Needs name earlier steps only, so the steps form a DAG in
		// declaration order.
		anc := map[string]bool{}
		for _, n := range w.StepNeeds(i) {
			if n == s.ID || !ids[n] {
				return fmt.Errorf("%s: needs %q, which is not an earlier step", where, n)
			}
			anc[n] = true
			for a := range ancestors[n] {
				anc[a] = true
			}
		}
		if s.When != nil {
			if !anc[s.When.Step] {
				return fmt.Errorf("%s: when.step %q is not a step it needs (directly or through others)", where, s.When.Step)
			}
			if !OutcomePattern.MatchString(s.When.Outcome) {
				return fmt.Errorf("%s: when.outcome %q must match %s", where, s.When.Outcome, OutcomePattern)
			}
		}
		ancestors[s.ID] = anc
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

// EscalationChat is the escalation chat policy (#379): a workflow's
// escalate block, or the top-level escalation block for runs without one
// and for pool quarantines. The controller starts one session in
// ChatPool keyed escalation:<ticket> and sends the rendered prompt as its
// first message. An empty ChatPool means no chat, ticket only.
type EscalationChat struct {
	ChatPool string `yaml:"chat_pool"`
	// Prompt is the first message of a run escalation chat. Vars:
	// EscalationVars, plus the declaring pack's vars on a workflow.
	Prompt     string `yaml:"prompt"`
	PromptFile string `yaml:"prompt_file"`
	// PoolPrompt is the first message of a pool quarantine escalation
	// chat. Vars: EscalationPoolVars.
	PoolPrompt     string `yaml:"pool_prompt"`
	PoolPromptFile string `yaml:"pool_prompt_file"`
	// Template is the declaring pack's syntax (#383), set by the loader.
	Template string `yaml:"template,omitempty"`
}

// Escalation chat template vars, filled when the chat starts.
var (
	EscalationVars     = []string{"ticket", "run", "workflow", "issue_ref", "body"}
	EscalationPoolVars = []string{"ticket", "pool", "error", "body"}
)

// escalationPassthrough returns vars plus each escalation var mapped to
// its own reference, so pack substitution leaves them for the controller.
// Escalation vars win over pack vars of the same name.
func escalationPassthrough(vars map[string]any, syn string) map[string]any {
	out := make(map[string]any, len(vars)+len(EscalationVars)+len(EscalationPoolVars))
	for k, v := range vars {
		out[k] = v
	}
	for _, k := range append(slices.Clone(EscalationVars), EscalationPoolVars...) {
		out[k] = placeholder(k, syn)
	}
	return out
}

// stepRef matches ${steps.<id>.output} and ${steps.<id>.json.<path>},
// the outputs of earlier steps, filled when a command or http step runs.
var stepRef = regexp.MustCompile(`\$\{steps\.([a-z][a-z0-9_-]*)\.(output|outcome|json\.[A-Za-z0-9_.-]+)\}`)

// StepRefs returns the step ids s refers to with ${steps.<id>...}.
func StepRefs(s string) []string {
	var out []string
	for _, m := range stepRef.FindAllStringSubmatch(s, -1) {
		out = append(out, m[1])
	}
	return out
}

// holdEscalationPrompts removes inline escalation prompts so pack var
// substitution of the whole file leaves them alone; the returned func
// puts them back. resolveEscalationFiles then substitutes them with the
// escalation vars passed through.
func holdEscalationPrompts(m map[string]any) func() {
	type held struct {
		e map[string]any
		k string
		v any
	}
	var hs []held
	for _, b := range escalationBlocks(m) {
		e, _ := b.(map[string]any)
		for _, k := range []string{"prompt", "pool_prompt"} {
			if v, ok := e[k]; ok {
				hs = append(hs, held{e, k, v})
				delete(e, k)
			}
		}
	}
	return func() {
		for _, h := range hs {
			h.e[h.k] = h.v
		}
	}
}

// escalationBlocks returns the raw escalation maps of m: the top-level
// escalation block and each workflow's escalate block.
func escalationBlocks(m map[string]any) map[string]any {
	out := map[string]any{}
	if e, ok := m["escalation"].(map[string]any); ok {
		out["escalation"] = e
	}
	wfs, _ := m["workflows"].(map[string]any)
	for n, w := range wfs {
		wm, _ := w.(map[string]any)
		if e, ok := wm["escalate"].(map[string]any); ok {
			out["workflows."+n+".escalate"] = e
		}
	}
	return out
}

// checkEscalation validates an escalation block.
func (c *Config) checkEscalation(e *EscalationChat) error {
	if e.PromptFile != "" || e.PoolPromptFile != "" {
		return errors.New("prompt_file needs a config file location")
	}
	if e.ChatPool == "" {
		return nil
	}
	p, ok := c.Pools[e.ChatPool]
	if !ok {
		return fmt.Errorf("chat_pool: unknown pool %q", e.ChatPool)
	}
	if !p.Chat() {
		return fmt.Errorf("chat_pool: pool %s is not a chat pool (type: chat)", e.ChatPool)
	}
	if e.Prompt == "" && e.PoolPrompt == "" {
		return errors.New("chat_pool needs prompt or prompt_file")
	}
	for _, f := range []struct {
		key, v string
		vars   []string
	}{{"prompt", e.Prompt, EscalationVars}, {"pool_prompt", e.PoolPrompt, EscalationPoolVars}} {
		vs, err := refs(f.v, e.Template)
		if err != nil {
			return fmt.Errorf("%s: %w", f.key, err)
		}
		for _, v := range vs {
			if !slices.Contains(f.vars, v) {
				return fmt.Errorf("%s uses %s; available: %v", f.key, placeholder(v, e.Template), f.vars)
			}
		}
	}
	return nil
}

// FillStepTemplate fills a command or http step template of w. In a
// ${name} pack it fills ${steps.<id>.output|outcome|json.<path>} from
// look, then ${name} from vars. In a template: go pack the outputs are
// the steps var: {{.steps.<id>.output}}, {{.steps.<id>.json.sha}}
// (index for ids with a dash: {{(index .steps "wait-ci").output}}).
func (w *Workflow) FillStepTemplate(s string, vars map[string]any, steps map[string]any, look func(id, field string) (string, bool)) (string, error) {
	if w.Template == SyntaxGo {
		all := maps.Clone(vars)
		all[StepsVar] = steps
		return Fill(s, SyntaxGo, all)
	}
	var missing string
	s = stepRef.ReplaceAllStringFunc(s, func(m string) string {
		p := stepRef.FindStringSubmatch(m)
		v, ok := look(p[1], p[2])
		if !ok {
			if missing == "" {
				missing = m
			}
			return ""
		}
		// Escape so a $ in the output is not read as a var.
		return strings.ReplaceAll(v, "$", "$$")
	})
	if missing != "" {
		return "", fmt.Errorf("%s: no such step output", missing)
	}
	return FillTemplate(s, vars)
}

// StepsVar is the template: go var holding earlier steps' outputs.
const StepsVar = "steps"

// checkAutoStep checks a command or http step.
func (c *Config) checkAutoStep(w *Workflow, s *Step) error {
	var templates []string
	switch s.Type {
	case StepCommand:
		argv, err := s.Argv()
		if err != nil {
			return err
		}
		templates = argv
		if s.URL != "" || s.Body != "" || s.Method != "" || len(s.Headers) > 0 || len(s.HeaderFiles) > 0 {
			return errors.New("url, method, headers, header_files and body are for http steps")
		}
	case StepHTTP:
		if s.URL == "" {
			return errors.New("url is required")
		}
		if s.Run != nil {
			return errors.New("run is for command steps")
		}
		switch s.Method {
		case "":
			s.Method = "POST"
		case "GET", "POST", "PUT", "PATCH", "DELETE":
		default:
			return fmt.Errorf("method %q must be GET, POST, PUT, PATCH or DELETE", s.Method)
		}
		templates = append(templates, s.URL, s.Body)
		for _, v := range s.Headers {
			templates = append(templates, v)
		}
	}
	if s.Prompt != "" {
		return errors.New("prompt is for agent steps")
	}
	for _, f := range []struct{ key, v string }{{"timeout", s.Timeout}, {"retry_after", s.RetryAfter}} {
		if f.v == "" {
			continue
		}
		if d, err := time.ParseDuration(f.v); err != nil || d <= 0 {
			return fmt.Errorf("%s %q must be a positive duration", f.key, f.v)
		}
	}
	if s.Concurrency < 0 {
		return errors.New("concurrency must be >= 0")
	}
	switch s.OnRestart {
	case "", OnRestartRetry, OnRestartFail:
	default:
		return fmt.Errorf("on_restart %q must be %s or %s", s.OnRestart, OnRestartRetry, OnRestartFail)
	}
	self := slices.IndexFunc(w.Steps, func(o Step) bool { return o.ID == s.ID })
	for _, t := range templates {
		if w.Template == SyntaxGo {
			vs, err := refs(t, SyntaxGo)
			if err != nil {
				return err
			}
			for _, v := range vs {
				if _, ok := w.PackVars[v]; !ok && !slices.Contains(w.Vars, v) && v != StepsVar {
					return fmt.Errorf("{{.%s}} is not in workflow vars or pack vars", v)
				}
			}
			continue
		}
		for _, id := range StepRefs(t) {
			if i := slices.IndexFunc(w.Steps, func(o Step) bool { return o.ID == id }); i < 0 || i >= self {
				return fmt.Errorf("${steps.%s...}: no earlier step %q", id, id)
			}
		}
		for _, v := range templateVars(stepRef.ReplaceAllString(t, "")) {
			if _, ok := w.PackVars[v]; !ok && !slices.Contains(w.Vars, v) {
				return fmt.Errorf("${%s} is not in workflow vars or pack vars (write $$ for a literal $)", v)
			}
		}
	}
	return nil
}
