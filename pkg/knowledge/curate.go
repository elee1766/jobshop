package knowledge

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/knowledge/memory"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// The mechanics of curation. pkg/orchestration/control decides who may write memory,
// which cap and thresholds apply and when an evaluation runs, works out
// which pools a learning points at from its tickets, and logs. The calls
// here take those answers as arguments.

// StepStore is a Store that can also list tickets, which the gate reads to
// measure a window of steps, and count the runs in one.
type StepStore interface {
	Store
	List(ctx context.Context, f store.Filter) ([]domain.Ticket, error)
	CountRuns(ctx context.Context, pools []string, r store.TimeRange) (int, error)
}

// Measure reads the steps and learnings of [from, to) and computes the
// metrics of pools: the window after an edit's merge when after is true,
// else the baseline before it.
func Measure(ctx context.Context, st StepStore, cited []domain.Learning, pools []string, from, to time.Time, after bool) (domain.PoolMetrics, error) {
	steps, err := st.List(ctx, store.Filter{Kind: domain.KindStep, Created: &store.TimeRange{Since: from, Until: to}, Order: store.OrderOldest, Limit: store.NoLimit})
	if err != nil {
		return domain.PoolMetrics{}, err
	}
	ls, err := LearningsBetween(ctx, st, from, to)
	if err != nil {
		return domain.PoolMetrics{}, err
	}
	return EditMetrics(steps, ls, cited, pools, from, to, after), nil
}

// CitedForPR returns the learnings a merged PR cites: those named by ids,
// or, with no ids, the ones whose pr is pr.
func CitedForPR(ctx context.Context, st Store, pr string, ids []int64) ([]domain.Learning, error) {
	if len(ids) == 0 {
		return LearningsByPR(ctx, st, pr)
	}
	return GetLearnings(ctx, st, ids)
}

// RecordPREdit records a merged curator PR with its baseline, the metrics
// of pools over the baseline window before mergedAt.
func RecordPREdit(ctx context.Context, st StepStore, pr string, cited []domain.Learning, pools []string, mergedAt time.Time, baseline time.Duration) (domain.MemoryEdit, error) {
	base, err := Measure(ctx, st, cited, pools, mergedAt.Add(-baseline), mergedAt, false)
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	e := domain.MemoryEdit{PR: pr, Pools: pools, MergedAt: mergedAt, Baseline: base, Learnings: learningIDs(cited)}
	return AddPREdit(ctx, st, e)
}

// EditDue reports whether a pending edit has enough evidence: age since
// the merge, or runs runs in its pools since it. The age is checked first
// with no query, and the runs are counted by the store, not by loading
// every step and learning of the window (JOB-191): the gate polls this.
// Measure takes the metrics once an edit is due.
func EditDue(ctx context.Context, st StepStore, e domain.MemoryEdit, runs int, age time.Duration, now time.Time) (bool, error) {
	if !now.Before(e.MergedAt.Add(age)) || runs <= 0 {
		return true, nil
	}
	n, err := CountEditRuns(ctx, st, e, now)
	return n >= runs, err
}

// CountEditRuns counts the runs with a step in e's pools between its merge
// and now: PoolMetrics.Runs of the window after the edit.
func CountEditRuns(ctx context.Context, st StepStore, e domain.MemoryEdit, now time.Time) (int, error) {
	return st.CountRuns(ctx, e.Pools, store.TimeRange{Since: e.MergedAt, Until: now})
}

// MeasureEdit measures the window after e's merge, to settle it.
func MeasureEdit(ctx context.Context, st StepStore, e domain.MemoryEdit, now time.Time) (domain.PoolMetrics, error) {
	cited, err := GetLearnings(ctx, st, e.Learnings)
	if err != nil {
		return domain.PoolMetrics{}, err
	}
	return Measure(ctx, st, cited, e.Pools, e.MergedAt, now, true)
}

// SettleEdit decides e from its baseline and the measured after, stores
// the decision on the edit and copies it to each cited learning's gate.
//
// The edit's result is the record; the gate line on each cited learning is
// a copy for `learn list`, so a failure on a learning leaves the edit
// decided and the next evaluation does not redo it.
func SettleEdit(ctx context.Context, st Store, e domain.MemoryEdit, after domain.PoolMetrics, now time.Time) (domain.MemoryEdit, error) {
	state, reason := DecideEdit(e.Baseline, after)
	u, err := SetEditResult(ctx, st, e.ID, state, after, reason, now)
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	gate := fmt.Sprintf("%s: %s (%s)", state, reason, EditRef(u))
	for _, l := range u.Learnings {
		if err := SetLearningGate(ctx, st, l, gate); err != nil {
			return domain.MemoryEdit{}, err
		}
	}
	return u, nil
}

// LocalMemory serializes local memory writes and restores (#592), so a
// write's before text is the file it replaced. The zero value is ready.
// The files are not in the store's transaction: a write puts the file
// down, records the edit, and undoes the file if the record fails.
type LocalMemory struct {
	mu sync.Mutex
}

// LocalWrite is one write to a memory file. Dir is the memory directory,
// Rel the file under it (memory.Target), Cap the character cap of a memory
// file (0 means memory.DefaultCap). Cited are the learnings the write
// answers and Pools the pools they point at.
type LocalWrite struct {
	Dir, Rel, Text string
	Cap            int
	Cited          []domain.Learning
	Pools          []string
	Now            time.Time
	Baseline       time.Duration
}

// Write writes the file, records the edit with the text it replaced and
// its baseline, and marks the cited learnings applied. memory.ErrOverCap
// reports a file over the cap, with nothing written or recorded. A learning
// that cannot be marked leaves the file and the edit in place.
func (m *LocalMemory) Write(ctx context.Context, st StepStore, w LocalWrite) (domain.MemoryEdit, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	before, existed, err := memory.ReadRaw(w.Dir, w.Rel)
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	base, err := Measure(ctx, st, w.Cited, w.Pools, w.Now.Add(-w.Baseline), w.Now, false)
	if err != nil {
		return domain.MemoryEdit{}, err
	}
	if err := memory.Write(w.Dir, w.Rel, w.Text, w.Cap); err != nil {
		return domain.MemoryEdit{}, err
	}
	e := domain.MemoryEdit{Path: w.Rel, Before: before, Existed: existed, Pools: w.Pools, MergedAt: w.Now, Baseline: base, Learnings: learningIDs(w.Cited)}
	out, err := AddWriteEdit(ctx, st, e)
	if err != nil {
		// Keep the file and the record in step: undo the write.
		_ = memory.Restore(w.Dir, w.Rel, before, existed)
		return domain.MemoryEdit{}, err
	}
	for _, id := range out.Learnings {
		if _, err := UpdateLearning(ctx, st, id, domain.LearningApplied, MemoryRef(out.ID)); err != nil {
			return domain.MemoryEdit{}, err
		}
	}
	return out, nil
}

// Restored says what Restore did.
type Restored struct {
	// Newer is the id of a later write to the same path, which wins: the
	// file was left alone. Zero when the file was put back.
	Newer int64
}

// Restore undoes a local write the gate found not helping: the file gets
// its previous text back, or is removed when the write created it, and the
// cited learnings become rejected so the curator does not write the same
// thing again. A newer write to the same path wins: restoring under it
// would drop that write, so the file is skipped. A PR edit has no path and
// nothing to restore.
func (m *LocalMemory) Restore(ctx context.Context, st Store, dir string, e domain.MemoryEdit) (Restored, error) {
	if e.Path == "" {
		return Restored{}, nil
	}
	newer, err := ListEdits(ctx, st, "", 0)
	if err != nil {
		return Restored{}, err
	}
	for _, n := range newer {
		if n.Path == e.Path && n.ID > e.ID {
			return Restored{Newer: n.ID}, rejectCited(ctx, st, e)
		}
	}
	m.mu.Lock()
	err = memory.Restore(dir, e.Path, e.Before, e.Existed)
	m.mu.Unlock()
	if err != nil {
		return Restored{}, err
	}
	return Restored{}, rejectCited(ctx, st, e)
}

func rejectCited(ctx context.Context, st Store, e domain.MemoryEdit) error {
	for _, id := range e.Learnings {
		if _, err := UpdateLearning(ctx, st, id, domain.LearningRejected, MemoryRef(e.ID)); err != nil {
			return err
		}
	}
	return nil
}

func learningIDs(ls []domain.Learning) []int64 {
	var out []int64
	for _, l := range ls {
		out = append(out, l.ID)
	}
	return out
}

// IsOverCap reports whether err is a write refused for the cap.
func IsOverCap(err error) bool { return errors.Is(err, memory.ErrOverCap) }
