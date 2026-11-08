package config

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

// DefaultStallMaxRuns is a stall trigger's max_runs when unset.
const DefaultStallMaxRuns = 3

// DefaultStallMaxOpen is a stall trigger's max_open when unset.
const DefaultStallMaxOpen = 3

// StallVars are the template vars a stall trigger's run vars may use.
var StallVars = []string{StallSignalVar, StallSubjectVar, StallDetailVar}

// checkStallTrigger validates a when.stall trigger (#143). It has no other
// condition, starts a run, and its run vars and key may only use the
// stall vars. An empty key is <key_prefix>:<signal>:<subject>, key_prefix
// defaulting to stall (JOB-95); a set one (JOB-33, pr-<n>) lets the run
// dedup with runs started by hand for the subject.
// With on:, a check may use {{.event.<field>}} (JOB-72).
func (c *Config) checkStallTrigger(tr *Trigger) error {
	w := tr.When
	if w.Cron != "" || w.HasCounts() || w.Check != "" || w.Pool != "" || w.Kind != "" || w.Label != "" {
		return errors.New("when: stall takes no other condition")
	}
	if err := checkTriggerOn(tr); err != nil {
		return err
	}
	if w.CheckTimeout != "" {
		if d, err := time.ParseDuration(w.CheckTimeout); err != nil || d <= 0 {
			return fmt.Errorf("when: check_timeout %q must be a positive duration", w.CheckTimeout)
		}
	}
	seen := map[string]bool{}
	for i, s := range w.Stall {
		if !namePattern.MatchString(s.Signal) {
			return fmt.Errorf("when.stall[%d]: signal must match %s", i, namePattern)
		}
		if seen[s.Signal] {
			return fmt.Errorf("when.stall: signal %s listed twice", s.Signal)
		}
		seen[s.Signal] = true
		if s.Signal == StallLinearIssueNoRun {
			if err := checkLinearStall(s); err != nil {
				return err
			}
			continue
		}
		if s.Team != "" || s.States != "" || s.Labels != "" || s.SkipLabels != "" || s.Poll != "" {
			return fmt.Errorf("when.stall: team, states, labels, skip_labels and poll are only for %s", StallLinearIssueNoRun)
		}
		builtin := slices.Contains(StoreStallSignals, s.Signal)
		switch {
		case builtin && s.Check != "":
			return fmt.Errorf("when.stall: %s is a store signal and takes no check", s.Signal)
		case !builtin && s.Check == "":
			return fmt.Errorf("when.stall: %s is not a store signal %v and needs a check", s.Signal, StoreStallSignals)
		}
		if len(tr.On) > 0 {
			if err := checkStallCheckTemplate(s); err != nil {
				return err
			}
		}
		switch {
		case s.Signal == StallIssueNoRun && !RepoPattern.MatchString(s.Repo):
			return fmt.Errorf("when.stall: %s needs repo (owner/name), got %q", s.Signal, s.Repo)
		case s.Signal != StallIssueNoRun && s.Repo != "":
			return fmt.Errorf("when.stall: repo is only for %s", StallIssueNoRun)
		}
		if s.After != "" {
			if d, err := time.ParseDuration(s.After); err != nil || d < 0 {
				return fmt.Errorf("when.stall: %s after %q must be a non-negative duration", s.Signal, s.After)
			}
		} else if builtin && s.Signal != StallPoolQuarantined {
			return fmt.Errorf("when.stall: %s needs after", s.Signal)
		}
	}
	if tr.Cooldown != "" {
		if d, err := time.ParseDuration(tr.Cooldown); err != nil || d < 0 {
			return fmt.Errorf("cooldown %q must be a non-negative duration", tr.Cooldown)
		}
	}
	if tr.MaxRuns < 0 {
		return errors.New("max_runs must be >= 0")
	}
	if tr.MaxOpen < 0 {
		return errors.New("max_open must be >= 0")
	}
	if tr.Handler != "" {
		// The handler gets each stall as an event and returns its own
		// actions; args.workflow names the workflow whose tickets it skips.
		if tr.Ticket != nil || tr.Run != nil || tr.Command != nil {
			return errors.New("a stall trigger with handler takes no run, ticket or command")
		}
		if w, ok := tr.Args["workflow"]; ok {
			name, isString := w.(string)
			if _, known := c.Workflows[name]; !isString || !known {
				return fmt.Errorf("args.workflow: unknown workflow %v", w)
			}
		}
		return nil
	}
	if tr.Ticket != nil || tr.Run == nil {
		return errors.New("a stall trigger needs run and no ticket")
	}
	r := tr.Run
	if r.KeyPrefix != "" {
		if !namePattern.MatchString(r.KeyPrefix) {
			return fmt.Errorf("run: key_prefix must match %s", namePattern)
		}
		if r.Key != "" {
			return errors.New("run: set key or key_prefix, not both")
		}
	}
	wf, ok := c.Workflows[r.Workflow]
	if !ok {
		return fmt.Errorf("run: unknown workflow %q", r.Workflow)
	}
	for k := range r.Vars {
		if !slices.Contains(wf.Vars, k) {
			return fmt.Errorf("run: workflow %s has no var %q", r.Workflow, k)
		}
	}
	for _, k := range wf.Vars {
		_, set := r.Vars[k]
		_, packed := wf.PackVars[k]
		if !set && !packed {
			return fmt.Errorf("run: workflow %s needs var %s", r.Workflow, k)
		}
	}
	fields := map[string]string{"key": r.Key}
	for k, v := range r.Vars {
		fields["vars."+k] = v
	}
	for k, v := range fields {
		vs, err := refs(v)
		if err != nil {
			return fmt.Errorf("run: %s: %w", k, err)
		}
		for _, ref := range vs {
			if !slices.Contains(StallVars, ref) {
				return fmt.Errorf("run: %s uses %s; only %v are available", k, placeholder(ref), StallVars)
			}
		}
	}
	return nil
}

// checkStallCheckTemplate checks that the check of a stall signal on a
// trigger with on: references only {{.event.<field>}}: it is rendered
// when the signal runs. Without on: it is run as written.
func checkStallCheckTemplate(s StallSignal) error {
	vs, err := refs(s.Check)
	if err != nil {
		return fmt.Errorf("when.stall: %s check: %w", s.Signal, err)
	}
	for _, v := range vs {
		if v != EventVar {
			return fmt.Errorf("when.stall: %s check uses %s; only %s is available", s.Signal, placeholder(v), placeholder(EventVar+".<field>"))
		}
	}
	return nil
}

// checkLinearStall validates a linear_issue_no_run entry: the controller
// reads Linear itself, so it takes no check, repo or after.
func checkLinearStall(s StallSignal) error {
	if s.Check != "" || s.Repo != "" || s.After != "" {
		return fmt.Errorf("when.stall: %s takes team, states, labels, skip_labels, poll and grace, not check, repo or after", s.Signal)
	}
	if s.Team != "" && !linearTeamPattern.MatchString(s.Team) {
		return fmt.Errorf("when.stall: %s team %q is not a Linear team key like ENG", s.Signal, s.Team)
	}
	if s.Poll != "" {
		if d, err := time.ParseDuration(s.Poll); err != nil || d <= 0 {
			return fmt.Errorf("when.stall: %s poll %q must be a positive duration", s.Signal, s.Poll)
		}
	}
	if s.Grace != "" {
		if d, err := time.ParseDuration(s.Grace); err != nil || d < 0 {
			return fmt.Errorf("when.stall: %s grace %q must be a non-negative duration", s.Signal, s.Grace)
		}
	}
	return nil
}

var linearTeamPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]*$`)

// SplitList splits a comma-separated var into trimmed, non-empty names.
func SplitList(s string) []string {
	var out []string
	for f := range strings.SplitSeq(s, ",") {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}

// DefaultStallKeyPrefix starts a stall run's key when run.key_prefix is
// unset.
const DefaultStallKeyPrefix = "stall"

// StallKeyPrefix is tr.Run.KeyPrefix or the default.
func (tr Trigger) StallKeyPrefix() string {
	if tr.Run != nil && tr.Run.KeyPrefix != "" {
		return tr.Run.KeyPrefix
	}
	return DefaultStallKeyPrefix
}

// StallMaxRuns is tr.MaxRuns or the default.
func (tr Trigger) StallMaxRuns() int {
	if tr.MaxRuns > 0 {
		return tr.MaxRuns
	}
	return DefaultStallMaxRuns
}

// StallMaxOpen is tr.MaxOpen or the default.
func (tr Trigger) StallMaxOpen() int {
	if tr.MaxOpen > 0 {
		return tr.MaxOpen
	}
	return DefaultStallMaxOpen
}

// StallWorkflow is the workflow a stall trigger starts, whose own runs and
// steps it does not watch: run.workflow, or a handler's args.workflow.
func (tr Trigger) StallWorkflow() string {
	if tr.Run != nil {
		return tr.Run.Workflow
	}
	w, _ := tr.Args["workflow"].(string)
	return w
}
