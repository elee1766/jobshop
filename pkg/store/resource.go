package store

import "strings"

// Resource pools (#400): a step names resource keys in fields.resources.
// A key is a resource name ("emulator") or a keyed form "name:key"
// ("pr:acme/widget#12") that takes the capacity of "name" per key.
// Occupancy is the in-progress tickets naming a key, so a slot frees on
// every claim end (close, release, lease expiry, dead session) with no
// separate bookkeeping. A key with no capacity is unlimited.

// ResourceName is the configured resource a key belongs to: the part
// before the first ':'.
func ResourceName(key string) string {
	name, _, _ := strings.Cut(key, ":")
	return name
}

// ResourceUse is one resource key's occupancy.
type ResourceUse struct {
	Key      string `json:"key"`
	Capacity int    `json:"capacity,omitempty"` // 0: unlimited
	Held     int    `json:"held"`
	Waiting  int    `json:"waiting"`
}

// ResourceWait is a ticket held back only by a full resource key (#408):
// open, unblocked and not deferred, but every slot of Key is taken.
// Ahead counts the waiters on Key that claim order puts first.
type ResourceWait struct {
	Ticket   string `json:"ticket"`
	Key      string `json:"key"`
	Ahead    int    `json:"ahead"`
	Held     int    `json:"held"`
	Capacity int    `json:"capacity"`
}
