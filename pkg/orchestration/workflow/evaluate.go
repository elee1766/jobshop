package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/integrations/github"
)

// Decision is the outcome of evaluating one wait against a Pass. It is
// advice computed from the wait snapshot passed to Evaluate: the engine
// writes nothing. The controller applies it with a revision-checked
// update on the ticket the snapshot came from, so a decision made from a
// stale observation or snapshot loses to a newer write (store.ErrConflict)
// and the next pass or wake evaluates again.
type Decision struct {
	// Clear: end the wait. Why is the reason for the ticket comment,
	// "condition met: ..." or "invalidated: ...".
	Clear bool
	Why   string
	// Wait is the snapshot with the evaluation applied: met flags and
	// notes, PR heads bound on first sight, TimedOut set. Unset when
	// Clear.
	Wait domain.Wait
	// Save: Wait differs from the snapshot and must be persisted.
	Save bool
	// TimedOut: the timeout passed in this evaluation and had not been
	// escalated yet; the controller escalates once.
	TimedOut bool
	// Errs are condition errors (a failed provider read). The wait keeps
	// its other conditions' progress.
	Errs []error
}

// invalidated means a condition can no longer hold as stated.
type invalidated struct{ why string }

func (e invalidated) Error() string { return e.why }

// Evaluate evaluates every unmet condition of w at now. It does not
// modify w.
func Evaluate(ctx context.Context, pass *Pass, w domain.Wait, now time.Time) Decision {
	before, _ := json.Marshal(w)
	w.Conds = slices.Clone(w.Conds)
	var errs []error
	for i := range w.Conds {
		cd := &w.Conds[i]
		if cd.Met {
			continue
		}
		met, note, err := evalCond(ctx, pass, &w, cd, now)
		var inv invalidated
		if errors.As(err, &inv) {
			return Decision{Clear: true, Why: "invalidated: " + inv.why}
		}
		if err != nil {
			errs = append(errs, err)
			continue
		}
		cd.Note = note
		if met {
			cd.Met, cd.MetAt = true, now
		}
	}
	if w.Satisfied() {
		var notes []string
		for _, cd := range w.Conds {
			if cd.Met && cd.Note != "" {
				notes = append(notes, cd.Note)
			}
		}
		return Decision{Clear: true, Why: "condition met: " + strings.Join(notes, "; ")}
	}
	d := Decision{Errs: errs}
	d.TimedOut = !w.Timeout.IsZero() && !now.Before(w.Timeout) && !w.TimedOut
	if d.TimedOut {
		w.TimedOut = true
	}
	after, _ := json.Marshal(w)
	d.Save = !bytes.Equal(before, after)
	d.Wait = w
	return d
}

// evalCond reports whether cd holds now. It records cd.Head the first
// time a PR condition is evaluated without one.
func evalCond(ctx context.Context, pass *Pass, w *domain.Wait, cd *domain.WaitCond, now time.Time) (bool, string, error) {
	switch cd.Type {
	case domain.WaitTime:
		return !now.Before(cd.At), "time " + cd.At.Format(time.RFC3339), nil
	case domain.WaitReply:
		return pass.src.Replies.Reply(ctx, *cd, w.Since)
	case domain.WaitLabel:
		return evalLabel(ctx, pass, cd)
	case domain.WaitTag:
		ok, err := pass.src.GitHub.Tag(ctx, cd.Repo, cd.Tag)
		// The note is stored on every pass, met or not (#716).
		if !ok {
			return false, fmt.Sprintf("tag %s not pushed to %s yet", cd.Tag, cd.Repo), err
		}
		return true, fmt.Sprintf("tag %s pushed to %s", cd.Tag, cd.Repo), err
	case domain.WaitDeploy:
		if pass.src.Commit == "" {
			return false, "", invalidated{"this controller reports no build commit"}
		}
		ok, err := pass.src.GitHub.Contains(ctx, cd.Repo, cd.SHA, pass.src.Commit)
		if !ok {
			return false, fmt.Sprintf("deployed %s does not contain %s yet", short(pass.src.Commit), short(cd.SHA)), err
		}
		return true, fmt.Sprintf("deployed %s contains %s", short(pass.src.Commit), short(cd.SHA)), err
	case domain.WaitRun:
		return evalRun(ctx, pass, cd)
	case domain.WaitLinearState, domain.WaitLinearLabel:
		return evalLinear(ctx, pass, cd)
	case domain.WaitChecks:
		if cd.PR == 0 {
			return evalChecks(ctx, pass, cd, cd.SHA)
		}
	}
	pr, err := pass.pr(ctx, cd.Repo, cd.PR)
	if err != nil {
		return false, "", err
	}
	ref := PRKey(cd.Repo, cd.PR)
	if cd.Head == "" {
		cd.Head = pr.Head
	}
	if cd.Type == domain.WaitPR {
		switch cd.State {
		case "merged":
			if pr.State == "closed" && !pr.Merged {
				return false, "", invalidated{ref + " closed without merging"}
			}
			return pr.Merged, ref + " merged", nil
		case "closed":
			return pr.State == "closed", ref + " closed", nil
		}
		return pr.Head != cd.Head || pr.State == "closed", fmt.Sprintf("%s head moved %s -> %s", ref, short(cd.Head), short(pr.Head)), nil
	}
	if pr.Merged {
		return false, "", invalidated{ref + " merged"}
	}
	if pr.State == "closed" {
		return false, "", invalidated{ref + " closed"}
	}
	if pr.Head != cd.Head {
		return false, "", invalidated{fmt.Sprintf("%s head moved %s -> %s", ref, short(cd.Head), short(pr.Head))}
	}
	if cd.Type == domain.WaitChecks {
		if pr.Conflicting {
			return false, "", invalidated{ref + " conflicts with its base; CI will not run until it is rebased"}
		}
		return evalChecks(ctx, pass, cd, pr.Head)
	}
	rs, err := pass.reviewsOf(ctx, cd.Repo, cd.PR)
	if err != nil {
		return false, "", err
	}
	want := strings.ToUpper(cd.State)
	if want == "" {
		want = "APPROVED"
	}
	// The login's latest review on the exact head decides.
	var last *github.Review
	for i := range rs {
		r := &rs[i]
		if !strings.EqualFold(r.Login, cd.Login) || r.Commit != cd.Head {
			continue
		}
		if last == nil || !r.At.Before(last.At) {
			last = r
		}
	}
	observed := "no review"
	if last != nil {
		observed = last.State
	}
	return last != nil && last.State == want, fmt.Sprintf("%s wants %s by %s on %s, observed %s", ref, want, cd.Login, short(cd.Head), observed), nil
}

// evalRun: the run is done (state done) or ended any way (closed). A
// done wait on a run that was cancelled can never hold.
func evalRun(ctx context.Context, pass *Pass, cd *domain.WaitCond) (bool, string, error) {
	r, err := pass.src.Runs.Run(ctx, cd.Run)
	if err != nil {
		return false, "", err
	}
	if cd.State == "closed" {
		return r.Ended(), fmt.Sprintf("run %s ended (%s)", cd.Run, r.State), nil
	}
	if r.State == domain.RunCancelled {
		return false, "", invalidated{"run " + cd.Run + " was cancelled"}
	}
	return r.State == domain.RunDone, "run " + cd.Run + " done", nil
}

// evalLinear: linear.state matches the state name or type
// (case-insensitive); linear.label has the label, or lacks it with Absent.
func evalLinear(ctx context.Context, pass *Pass, cd *domain.WaitCond) (bool, string, error) {
	is, err := pass.linearIssue(ctx, cd.Thread)
	if err != nil {
		return false, "", err
	}
	if cd.Type == domain.WaitLinearState {
		ok := strings.EqualFold(is.State.Name, cd.State) || strings.EqualFold(is.State.Type, cd.State)
		return ok, fmt.Sprintf("%s is %s", cd.Thread, is.State.Name), nil
	}
	has := slices.ContainsFunc(is.Labels, func(l string) bool { return strings.EqualFold(l, cd.Label) })
	if cd.Absent {
		return !has, fmt.Sprintf("%s no longer labeled %s", cd.Thread, cd.Label), nil
	}
	return has, fmt.Sprintf("%s labeled %s", cd.Thread, cd.Label), nil
}

// evalLabel: the issue carries the label, or is closed (a rejected
// proposal also ends the wait).
func evalLabel(ctx context.Context, pass *Pass, cd *domain.WaitCond) (bool, string, error) {
	is, err := pass.src.GitHub.Issue(ctx, cd.Repo, cd.Issue)
	if err != nil {
		return false, "", err
	}
	ref := fmt.Sprintf("%s#%d", cd.Repo, cd.Issue)
	if is.State == "closed" {
		return true, ref + " closed", nil
	}
	return slices.ContainsFunc(is.Labels, func(l string) bool { return strings.EqualFold(l, cd.Label) }), ref + " labeled " + cd.Label, nil
}

// evalChecks: success needs at least one check and none pending or
// failed; failure is any failed check.
//
// A cancelled check is not a failure (#831, JOB-75): a newer push or a
// concurrency group superseded the run, and the agent has nothing to fix.
// Once nothing is pending or failed and something was cancelled, the head
// will never go green, so the wait is invalidated with that reason and
// the agent decides again. Skipped checks count for neither side. On a
// commit success needs at least one passed check, so the release
// workflow's skipped jobs alone are not green.
//
// With cd.Workflow set only that workflow's newest run on the commit
// counts (evalWorkflow).
func evalChecks(ctx context.Context, pass *Pass, cd *domain.WaitCond, head string) (bool, string, error) {
	if cd.Workflow != "" {
		return evalWorkflow(ctx, pass, cd, head)
	}
	cs, err := pass.checksOf(ctx, cd.Repo, cd.PR, head, cd.Required)
	if err != nil {
		return false, "", err
	}
	what := "checks"
	if cd.Required {
		what = "required checks"
	}
	at := short(head)
	if cd.Check != "" {
		what = cd.Check
		cs = slices.DeleteFunc(slices.Clone(cs), func(ch github.Check) bool { return !checkMatches(cd.Check, ch.Name) })
		if len(cs) == 0 {
			return false, fmt.Sprintf("%s not reported on %s", cd.Check, at), nil
		}
	}
	commit := cd.PR == 0
	var failed, pending, cancelled []string
	passed := 0
	for _, ch := range cs {
		switch ch.Bucket {
		case "cancel":
			cancelled = append(cancelled, ch.Name)
		case "fail":
			failed = append(failed, ch.Name)
		case "pending":
			pending = append(pending, ch.Name)
		case "pass":
			passed++
		}
	}
	if len(cancelled) > 0 && len(failed) == 0 && len(pending) == 0 {
		return false, "", invalidated{fmt.Sprintf("%s cancelled on %s: %s; this head will not go green, push or wait on a newer commit", what, at, strings.Join(cancelled, ", "))}
	}
	if cd.State == "failure" {
		return len(failed) > 0, fmt.Sprintf("%s failed on %s: %s", what, at, strings.Join(failed, ", ")), nil
	}
	green := len(cs) > 0 && len(failed) == 0 && len(pending) == 0 && len(cancelled) == 0
	if commit {
		green = green && passed > 0
	}
	return green, fmt.Sprintf("%s green on %s", what, at), nil
}

// checkMatches: a check name equals pattern, or matches it as a
// path.Match glob (`go (*)` covers a matrix). A bad pattern only matches
// by equality; domain validation rejects it before it is stored.
func checkMatches(pattern, name string) bool {
	if pattern == name {
		return true
	}
	ok, err := path.Match(pattern, name)
	return err == nil && ok
}

// evalWorkflow follows the newest run of cd.Workflow on head:
// none yet or in progress is not met, success meets state=success,
// failure meets state=failure, and a cancelled or skipped run can never
// become green, so the wait is invalidated and the agent resumes now.
func evalWorkflow(ctx context.Context, pass *Pass, cd *domain.WaitCond, head string) (bool, string, error) {
	rs, err := pass.runsOf(ctx, cd.Repo, head)
	if err != nil {
		return false, "", err
	}
	at := short(head)
	run, ok := github.NewestRun(rs, cd.Workflow)
	if !ok {
		return false, fmt.Sprintf("%s not started on %s", cd.Workflow, at), nil
	}
	switch github.CheckBucket(run.Status, run.Conclusion) {
	case "pending":
		return false, fmt.Sprintf("%s in progress on %s", cd.Workflow, at), nil
	case "pass":
		return cd.State == "success", fmt.Sprintf("%s green on %s", cd.Workflow, at), nil
	case "fail":
		return cd.State == "failure", fmt.Sprintf("%s failed on %s", cd.Workflow, at), nil
	case "cancel":
		return false, "", invalidated{fmt.Sprintf("%s cancelled on %s; this head will not go green, push or wait on a newer commit", cd.Workflow, at)}
	}
	return false, "", invalidated{fmt.Sprintf("%s skipped on %s", cd.Workflow, at)}
}

func short(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
