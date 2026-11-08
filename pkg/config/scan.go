package config

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

// DefaultScanMaxRuns is a scan trigger's max_runs when unset.
const DefaultScanMaxRuns = 3

// DefaultScanMaxOpen is a scan trigger's max_open when unset.
const DefaultScanMaxOpen = 3

// ScanVars are the template vars a scan trigger's run vars may use.
var ScanVars = []string{ScanSignalVar, ScanSubjectVar, ScanDetailVar}

// checkScanTrigger validates a when.scan trigger (#143). It has no other
// condition, starts a run, and its run vars and key may only use the
// scan vars. An empty key is <key_prefix>:<signal>:<subject>, key_prefix
// defaulting to finding (JOB-95); a set one (JOB-33, pr-<n>) lets the run
// dedup with runs started by hand for the subject.
// With on:, a check may use {{.event.<field>}} (JOB-72).
func (c *Config) checkScanTrigger(tr *Trigger) error {
	w := tr.When
	if w.Cron != "" || w.HasCounts() || w.Check != "" || w.Pool != "" || w.Kind != "" || w.Label != "" {
		return errors.New("when: finding takes no other condition")
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
	for i, s := range w.Scan {
		if !namePattern.MatchString(s.Signal) {
			return fmt.Errorf("when.scan[%d]: signal must match %s", i, namePattern)
		}
		if seen[s.Signal] {
			return fmt.Errorf("when.scan: signal %s listed twice", s.Signal)
		}
		seen[s.Signal] = true
		if s.Signal == SignalLinearIssueNoRun {
			if err := checkLinearSignal(s); err != nil {
				return err
			}
			continue
		}
		if s.Team != "" || s.States != "" || s.Labels != "" || s.SkipLabels != "" {
			return fmt.Errorf("when.scan: team, states, labels and skip_labels are only for %s", SignalLinearIssueNoRun)
		}
		builtin := slices.Contains(StoreSignals, s.Signal)
		switch {
		case builtin && (s.Check != "" || s.Script != ""):
			return fmt.Errorf("when.scan: %s is a store signal and takes no check or script", s.Signal)
		case !builtin && (s.Check == "") == (s.Script == ""):
			return fmt.Errorf("when.scan: %s is not a store signal %v and needs a check or a script", s.Signal, StoreSignals)
		case s.Script == "" && (s.Poll != "" || len(s.Args) > 0):
			return fmt.Errorf("when.scan: %s: poll and args are for script signals and %s", s.Signal, SignalLinearIssueNoRun)
		}
		if s.Poll != "" {
			if d, err := time.ParseDuration(s.Poll); err != nil || d <= 0 {
				return fmt.Errorf("when.scan: %s poll %q must be a positive duration", s.Signal, s.Poll)
			}
		}
		if len(tr.On) > 0 && s.Check != "" {
			if err := checkSignalCheckTemplate(s); err != nil {
				return err
			}
		}
		switch {
		case s.Signal == SignalIssueNoRun && !RepoPattern.MatchString(s.Repo):
			return fmt.Errorf("when.scan: %s needs repo (owner/name), got %q", s.Signal, s.Repo)
		case s.Signal != SignalIssueNoRun && s.Repo != "":
			return fmt.Errorf("when.scan: repo is only for %s", SignalIssueNoRun)
		}
		if s.After != "" {
			if d, err := time.ParseDuration(s.After); err != nil || d < 0 {
				return fmt.Errorf("when.scan: %s after %q must be a non-negative duration", s.Signal, s.After)
			}
		} else if builtin && s.Signal != SignalPoolQuarantined {
			return fmt.Errorf("when.scan: %s needs after", s.Signal)
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
		// The handler gets each finding as an event and returns its own
		// actions; args.workflow names the workflow whose tickets it skips.
		if tr.Ticket != nil || tr.Run != nil || tr.Command != nil {
			return errors.New("a scan trigger with handler takes no run, ticket or command")
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
		return errors.New("a scan trigger needs run and no ticket")
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
			if !slices.Contains(ScanVars, ref) {
				return fmt.Errorf("run: %s uses %s; only %v are available", k, placeholder(ref), ScanVars)
			}
		}
	}
	return nil
}

// checkSignalCheckTemplate checks that the check of a scan signal on a
// trigger with on: references only {{.event.<field>}}: it is rendered
// when the signal runs. Without on: it is run as written.
func checkSignalCheckTemplate(s ScanSignal) error {
	vs, err := refs(s.Check)
	if err != nil {
		return fmt.Errorf("when.scan: %s check: %w", s.Signal, err)
	}
	for _, v := range vs {
		if v != EventVar {
			return fmt.Errorf("when.scan: %s check uses %s; only %s is available", s.Signal, placeholder(v), placeholder(EventVar+".<field>"))
		}
	}
	return nil
}

// checkLinearSignal validates a linear_issue_no_run entry: the controller
// reads Linear itself, so it takes no check, repo or after.
func checkLinearSignal(s ScanSignal) error {
	if s.Script != "" || len(s.Args) > 0 {
		return fmt.Errorf("when.scan: %s takes no script or args", s.Signal)
	}
	if s.Check != "" || s.Repo != "" || s.After != "" {
		return fmt.Errorf("when.scan: %s takes team, states, labels, skip_labels, poll and grace, not check, repo or after", s.Signal)
	}
	if s.Team != "" && !linearTeamPattern.MatchString(s.Team) {
		return fmt.Errorf("when.scan: %s team %q is not a Linear team key like ENG", s.Signal, s.Team)
	}
	if s.Poll != "" {
		if d, err := time.ParseDuration(s.Poll); err != nil || d <= 0 {
			return fmt.Errorf("when.scan: %s poll %q must be a positive duration", s.Signal, s.Poll)
		}
	}
	if s.Grace != "" {
		if d, err := time.ParseDuration(s.Grace); err != nil || d < 0 {
			return fmt.Errorf("when.scan: %s grace %q must be a non-negative duration", s.Signal, s.Grace)
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

// DefaultScanKeyPrefix starts a scan run's key when run.key_prefix is
// unset.
const DefaultScanKeyPrefix = "scan"

// ScanKeyPrefix is tr.Run.KeyPrefix or the default.
func (tr Trigger) ScanKeyPrefix() string {
	if tr.Run != nil && tr.Run.KeyPrefix != "" {
		return tr.Run.KeyPrefix
	}
	return DefaultScanKeyPrefix
}

// ScanMaxRuns is tr.MaxRuns or the default.
func (tr Trigger) ScanMaxRuns() int {
	if tr.MaxRuns > 0 {
		return tr.MaxRuns
	}
	return DefaultScanMaxRuns
}

// ScanMaxOpen is tr.MaxOpen or the default.
func (tr Trigger) ScanMaxOpen() int {
	if tr.MaxOpen > 0 {
		return tr.MaxOpen
	}
	return DefaultScanMaxOpen
}

// ScanWorkflow is the workflow a scan trigger starts, whose own runs and
// steps it does not watch: run.workflow, or a handler's args.workflow.
func (tr Trigger) ScanWorkflow() string {
	if tr.Run != nil {
		return tr.Run.Workflow
	}
	w, _ := tr.Args["workflow"].(string)
	return w
}
