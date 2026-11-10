// Package storerules holds the I/O-free ticket rules every store must
// apply the same way (JOB-40): claim and fence decisions, release,
// handoff, close and step failure transitions, the revision bump, update
// diffs and the field-derived columns. A store loads a ticket, calls a
// rule, then writes the result in its own transaction. Plain glue stays
// in each store; nothing here touches a database.
package storerules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Stamp is the time stored for now: UTC, nanosecond precision, as the
// stores keep timestamps (JOB-58). Round(0) drops the monotonic reading.
func Stamp(now time.Time) time.Time { return now.UTC().Round(0) }

// Clock hands out strictly increasing creation times (JOB-58): a batch
// insert in one transaction gets distinct, ordered created_at values,
// so ticket order reaches the id tiebreak only across controllers.
type Clock struct {
	mu   sync.Mutex
	last int64
}

// Next returns now, or 1ns after the last value when the clock returned
// the same or an earlier time.
func (c *Clock) Next(now time.Time) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := now.UnixNano()
	if n <= c.last {
		n = c.last + 1
	}
	c.last = n
	return time.Unix(0, n).UTC()
}

// Floor returns now, or the last value Next handed out when now is
// earlier, so a write after a ticket creation is never stamped before it.
func (c *Clock) Floor(now time.Time) time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	if now.UnixNano() < c.last {
		return time.Unix(0, c.last).UTC()
	}
	return now
}

// Seed starts the clock at last (unix ns), the newest stored creation
// time, so a restarted store does not hand out times below it.
func (c *Clock) Seed(last int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.last = max(c.last, last)
}

// Bump moves t to its next revision at now and returns the revision the
// guarded write must still find.
func Bump(t *domain.Ticket, now time.Time) (prev int64) {
	prev = t.Revision
	t.Revision++
	t.UpdatedAt = Stamp(now)
	return prev
}

// CheckLease rejects a claim or heartbeat without a holder or lease.
func CheckLease(holder string, lease time.Duration) error {
	if holder == "" || lease <= 0 {
		return fmt.Errorf("%w: holder and a positive lease are required", domain.ErrInvalid)
	}
	return nil
}

// CanClaim is the check for claiming one named ticket: it must be open
// and unassigned or assigned to holder (or already held by holder, a
// resume), unblocked, and an open one must not wait for a resource slot.
func CanClaim(t domain.Ticket, holder string, blocked, waitsSlot bool) error {
	switch {
	case t.Status == domain.StatusClosed:
		return fmt.Errorf("%w: %s", domain.ErrClosed, t.ID)
	case t.Status == domain.StatusInProgress && t.Holder != holder:
		return fmt.Errorf("%w: %s is held by %s", domain.ErrNotHolder, t.ID, t.Holder)
	case t.Status == domain.StatusOpen && t.Holder != "" && t.Holder != holder:
		return fmt.Errorf("%w: %s is assigned to %s", domain.ErrNotHolder, t.ID, t.Holder)
	case blocked:
		return fmt.Errorf("%w: %s is blocked", domain.ErrNoWork, t.ID)
	case waitsSlot && t.Status == domain.StatusOpen:
		return fmt.Errorf("%w: %s waits for a resource slot", domain.ErrNoWork, t.ID)
	}
	return nil
}

// StillCandidate rechecks a ClaimNext candidate inside the claiming
// transaction: without a single writer another claimer may have taken it.
func StillCandidate(t domain.Ticket, holder string) error {
	if t.Status == domain.StatusClosed || (t.Holder != "" && t.Holder != holder) {
		return fmt.Errorf("%w: %s was taken", domain.ErrConflict, t.ID)
	}
	return nil
}

// Claim gives t to holder until now+lease. A fresh claim (from open)
// bumps the fence; a resume by the same holder keeps it. It returns the
// claim and the history diff; fresh reports whether slots must be taken.
func Claim(t *domain.Ticket, holder string, lease time.Duration, now time.Time) (c store.Claim, diff map[string]any, fresh bool) {
	resumed := t.Status == domain.StatusInProgress
	t.Status = domain.StatusInProgress
	t.Holder = holder
	t.LeaseUntil = Stamp(now.Add(lease))
	if !resumed {
		t.Fence++
	}
	return store.Claim{TicketID: t.ID, Holder: holder, Fence: t.Fence, LeaseUntil: t.LeaseUntil},
		map[string]any{"holder": holder, "fence": t.Fence, "resumed": resumed}, !resumed
}

// Holds reports ErrNotHolder unless c is t's current claim.
func Holds(t domain.Ticket, c store.Claim) error {
	if t.Status != domain.StatusInProgress || t.Holder != c.Holder || t.Fence != c.Fence {
		return fmt.Errorf("%w: %s", domain.ErrNotHolder, t.ID)
	}
	return nil
}

// Reopen returns t to open and unassigned. Resolution and closed time
// are the caller's to clear.
func Reopen(t *domain.Ticket) {
	t.Status = domain.StatusOpen
	t.Holder = ""
	t.LeaseUntil = time.Time{}
}

func reopenClosed(t *domain.Ticket) {
	Reopen(t)
	t.Resolution = ""
	t.ClosedAt = time.Time{}
}

// StepFailure counts a failure when t is a step. At the retry limit the
// step is closed failed; below it a failed close becomes another
// attempt. It returns the comment to add ("" for non-steps) and records
// the count in d.
func StepFailure(t *domain.Ticket, how string, d map[string]any, now time.Time) (string, error) {
	f, exhausted, isStep, err := domain.StepFailure(t)
	if err != nil || !isStep {
		return "", err
	}
	d["failures"] = f.Failures
	if exhausted && t.Status != domain.StatusClosed {
		t.Status = domain.StatusClosed
		t.Resolution = "failed"
		t.Holder = ""
		t.LeaseUntil = time.Time{}
		t.ClosedAt = Stamp(now)
	}
	if !exhausted && t.Status == domain.StatusClosed {
		reopenClosed(t)
	}
	return domain.StepFailureComment(f, how, exhausted), nil
}

// CreateDiff is the history diff of a new ticket: the labels it was born
// with, so a replay of the history knows them (#1103). Nil for none.
func CreateDiff(nt store.NewTicket) any {
	if len(nt.Labels) == 0 {
		return nil
	}
	return map[string]any{"labels": nt.Labels}
}

// Release returns a ticket held by c to open. With until set it is a
// deferral ("not yet"), not a failed attempt; otherwise a step counts a
// failure. It returns the history diff and the controller note.
func Release(t *domain.Ticket, c store.Claim, reason string, until, now time.Time) (map[string]any, string, error) {
	if err := Holds(*t, c); err != nil {
		return nil, "", err
	}
	Reopen(t)
	d := map[string]any{"reason": reason}
	if !until.IsZero() {
		t.DeferUntil = Stamp(until)
		d["defer_until"] = t.DeferUntil
		if err := noteDefer(t, reason); err != nil {
			return nil, "", err
		}
		return d, fmt.Sprintf("Released until %s: %s", t.DeferUntil.Format(time.RFC3339), reason), nil
	}
	note, err := StepFailure(t, "released: "+reason, d, now)
	return d, note, err
}

// noteDefer records why a step was released until t.DeferUntil, so Desk
// can show the reason beside the resume time (#1185). Other tickets keep
// the reason in the history and the "Released until" comment only.
func noteDefer(t *domain.Ticket, reason string) error {
	if t.Kind != domain.KindStep {
		return nil
	}
	f, err := domain.DecodeFields[domain.StepFields](*t)
	if err != nil {
		return err
	}
	f.Defer = &domain.StepDefer{Reason: reason, Until: t.DeferUntil}
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	t.Fields = b
	return nil
}

// Handoff moves t to pool. A held ticket is released and its fence
// bumped, so the old holder's claim stops working. A non-nil c must
// still hold it. It is not a step failure.
func Handoff(t *domain.Ticket, c *store.Claim, pool, note string) (map[string]any, error) {
	if pool == "" {
		return nil, fmt.Errorf("%w: handoff needs a pool", domain.ErrInvalid)
	}
	if t.Status == domain.StatusClosed {
		return nil, fmt.Errorf("%w: %s", domain.ErrClosed, t.ID)
	}
	if c != nil {
		if err := Holds(*t, *c); err != nil {
			return nil, err
		}
	}
	d := map[string]any{"from_pool": t.Pool, "pool": pool, "holder": t.Holder}
	t.Pool = pool
	if t.Status == domain.StatusInProgress {
		Reopen(t)
		t.Fence++
	}
	if note != "" {
		d["note"] = note
	}
	return d, nil
}

// Close closes t with resolution, fenced by c when non-nil. A failed
// close counts a step failure (and below the limit reopens it); a
// rebase close of a step reopens it without counting. It returns the
// history diff and the controller note.
func Close(t *domain.Ticket, c *store.Claim, resolution string, now time.Time, results ...*store.StepResult) (map[string]any, string, error) {
	if t.Status == domain.StatusClosed {
		return nil, "", fmt.Errorf("%w: %s", domain.ErrClosed, t.ID)
	}
	if c != nil {
		if err := Holds(*t, *c); err != nil {
			return nil, "", err
		}
	}
	if len(results) > 0 && results[0] != nil {
		if results[0].Revision != 0 && results[0].Revision != t.Revision {
			return nil, "", fmt.Errorf("%w: step changed before close", domain.ErrConflict)
		}
		if t.Kind != domain.KindStep {
			return nil, "", fmt.Errorf("%w: structured results require a step", domain.ErrInvalid)
		}
		f, err := domain.DecodeFields[domain.StepFields](*t)
		if err != nil {
			return nil, "", err
		}
		f.Outcome, f.StructuredOutput = results[0].Outcome, results[0].Output
		t.Fields, err = json.Marshal(f)
		if err != nil {
			return nil, "", err
		}
	}
	t.Status = domain.StatusClosed
	t.Resolution = resolution
	t.LeaseUntil = time.Time{}
	t.ClosedAt = Stamp(now)
	d := map[string]any{"resolution": resolution}
	if resolution == "failed" {
		note, err := StepFailure(t, "closed failed", d, now)
		return d, note, err
	}
	if resolution == domain.ResolutionRebase && t.Kind == domain.KindStep {
		reopenClosed(t)
		return d, "Closed for a rebase: reopened without counting a failed attempt.", nil
	}
	if resolution == domain.ResolutionRequestChanges && t.Kind == domain.KindStep {
		return requestChangesRound(t, d)
	}
	if resolution == "done" && t.Kind == domain.KindStep {
		// The change request was answered: later prompts drop it.
		if f, err := domain.DecodeFields[domain.StepFields](*t); err == nil && f.Changes != nil {
			f.Changes = nil
			if t.Fields, err = json.Marshal(f); err != nil {
				return nil, "", err
			}
		}
	}
	return d, "", nil
}

// requestChangesRound counts a review round on a reviewing step closed
// request_changes (JOB-98). Within the cap it reopens the step, without
// counting a failure; past it the step stays closed and the run needs a
// human.
func requestChangesRound(t *domain.Ticket, d map[string]any) (map[string]any, string, error) {
	f, err := domain.DecodeFields[domain.StepFields](*t)
	if err != nil {
		return nil, "", err
	}
	f.ReviewRounds++
	f.StructuredOutput = nil
	if t.Fields, err = json.Marshal(f); err != nil {
		return nil, "", err
	}
	d["review_rounds"] = f.ReviewRounds
	if f.ReviewRounds > f.ReviewRoundLimit() {
		return d, fmt.Sprintf("Changes requested in review round %d, over the limit of %d: step closed, run needs a human.", f.ReviewRounds, f.ReviewRoundLimit()), nil
	}
	reopenClosed(t)
	return d, fmt.Sprintf("Changes requested (review round %d of %d): reopened behind the step it reviews, without counting a failed attempt.", f.ReviewRounds, f.ReviewRoundLimit()), nil
}

// RequestChanges reopens target, a closed step of reviewer's run, with
// the reviewer's findings in fields.changes (JOB-98). reviewer is the
// reviewing step after its close counted the round. It returns the
// history diff and the comment for target.
func RequestChanges(target *domain.Ticket, reviewer domain.Ticket, findings string) (map[string]any, string, error) {
	if target.Kind != domain.KindStep || target.ID == reviewer.ID || target.ParentID == "" || target.ParentID != reviewer.ParentID {
		return nil, "", fmt.Errorf("%w: request_changes: %s is not another step of run %s", domain.ErrInvalid, target.ID, reviewer.ParentID)
	}
	if target.Status != domain.StatusClosed {
		return nil, "", fmt.Errorf("%w: request_changes: step %s is %s, not closed", domain.ErrInvalid, target.ID, target.Status)
	}
	rf, err := domain.DecodeFields[domain.StepFields](reviewer)
	if err != nil {
		return nil, "", err
	}
	tf, err := domain.DecodeFields[domain.StepFields](*target)
	if err != nil {
		return nil, "", err
	}
	tf.Changes = &domain.StepChanges{
		Step: rf.Step, Ticket: reviewer.ID, Round: rf.ReviewRounds, Max: rf.ReviewRoundLimit(), Findings: findings,
	}
	tf.StructuredOutput = nil
	if target.Fields, err = json.Marshal(tf); err != nil {
		return nil, "", err
	}
	reopenClosed(target)
	d := map[string]any{"request_changes": reviewer.ID, "round": rf.ReviewRounds}
	note := fmt.Sprintf("Changes requested by step %s (%s), review round %d of %d; reopened. The findings are in the prompt and on %s.",
		rf.Step, reviewer.ID, rf.ReviewRounds, rf.ReviewRoundLimit(), reviewer.ID)
	return d, note, nil
}

// ReopenClosed returns a closed ticket to open, checking ifRev when set.
func ReopenClosed(t *domain.Ticket, ifRev int64) error {
	if ifRev != 0 && ifRev != t.Revision {
		return fmt.Errorf("%w: %s at revision %d, want %d", domain.ErrConflict, t.ID, t.Revision, ifRev)
	}
	if t.Status != domain.StatusClosed {
		return fmt.Errorf("%w: %s is not closed", domain.ErrInvalid, t.ID)
	}
	if t.Kind == domain.KindStep {
		fields, err := domain.DecodeFields[domain.StepFields](*t)
		if err != nil {
			return err
		}
		fields.StructuredOutput = nil
		if t.Fields, err = json.Marshal(fields); err != nil {
			return err
		}
	}
	reopenClosed(t)
	return nil
}

// Expire reopens a ticket a dead holder held or was assigned, or whose
// lease ran out (op "release" or "lease_expired"). extra is merged into
// the diff. A pre-assigned ticket that was never claimed has not failed;
// a claimed step counts a failure. It returns the diff and the note.
func Expire(t *domain.Ticket, op string, extra map[string]any, now time.Time) (map[string]any, string, error) {
	d := map[string]any{"holder": t.Holder, "status": t.Status}
	for k, v := range extra {
		d[k] = v
	}
	claimed := t.Status == domain.StatusInProgress
	Reopen(t)
	if !claimed {
		return d, "", nil
	}
	how := op
	if r, ok := d["reason"].(string); ok && r != "" {
		how += ": " + r
	}
	note, err := StepFailure(t, how, d, now)
	return d, note, err
}

// ApplyUpdate applies u's field changes to t and returns the history
// diff: only what changed. Holder changes (pre-assignment) are allowed
// only while open.
func ApplyUpdate(t *domain.Ticket, u store.Update) (map[string]any, error) {
	if u.IfRevision != 0 && u.IfRevision != t.Revision {
		return nil, fmt.Errorf("%w: %s at revision %d, want %d", domain.ErrConflict, t.ID, t.Revision, u.IfRevision)
	}
	d := map[string]any{}
	set := func(name string, dst *string, v *string) {
		if v != nil && *v != *dst {
			*dst = *v
			d[name] = *v
		}
	}
	set("title", &t.Title, u.Title)
	set("body", &t.Body, u.Body)
	set("pool", &t.Pool, u.Pool)
	if u.Holder != nil && *u.Holder != t.Holder {
		if t.Status != domain.StatusOpen {
			return nil, fmt.Errorf("%w: %s: assign only while open (status %s)", domain.ErrInvalid, t.ID, t.Status)
		}
		t.Holder = *u.Holder
		d["holder"] = *u.Holder
	}
	if u.Priority != nil && *u.Priority != t.Priority {
		t.Priority = *u.Priority
		d["priority"] = *u.Priority
	}
	if u.DeferUntil != nil {
		t.DeferUntil = Stamp(*u.DeferUntil)
		d["defer_until"] = t.DeferUntil
	}
	if u.Fields != nil {
		f, err := domain.NormalizeFields(t.Kind, u.Fields)
		if err != nil {
			return nil, err
		}
		t.Fields = f
		d["fields"] = f
	}
	if u.Meta != nil {
		m, err := domain.NormalizeMeta(u.Meta)
		if err != nil {
			return nil, err
		}
		t.Meta = m
		d["meta"] = m
	}
	return d, nil
}

// Columns are the fields keys queries filter and sort on, copied to plain
// columns so SQL needs no JSON functions (#405). A key that is absent, or
// not the expected JSON type, is nil (NULL).
type Columns struct {
	Key, To, State, Workflow *string
	Index                    *int64
	Waiting                  bool
}

// FieldColumns derives Columns from a ticket's fields JSON.
func FieldColumns(fields json.RawMessage) Columns {
	var c Columns
	var m map[string]json.RawMessage
	if json.Unmarshal(fields, &m) != nil {
		return c
	}
	str := func(k string) *string {
		var s string
		if v, ok := m[k]; ok && json.Unmarshal(v, &s) == nil && bytes.HasPrefix(bytes.TrimSpace(v), []byte(`"`)) {
			return &s
		}
		return nil
	}
	c.Key, c.To, c.State, c.Workflow = str("key"), str("to"), str("state"), str("workflow")
	var f float64
	if v, ok := m["index"]; ok && json.Unmarshal(v, &f) == nil {
		i := int64(f)
		c.Index = &i
	}
	if v, ok := m["wait"]; ok && !bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
		c.Waiting = true
	}
	return c
}
