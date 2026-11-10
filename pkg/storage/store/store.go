package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gfx-labs/jobshop/pkg/domain"
)

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
	Edges    []domain.Edge
	Comments []domain.Comment
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
	Kind       domain.Kind
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
	Edges       []domain.Edge
	RemoveEdges []domain.Edge
	Unclaimed   []string
	// Comments are appended after the rest, so a comment may name a ticket
	// created in the same graph.
	Comments []GraphComment
}

// GraphComment is one comment a Graph adds.
type GraphComment struct {
	TicketID string
	Body     string
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
	Kind  domain.Kind
	Kinds []domain.Kind
	// Status and Statuses match any of the statuses named.
	Status   domain.Status
	Statuses []domain.Status
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
	Kinds   []domain.Kind // empty means ClaimableKinds
	Lease   time.Duration
}

// Store is the only persistence interface. Every write records history
// (except heartbeats) and bumps the ticket revision.
type Store interface {
	Get(ctx context.Context, id string) (domain.Ticket, error)
	// List returns tickets matching f. It replaces the per-shape reads:
	// ready, orphaned, children, runs, by key, by creation time (#579).
	List(ctx context.Context, f Filter) ([]domain.Ticket, error)
	Update(ctx context.Context, actor, id string, u Update) (domain.Ticket, error)

	// SetResources replaces the resource capacities that gate
	// List(Filter{Ready: true}) and ClaimNext (#400) and the provider
	// resource names (#407), in one transaction.
	SetResources(ctx context.Context, r domain.ResourceConfig) error
	// Slot applies one provider slot event (add, remove, unhealthy,
	// reset) and returns the slot. Unhealthy on a held slot reopens its
	// claim without counting a step failure and returns the ticket id.
	Slot(ctx context.Context, actor string, e domain.SlotEvent) (domain.Slot, string, error)
	// ReleaseSlots moves held slots whose claim ended to releasing, then
	// returns every releasing slot.
	ReleaseSlots(ctx context.Context) ([]domain.Slot, error)
	// Resources reports key use, resource waits and slots; with a pool,
	// only that pool's slots, plus its demand.
	Resources(ctx context.Context, pool string) (domain.ResourceReport, error)

	// ClaimNext atomically claims the next ticket: the holder's in-progress
	// work, then tickets pre-assigned to it, then the pool queue.
	ClaimNext(ctx context.Context, req ClaimNext) (domain.Ticket, Claim, error)
	// Claim claims one specific open ticket.
	Claim(ctx context.Context, id, holder string, lease time.Duration) (domain.Ticket, Claim, error)
	// Heartbeat extends a lease. It never bumps the revision or writes history.
	Heartbeat(ctx context.Context, c Claim, lease time.Duration) (Claim, error)
	// Release lets go of ticket id as o says: close, hand off (pool,
	// release, fence and note in one transaction, #492), or return to
	// open, deferred when o.Until is set (#139). A non-nil claim must
	// still hold the ticket; a nil claim is administrative and allowed
	// only for close and handoff. An empty actor is the claim's holder.
	Release(ctx context.Context, actor, id string, c *Claim, o Outcome) (domain.Ticket, error)

	// Expire reopens what a dead holder held or was assigned, or, with
	// holder "", every ticket whose lease expired, and returns the ids.
	Expire(ctx context.Context, actor, holder, reason string) ([]string, error)

	// Thread reads the parts of a ticket's thread that p asks for: edges
	// in either direction, comments oldest first, history oldest first.
	// It does not check that id exists.
	Thread(ctx context.Context, id string, p Parts) (Thread, error)
	// CommentsFor reads the comments of all ids in one query, oldest
	// first per ticket, keyed by ticket id. A ticket without comments
	// has no key. It does not check that an id exists.
	CommentsFor(ctx context.Context, ids []string) (map[string][]domain.Comment, error)
	// HeldAt lists the ids of the tickets actor may have held at time at,
	// by first claim: a superset, for the caller to confirm with each
	// ticket's hold windows. The trace uses it to connect a ticket to
	// those its creator held when it was created (JOB-190).
	HeldAt(ctx context.Context, actor string, at time.Time) ([]string, error)
	// Record adds a history row for an outside action on a ticket without
	// changing it (#542: desk_exec). The row carries the ticket's current
	// revision.
	Record(ctx context.Context, actor, id, op string, diff any) error
	// Changes returns history rows of all tickets with seq > after, oldest
	// first, at most n (0 means DefaultLimit), and the newest seq. It
	// feeds events.subscribe; n < 0 reads only Last.
	Changes(ctx context.Context, after int64, n int) (ChangePage, error)

	// CountRuns counts the distinct runs that own a step created in r,
	// in pools (any pool when empty): the run count of the memory-edit
	// gate, without loading the steps (JOB-191).
	CountRuns(ctx context.Context, pools []string, r TimeRange) (int, error)

	// Maintain refreshes what the backend's planner needs to pick
	// indexes (SQLite statistics, JOB-191). It is cheap when nothing
	// changed much; the controller calls it on a slow interval. Open
	// already ran it once.
	Maintain(ctx context.Context) error

	// Apply writes g in one transaction and returns the created tickets
	// in order (#579: replaces Create, CreateGraph, ApplyGraph, AddEdge
	// and RemoveEdge). Each Unclaimed id must be open and unheld inside
	// that transaction, or nothing is written and the error wraps
	// ErrNotHolder (#169: an edge added to a claimed step has no effect).
	Apply(ctx context.Context, actor string, g Graph) ([]domain.Ticket, error)
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

// LearningFilter selects learnings. Zero values match everything.
type LearningFilter struct {
	Since  time.Time
	Scope  string
	Kind   string
	Status string
	Ticket string
	Limit  int // 0 means DefaultLimit
}
