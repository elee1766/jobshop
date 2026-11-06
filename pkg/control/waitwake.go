package control

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/store"
)

// Targeted wait wake (#679). An event wakes only the waits it can
// change: an inbox item names a repo and number or a Linear issue; a run
// end names the run. The pump evaluates just those waits, so a CI
// webhook costs one GitHub read for the PR it names instead of a pass
// over every wait. A full pass still runs on the interval, and at the
// earliest time condition, as the reconcile backstop.

// waitWakes is what the wait pump owes: a full pass, or the waits whose
// subjects an event named.
type waitWakes struct {
	all  bool
	subj map[string]bool
}

// waitSubject keys. Lowercase: GitHub repo names compare case-insensitive.
func repoSubject(repo string) string { return "repo:" + strings.ToLower(repo) }

func issueSubject(repo string, n int) string {
	return fmt.Sprintf("gh:%s#%d", strings.ToLower(repo), n)
}
func linearSubject(key string) string { return "linear:" + strings.ToUpper(key) }
func runSubject(id string) string     { return "run:" + id }

// condSubjects are the subjects whose events can change cd. Checks on
// a PR also match repo-wide events: status events and some check runs
// name only the sha. Deploy has none: the startup pass (a deploy
// restarts the controller) and the interval cover it; time conditions
// set the pump's timer.
func condSubjects(cd store.WaitCond) []string {
	switch cd.Type {
	case store.WaitReview, store.WaitPR:
		return []string{issueSubject(cd.Repo, cd.PR)}
	case store.WaitChecks:
		if cd.PR == 0 {
			return []string{repoSubject(cd.Repo)}
		}
		return []string{issueSubject(cd.Repo, cd.PR), repoSubject(cd.Repo)}
	case store.WaitLabel:
		return []string{issueSubject(cd.Repo, cd.Issue)}
	case store.WaitTag:
		return []string{repoSubject(cd.Repo)}
	case store.WaitRun:
		return []string{runSubject(cd.Run)}
	case store.WaitLinearState, store.WaitLinearLabel:
		return []string{linearSubject(cd.Thread)}
	case store.WaitReply:
		if repo, num, ok := strings.Cut(cd.Thread, "#"); ok {
			var n int
			_, _ = fmt.Sscan(num, &n)
			return []string{issueSubject(repo, n)}
		}
		return []string{linearSubject(cd.Thread)}
	}
	return nil
}

// itemSubjects are the subjects an inbox item names: repo#number, or
// the repo alone for an event with no number (status, push, a check run
// on no PR), or a Linear key.
func itemSubjects(it store.InboxItem) []string {
	if it.RepoOrTeam == "" {
		return nil
	}
	if it.Source == "linear" {
		if it.Number == 0 {
			return nil
		}
		return []string{linearSubject(fmt.Sprintf("%s-%d", it.RepoOrTeam, it.Number))}
	}
	if it.Number > 0 {
		return []string{issueSubject(it.RepoOrTeam, int(it.Number))}
	}
	return []string{repoSubject(it.RepoOrTeam)}
}

// wakeWaits asks the wait pump for a full pass now.
func (c *Controller) wakeWaits() {
	c.waitWakeMu.Lock()
	c.waitPending.all = true
	c.waitWakeMu.Unlock()
	c.signalWaits()
}

// wakeWaitsSubjects asks the pump to evaluate the waits on subjects.
func (c *Controller) wakeWaitsSubjects(subjects ...string) {
	if len(subjects) == 0 {
		return
	}
	c.waitWakeMu.Lock()
	if c.waitPending.subj == nil {
		c.waitPending.subj = map[string]bool{}
	}
	for _, s := range subjects {
		c.waitPending.subj[s] = true
	}
	c.waitWakeMu.Unlock()
	c.signalWaits()
}

// wakeWaitsFor wakes the waits new inbox items name. An item with no
// subject (a generic endpoint without a map) wakes a full pass: it may
// still be what a wait is on.
func (c *Controller) wakeWaitsFor(items []store.InboxItem) {
	var subj []string
	for _, it := range items {
		s := itemSubjects(it)
		if len(s) == 0 {
			c.wakeWaits()
			return
		}
		subj = append(subj, s...)
	}
	c.wakeWaitsSubjects(subj...)
}

func (c *Controller) signalWaits() {
	if c.waitWake == nil {
		return
	}
	select {
	case c.waitWake <- struct{}{}:
	default:
	}
}

// takeWaitWakes returns and clears what the pump owes.
func (c *Controller) takeWaitWakes() waitWakes {
	c.waitWakeMu.Lock()
	defer c.waitWakeMu.Unlock()
	w := c.waitPending
	c.waitPending = waitWakes{}
	return w
}

// pumpWaits evaluates waits: all of them on the interval, at the
// earliest time condition and on a full wake; only the named ones on a
// targeted wake.
func (c *Controller) pumpWaits() {
	full := true
	var subj map[string]bool
	for {
		if full {
			c.pollWaits(c.ctx, c.waitSrc(), c.threadSrc())
		} else {
			c.pollWaitsOn(c.ctx, c.waitSrc(), c.threadSrc(), subj)
		}
		next := c.untilNextTime(c.ctx)
		select {
		case <-c.ctx.Done():
			return
		case <-c.after(next):
			full, subj = true, nil
		case <-c.waitWake:
			w := c.takeWaitWakes()
			full, subj = w.all, w.subj
		}
	}
}

// pollWaitsOn evaluates the waits with a condition on one of subjects.
func (c *Controller) pollWaitsOn(ctx context.Context, src waitSource, threads threadSource, subjects map[string]bool) {
	if len(subjects) == 0 {
		return
	}
	c.waitMu.Lock()
	defer c.waitMu.Unlock()
	ts, err := c.waiting(ctx, "")
	if err != nil {
		if ctx.Err() == nil {
			c.log.WarnContext(ctx, "waits: list", "err", err)
		}
		return
	}
	pass := c.newWaitPass(src, threads)
	for _, t := range ts {
		if !waitNames(waitOf(t), subjects) {
			continue
		}
		if err := c.evalWait(ctx, pass, t); err != nil && ctx.Err() == nil {
			c.log.WarnContext(ctx, "waits: evaluate", "ticket", t.ID, "err", err)
		}
	}
}

// waitNames reports whether an unmet condition of w is on a subject.
func waitNames(w *store.Wait, subjects map[string]bool) bool {
	if w == nil {
		return false
	}
	for _, cd := range w.Conds {
		if cd.Met {
			continue
		}
		for _, s := range condSubjects(cd) {
			if subjects[s] {
				return true
			}
		}
	}
	return false
}

func (c *Controller) newWaitPass(src waitSource, threads threadSource) *waitPass {
	return &waitPass{src: src, threads: threads, linear: c.linearWaitSrc(), prs: map[string]prInfo{}, reviews: map[string][]prReview{}, checks: map[string][]prCheck{}, issues: map[string]linear.Context{}}
}

// untilNextTime is how long until the earliest unmet time condition or
// wait timeout, capped at the interval.
func (c *Controller) untilNextTime(ctx context.Context) time.Duration {
	ts, err := c.waiting(ctx, "")
	if err != nil {
		return c.waitEvery()
	}
	return c.nextWaitTime(ts)
}

// nextWaitTime is untilNextTime over ts.
func (c *Controller) nextWaitTime(ts []store.Ticket) time.Duration {
	d := c.waitEvery()
	now := c.now()
	at := func(t time.Time) {
		if !t.IsZero() {
			d = min(d, max(t.Sub(now), 0))
		}
	}
	for _, t := range ts {
		w := waitOf(t)
		if w == nil {
			continue
		}
		if !w.TimedOut {
			at(w.Timeout)
		}
		for _, cd := range w.Conds {
			if cd.Type == store.WaitTime && !cd.Met {
				at(cd.At)
			}
		}
	}
	return d
}
