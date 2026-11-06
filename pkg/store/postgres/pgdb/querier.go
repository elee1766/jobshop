package pgdb

import (
	"context"
)

// Querier is the query set the store core runs; backends wrap it (Driver.Bind).
type Querier interface {
	AddTicketMeta(ctx context.Context, arg AddTicketMetaParams) error
	AddTicketResource(ctx context.Context, arg AddTicketResourceParams) error
	// BlockingPathExists reports whether start reaches target through blocking
	// edges. Adding (A blocked on B) would close a cycle if B already reaches A.
	BlockingPathExists(ctx context.Context, arg BlockingPathExistsParams) (bool, error)
	// ClaimWriteTicket is WriteTicket for a claim (#406): the not-blocked
	// check runs in the same statement as the write, so a blocker that
	// reopens between the caller's read and this write fails the claim
	// instead of leaving a blocked ticket claimed. Blockers are other rows,
	// which the revision check alone does not cover.
	ClaimWriteTicket(ctx context.Context, arg ClaimWriteTicketParams) (int64, error)
	// ClearTicketMeta and AddTicketMeta keep ticket_meta in step with meta.
	ClearTicketMeta(ctx context.Context, ticketID string) error
	// ClearTicketResources and AddTicketResource keep ticket_resources in
	// step with fields.resources (#426).
	ClearTicketResources(ctx context.Context, ticketID string) error
	DeleteEdge(ctx context.Context, arg DeleteEdgeParams) (int64, error)
	DeleteLabel(ctx context.Context, arg DeleteLabelParams) (int64, error)
	DeleteResourceCaps(ctx context.Context) error
	DeleteResourceProviders(ctx context.Context) error
	DeleteSlot(ctx context.Context, arg DeleteSlotParams) error
	// ExpireWriteTicket reopens a ticket only if its lease is still expired
	// at write time (#406). Heartbeat renews a lease without a revision
	// bump, so the revision check alone would let a sweep reopen a ticket
	// whose holder just renewed.
	ExpireWriteTicket(ctx context.Context, arg ExpireWriteTicketParams) (int64, error)
	ExpiredLeases(ctx context.Context, now int64) ([]string, error)
	GetSlot(ctx context.Context, arg GetSlotParams) (ResourceSlot, error)
	GetTicket(ctx context.Context, id string) (Ticket, error)
	// Heartbeat renews a lease without a revision bump or history row.
	Heartbeat(ctx context.Context, arg HeartbeatParams) (int64, error)
	// HeldBy covers claimed and pre-assigned tickets, so a dead session strands neither.
	HeldBy(ctx context.Context, holder string) ([]string, error)
	InsertComment(ctx context.Context, arg InsertCommentParams) (Comment, error)
	InsertEdge(ctx context.Context, arg InsertEdgeParams) error
	InsertHistory(ctx context.Context, arg InsertHistoryParams) error
	InsertLabel(ctx context.Context, arg InsertLabelParams) (int64, error)
	InsertResourceCap(ctx context.Context, arg InsertResourceCapParams) error
	InsertResourceProvider(ctx context.Context, name string) error
	InsertSlot(ctx context.Context, arg InsertSlotParams) error
	InsertTicket(ctx context.Context, arg InsertTicketParams) error
	BumpSeq(ctx context.Context, prefix string) (int64, error)
	StartSeq(ctx context.Context, prefix string, last int64) error
	TicketExists(ctx context.Context, id string) (bool, error)
	TicketIDsWithPrefix(ctx context.Context, p string) ([]string, error)
	LabelsFor(ctx context.Context, ids []string) ([]Label, error)
	ListEdges(ctx context.Context, id string) ([]Edge, error)
	ListSlots(ctx context.Context) ([]ResourceSlot, error)
	// MarkSlotsReleasing moves held slots whose claim ended (the ticket is
	// not in progress, or was claimed again with a new fence) to releasing.
	MarkSlotsReleasing(ctx context.Context, updatedAt int64) error
	// ProviderDemand is the waiting side of a provider's demand (#407): open
	// tickets naming it and the oldest one's created_at.
	ProviderDemand(ctx context.Context, pool string) (ProviderDemandRow, error)
	// ProviderEmpty reports whether ticket id needs a provider resource with
	// no ready slot (#407), for Claim by id. Same rule as ReadyTickets.
	ProviderEmpty(ctx context.Context, id string) (int64, error)
	ResourceCaps(ctx context.Context) ([]ResourceCap, error)
	// ResourceFull reports whether ticket id names a resource key with no
	// free slot (#400), for Claim by id. Same rule as ReadyTickets.
	ResourceFull(ctx context.Context, id string) (int64, error)
	// ResourceUse counts in-progress tickets per resource key (#400).
	ResourceUse(ctx context.Context) ([]ResourceUseRow, error)
	// ResourceWaiters lists open, unblocked, non-deferred tickets per
	// resource key they name, in claim order (#408). The caller keeps the
	// keys that are full: those tickets wait on the resource alone.
	ResourceWaiters(ctx context.Context, now int64) ([]ResourceWaitersRow, error)
	// ResourceWaiting counts open tickets per resource key (#400).
	ResourceWaiting(ctx context.Context) ([]ResourceWaitingRow, error)
	// SetParent re-parents a ticket (batch adoption, #169). The caller bumps
	// the revision and writes history through mutate.
	SetParent(ctx context.Context, arg SetParentParams) error
	// TakeSlot holds pool's longest-idle ready slot for a claim (#407).
	TakeSlot(ctx context.Context, arg TakeSlotParams) (ResourceSlot, error)
	// TicketProviderResources lists the provider resources ticket id needs.
	TicketProviderResources(ctx context.Context, id string) ([]string, error)
	UpdateSlot(ctx context.Context, arg UpdateSlotParams) error
	// WriteTicket stores every mutable column, guarded by the previous revision.
	WriteTicket(ctx context.Context, arg WriteTicketParams) (int64, error)
}

var _ Querier = (*Queries)(nil)
