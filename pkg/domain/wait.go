package domain

import (
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"time"
)

// Waits (#381): a task or step parked until outside state holds. The
// wait lives in fields.wait; while it is set the ticket is out of the
// ready frontier. The controller evaluates the conditions and clears it.

// Wait condition types.
const (
	WaitReview = "github.review"
	WaitChecks = "github.checks"
	WaitPR     = "github.pr"
	WaitReply  = "thread.reply"
	WaitTime   = "time"
	// WaitLabel: an issue has a label, or is closed (#603). Repo, Issue
	// and Label are required.
	WaitLabel = "github.label"
	// WaitTag: a tag exists in Repo (JOB-11).
	WaitTag = "github.tag"
	// WaitDeploy: the deployed commit (the controller's own build) is SHA
	// or a descendant of it in Repo (JOB-11).
	WaitDeploy = "deploy"
	// WaitRun: run Run reached State done (finished) or closed (any end).
	WaitRun = "run"
	// WaitLinearState: Linear issue Thread is in state State (a name, or
	// a type such as completed).
	WaitLinearState = "linear.state"
	// WaitLinearLabel: Linear issue Thread has Label, or with Absent lacks it.
	WaitLinearLabel = "linear.label"
)

// Wait modes.
const (
	WaitAll = "all"
	WaitAny = "any"
)

// Wait is fields.wait on a task or step.
type Wait struct {
	Conds  []WaitCond `json:"conds"`
	Mode   string     `json:"mode,omitempty"` // all (default) or any
	Reason string     `json:"reason"`
	Owner  string     `json:"owner,omitempty"` // who must act
	// Resume pre-assigns the ticket to Session when the wait clears, so
	// the same conversation continues.
	Resume  bool      `json:"resume,omitempty"`
	Session string    `json:"session,omitempty"`
	Timeout time.Time `json:"timeout,omitzero"`
	// TimedOut is set once the timeout escalated (needs-human label).
	TimedOut bool      `json:"timed_out,omitempty"`
	Since    time.Time `json:"since"`
	By       string    `json:"by,omitempty"`
	// Checked is the last evaluation, for the poll interval.
	Checked time.Time `json:"checked,omitzero"`
}

// WaitCond is one condition. Which fields apply depends on Type.
type WaitCond struct {
	Type string `json:"type"`
	Repo string `json:"repo,omitempty"`
	PR   int    `json:"pr,omitempty"`
	// Issue and Label are for github.label.
	Issue int    `json:"issue,omitempty"`
	Label string `json:"label,omitempty"`
	// Absent: linear.label waits for the label to be gone.
	Absent bool `json:"absent,omitempty"`
	// Tag is for github.tag, Run for run.
	Tag string `json:"tag,omitempty"`
	Run string `json:"run,omitempty"`
	// Head is the PR head the condition is about. Empty means the head
	// seen on the first evaluation, which is then recorded.
	Head  string `json:"head,omitempty"`
	SHA   string `json:"sha,omitempty"`
	Login string `json:"login,omitempty"`
	// State: review state (default approved), checks success|failure, pr
	// merged|closed|head_changed.
	State    string `json:"state,omitempty"`
	Required bool   `json:"required,omitempty"`
	// Workflow is for github.checks: the name or file of a GitHub Actions
	// workflow. Set, the wait follows that workflow's newest run on the
	// commit instead of every check (JOB-75).
	Workflow string `json:"workflow,omitempty"`
	// Check is for github.checks: a check name, or a path.Match glob such
	// as "go (*)". Set, the wait follows only the checks of that name
	// (JOB-200). On a PR a check name is the job name.
	Check string `json:"check,omitempty"`
	// Thread is owner/name#n or a Linear key, for thread.reply.
	Thread string    `json:"thread,omitempty"`
	From   string    `json:"from,omitempty"`
	At     time.Time `json:"at,omitzero"`
	Met    bool      `json:"met,omitempty"`
	MetAt  time.Time `json:"met_at,omitzero"`
	Note   string    `json:"note,omitempty"`
}

var (
	repoRe   = regexp.MustCompile(`^[\w.-]+/[\w.-]+$`)
	threadRe = regexp.MustCompile(`^([\w.-]+/[\w.-]+#\d+|[A-Z][A-Z0-9]*-\d+)$`)
	linearRe = regexp.MustCompile(`^[A-Z][A-Z0-9]*-\d+$`)
	shaRe    = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
)

// Validate checks the wait's shape.
func (w *Wait) Validate() error {
	if w == nil {
		return nil
	}
	if len(w.Conds) == 0 {
		return fmt.Errorf("%w: wait: at least one condition", ErrInvalid)
	}
	if w.Mode != "" && w.Mode != WaitAll && w.Mode != WaitAny {
		return fmt.Errorf("%w: wait: mode must be all or any", ErrInvalid)
	}
	if w.Reason == "" {
		return fmt.Errorf("%w: wait: reason is required", ErrInvalid)
	}
	for i, c := range w.Conds {
		if err := c.validate(); err != nil {
			return fmt.Errorf("%w: wait condition %d: %w", ErrInvalid, i+1, err)
		}
	}
	return nil
}

func (c WaitCond) validate() error {
	if c.Workflow != "" && c.Type != WaitChecks {
		return fmt.Errorf("%s: workflow applies only to %s", c.Type, WaitChecks)
	}
	if c.Check != "" {
		if c.Type != WaitChecks {
			return fmt.Errorf("%s: check applies only to %s", c.Type, WaitChecks)
		}
		if c.Workflow != "" {
			return fmt.Errorf("%s check does not combine with workflow", c.Type)
		}
		if _, err := path.Match(c.Check, ""); err != nil {
			return fmt.Errorf("%s check %q: %w", c.Type, c.Check, err)
		}
	}
	needPR := func() error {
		if !repoRe.MatchString(c.Repo) || c.PR <= 0 {
			return fmt.Errorf("%s needs repo owner/name and pr", c.Type)
		}
		return nil
	}
	switch c.Type {
	case WaitReview:
		if c.Login == "" {
			return fmt.Errorf("%s needs login", c.Type)
		}
		return needPR()
	case WaitChecks:
		if c.State != "success" && c.State != "failure" {
			return fmt.Errorf("%s state must be success or failure", c.Type)
		}
		if c.Workflow != "" && c.Required {
			return fmt.Errorf("%s required does not apply with workflow", c.Type)
		}
		if c.PR == 0 && c.SHA != "" && repoRe.MatchString(c.Repo) {
			if c.Required {
				return fmt.Errorf("%s required needs pr", c.Type)
			}
			return nil
		}
		return needPR()
	case WaitPR:
		switch c.State {
		case "merged", "closed", "head_changed":
		default:
			return fmt.Errorf("%s state must be merged, closed or head_changed", c.Type)
		}
		return needPR()
	case WaitReply:
		if !threadRe.MatchString(c.Thread) {
			return fmt.Errorf("%s thread must be owner/name#n or a Linear key", c.Type)
		}
		return nil
	case WaitTime:
		if c.At.IsZero() {
			return fmt.Errorf("%s needs at", c.Type)
		}
		return nil
	case WaitLabel:
		if !repoRe.MatchString(c.Repo) || c.Issue <= 0 || c.Label == "" {
			return fmt.Errorf("%s needs repo owner/name, issue and label", c.Type)
		}
		return nil
	case WaitTag:
		if !repoRe.MatchString(c.Repo) || c.Tag == "" {
			return fmt.Errorf("%s needs repo owner/name and tag", c.Type)
		}
		return nil
	case WaitDeploy:
		if !repoRe.MatchString(c.Repo) || !shaRe.MatchString(c.SHA) {
			return fmt.Errorf("%s needs repo owner/name and a hex sha", c.Type)
		}
		return nil
	case WaitRun:
		if c.Run == "" || c.State != "done" && c.State != "closed" {
			return fmt.Errorf("%s needs id and state done or closed", c.Type)
		}
		return nil
	case WaitLinearState:
		if !linearRe.MatchString(c.Thread) || c.State == "" {
			return fmt.Errorf("%s needs a Linear issue key and state", c.Type)
		}
		return nil
	case WaitLinearLabel:
		if !linearRe.MatchString(c.Thread) || c.Label == "" {
			return fmt.Errorf("%s needs a Linear issue key and label", c.Type)
		}
		return nil
	}
	return fmt.Errorf("unknown type %q", c.Type)
}

// Satisfied reports whether the met conditions satisfy the mode.
func (w *Wait) Satisfied() bool {
	n := 0
	for _, c := range w.Conds {
		if c.Met {
			n++
		}
	}
	if w.Mode == WaitAny {
		return n > 0
	}
	return n == len(w.Conds)
}

// WaitOf returns fields.wait of a task or step, nil for none.
func WaitOf(t Ticket) *Wait {
	if t.Kind != KindTask && t.Kind != KindStep || len(t.Fields) == 0 {
		return nil
	}
	var f struct {
		Wait *Wait `json:"wait"`
	}
	if json.Unmarshal(t.Fields, &f) != nil {
		return nil
	}
	return f.Wait
}
