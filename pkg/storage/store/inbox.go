package store

import "encoding/json"

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
