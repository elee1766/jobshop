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
