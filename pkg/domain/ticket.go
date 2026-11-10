package domain

import (
	"encoding/json"
	"errors"
	"time"
)

// Status is a ticket's lifecycle state. "Blocked" is derived from edges and
// never stored.
type Status string

// Ticket statuses.
const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusClosed     Status = "closed"
)

// OpenStatuses are the statuses of work not yet closed.
var OpenStatuses = []Status{StatusOpen, StatusInProgress}

// EdgeKind is an engine-defined relation. Blocks and WaitsFor gate readiness.
type EdgeKind string

// Edge kinds.
const (
	EdgeBlocks   EdgeKind = "blocks"
	EdgeWaitsFor EdgeKind = "waits_for"
	EdgeTracks   EdgeKind = "tracks"
)

// Blocking reports whether the edge kind gates readiness.
func (k EdgeKind) Blocking() bool { return k == EdgeBlocks || k == EdgeWaitsFor }

// Valid reports whether k is an engine edge kind.
func (k EdgeKind) Valid() bool { return k == EdgeBlocks || k == EdgeWaitsFor || k == EdgeTracks }

// Ticket is one row of any kind. Fields holds engine data typed per kind;
// Meta holds pack data grouped by pack name and is never read by the engine.
type Ticket struct {
	ID         string          `json:"id"`
	Project    string          `json:"project"`
	Kind       Kind            `json:"kind"`
	Title      string          `json:"title,omitempty"`
	Body       string          `json:"body,omitempty"`
	Status     Status          `json:"status"`
	Resolution string          `json:"resolution,omitempty"`
	Priority   int             `json:"priority"`
	Pool       string          `json:"pool,omitempty"`
	Holder     string          `json:"holder,omitempty"`
	LeaseUntil time.Time       `json:"lease_until,omitzero"`
	Fence      int64           `json:"fence"`
	ParentID   string          `json:"parent_id,omitempty"`
	DeferUntil time.Time       `json:"defer_until,omitzero"`
	Ephemeral  bool            `json:"ephemeral,omitempty"`
	Fields     json.RawMessage `json:"fields,omitempty"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	Labels     []string        `json:"labels,omitempty"`
	Revision   int64           `json:"revision"`
	CreatedBy  string          `json:"created_by,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	ClosedAt   time.Time       `json:"closed_at,omitzero"`
}

// Edge is a typed relation. From is the dependent: (A, B, blocks) means A is
// blocked until B closes.
type Edge struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Kind      EdgeKind  `json:"kind"`
	CreatedAt time.Time `json:"created_at"`
}

// Comment is a note on a ticket.
type Comment struct {
	ID        int64     `json:"id"`
	TicketID  string    `json:"ticket_id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// History ops the controller records with store.Record for a session's
// agent turns on the tickets it holds (JOB-184). They change nothing on the
// ticket and do not bump its revision. turn_end carries TurnDiff.
const (
	OpTurnStart = "turn_start"
	OpTurnEnd   = "turn_end"
)

// TurnDiff is the diff of a turn_end row: how the turn stopped.
type TurnDiff struct {
	Stop string `json:"stop"`
}

// Errors returned by every store backend and by the validation of the types
// in this package. Callers match with errors.Is. The messages keep the
// "store:" prefix they have always had: they reach users and logs, and
// moving the values here must not change them.
var (
	ErrNotFound     = errors.New("store: not found")
	ErrConflict     = errors.New("store: revision conflict")
	ErrNotHolder    = errors.New("store: caller does not hold the claim")
	ErrNoWork       = errors.New("store: no claimable ticket")
	ErrInvalid      = errors.New("store: invalid input")
	ErrCycle        = errors.New("store: edge would create a blocking cycle")
	ErrClosed       = errors.New("store: ticket is closed")
	ErrAlreadyExist = errors.New("store: already exists")
)
