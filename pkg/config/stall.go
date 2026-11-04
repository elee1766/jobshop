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

// StallVars are the template vars a stall trigger's run vars may use.
var StallVars = []string{StallSignalVar, StallSubjectVar, StallDetailVar}

// checkStallTrigger validates a when.stall trigger (#143). It has no other
// condition, starts a run, sets the run key itself, and its run vars may
// only use the stall vars.
func (c *Config) checkStallTrigger(tr *Trigger) error {
	w := tr.When
	if w.Cron != "" || w.HasCounts() || w.Check != "" || w.Pool != "" || w.Kind != "" || w.Label != "" {
		return errors.New("when: stall takes no other condition")
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
	if tr.Ticket != nil || tr.Run == nil {
		return errors.New("a stall trigger needs run and no ticket")
	}
	r := tr.Run
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
	if r.Key != "" {
		return errors.New("run: a stall trigger sets the key itself (stall:<signal>:<subject>)")
	}
	for k, v := range r.Vars {
		vs, err := refs(v, r.Template)
		if err != nil {
			return fmt.Errorf("run: vars.%s: %w", k, err)
		}
		for _, ref := range vs {
			if !slices.Contains(StallVars, ref) {
				return fmt.Errorf("run: vars.%s uses %s; only %v are available", k, placeholder(ref, r.Template), StallVars)
			}
		}
	}
	return nil
}

// checkLinearStall validates a linear_issue_no_run entry: the controller
// reads Linear itself, so it takes no check, repo or after.
func checkLinearStall(s StallSignal) error {
	if s.Check != "" || s.Repo != "" || s.After != "" {
		return fmt.Errorf("when.stall: %s takes team, states, labels, skip_labels and poll, not check, repo or after", s.Signal)
	}
	if s.Team != "" && !linearTeamPattern.MatchString(s.Team) {
		return fmt.Errorf("when.stall: %s team %q is not a Linear team key like ENG", s.Signal, s.Team)
	}
	if s.Poll != "" {
		if d, err := time.ParseDuration(s.Poll); err != nil || d <= 0 {
			return fmt.Errorf("when.stall: %s poll %q must be a positive duration", s.Signal, s.Poll)
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

// StallMaxRuns is tr.MaxRuns or the default.
func (tr Trigger) StallMaxRuns() int {
	if tr.MaxRuns > 0 {
		return tr.MaxRuns
	}
	return DefaultStallMaxRuns
}
