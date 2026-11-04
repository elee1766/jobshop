package store

import "time"

// WriteKey is the idempotency record of one tracker write (#280).
// Pending (Done false) from the claim until the write's result is stored.
type WriteKey struct {
	Key       string    `json:"key"`
	Op        string    `json:"op"`
	Result    string    `json:"result,omitempty"`
	Done      bool      `json:"done"`
	Actor     string    `json:"actor,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// New is set by ClaimWriteKey when this call inserted the key. A
	// pending key that is not new had an earlier attempt with an unknown
	// outcome.
	New bool `json:"new,omitempty"`
	// Scoped is the project-scoped name the controller stores the key
	// under; GitHub markers carry it (#415).
	Scoped string `json:"scoped,omitempty"`
}
