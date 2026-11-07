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

// RecallQuery selects comment hits. Empty fields match everything.
type RecallQuery struct {
	Text     string
	Project  string
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

// Parts selects what Thread reads.
type Parts struct {
	Edges, Comments, History bool
}

// AllParts reads the whole thread.
var AllParts = Parts{Edges: true, Comments: true, History: true}

// Thread is a ticket's edges, comments and history. A part not asked
// for is nil.
type Thread struct {
	Edges    []Edge
	Comments []Comment
	History  []HistoryEntry
}

// ChangePage is one page of the global history and the newest seq at
// the time of the read.
type ChangePage struct {
	Changes []HistoryEntry
	Last    int64
}

// Claim is what a holder needs to heartbeat, close or release a ticket.
type Claim struct {
	TicketID   string    `json:"ticket_id"`
	Holder     string    `json:"holder"`
	Fence      int64     `json:"fence"`
	LeaseUntil time.Time `json:"lease_until"`
}

// Outcome says how Release lets go of a ticket. Close set closes it
// with that resolution; Pool set hands it off to that pool with Note as
// a comment; otherwise it returns to open, deferred to Until when set
// (not a step failure) and counted as a failed attempt with Reason when
// not. Close and Pool are exclusive.
//
// Reopen goes with Close ResolutionRequestChanges on a step (JOB-98): the
// step of the same run that the request reopens in the same transaction,
// with Note as the findings. Past the closing step's review round cap
// nothing is reopened.
type Outcome struct {
	// StepResult is applied only after the closing claim has been fenced.
	StepResult *StepResult
	Close      string
	Pool       string
	Note       string
	Until      time.Time
	Reason     string
	Reopen     string
}

// StepResult is an attempt's structured result, committed with its close.
type StepResult struct {
	Revision int64
	Outcome  string
	Output   map[string]any
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

// Graph is the input to Apply. It runs in this order: check Unclaimed,
// insert Tickets (a parent before its children), re-parent, remove
// RemoveEdges (a missing edge is ErrNotFound), add Edges.
type Graph struct {
	// Revisions checks the entire read set before any graph mutation.
	Revisions map[string]int64
	Children  map[string]int64
	// Fields and Close are applied atomically with newly created tickets.
	Fields      map[string]json.RawMessage
	Close       map[string]string
	Tickets     []NewTicket
	Reparent    map[string]string // existing ticket id -> new parent id
	Edges       []Edge
	RemoveEdges []Edge
	Unclaimed   []string
}

// Update changes mutable, non-claim fields, labels and comments, and
// reopens a closed ticket, in one transaction (#579). Nil pointers and
// empty lists are left unchanged. Other claim state (holder, lease,
// fence) changes only through Claim, Release and Expire.
//
// Order: Reopen, then the fields, then labels, then Comment. Label and
// comment changes do not bump the revision; each records its own
// history row (label_add, label_remove, comment).
type Update struct {
	// Reopen returns a closed ticket to open; ErrInvalid if not closed.
	Reopen bool
	// AddLabels adds labels (existing ones are a no-op); RemoveLabels
	// removes them (missing ones are a no-op).
	AddLabels    []string
	RemoveLabels []string
	// Comment, when non-empty, appends a comment by the actor.
	Comment string

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
	// Query matches literal case-insensitive terms in id, title, body, fields and metadata.
	Query   string
	Project string
	// Kind and Kinds match any of the kinds named.
	Kind  Kind
	Kinds []Kind
	// Status and Statuses match any of the statuses named.
	Status   Status
	Statuses []Status
	Pool     string
	// PoolNotIn keeps tickets with no pool or a pool not in the list.
	PoolNotIn []string
	// Unheld keeps tickets with no holder.
	Unheld bool
	// Ready keeps open, unblocked, non-deferred, non-waiting tickets
	// whose resources have a free slot: the rule ClaimNext uses. With no
	// Kind it means the claimable kinds.
	Ready    bool
	Holder   string
	ParentID string
	Label    string
	// CreatedBy matches the actor that created the ticket (a session id,
	// "admin" or "controller").
	CreatedBy string
	// MetaPack and MetaKey with MetaValue match tickets whose
	// meta[MetaPack][MetaKey] is the string MetaValue.
	MetaPack, MetaKey, MetaValue string
	// Meta adds more meta matches, all required (#541). With MetaPack
	// at most MaxMetaFilters in total.
	Meta []MetaMatch
	// Key, Workflow and State match fields.key, fields.workflow and
	// fields.state.
	Key, Workflow, State string
	// Since keeps tickets updated at or after it.
	Since time.Time
	// Created keeps tickets created in the range.
	Created *TimeRange
	// After is a cursor: the id of the last ticket of the previous page,
	// in the same order. OrderIndex has none.
	After string
	// Order is one of the Order constants; empty is OrderPriority.
	Order Order
	// Limit 0 means DefaultLimit; NoLimit means none.
	Limit int
}

// TimeRange is [Since, Until); a zero end is open.
type TimeRange struct{ Since, Until time.Time }

// Order is a List order.
type Order string

// List orders. Each ends with id, so a cursor is exact.
const (
	OrderPriority Order = "priority" // priority, then oldest
	OrderNewest   Order = "newest"   // created_at descending
	OrderOldest   Order = "oldest"   // created_at ascending
	OrderIndex    Order = "index"    // fields.index, then oldest (a run's steps)
)

// NoLimit as Filter.Limit returns every match.
const NoLimit = -1

// MetaMatch is one meta[Pack][Key] == Value condition.
type MetaMatch struct{ Pack, Key, Value string }

// MaxMetaFilters bounds the meta matches in one List.
const MaxMetaFilters = 3

// DefaultLimit bounds List when no limit is given.
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
	Get(ctx context.Context, id string) (Ticket, error)
	// List returns tickets matching f. It replaces the per-shape reads:
	// ready, orphaned, children, runs, by key, by creation time (#579).
	List(ctx context.Context, f Filter) ([]Ticket, error)
	Update(ctx context.Context, actor, id string, u Update) (Ticket, error)

	// SetResources replaces the resource capacities that gate
	// List(Filter{Ready: true}) and ClaimNext (#400) and the provider
	// resource names (#407), in one transaction.
	SetResources(ctx context.Context, r ResourceConfig) error
	// Slot applies one provider slot event (add, remove, unhealthy,
	// reset) and returns the slot. Unhealthy on a held slot reopens its
	// claim without counting a step failure and returns the ticket id.
	Slot(ctx context.Context, actor string, e SlotEvent) (Slot, string, error)
	// ReleaseSlots moves held slots whose claim ended to releasing, then
	// returns every releasing slot.
	ReleaseSlots(ctx context.Context) ([]Slot, error)
	// Resources reports key use, resource waits and slots; with a pool,
	// only that pool's slots, plus its demand.
	Resources(ctx context.Context, pool string) (ResourceReport, error)

	// ClaimNext atomically claims the next ticket: the holder's in-progress
	// work, then tickets pre-assigned to it, then the pool queue.
	ClaimNext(ctx context.Context, req ClaimNext) (Ticket, Claim, error)
	// Claim claims one specific open ticket.
	Claim(ctx context.Context, id, holder string, lease time.Duration) (Ticket, Claim, error)
	// Heartbeat extends a lease. It never bumps the revision or writes history.
	Heartbeat(ctx context.Context, c Claim, lease time.Duration) (Claim, error)
	// Release lets go of ticket id as o says: close, hand off (pool,
	// release, fence and note in one transaction, #492), or return to
	// open, deferred when o.Until is set (#139). A non-nil claim must
	// still hold the ticket; a nil claim is administrative and allowed
	// only for close and handoff. An empty actor is the claim's holder.
	Release(ctx context.Context, actor, id string, c *Claim, o Outcome) (Ticket, error)

	// Expire reopens what a dead holder held or was assigned, or, with
	// holder "", every ticket whose lease expired, and returns the ids.
	Expire(ctx context.Context, actor, holder, reason string) ([]string, error)

	// Thread reads the parts of a ticket's thread that p asks for: edges
	// in either direction, comments oldest first, history oldest first.
	// It does not check that id exists.
	Thread(ctx context.Context, id string, p Parts) (Thread, error)
	// Record adds a history row for an outside action on a ticket without
	// changing it (#542: desk_exec). The row carries the ticket's current
	// revision.
	Record(ctx context.Context, actor, id, op string, diff any) error
	// Changes returns history rows of all tickets with seq > after, oldest
	// first, at most n (0 means DefaultLimit), and the newest seq. It
	// feeds events.subscribe; n < 0 reads only Last.
	Changes(ctx context.Context, after int64, n int) (ChangePage, error)

	// Apply writes g in one transaction and returns the created tickets
	// in order (#579: replaces Create, CreateGraph, ApplyGraph, AddEdge
	// and RemoveEdge). Each Unclaimed id must be open and unheld inside
	// that transaction, or nothing is written and the error wraps
	// ErrNotHolder (#169: an edge added to a claimed step has no effect).
	Apply(ctx context.Context, actor string, g Graph) ([]Ticket, error)
	// NewID mints a ticket id for project, so a caller can wire parent ids
	// and edges before Apply.
	NewID(project string) (string, error)

	// Recall searches ticket comments, best match first. The backend
	// indexes comments as they are written; session journals and
	// learnings are collections with their own Search (#579).
	Recall(ctx context.Context, q RecallQuery) ([]RecallHit, error)

	// Collection registers a collection of records (#579) and returns
	// its handle. Registration is idempotent; call it once at startup.
	Collection(ctx context.Context, spec CollectionSpec) (Collection, error)

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

// LearningKinds are the default learning kinds, used when the config
// sets no learnings.kinds (#524). The controller's own learnings use
// correction and failure and are always allowed.
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
