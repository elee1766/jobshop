package waitengine

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/domain"
)

// Targeted wait wake (#679). An event wakes only the waits it can
// change: an inbox item names a repo and number or a Linear issue; a run
// end names the run. The pump evaluates just those waits, so a CI
// webhook costs one GitHub read for the PR it names instead of a pass
// over every wait. A full pass still runs on the interval, and at the
// earliest time condition, as the reconcile backstop.

// Subject keys are lowercase: GitHub repo names compare case-insensitive.

// RepoSubject names every wait on repo.
func RepoSubject(repo string) string { return "repo:" + strings.ToLower(repo) }

// IssueSubject names the waits on one GitHub issue or PR.
func IssueSubject(repo string, n int) string {
	return fmt.Sprintf("gh:%s#%d", strings.ToLower(repo), n)
}

// LinearSubject names the waits on one Linear issue.
func LinearSubject(key string) string { return "linear:" + strings.ToUpper(key) }

// RunSubject names the waits on one run.
func RunSubject(id string) string { return "run:" + id }

// CondSubjects are the subjects whose events can change cd. Checks on
// a PR also match repo-wide events: status events and some check runs
// name only the sha. Deploy has none: the startup pass (a deploy
// restarts the controller) and the interval cover it; time conditions
// set the pump's timer.
func CondSubjects(cd domain.WaitCond) []string {
	switch cd.Type {
	case domain.WaitReview, domain.WaitPR:
		return []string{IssueSubject(cd.Repo, cd.PR)}
	case domain.WaitChecks:
		if cd.PR == 0 {
			return []string{RepoSubject(cd.Repo)}
		}
		return []string{IssueSubject(cd.Repo, cd.PR), RepoSubject(cd.Repo)}
	case domain.WaitLabel:
		return []string{IssueSubject(cd.Repo, cd.Issue)}
	case domain.WaitTag:
		return []string{RepoSubject(cd.Repo)}
	case domain.WaitRun:
		return []string{RunSubject(cd.Run)}
	case domain.WaitLinearState, domain.WaitLinearLabel:
		return []string{LinearSubject(cd.Thread)}
	case domain.WaitReply:
		if repo, num, ok := strings.Cut(cd.Thread, "#"); ok {
			var n int
			_, _ = fmt.Sscan(num, &n)
			return []string{IssueSubject(repo, n)}
		}
		return []string{LinearSubject(cd.Thread)}
	}
	return nil
}

// ItemSubjects are the subjects an inbox item names: repo#number, or
// the repo alone for an event with no number (status, push, a check run
// on no PR), or a Linear key.
func ItemSubjects(it domain.InboxItem) []string {
	if it.RepoOrTeam == "" {
		return nil
	}
	if it.Source == "linear" {
		if it.Number == 0 {
			return nil
		}
		return []string{LinearSubject(fmt.Sprintf("%s-%d", it.RepoOrTeam, it.Number))}
	}
	if it.Number > 0 {
		return []string{IssueSubject(it.RepoOrTeam, int(it.Number))}
	}
	return []string{RepoSubject(it.RepoOrTeam)}
}

// Names reports whether an unmet condition of w is on one of subjects.
func Names(w *domain.Wait, subjects map[string]bool) bool {
	if w == nil {
		return false
	}
	for _, cd := range w.Conds {
		if cd.Met {
			continue
		}
		for _, s := range CondSubjects(cd) {
			if subjects[s] {
				return true
			}
		}
	}
	return false
}

// Pending is what the pump owes: a full pass, or the waits whose
// subjects an event named.
type Pending struct {
	All      bool
	Subjects map[string]bool
}

// Wakes accumulates wake requests between pump passes. The zero value is
// ready. The pump calls Take before it lists the waits, so a request that
// arrives during a pass stays queued for the next one: a wake is never
// lost to an observation the pass had already cached.
type Wakes struct {
	mu      sync.Mutex
	pending Pending
}

// All asks for a full pass.
func (w *Wakes) All() {
	w.mu.Lock()
	w.pending.All = true
	w.mu.Unlock()
}

// Subjects asks for the waits on subjects. It reports whether it queued
// anything.
func (w *Wakes) Subjects(subjects ...string) bool {
	if len(subjects) == 0 {
		return false
	}
	w.mu.Lock()
	if w.pending.Subjects == nil {
		w.pending.Subjects = map[string]bool{}
	}
	for _, s := range subjects {
		w.pending.Subjects[s] = true
	}
	w.mu.Unlock()
	return true
}

// Items queues the waits new inbox items name. An item with no subject (a
// generic endpoint without a map) queues a full pass: it may still be
// what a wait is on. It reports whether it queued anything.
func (w *Wakes) Items(items []domain.InboxItem) bool {
	var subj []string
	for _, it := range items {
		s := ItemSubjects(it)
		if len(s) == 0 {
			w.All()
			return true
		}
		subj = append(subj, s...)
	}
	return w.Subjects(subj...)
}

// Take returns and clears what the pump owes.
func (w *Wakes) Take() Pending {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := w.pending
	w.pending = Pending{}
	return p
}

// NextDue is how long until the earliest unmet time condition or wait
// timeout, capped at every.
func NextDue(waits []domain.Wait, now time.Time, every time.Duration) time.Duration {
	d := every
	at := func(t time.Time) {
		if !t.IsZero() {
			d = min(d, max(t.Sub(now), 0))
		}
	}
	for _, w := range waits {
		if !w.TimedOut {
			at(w.Timeout)
		}
		for _, cd := range w.Conds {
			if cd.Type == domain.WaitTime && !cd.Met {
				at(cd.At)
			}
		}
	}
	return d
}
