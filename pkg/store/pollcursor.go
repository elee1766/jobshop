package store

import "time"

// PollCursor is an ingest poller's read position on one repo and event
// stream (#341). Since is the newest updated_at handled; ETag is the
// last response's, sent as If-None-Match.
type PollCursor struct {
	Poller string    `json:"poller"`
	Repo   string    `json:"repo"`
	Stream string    `json:"stream"`
	Since  time.Time `json:"since,omitzero"`
	ETag   string    `json:"etag,omitempty"`
	// Found is false when no cursor was stored yet.
	Found bool `json:"-"`
}
