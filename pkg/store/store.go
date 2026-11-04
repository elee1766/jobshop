package store

import (
	"context"
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

// Recall source kinds (#265).
const (
	RecallComment  = "comment"
	RecallJournal  = "journal"
	RecallLearning = "learning"
)

// RecallDoc is one searchable document. Ref identifies it within Kind: a
// comment id, "<session>:<seq>" for a journal turn, a learning id.
type RecallDoc struct {
	Kind     string
	Ref      string
	TicketID string
	Project  string
	At       time.Time
	Body     string
}

// RecallQuery selects recall hits. Empty fields match everything.
type RecallQuery struct {
	Text     string
	Project  string
	Kind     string
	TicketID string
	Limit    int // 0 means 10
}

// RecallHit is one ranked search result. Lower Score is a better match
// (SQLite bm25).
type RecallHit struct {
	Kind     string    `json:"kind"`
	Ref      string    `json:"ref"`
	TicketID string    `json:"ticket_id,omitempty"`
	Project  string    `json:"project,omitempty"`
	At       time.Time `json:"at,omitzero"`
	Snippet  string    `json:"snippet"`
	Score    float64   `json:"score"`
}

// HistoryEntry is one audit-log row.
type HistoryEntry struct {
	Seq      int64           `json:"seq"`
	TicketID string          `json:"ticket_id"`
	At       time.Time       `json:"at"`
	Actor    string          `json:"actor"`
	Op       string          `json:"op"`
	Revision int64           `json:"revision"`
	Diff     json.RawMessage `json:"diff,omitempty"`
}

// Claim is what a holder needs to heartbeat, close or release a ticket.
type Claim struct {
	TicketID   string    `json:"ticket_id"`
	Holder     string    `json:"holder"`
	Fence      int64     `json:"fence"`
	LeaseUntil time.Time `json:"lease_until"`
}

// NewTicket is the input to Create. ID, if empty, is generated.
type NewTicket struct {
	ID         string
	Project    string
	Kind       Kind
	Title      string
	Body       string
	Priority   *int // nil means DefaultPriority
	Pool       string
	Holder     string // pre-assignment; the ticket stays open
	ParentID   string
	DeferUntil time.Time
	Ephemeral  bool
	Fields     json.RawMessage
	Meta       json.RawMessage
	Labels     []string
}

// DefaultPriority is used when NewTicket.Priority is nil. Lower runs first.
const DefaultPriority = 2

// Graph is the input to ApplyGraph.
type Graph struct {
	Tickets   []NewTicket
	Reparent  map[string]string // existing ticket id -> new parent id
	Edges     []Edge
	Unclaimed []string
}

// Update changes mutable, non-claim fields. Nil pointers are left unchanged.
// Claim state (status, holder, lease, fence) changes only through Claim*,
// Release* and Close.
type Update struct {
	Title      *string
	Body       *string
	Priority   *int
	Pool       *string
	Holder     *string // pre-assign or clear; only while open
	DeferUntil *time.Time
	Fields     json.RawMessage // replaces fields when non-nil; validated per kind
	Meta       json.RawMessage // replaces meta when non-nil
	// IfRevision, when non-zero, makes the update conditional (compare-and-swap).
	IfRevision int64
}

// Filter selects tickets for List. Zero values match everything.
type Filter struct {
	Project  string
	Kind     Kind
	Status   Status
	Pool     string
	Holder   string
	ParentID string
	Label    string
	// CreatedBy matches the actor that created the ticket (a session id,
	// "admin" or "controller").
	CreatedBy string
	// MetaPack and MetaKey with MetaValue match tickets whose
	// meta[MetaPack][MetaKey] is the string MetaValue.
	MetaPack, MetaKey, MetaValue string
	// Newest orders newest first instead of by priority.
	Newest bool
	Limit  int // 0 means DefaultLimit
}

// DefaultLimit bounds List and Ready when no limit is given.
const DefaultLimit = 500

// ClaimNext describes a claim-next request (the hook).
type ClaimNext struct {
	Project string
	Pool    string
	Holder  string
	Kinds   []Kind // empty means ClaimableKinds
	Lease   time.Duration
}

// Errors returned by every backend. Callers match with errors.Is.
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

// Store is the only persistence interface. Every write records history
// (except heartbeats) and bumps the ticket revision.
type Store interface {
	Create(ctx context.Context, actor string, t NewTicket) (Ticket, error)
	Get(ctx context.Context, id string) (Ticket, error)
	List(ctx context.Context, f Filter) ([]Ticket, error)
	Update(ctx context.Context, actor, id string, u Update) (Ticket, error)

	// Ready returns open, unblocked, non-deferred claimable tickets.
	Ready(ctx context.Context, project, pool string, limit int) ([]Ticket, error)
	// Orphaned returns open, unclaimed claimable tickets whose pool is
	// empty or not in pools, so no session will ever claim them. Empty
	// project means every project.
	Orphaned(ctx context.Context, project string, pools []string, limit int) ([]Ticket, error)

	// ClaimNext atomically claims the next ticket: the holder's in-progress
	// work, then tickets pre-assigned to it, then the pool queue.
	ClaimNext(ctx context.Context, req ClaimNext) (Ticket, Claim, error)
	// Claim claims one specific open ticket.
	Claim(ctx context.Context, id, holder string, lease time.Duration) (Ticket, Claim, error)
	// Heartbeat extends a lease. It never bumps the revision or writes history.
	Heartbeat(ctx context.Context, c Claim, lease time.Duration) (Claim, error)
	// Release returns a held ticket to open. A non-zero until defers the
	// ticket to that time and is not counted as a step failure (#139).
	Release(ctx context.Context, c Claim, reason string, until time.Time) (Ticket, error)
	// CloseTicket closes a ticket. A non-nil claim must still hold it; a nil
	// claim is an administrative close.
	CloseTicket(ctx context.Context, actor, id string, c *Claim, resolution string) (Ticket, error)
	Reopen(ctx context.Context, actor, id string) (Ticket, error)

	// ReleaseHolder reopens everything held by a dead session (claims and
	// pre-assignments) and returns the affected ids.
	ReleaseHolder(ctx context.Context, actor, holder, reason string) ([]string, error)
	// SweepLeases reopens tickets whose lease has expired.
	SweepLeases(ctx context.Context, actor string) ([]string, error)

	AddEdge(ctx context.Context, actor string, e Edge) error
	RemoveEdge(ctx context.Context, actor string, e Edge) error
	Edges(ctx context.Context, id string) ([]Edge, error)

	AddLabel(ctx context.Context, actor, id, label string) error
	RemoveLabel(ctx context.Context, actor, id, label string) error

	AddComment(ctx context.Context, author, id, body string) (Comment, error)
	Comments(ctx context.Context, id string) ([]Comment, error)
	History(ctx context.Context, id string) ([]HistoryEntry, error)
	// Changes returns history rows of all tickets with seq > after, oldest
	// first, at most n (0 means DefaultLimit). It feeds events.subscribe.
	Changes(ctx context.Context, after int64, n int) ([]HistoryEntry, error)
	// LastSeq is the newest history seq.
	LastSeq(ctx context.Context) (int64, error)

	// CreateGraph inserts tickets and edges in one transaction.
	CreateGraph(ctx context.Context, actor string, tickets []NewTicket, edges []Edge) ([]Ticket, error)
	// ApplyGraph is CreateGraph that may also re-parent existing tickets,
	// in one transaction. Each Unclaimed id must be open and unheld inside
	// that transaction, or nothing is written and the error wraps
	// ErrNotHolder (#169: an edge added to a claimed step has no effect).
	ApplyGraph(ctx context.Context, actor string, g Graph) ([]Ticket, error)
	// NewID mints a ticket id for project, so a caller can wire parent ids
	// and edges before CreateGraph.
	NewID(project string) (string, error)

	// ListChildren returns tickets whose parent is id, ordered by
	// fields.index then age (a run's steps in order).
	ListChildren(ctx context.Context, id string) ([]Ticket, error)
	// FindOpenRunByKey returns the open run of workflow with key, or
	// ErrNotFound.
	FindOpenRunByKey(ctx context.Context, workflow, key string) (Ticket, error)
	// FindOpenBatchByKey returns the open batch of project with key, or
	// ErrNotFound.
	FindOpenBatchByKey(ctx context.Context, project, key string) (Ticket, error)
	// ListRuns returns runs newest first. Empty project or state matches all.
	ListRuns(ctx context.Context, project, state string, limit int) ([]Ticket, error)

	// AddLearning records a learning (#264). Kind and Source are
	// validated; Status defaults to open.
	AddLearning(ctx context.Context, l Learning) (Learning, error)
	// Learnings returns learnings newest first.
	Learnings(ctx context.Context, f LearningFilter) ([]Learning, error)
	// UpdateLearning sets a learning's status and PR (#273).
	UpdateLearning(ctx context.Context, id int64, status, pr string) (Learning, error)
	// GetLearning reads one learning; ErrNotFound if missing.
	GetLearning(ctx context.Context, id int64) (Learning, error)
	// CountLearnings counts learnings created at or after since with
	// status (empty: any).
	CountLearnings(ctx context.Context, since time.Time, status string) (int, error)
	// LearningsBetween returns learnings recorded in [since, until),
	// oldest first (#274).
	LearningsBetween(ctx context.Context, since, until time.Time) ([]Learning, error)
	// LearningsByPR returns the learnings whose pr is pr (#274).
	LearningsByPR(ctx context.Context, pr string) ([]Learning, error)
	// StepsCreated returns step tickets created in [since, until),
	// oldest first (#274).
	StepsCreated(ctx context.Context, since, until time.Time) ([]Ticket, error)
	// AddMemoryEdit records a merged curator PR; an already recorded PR
	// returns its row unchanged (#274).
	AddMemoryEdit(ctx context.Context, e MemoryEdit) (MemoryEdit, error)
	// MemoryEdits lists memory edits newest merge first; empty state is all.
	MemoryEdits(ctx context.Context, state string, limit int) ([]MemoryEdit, error)
	// SetMemoryEditResult stores the gate's decision and writes it on
	// each cited learning's gate.
	SetMemoryEditResult(ctx context.Context, id int64, state string, after PoolMetrics, reason string) (MemoryEdit, error)
	// ClaimWriteKey returns the write stored under key, inserting it
	// pending when new (#280).
	ClaimWriteKey(ctx context.Context, key, op, actor string) (WriteKey, error)
	// FinishWriteKey stores a pending write's result; a done key keeps
	// its first result.
	FinishWriteKey(ctx context.Context, key, result string) (WriteKey, error)
	// InboxPut stores an ingested item (#321). A delivery id already
	// stored is a no-op that returns the existing row and dup true.
	InboxPut(ctx context.Context, it InboxItem) (out InboxItem, dup bool, err error)
	// InboxClaim leases the oldest item in one of states whose lease is
	// empty or expired, adding one to attempts. ErrNotFound if none.
	InboxClaim(ctx context.Context, states []string, holder string, lease time.Duration) (InboxItem, error)
	// InboxRenew extends holder's live lease; ErrConflict if not held.
	InboxRenew(ctx context.Context, id int64, holder string, lease time.Duration) (InboxItem, error)
	// InboxSetState moves a held item to a new state and drops the
	// lease. ErrConflict if holder no longer holds a live lease.
	InboxSetState(ctx context.Context, id int64, holder string, u InboxUpdate) (InboxItem, error)
	// InboxGet returns one item.
	InboxGet(ctx context.Context, id int64) (InboxItem, error)
	// InboxList lists items newest first.
	InboxList(ctx context.Context, f InboxFilter) ([]InboxItem, error)
	// InboxPrune deletes terminal items received before before and
	// returns how many.
	InboxPrune(ctx context.Context, before time.Time) (int64, error)
	// IndexRecall adds or replaces (by kind and ref) one document in the
	// recall index. Comments are indexed by the backend itself.
	IndexRecall(ctx context.Context, d RecallDoc) error
	// Recall searches the index, best match first.
	Recall(ctx context.Context, q RecallQuery) ([]RecallHit, error)

	// Close releases backend resources.
	Close() error
}

// Learning kinds, sources and statuses (#264).
const (
	LearningCorrection = "correction"
	LearningFailure    = "failure"
	LearningWorkaround = "workaround"

	LearningSourceAgent      = "agent"
	LearningSourceController = "controller"

	LearningOpen     = "open"
	LearningApplied  = "applied"
	LearningRejected = "rejected"
)

// LearningKinds lists the valid learning kinds.
var LearningKinds = []string{LearningCorrection, LearningFailure, LearningWorkaround}

// Learning is a correction, failure or workaround recorded for the curator.
// Scope is "global", "repo:<owner/name>" or "pool:<name>".
type Learning struct {
	ID      int64  `json:"id"`
	Kind    string `json:"kind"`
	Summary string `json:"summary"`
	Fix     string `json:"fix,omitempty"`
	Scope   string `json:"scope"`
	Ticket  string `json:"ticket,omitempty"`
	Session string `json:"session,omitempty"`
	Source  string `json:"source"`
	Status  string `json:"status"`
	// PR is the curator's PR that addresses it (#273): pending while
	// open, kept when applied or rejected.
	PR string `json:"pr,omitempty"`
	// Gate is the curator gate's result on the edit that cited it
	// (#274): "helped: ..." or "not_helped: ...".
	Gate      string    `json:"gate,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// LearningFilter selects learnings. Zero values match everything.
type LearningFilter struct {
	Since  time.Time
	Scope  string
	Kind   string
	Status string
	Ticket string
	Limit  int // 0 means DefaultLimit
}
