package knowledge

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/elee1766/jobshop/pkg/domain"
)

// The curator gate's judgement (#274), moved out of pkg/control by the
// JOB-155 audit. It reads values only: the controller lists the steps and
// learnings, passes them in and records the answer with SetResult.

const (
	// GateWorse is how much the failure rate may grow before an edit counts
	// as harmful: noise between windows of a few dozen steps.
	GateWorse = 1.25
	// GateSimilarMin is the word overlap (Jaccard) at which a learning
	// repeats a cited one.
	GateSimilarMin = 0.5
	// ReviewFindingPrefix starts the correction the controller records
	// when a review sends a step back; the gate counts those as findings.
	ReviewFindingPrefix = "review resumed "
)

// EditMetrics computes the metrics of steps in pools (empty: all) and
// learnings recorded in [from, to). Recurrence counts learnings that
// repeat one of cited (same kind, similar summary). Before the merge
// (after false) the cited learnings are the problem itself and count.
// After it (after true) a cited learning is not a recurrence, however
// it matches: the session that ships the fix records a learning about
// it later, and that is not the problem coming back (#1061).
func EditMetrics(steps []domain.Ticket, ls []domain.Learning, cited []domain.Learning, pools []string, from, to time.Time, after bool) domain.PoolMetrics {
	m := domain.PoolMetrics{From: from, To: to}
	inWin := func(t time.Time) bool { return !t.Before(from) && t.Before(to) }
	runs := map[string]bool{}
	ids := map[string]bool{}
	for _, s := range steps {
		if s.Kind != domain.KindStep || !inWin(s.CreatedAt) || (len(pools) > 0 && !slices.Contains(pools, s.Pool)) {
			continue
		}
		m.Steps++
		ids[s.ID] = true
		if s.ParentID != "" {
			runs[s.ParentID] = true
		}
		if sf, err := domain.DecodeFields[domain.StepFields](s); err == nil {
			m.Retries += sf.Failures
		}
		if s.Status == domain.StatusClosed && s.Resolution == "failed" {
			m.Escalations++
		}
	}
	m.Runs = len(runs)
	words := make([]map[string]bool, len(cited))
	citedID := map[int64]bool{}
	for i, c := range cited {
		words[i] = wordSet(c.Summary)
		citedID[c.ID] = true
	}
	for _, l := range ls {
		if !inWin(l.CreatedAt) {
			continue
		}
		if l.Kind == domain.LearningCorrection && strings.HasPrefix(l.Summary, ReviewFindingPrefix) && ids[l.Ticket] {
			m.Findings++
		}
		if after && citedID[l.ID] {
			continue
		}
		w := wordSet(l.Summary)
		for i, c := range cited {
			if c.ID == l.ID || (c.Kind == l.Kind && jaccard(words[i], w) >= GateSimilarMin) {
				m.Recurrence++
				break
			}
		}
	}
	return m
}

// DecideEdit compares before and after. The edit did not help when the
// cited problem recurs at least as often per step as before, or when
// retries, escalations and review findings per step grew by more than
// GateWorse. No steps after the merge is no evidence: it helped.
func DecideEdit(base, after domain.PoolMetrics) (string, string) {
	if after.Steps == 0 {
		return domain.EditHelped, "no steps in the scoped pools since the merge"
	}
	rb, ra := base.Rate(base.Recurrence), after.Rate(after.Recurrence)
	if after.Recurrence > 0 && ra >= rb {
		return domain.EditNotHelped, fmt.Sprintf("cited problem still recurs: %d in %d steps after, %d in %d before",
			after.Recurrence, after.Steps, base.Recurrence, base.Steps)
	}
	fb := base.Rate(base.Retries + base.Escalations + base.Findings)
	fa := after.Rate(after.Retries + after.Escalations + after.Findings)
	if fa > fb*GateWorse && fa > 0 && after.Retries+after.Escalations+after.Findings > 1 {
		return domain.EditNotHelped, fmt.Sprintf("failures per step rose from %.2f to %.2f (retries %d, escalations %d, findings %d in %d steps)",
			fb, fa, after.Retries, after.Escalations, after.Findings, after.Steps)
	}
	return domain.EditHelped, fmt.Sprintf("recurrence %d in %d steps (before %d in %d), failures per step %.2f (before %.2f)",
		after.Recurrence, after.Steps, base.Recurrence, base.Steps, fa, fb)
}

func wordSet(s string) map[string]bool {
	out := map[string]bool{}
	for _, w := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		if len(w) >= 4 {
			out[w] = true
		}
	}
	return out
}

func jaccard(a, b map[string]bool) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	n := 0
	for w := range a {
		if b[w] {
			n++
		}
	}
	return float64(n) / float64(len(a)+len(b)-n)
}
