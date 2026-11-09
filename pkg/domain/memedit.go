package domain

import "time"

// Memory edit states (#274): pending until `learn evaluate` decides.
const (
	EditPending   = "pending"
	EditHelped    = "helped"
	EditNotHelped = "not_helped"
)

// PoolMetrics is what the curator gate compares before and after a
// memory edit merges, over the steps of the scoped pools created in a
// window and the learnings recorded in it.
type PoolMetrics struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
	// Runs is the number of distinct runs with a step in the window.
	Runs  int `json:"runs"`
	Steps int `json:"steps"`
	// Retries counts failed attempts (fields.failures summed).
	Retries int `json:"retries"`
	// Escalations counts steps that closed failed at their retry limit.
	Escalations int `json:"escalations"`
	// Findings counts review findings ("review resumed" corrections).
	Findings int `json:"findings"`
	// Recurrence counts learnings in the window that repeat a cited one.
	Recurrence int `json:"recurrence"`
}

// Rate divides n by Steps (0 with no steps).
func (m PoolMetrics) Rate(n int) float64 {
	if m.Steps == 0 {
		return 0
	}
	return float64(n) / float64(m.Steps)
}

// MemoryEdit is a memory change the gate evaluates (#274): a merged
// curator PR (PR set) or, since #592, a local write under <root>/memory
// (Path set, Before the text it replaced, Existed false for a new file).
type MemoryEdit struct {
	ID          int64       `json:"id"`
	PR          string      `json:"pr,omitempty"`
	Path        string      `json:"path,omitempty"`
	Before      string      `json:"before,omitempty"`
	Existed     bool        `json:"existed,omitempty"`
	Learnings   []int64     `json:"learnings"`
	Pools       []string    `json:"pools"`
	MergedAt    time.Time   `json:"merged_at"`
	Baseline    PoolMetrics `json:"baseline"`
	State       string      `json:"state"`
	After       PoolMetrics `json:"after,omitzero"`
	Reason      string      `json:"reason,omitempty"`
	EvaluatedAt time.Time   `json:"evaluated_at,omitzero"`
	CreatedAt   time.Time   `json:"created_at"`
}
