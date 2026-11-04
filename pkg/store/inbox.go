package store

import (
	"encoding/json"
	"time"
)

// Inbox states (#321). pending is a stored, not yet admitted item;
// dispatched, denied and failed are terminal and pruned after retention.
const (
	InboxPending    = "pending"
	InboxAdmitted   = "admitted"
	InboxDenied     = "denied"
	InboxHeld       = "held"
	InboxDispatched = "dispatched"
	InboxFailed     = "failed"
)

// InboxStates lists every inbox state.
var InboxStates = []string{InboxPending, InboxAdmitted, InboxDenied, InboxHeld, InboxDispatched, InboxFailed}

// InboxItem is one ingested webhook delivery or poller find (#318).
type InboxItem struct {
	ID         int64           `json:"id"`
	Source     string          `json:"source"` // github or linear
	DeliveryID string          `json:"delivery_id"`
	Kind       string          `json:"kind,omitempty"`
	RepoOrTeam string          `json:"repo_or_team,omitempty"`
	Number     int64           `json:"number,omitempty"`
	CommentID  string          `json:"comment_id,omitempty"`
	Actor      string          `json:"actor,omitempty"`
	ReceivedAt time.Time       `json:"received_at"`
	Payload    json.RawMessage `json:"payload"`
	// Admitted is the payload after admission hooks; nil until admitted.
	Admitted    json.RawMessage `json:"admitted,omitempty"`
	State       string          `json:"state"`
	HookResults json.RawMessage `json:"hook_results"`
	Holder      string          `json:"holder,omitempty"`
	LeaseUntil  time.Time       `json:"lease_until,omitzero"`
	Attempts    int64           `json:"attempts"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// InboxFilter narrows InboxList. Empty fields match all; Limit 0 is 100.
type InboxFilter struct {
	State  string
	Source string
	Limit  int
}

// InboxUpdate is an InboxSetState change. Nil Admitted or HookResults
// keep the stored value.
type InboxUpdate struct {
	State       string
	Admitted    json.RawMessage
	HookResults json.RawMessage
}
