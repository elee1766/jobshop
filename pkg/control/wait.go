package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// Waits (#381). A task or step with fields.wait is out of the ready
// frontier. The controller evaluates its conditions on the subscription
// poll interval and when the ingest drain moves a webhook delivery, then
// clears the wait: the ticket becomes ready, pre-assigned to the session
// that parked it when resume is set. A PR that merges, closes or moves
// its head invalidates review and checks conditions bound to the old
// head; that also clears the wait, with the reason, so the agent decides
// again. Past its timeout a wait is labeled needs-human once.

// waitSource reads PR state from GitHub. Tests replace it.
type waitSource interface {
	PR(ctx context.Context, repo string, n int) (prInfo, error)
	Reviews(ctx context.Context, repo string, n int) ([]prReview, error)
	// Checks lists the checks of a PR's current head (required only when
	// asked), or of sha when n is 0.
	Checks(ctx context.Context, repo string, n int, sha string, required bool) ([]prCheck, error)
	// WorkflowRuns lists the GitHub Actions runs on commit sha, every
	// workflow (github.checks with workflow=, JOB-75).
	WorkflowRuns(ctx context.Context, repo, sha string) ([]actionsRun, error)
	// Issue reads an issue's state and labels (github.label, #603).
	Issue(ctx context.Context, repo string, n int) (issueInfo, error)
	// Tag reports whether tag exists in repo (github.tag, JOB-11).
	Tag(ctx context.Context, repo, tag string) (bool, error)
	// Contains reports whether head is sha or a descendant of it in repo
	// (deploy, JOB-11).
	Contains(ctx context.Context, repo, sha, head string) (bool, error)
}

// linearWaitSource reads a Linear issue for linear.state and
// linear.label. Tests replace it.
type linearWaitSource interface {
	IssueContext(ctx context.Context, key string) (linear.Context, error)
}

// liveLinearWait reads through the identity covering the issue's team.
type liveLinearWait struct{ c *Controller }

func (l liveLinearWait) IssueContext(ctx context.Context, key string) (linear.Context, error) {
	cl, err := l.c.linearClient(ctx, key)
	if err != nil {
		return linear.Context{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	return cl.GetContext(ctx, key, false)
}

type issueInfo struct {
	State  string // open or closed
	Labels []string
}

type prInfo struct {
	Head   string
	State  string // open or closed
	Merged bool
	// Conflicting: GitHub reports mergeable false. No pull_request CI runs
	// on such a PR, so a checks wait on it would never resolve.
	Conflicting bool
}

type prReview struct {
	Login  string
	State  string // APPROVED, CHANGES_REQUESTED, COMMENTED, DISMISSED
	Commit string
	At     time.Time
}

// prCheck bucket is gh's: pass, fail, pending, skipping, cancel.
type prCheck struct {
	Name   string
	Bucket string
}

// waitEvery is the poll interval: the subscriptions interval when set.
func (c *Controller) waitEvery() time.Duration {
	if s := c.config().Subscriptions; s != nil {
		return config.Duration(s.Interval, defaultSubInterval)
	}
	return defaultSubInterval
}

func (c *Controller) waitSrc() waitSource {
	if c.waitSource != nil {
		return c.waitSource
	}
	return ghWaitSource{}
}

func (c *Controller) linearWaitSrc() linearWaitSource {
	if c.linearWait != nil {
		return c.linearWait
	}
	return liveLinearWait{c: c}
}

func (c *Controller) threadSrc() threadSource {
	if c.subSource != nil {
		return c.subSource
	}
	return muxSource{gh: ghSource{}, linear: linearSource{c: c, url: c.linearURL}}
}

// waitOf returns fields.wait of a task or step, nil for none.
func waitOf(t store.Ticket) *store.Wait {
	if t.Kind != store.KindTask && t.Kind != store.KindStep || len(t.Fields) == 0 {
		return nil
	}
	var f struct {
		Wait *store.Wait `json:"wait"`
	}
	if json.Unmarshal(t.Fields, &f) != nil {
		return nil
	}
	return f.Wait
}

// withWait returns t's fields with wait set (nil removes it), keeping
// the other fields.
func withWait(t store.Ticket, w *store.Wait) (json.RawMessage, error) {
	m := map[string]json.RawMessage{}
	if len(t.Fields) > 0 {
		if err := json.Unmarshal(t.Fields, &m); err != nil {
			return nil, err
		}
	}
	if w == nil {
		delete(m, "wait")
	} else {
		b, err := json.Marshal(w)
		if err != nil {
			return nil, err
		}
		m["wait"] = b
	}
	return json.Marshal(m)
}

// waitAdd parks a ticket. A held ticket is released as a deferral, not a
// failed attempt; with resume the wait remembers the session.
func (c *Controller) waitAdd(ctx context.Context, sc scope, p rpc.WaitAddParams) (any, error) {
	id, err := c.current(ctx, sc, p.ID)
	if err != nil {
		return nil, err
	}
	t, err := c.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if t.Kind != store.KindTask && t.Kind != store.KindStep {
		return nil, fmt.Errorf("%w: only tasks and steps wait (%s is a %s)", store.ErrInvalid, id, t.Kind)
	}
	if t.Status == store.StatusClosed {
		return nil, fmt.Errorf("%w: %s", store.ErrClosed, id)
	}
	if !sc.admin && t.Project != sc.project {
		return nil, forbidden("%s is in another project", id)
	}
	var held *store.Claim
	if t.Status == store.StatusInProgress {
		cl, err := c.heldClaim(ctx, sc, id)
		if err != nil {
			return nil, err
		}
		held = &cl
	}
	now := c.now().UTC()
	w := &store.Wait{Conds: p.Conds, Reason: strings.TrimSpace(p.Reason), Owner: p.Owner, Resume: p.Resume, Since: now, By: sc.actor}
	if p.Any {
		w.Mode = store.WaitAny
	}
	if p.Timeout != nil {
		if !p.Timeout.After(now) {
			return nil, fmt.Errorf("%w: timeout must be in the future", store.ErrInvalid)
		}
		w.Timeout = p.Timeout.UTC()
	}
	if p.Resume {
		w.Session = t.Holder
		if w.Session == "" {
			w.Session = sc.session
		}
	}
	if err := w.Validate(); err != nil {
		return nil, err
	}
	for _, cd := range w.Conds {
		if cd.Type == store.WaitDeploy && c.opt.Commit == "" {
			return nil, fmt.Errorf("%w: deploy wait: this controller reports no build commit, so it cannot tell what is deployed", store.ErrInvalid)
		}
	}
	fields, err := withWait(t, w)
	if err != nil {
		return nil, err
	}
	if t, err = c.store.Update(ctx, sc.actor, id, store.Update{Fields: fields, IfRevision: t.Revision}); err != nil {
		return nil, err
	}
	if held != nil {
		held.Fence = t.Fence
		// A release with until is a deferral: not a failed attempt.
		if t, err = c.store.Release(ctx, "", held.TicketID, held, store.Outcome{Reason: "waiting: " + w.Reason, Until: now}); err != nil {
			return nil, err
		}
	}
	c.wakeWaits()
	return t, nil
}

// waitList lists waiting tickets.
func (c *Controller) waitList(ctx context.Context, sc scope, p rpc.WaitListParams) (any, error) {
	proj := p.Project
	if !sc.admin {
		proj = sc.project
	}
	ts, err := c.waiting(ctx, proj)
	if err != nil {
		return nil, err
	}
	out := []rpc.WaitItem{}
	for _, t := range ts {
		out = append(out, rpc.WaitItem{ID: t.ID, Project: t.Project, Title: t.Title, Wait: *waitOf(t)})
	}
	return out, nil
}

// waitClear removes a wait by hand.
func (c *Controller) waitClear(ctx context.Context, sc scope, p rpc.WaitClearParams) (any, error) {
	if p.ID == "" {
		return nil, fmt.Errorf("%w: ticket id is required", store.ErrInvalid)
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if !sc.admin && t.Project != sc.project {
		return nil, forbidden("%s is in another project", p.ID)
	}
	w := waitOf(t)
	if w == nil {
		return t, nil
	}
	why := "cleared by " + sc.actor
	if p.Reason != "" {
		why += ": " + p.Reason
	}
	return c.clearWait(ctx, t, w, why)
}

// waiting is every open task or step with a wait, in project (empty: all).
func (c *Controller) waiting(ctx context.Context, project string) ([]store.Ticket, error) {
	var out []store.Ticket
	for _, k := range []store.Kind{store.KindTask, store.KindStep} {
		ts, err := c.store.List(ctx, store.Filter{Project: project, Kind: k, Status: store.StatusOpen, Limit: math.MaxInt32})
		if err != nil {
			return nil, err
		}
		for _, t := range ts {
			if waitOf(t) != nil {
				out = append(out, t)
			}
		}
	}
	return out, nil
}

// clearWait removes the wait, notes why on the ticket and, with resume,
// pre-assigns it to the session that parked it while that is open and
// idle. A busy one leaves the ticket in its pool (JOB-120).
func (c *Controller) clearWait(ctx context.Context, t store.Ticket, w *store.Wait, why string) (store.Ticket, error) {
	fields, err := withWait(t, nil)
	if err != nil {
		return t, err
	}
	// One Update so no reader sees the wait cleared without its comment
	// (JOB-39).
	u := store.Update{
		Fields:     fields,
		IfRevision: t.Revision,
		Comment:    fmt.Sprintf("Wait (%s) ended: %s", w.Reason, why),
	}
	if w.TimedOut && slices.Contains(t.Labels, needsHumanLabel) {
		u.RemoveLabels = []string{needsHumanLabel}
	}
	if w.Resume && w.Session != "" && t.Holder == "" {
		switch holder, busy := c.resumeHolder(ctx, w.Session); {
		case holder != "":
			u.Holder = &holder
		case busy != "":
			u.Comment += fmt.Sprintf("\n\nSession %s is busy on %s, so this goes back to pool %s for any idle session.", w.Session, busy, t.Pool)
		}
	}
	if t, err = c.store.Update(ctx, "controller", t.ID, u); err != nil {
		return t, err
	}
	c.log.InfoContext(ctx, "wait cleared", "ticket", t.ID, "why", why)
	return t, nil
}

// resumeHolder is the session a resumed ticket is pre-assigned to: the
// parking session, while it is open and holds no claimed ticket. A session
// claims its in-progress work before anything pre-assigned, so a busy one
// would leave the resumed ticket behind unrelated work for as long as that
// takes (JOB-120). Then holder is empty and busy names the work; the
// ticket stays in its pool, and the PR and the ticket's comments carry
// the context to whichever session claims it.
func (c *Controller) resumeHolder(ctx context.Context, id string) (holder, busy string) {
	s, err := c.store.Get(ctx, id)
	if err != nil || s.Kind != store.KindSession || s.Status == store.StatusClosed {
		return "", ""
	}
	held, err := c.store.List(ctx, store.Filter{Holder: id, Status: store.StatusInProgress, Limit: 1})
	if err != nil {
		return "", "" // unknown: the pool is never wrong, only slower
	}
	if len(held) > 0 {
		return "", held[0].ID
	}
	return id, ""
}

// waitPass caches GitHub reads within one pass, so waits on one PR
// share a fetch.
type waitPass struct {
	src     waitSource
	threads threadSource
	linear  linearWaitSource
	prs     map[string]prInfo
	reviews map[string][]prReview
	checks  map[string][]prCheck
	runs    map[string][]actionsRun
	issues  map[string]linear.Context
}

func (p *waitPass) linearIssue(ctx context.Context, key string) (linear.Context, error) {
	if v, ok := p.issues[key]; ok {
		return v, nil
	}
	v, err := p.linear.IssueContext(ctx, key)
	if err == nil {
		p.issues[key] = v
	}
	return v, err
}

func prKey(repo string, n int) string { return fmt.Sprintf("%s#%d", repo, n) }

func (p *waitPass) pr(ctx context.Context, repo string, n int) (prInfo, error) {
	k := prKey(repo, n)
	if v, ok := p.prs[k]; ok {
		return v, nil
	}
	v, err := p.src.PR(ctx, repo, n)
	if err == nil {
		p.prs[k] = v
	}
	return v, err
}

func (p *waitPass) reviewsOf(ctx context.Context, repo string, n int) ([]prReview, error) {
	k := prKey(repo, n)
	if v, ok := p.reviews[k]; ok {
		return v, nil
	}
	v, err := p.src.Reviews(ctx, repo, n)
	if err == nil {
		p.reviews[k] = v
	}
	return v, err
}

func (p *waitPass) checksOf(ctx context.Context, repo string, n int, sha string, required bool) ([]prCheck, error) {
	k := fmt.Sprintf("%s#%d@%s/%t", repo, n, sha, required)
	if v, ok := p.checks[k]; ok {
		return v, nil
	}
	v, err := p.src.Checks(ctx, repo, n, sha, required)
	if err == nil {
		p.checks[k] = v
	}
	return v, err
}

func (p *waitPass) runsOf(ctx context.Context, repo, sha string) ([]actionsRun, error) {
	k := repo + "@" + sha
	if v, ok := p.runs[k]; ok {
		return v, nil
	}
	v, err := p.src.WorkflowRuns(ctx, repo, sha)
	if err == nil {
		p.runs[k] = v
	}
	return v, err
}

// pollWaits runs one pass over every waiting ticket.
func (c *Controller) pollWaits(ctx context.Context, src waitSource, threads threadSource) {
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
		if err := c.evalWait(ctx, pass, t); err != nil && ctx.Err() == nil {
			c.log.WarnContext(ctx, "waits: evaluate", "ticket", t.ID, "err", err)
		}
	}
}

// errInvalidated means a condition can no longer hold as stated.
type errInvalidated struct{ why string }

func (e errInvalidated) Error() string { return e.why }

// evalWait evaluates one ticket's wait and clears, times out or saves it.
func (c *Controller) evalWait(ctx context.Context, pass *waitPass, t store.Ticket) error {
	w := waitOf(t)
	if w == nil {
		return nil
	}
	now := c.now().UTC()
	before, _ := json.Marshal(w)
	var errs []error
	for i := range w.Conds {
		cd := &w.Conds[i]
		if cd.Met {
			continue
		}
		met, note, err := c.evalCond(ctx, pass, w, cd, now)
		var inv errInvalidated
		if errors.As(err, &inv) {
			_, cerr := c.clearWait(ctx, t, w, "invalidated: "+inv.why)
			return cerr
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
		_, err := c.clearWait(ctx, t, w, "condition met: "+strings.Join(notes, "; "))
		return err
	}
	timedOut := !w.Timeout.IsZero() && !now.Before(w.Timeout) && !w.TimedOut
	if timedOut {
		w.TimedOut = true
	}
	if after, _ := json.Marshal(w); !bytes.Equal(before, after) {
		fields, err := withWait(t, w)
		if err != nil {
			return err
		}
		if t, err = c.store.Update(ctx, "controller", t.ID, store.Update{Fields: fields, IfRevision: t.Revision}); err != nil {
			return err
		}
	}
	if timedOut {
		if err := store.AddLabel(ctx, c.store, "controller", t.ID, needsHumanLabel); err != nil {
			return err
		}
		owner := w.Owner
		if owner == "" {
			owner = "unset"
		}
		_, _ = store.AddComment(ctx, c.store, "controller", t.ID, fmt.Sprintf("Wait (%s) passed its timeout %s; owner %s. Clear it with `jobshop wait clear %s`.", w.Reason, w.Timeout.Format(time.RFC3339), owner, t.ID))
		c.log.WarnContext(ctx, "wait timed out", "ticket", t.ID, "reason", w.Reason, "owner", w.Owner)
		c.holdChat(ctx, t.ID)
	}
	return errors.Join(errs...)
}

// evalCond reports whether cd holds now. It records cd.Head the first
// time a PR condition is evaluated without one.
func (c *Controller) evalCond(ctx context.Context, pass *waitPass, w *store.Wait, cd *store.WaitCond, now time.Time) (bool, string, error) {
	switch cd.Type {
	case store.WaitTime:
		return !now.Before(cd.At), "time " + cd.At.Format(time.RFC3339), nil
	case store.WaitReply:
		return c.evalReply(ctx, pass, w, cd)
	case store.WaitLabel:
		return evalLabel(ctx, pass, cd)
	case store.WaitTag:
		ok, err := pass.src.Tag(ctx, cd.Repo, cd.Tag)
		// The note is stored on every pass, met or not (#716).
		if !ok {
			return false, fmt.Sprintf("tag %s not pushed to %s yet", cd.Tag, cd.Repo), err
		}
		return true, fmt.Sprintf("tag %s pushed to %s", cd.Tag, cd.Repo), err
	case store.WaitDeploy:
		if c.opt.Commit == "" {
			return false, "", errInvalidated{"this controller reports no build commit"}
		}
		ok, err := pass.src.Contains(ctx, cd.Repo, cd.SHA, c.opt.Commit)
		if !ok {
			return false, fmt.Sprintf("deployed %s does not contain %s yet", short(c.opt.Commit), short(cd.SHA)), err
		}
		return true, fmt.Sprintf("deployed %s contains %s", short(c.opt.Commit), short(cd.SHA)), err
	case store.WaitRun:
		return c.evalRun(ctx, cd)
	case store.WaitLinearState, store.WaitLinearLabel:
		return evalLinear(ctx, pass, cd)
	case store.WaitChecks:
		if cd.PR == 0 {
			return evalChecks(ctx, pass, cd, cd.SHA)
		}
	}
	pr, err := pass.pr(ctx, cd.Repo, cd.PR)
	if err != nil {
		return false, "", err
	}
	ref := prKey(cd.Repo, cd.PR)
	if cd.Head == "" {
		cd.Head = pr.Head
	}
	if cd.Type == store.WaitPR {
		switch cd.State {
		case "merged":
			if pr.State == "closed" && !pr.Merged {
				return false, "", errInvalidated{ref + " closed without merging"}
			}
			return pr.Merged, ref + " merged", nil
		case "closed":
			return pr.State == "closed", ref + " closed", nil
		}
		return pr.Head != cd.Head || pr.State == "closed", fmt.Sprintf("%s head moved %s -> %s", ref, short(cd.Head), short(pr.Head)), nil
	}
	if pr.Merged {
		return false, "", errInvalidated{ref + " merged"}
	}
	if pr.State == "closed" {
		return false, "", errInvalidated{ref + " closed"}
	}
	if pr.Head != cd.Head {
		return false, "", errInvalidated{fmt.Sprintf("%s head moved %s -> %s", ref, short(cd.Head), short(pr.Head))}
	}
	if cd.Type == store.WaitChecks {
		if pr.Conflicting {
			return false, "", errInvalidated{ref + " conflicts with its base; CI will not run until it is rebased"}
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
	var last *prReview
	for i := range rs {
		r := &rs[i]
		if !strings.EqualFold(r.Login, cd.Login) || r.Commit != cd.Head {
			continue
		}
		if last == nil || !r.At.Before(last.At) {
			last = r
		}
	}
	return last != nil && last.State == want, fmt.Sprintf("%s %s by %s on %s", ref, strings.ToLower(want), cd.Login, short(cd.Head)), nil
}

// evalRun: the run is done (state done) or ended any way (closed). A
// done wait on a run that was cancelled can never hold.
func (c *Controller) evalRun(ctx context.Context, cd *store.WaitCond) (bool, string, error) {
	r, err := c.store.Get(ctx, cd.Run)
	if err != nil {
		return false, "", err
	}
	f, err := store.DecodeFields[store.RunFields](r)
	if err != nil {
		return false, "", err
	}
	ended := r.Status == store.StatusClosed || f.State == store.RunDone || f.State == store.RunCancelled
	if cd.State == "closed" {
		return ended, fmt.Sprintf("run %s ended (%s)", cd.Run, f.State), nil
	}
	if f.State == store.RunCancelled {
		return false, "", errInvalidated{"run " + cd.Run + " was cancelled"}
	}
	return f.State == store.RunDone, "run " + cd.Run + " done", nil
}

// evalLinear: linear.state matches the state name or type
// (case-insensitive); linear.label has the label, or lacks it with Absent.
func evalLinear(ctx context.Context, pass *waitPass, cd *store.WaitCond) (bool, string, error) {
	is, err := pass.linearIssue(ctx, cd.Thread)
	if err != nil {
		return false, "", err
	}
	if cd.Type == store.WaitLinearState {
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
func evalLabel(ctx context.Context, pass *waitPass, cd *store.WaitCond) (bool, string, error) {
	is, err := pass.src.Issue(ctx, cd.Repo, cd.Issue)
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
func evalChecks(ctx context.Context, pass *waitPass, cd *store.WaitCond, head string) (bool, string, error) {
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
	at := short(head)
	if len(cancelled) > 0 && len(failed) == 0 && len(pending) == 0 {
		return false, "", errInvalidated{fmt.Sprintf("%s cancelled on %s: %s; this head will not go green, push or wait on a newer commit", what, at, strings.Join(cancelled, ", "))}
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

// evalWorkflow follows the newest run of cd.Workflow on head:
// none yet or in progress is not met, success meets state=success,
// failure meets state=failure, and a cancelled or skipped run can never
// become green, so the wait is invalidated and the agent resumes now.
func evalWorkflow(ctx context.Context, pass *waitPass, cd *store.WaitCond, head string) (bool, string, error) {
	rs, err := pass.runsOf(ctx, cd.Repo, head)
	if err != nil {
		return false, "", err
	}
	at := short(head)
	run, ok := newestRun(rs, cd.Workflow)
	if !ok {
		return false, fmt.Sprintf("%s not started on %s", cd.Workflow, at), nil
	}
	switch checkBucket(run.Status, run.Conclusion) {
	case "pending":
		return false, fmt.Sprintf("%s in progress on %s", cd.Workflow, at), nil
	case "pass":
		return cd.State == "success", fmt.Sprintf("%s green on %s", cd.Workflow, at), nil
	case "fail":
		return cd.State == "failure", fmt.Sprintf("%s failed on %s", cd.Workflow, at), nil
	case "cancel":
		return false, "", errInvalidated{fmt.Sprintf("%s cancelled on %s; this head will not go green, push or wait on a newer commit", cd.Workflow, at)}
	}
	return false, "", errInvalidated{fmt.Sprintf("%s skipped on %s", cd.Workflow, at)}
}

// newestRun is the highest run id among the runs of workflow, matched
// case-insensitively by name or by workflow file (ci, ci.yml or
// .github/workflows/ci.yml), as trigger matching does: workflow=ci must
// find a workflow named CI in ci.yml (JOB-88).
func newestRun(rs []actionsRun, workflow string) (actionsRun, bool) {
	var best actionsRun
	found := false
	for _, r := range rs {
		base := path.Base(r.Path)
		stem := strings.TrimSuffix(base, path.Ext(base))
		if !slices.ContainsFunc([]string{r.Name, r.Path, base, stem}, func(s string) bool {
			return s != "" && strings.EqualFold(s, workflow)
		}) {
			continue
		}
		if !found || r.ID > best.ID {
			best, found = r, true
		}
	}
	return best, found
}

// evalReply: a human comment on the thread after the wait started, from
// From when set. It reads through the subscription source, so the same
// filters (bots, ignore, authors) apply.
func (c *Controller) evalReply(ctx context.Context, pass *waitPass, w *store.Wait, cd *store.WaitCond) (bool, string, error) {
	th := thread{Linear: cd.Thread}
	if repo, num, ok := strings.Cut(cd.Thread, "#"); ok {
		n, _ := strconv.Atoi(num)
		th = thread{Repo: repo, Number: n, PR: true}
	}
	cs, err := pass.threads.Comments(ctx, th, w.Since)
	if err != nil {
		return false, "", err
	}
	for _, cm := range cs {
		if !cm.At.After(w.Since) || c.ignored(cm) {
			continue
		}
		if cd.From != "" && !strings.EqualFold(cd.From, cm.Author) && (!cm.Linear || !c.linearMatch(cd.From, cm)) {
			continue
		}
		return true, fmt.Sprintf("%s replied on %s", cm.Author, cd.Thread), nil
	}
	return false, "", nil
}

func short(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

// ghWaitSource reads PRs with the gh CLI, like ghSource.
type ghWaitSource struct{}

// ghCmdName names a gh call for its exec span: the subcommand only ("api",
// "pr"), never the paths or flags after it.
func ghCmdName(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

func ghRun(ctx context.Context, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cmd, done := telemetry.ExecContext(ctx, "gh-"+ghCmdName(args), "gh", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	done(err)
	if err != nil {
		return b, fmt.Errorf("gh %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return b, nil
}

func (ghWaitSource) PR(ctx context.Context, repo string, n int) (prInfo, error) {
	b, err := ghRun(ctx, "api", fmt.Sprintf("repos/%s/pulls/%d", repo, n))
	if err != nil {
		return prInfo{}, err
	}
	return parsePR(b)
}

func (ghWaitSource) Reviews(ctx context.Context, repo string, n int) ([]prReview, error) {
	b, err := ghRun(ctx, "api", "--paginate", "--slurp", fmt.Sprintf("repos/%s/pulls/%d/reviews?per_page=100", repo, n))
	if err != nil {
		return nil, err
	}
	return parseReviews(b)
}

func (ghWaitSource) Checks(ctx context.Context, repo string, n int, sha string, required bool) ([]prCheck, error) {
	if n == 0 {
		return commitChecks(ctx, repo, sha)
	}
	args := []string{"pr", "checks", strconv.Itoa(n), "--repo", repo, "--json", "name,bucket"}
	if required {
		args = append(args, "--required")
	}
	// gh pr checks exits non-zero while checks are pending or failing;
	// its JSON is still the answer.
	b, err := ghRun(ctx, args...)
	if cs, perr := parsePRChecks(b); perr == nil && len(bytes.TrimSpace(b)) > 0 {
		return cs, nil
	}
	if err != nil && (strings.Contains(err.Error(), "no required checks") || strings.Contains(err.Error(), "no checks reported")) {
		return nil, nil // nothing reported yet: not green
	}
	return nil, err
}

// Tag reads GET /repos/{repo}/git/ref/tags/{tag}; 404 is not yet.
func (ghWaitSource) Tag(ctx context.Context, repo, tag string) (bool, error) {
	_, err := ghRun(ctx, "api", fmt.Sprintf("repos/%s/git/ref/tags/%s", repo, tag))
	if err != nil && strings.Contains(err.Error(), "404") {
		return false, nil
	}
	return err == nil, err
}

// Contains compares sha...head: identical or ahead means head has sha.
func (ghWaitSource) Contains(ctx context.Context, repo, sha, head string) (bool, error) {
	b, err := ghRun(ctx, "api", fmt.Sprintf("repos/%s/compare/%s...%s", repo, sha, head), "--jq", ".status")
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil // head not pushed (a local build): not deployed
		}
		return false, err
	}
	return compareContains(string(b)), nil
}

// compareContains reads the compare API's status.
func compareContains(status string) bool {
	s := strings.TrimSpace(status)
	return s == "identical" || s == "ahead"
}

// Issue reads GET /repos/{repo}/issues/{n}.
func (ghWaitSource) Issue(ctx context.Context, repo string, n int) (issueInfo, error) {
	b, err := ghRun(ctx, "api", fmt.Sprintf("repos/%s/issues/%d", repo, n))
	if err != nil {
		return issueInfo{}, err
	}
	return parseIssue(b)
}

func parseIssue(b []byte) (issueInfo, error) {
	var r struct {
		State  string `json:"state"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return issueInfo{}, fmt.Errorf("issue: %w", err)
	}
	out := issueInfo{State: r.State}
	for _, l := range r.Labels {
		out.Labels = append(out.Labels, l.Name)
	}
	return out, nil
}

// parsePR reads GET /repos/{repo}/pulls/{n}.
func parsePR(b []byte) (prInfo, error) {
	var r struct {
		State     string `json:"state"`
		Merged    bool   `json:"merged"`
		Mergeable *bool  `json:"mergeable"` // null while GitHub computes it
		Head      struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return prInfo{}, fmt.Errorf("pull request: %w", err)
	}
	return prInfo{Head: r.Head.SHA, State: r.State, Merged: r.Merged, Conflicting: r.Mergeable != nil && !*r.Mergeable}, nil
}

// parseReviews reads the slurped pages of GET .../pulls/{n}/reviews.
func parseReviews(b []byte) ([]prReview, error) {
	var pages [][]struct {
		State       string    `json:"state"`
		CommitID    string    `json:"commit_id"`
		SubmittedAt time.Time `json:"submitted_at"`
		User        struct {
			Login string `json:"login"`
		} `json:"user"`
	}
	if err := json.Unmarshal(b, &pages); err != nil {
		return nil, fmt.Errorf("reviews: %w", err)
	}
	var out []prReview
	for _, p := range pages {
		for _, r := range p {
			out = append(out, prReview{Login: r.User.Login, State: r.State, Commit: r.CommitID, At: r.SubmittedAt})
		}
	}
	return out, nil
}

// parsePRChecks reads `gh pr checks --json name,bucket`.
func parsePRChecks(b []byte) ([]prCheck, error) {
	var out []prCheck
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("pr checks: %w", err)
	}
	return out, nil
}

// actionsRun is one GitHub Actions run on a commit.
type actionsRun struct {
	ID         int64
	Name       string
	Path       string // .github/workflows/ci.yml
	Status     string
	Conclusion string
}

func (ghWaitSource) WorkflowRuns(ctx context.Context, repo, sha string) ([]actionsRun, error) {
	full := sha
	if len(full) < 40 { // head_sha takes only a full sha
		out, err := ghRun(ctx, "api", fmt.Sprintf("repos/%s/commits/%s", repo, sha), "--jq", ".sha")
		if err != nil {
			return nil, err
		}
		full = strings.TrimSpace(string(out))
	}
	b, err := ghRun(ctx, "api", fmt.Sprintf("repos/%s/actions/runs?head_sha=%s&per_page=100", repo, full))
	if err != nil {
		return nil, err
	}
	return parseRunList(b)
}

// commitChecks lists the check runs of sha plus its Actions workflow
// runs (#831). A workflow run queued behind a concurrency group has no
// check runs yet; without it a commit looks green once another workflow
// (release) finishes, and a run cancelled while queued would not show.
func commitChecks(ctx context.Context, repo, sha string) ([]prCheck, error) {
	b, err := ghRun(ctx, "api", fmt.Sprintf("repos/%s/commits/%s/check-runs?per_page=100", repo, sha))
	if err != nil {
		return nil, err
	}
	cs, err := parseCheckRuns(b)
	if err != nil {
		return nil, err
	}
	rs, err := ghWaitSource{}.WorkflowRuns(ctx, repo, sha)
	if err != nil {
		return nil, err
	}
	for _, r := range rs {
		cs = append(cs, prCheck{Name: r.Name + " (workflow)", Bucket: checkBucket(r.Status, r.Conclusion)})
	}
	return cs, nil
}

// checkBucket maps a check or workflow run's status and conclusion to
// gh's buckets.
func checkBucket(status, conclusion string) string {
	if status != "completed" {
		return "pending"
	}
	switch conclusion {
	case "success", "neutral":
		return "pass"
	case "skipped":
		return "skipping"
	case "cancelled":
		return "cancel"
	}
	return "fail"
}

// parseRunList reads GET .../actions/runs?head_sha=.
func parseRunList(b []byte) ([]actionsRun, error) {
	var r struct {
		WorkflowRuns []struct {
			ID         int64  `json:"id"`
			Name       string `json:"name"`
			Path       string `json:"path"`
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		} `json:"workflow_runs"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("workflow runs: %w", err)
	}
	var out []actionsRun
	for _, wr := range r.WorkflowRuns {
		out = append(out, actionsRun{ID: wr.ID, Name: wr.Name, Path: wr.Path, Status: wr.Status, Conclusion: wr.Conclusion})
	}
	return out, nil
}

// parseCheckRuns reads GET .../commits/{sha}/check-runs into gh buckets.
func parseCheckRuns(b []byte) ([]prCheck, error) {
	var r struct {
		CheckRuns []struct {
			Name       string `json:"name"`
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		} `json:"check_runs"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("check runs: %w", err)
	}
	var out []prCheck
	for _, cr := range r.CheckRuns {
		out = append(out, prCheck{Name: cr.Name, Bucket: checkBucket(cr.Status, cr.Conclusion)})
	}
	return out, nil
}
