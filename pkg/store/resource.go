package store

import (
	"strings"
	"time"
)

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

// Provider-owned slots (#407). A resource with a provider has concrete
// slots the provider adds and removes. A claim takes a ready slot in the
// claim transaction (ready -> held); when the claim ends the controller
// moves it to releasing and asks the provider to reset it: ok -> ready,
// failure -> gone.
const (
	SlotReady     = "ready"
	SlotHeld      = "held"
	SlotReleasing = "releasing"
	SlotGone      = "gone"
)

// Slot is one provider-owned resource slot.
type Slot struct {
	Pool  string            `json:"pool"`
	ID    string            `json:"id"`
	State string            `json:"state"`
	Env   map[string]string `json:"env,omitempty"`
	// Ticket and Fence name the claim holding the slot.
	Ticket string `json:"ticket,omitempty"`
	Fence  int64  `json:"fence,omitempty"`
	// Removing: removed while held; deleted without a reset once released.
	Removing  bool      `json:"removing,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SlotDemand is what the controller tells a provider (#407).
type SlotDemand struct {
	Waiting    int       `json:"waiting"`
	Ready      int       `json:"ready"`
	Held       int       `json:"held"`
	OldestWait time.Time `json:"oldest_wait,omitzero"`
}
