package sqldb

import (
	"database/sql"
)

// Comment is a row of the comment table.
type Comment struct {
	ID        int64
	TicketID  string
	Author    string
	Body      string
	CreatedAt int64
	EditedAt  int64
	Deleted   int64
}

// Edge is a row of the edge table.
type Edge struct {
	FromID    string
	ToID      string
	Kind      string
	CreatedAt int64
}

// FtsJournal is a row of the ftsjournal table.
type FtsJournal struct {
	Key  string
	Body string
}

// FtsLearning is a row of the ftslearning table.
type FtsLearning struct {
	Key  string
	Body string
}

// History is a row of the history table.
type History struct {
	Seq      int64
	TicketID string
	At       int64
	Actor    string
	Op       string
	Revision int64
	Diff     string
}

// Label is a row of the label table.
type Label struct {
	TicketID string
	Label    string
}

// Recall is a row of the recall table.
type Recall struct {
	Body     string
	Kind     string
	Ref      string
	TicketID string
	Project  string
	At       string
}

// Record is a row of the record table.
type Record struct {
	Collection string
	Key        string
	State      string
	Value      string
	Revision   int64
	Holder     string
	LeaseUntil int64
	Attempts   int64
	CreatedAt  int64
	UpdatedAt  int64
}

// ResourceCap is a row of the resourcecap table.
type ResourceCap struct {
	Name     string
	Capacity int64
}

// ResourceProvider is a row of the resourceprovider table.
type ResourceProvider struct {
	Name string
}

// ResourceSlot is a row of the resourceslot table.
type ResourceSlot struct {
	Pool      string
	ID        string
	State     string
	Env       string
	Ticket    string
	Fence     int64
	Removing  int64
	UpdatedAt int64
}

// Ticket is a row of the ticket table.
type Ticket struct {
	ID         string
	Project    string
	Kind       string
	Title      string
	Body       string
	Status     string
	Resolution string
	Priority   int64
	Pool       string
	Holder     string
	LeaseUntil int64
	Fence      int64
	ParentID   sql.NullString
	DeferUntil int64
	Ephemeral  int64
	Fields     string
	Meta       string
	Revision   int64
	CreatedBy  string
	CreatedAt  int64
	UpdatedAt  int64
	ClosedAt   int64
	FKey       sql.NullString
	FTo        sql.NullString
	FState     sql.NullString
	FWorkflow  sql.NullString
	FIndex     sql.NullInt64
	FWaiting   int64
}

// TicketMetum is a row of the ticketmetum table.
type TicketMetum struct {
	TicketID string
	Pack     string
	Mkey     string
	Value    string
}

// TicketResource is a row of the ticketresource table.
type TicketResource struct {
	TicketID string
	Rkey     string
	Rname    string
}
