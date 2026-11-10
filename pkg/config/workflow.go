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
	Project         string `yaml:"project,omitempty"`
	OnStart         string `yaml:"on_start,omitempty"`
	OnStepCompleted string `yaml:"on_step_completed,omitempty"`
	// PinCallbacks pins each run to the callbacks it last ran under (code,
	// grants, entry points, reference scope): a reload that changes them stops
	// the run until an operator repins it (JOB-188). Off by default, a run
	// follows the live callbacks. Only for a callback workflow.
	PinCallbacks bool `yaml:"pin_callbacks,omitempty"`
	// Enabled false keeps the workflow validated but not loaded (JOB-24).
	Enabled *bool `yaml:"enabled,omitempty"`
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
	// LinearWorkspace names the Linear workspace. With linear it may be
	// empty at load (a github instance whose Linear vars are unset, #652);
	// starting a run of the workflow then fails.
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
	// RetryPool is the pool a failed attempt or a review resume moves the
	// step to (JOB-59): a cheaper model's step escalates to a stronger
	// one. A template like pool. Empty, or equal to the step's pool, keeps
	// the step where it is. The move is one way.
	RetryPool string `yaml:"retry_pool,omitempty"`
	// OnRequestChanges is the step a `close -r request_changes` of this
	// step reopens with the findings (JOB-98). It must be one of the
	// step's needs, so the reviewer reopens blocked behind it. Empty: the
	// previous step, when the step needs it.
	OnRequestChanges string `yaml:"on_request_changes,omitempty"`
	// MaxReviewRounds caps the request_changes closes of this step in one
	// run; the next one sends the run to a human. Nil means
	// store.DefaultReviewRounds.
	MaxReviewRounds *int `yaml:"max_review_rounds,omitempty"`
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

// RequestChangesTarget is the step id a request_changes close of step i
// reopens (JOB-98): on_request_changes, else the previous step when step
// i needs it, else "" (step i cannot request changes).
func (w Workflow) RequestChangesTarget(i int) string {
	if t := w.Steps[i].OnRequestChanges; t != "" {
		return t
	}
	if i == 0 || w.Steps[i].Auto() {
		return ""
	}
	if prev := w.Steps[i-1].ID; slices.Contains(w.StepNeeds(i), prev) {
		return prev
	}
	return ""
}

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
var templateKeys = []string{"title", "prompt", "pool", "retry_pool", "needs_resources", "placement", "run", "url", "body", "headers"}

// checkTriggerRun validates a run: trigger. The workflow must exist and
// every workflow var must be set (by the trigger or by pack vars); vars
// and key may only reference ${check_output}, and only with a check.
func (c *Config) checkTriggerRun(tr *Trigger) error {
	r := tr.Run
	if r.KeyPrefix != "" {
		return errors.New("run: key_prefix needs when.scan")
	}
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
	return checkTriggerTemplates(tr, templates)
}

// checkTriggerTemplates checks that templates reference only
// {{.check_output}} (with a check) and {{.event.<field>}} (with on:).
func checkTriggerTemplates(tr *Trigger, templates map[string]string) error {
	for where, s := range templates {
		vs, err := refs(s)
		if err != nil {
			return fmt.Errorf("%s: %w", where, err)
		}
		for _, v := range vs {
			switch {
			case v == EventVar && len(tr.On) > 0:
			case v == CheckOutputVar && tr.When.Check != "":
			case v == CheckOutputVar:
				return fmt.Errorf("%s uses %s, which needs when.check", where, placeholder(CheckOutputVar))
			case v == EventVar:
				return fmt.Errorf("%s uses %s, which needs an on: list", where, placeholder(EventVar+".<field>"))
			default:
				return fmt.Errorf("%s uses %s; only %s and %s are available", where, placeholder(v), placeholder(CheckOutputVar), placeholder(EventVar+".<field>"))
			}
		}
	}
	return nil
}

// checkTriggerOn validates on: and fallback:.
func checkTriggerOn(tr *Trigger) error {
	for i, e := range tr.On {
		n := 0
		for _, set := range []bool{e.GitHub != nil, e.Linear != nil, e.Source != ""} {
			if set {
				n++
			}
		}
		if n != 1 {
			return fmt.Errorf("on[%d]: set exactly one of github, linear and source", i)
		}
		if e.Event != "" && e.Source == "" {
			return fmt.Errorf("on[%d]: event needs source", i)
		}
	}
	if tr.Fallback != "" {
		if d, err := time.ParseDuration(tr.Fallback); err != nil || d <= 0 {
			return fmt.Errorf("fallback %q must be a positive duration", tr.Fallback)
		}
	}
	return nil
}

// checkStepPool checks a step's pool. A pool naming a run var is checked
// at run start (ResolvePool); one filled from pack vars alone is checked
// here.
func (c *Config) checkStepPool(w *Workflow, pool string) error {
	runVar := false
	vs, err := refs(pool)
	if err != nil {
		return fmt.Errorf("pool: %w", err)
	}
	for _, v := range vs {
		_, packed := w.PackVars[v]
		switch {
		case slices.Contains(w.Vars, v):
			runVar = true
		case !packed:
			return fmt.Errorf("pool uses %s, not in workflow vars or pack vars", placeholder(v))
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
		if slices.Contains(c.Disabled.Pools, name) {
			return "", fmt.Errorf("pool %q is disabled", name)
		}
		return "", fmt.Errorf("unknown pool %q", name)
	}
	return name, nil
}

// Render fills a step title, prompt or pool template of w from vars.
func (w *Workflow) Render(s string, vars map[string]any) (string, error) {
	return Fill(s, vars)
}

// Fill fills a trigger run var or key template from vars.
func (r *TriggerRun) Fill(s string, vars map[string]any) (string, error) {
	return Fill(s, vars)
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
		if w.Issue == "" {
			return errors.New("tracker linear needs issue")
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
	if w.PinCallbacks && w.OnStart == "" {
		return errors.New("pin_callbacks needs a callback workflow (on_start and on_step_completed)")
	}
	if w.OnStart != "" || w.OnStepCompleted != "" {
		if w.OnStart == "" || w.OnStepCompleted == "" || len(w.Steps) != 0 {
			return errors.New("hook workflows require on_start and on_step_completed, without steps")
		}
		if w.Project == "" {
			w.Project = c.DefaultProject
		}
		if _, ok := c.Projects[w.Project]; !ok {
			return fmt.Errorf("unknown project %q", w.Project)
		}
		return nil
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
		if s.RetryPool != "" {
			if s.Auto() || s.Placement != nil {
				return fmt.Errorf("%s: retry_pool is for agent steps with a pool", where)
			}
			if err := c.checkStepPool(w, s.RetryPool); err != nil {
				return fmt.Errorf("%s: retry_pool: %w", where, err)
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
		if s.MaxReviewRounds != nil && *s.MaxReviewRounds < 0 {
			return fmt.Errorf("%s: max_review_rounds must be >= 0", where)
		}
		if t := s.OnRequestChanges; t != "" {
			// A direct need: the reviewer reopens blocked behind it. A
			// step further back would leave the reviewer unblocked by the
			// done steps in between.
			if s.Auto() || !slices.Contains(w.StepNeeds(i), t) {
				return fmt.Errorf("%s: on_request_changes %q must be a step this agent step needs directly", where, t)
			}
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
			vs, err := refs(f.v)
			if err != nil {
				return fmt.Errorf("%s: %s: %w", where, f.key, err)
			}
			for _, v := range vs {
				if _, ok := w.PackVars[v]; !ok && !slices.Contains(w.Vars, v) {
					return fmt.Errorf("%s: %s uses %s, not in workflow vars or pack vars", where, f.key, placeholder(v))
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
	// chat. Vars: EscalationVars.
	PoolPrompt     string `yaml:"pool_prompt"`
	PoolPromptFile string `yaml:"pool_prompt_file"`
	// HoldPrompt is the first message for every other hold that needs a
	// human: a wait past its timeout, a guard ask_human, a worker's
	// needs-human label. Empty: no chat for those. Vars: EscalationVars.
	HoldPrompt     string `yaml:"hold_prompt"`
	HoldPromptFile string `yaml:"hold_prompt_file"`
}

// EscalationVars are the template vars of an escalation chat's first
// message, besides the pack vars, filled when the chat starts. Every one
// is always set, empty when the cause has none. ticket and title are the
// escalation ticket, cause is run, pool, wait, guard or label, reason the
// short cause text and body the ticket body. held and held_title are the
// held ticket (the failed step, the waiting ticket, the labeled ticket),
// worker the session that raised the hold, run, workflow and step its
// place in a run, issue_ref how to read the issue, pool and quarantine a
// quarantined pool (error is quarantine, its older name).
var EscalationVars = []string{
	"ticket", "title", "cause", "reason", "body", "held", "held_title", "worker",
	"run", "workflow", "step", "issue_ref", "pool", "quarantine", "error",
}

// checkEscalation validates an escalation block.
func (c *Config) checkEscalation(e *EscalationChat) error {
	if e.PromptFile != "" || e.PoolPromptFile != "" || e.HoldPromptFile != "" {
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
	if e.Prompt == "" && e.PoolPrompt == "" && e.HoldPrompt == "" {
		return errors.New("chat_pool needs prompt or prompt_file")
	}
	for _, f := range []struct {
		key, v string
		vars   []string
	}{{"prompt", e.Prompt, EscalationVars}, {"pool_prompt", e.PoolPrompt, EscalationVars}, {"hold_prompt", e.HoldPrompt, EscalationVars}} {
		vs, err := refs(f.v)
		if err != nil {
			return fmt.Errorf("%s: %w", f.key, err)
		}
		for _, v := range vs {
			if !slices.Contains(f.vars, v) {
				return fmt.Errorf("%s uses %s; available: %v", f.key, placeholder(v), f.vars)
			}
		}
	}
	return nil
}

// FillStepTemplate fills a command or http step template of w. Earlier
// steps' outputs are the steps var: {{.steps.<id>.output}},
// {{.steps.<id>.json.sha}} (index for ids with a dash:
// {{(index .steps "wait-ci").output}}).
func (w *Workflow) FillStepTemplate(s string, vars map[string]any, steps map[string]any) (string, error) {
	all := maps.Clone(vars)
	all[StepsVar] = steps
	return Fill(s, all)
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
	for _, t := range templates {
		vs, err := refs(t)
		if err != nil {
			return err
		}
		for _, v := range vs {
			if _, ok := w.PackVars[v]; !ok && !slices.Contains(w.Vars, v) && v != StepsVar {
				return fmt.Errorf("{{.%s}} is not in workflow vars or pack vars", v)
			}
		}
	}
	return nil
}
