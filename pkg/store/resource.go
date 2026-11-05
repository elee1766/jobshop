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

// ResourceConfig is the resource part of the config the store gates
// claims with (#579: replaces SetResourceCaps and SetResourceProviders).
type ResourceConfig struct {
	// Caps maps a resource name to its capacity; each must be > 0.
	Caps map[string]int
	// Providers names the resources whose slots a provider owns (#407).
	Providers []string
}

// Slot event ops.
const (
	// SlotAdd adds a ready slot, or updates a known slot's env. A gone
	// slot comes back ready.
	SlotAdd = "add"
	// SlotRemove deletes an idle slot; a held one is deleted once its
	// claim ends, without a reset.
	SlotRemove = "remove"
	// SlotUnhealthy marks a slot gone. A held slot's claim is reopened
	// without counting a step failure.
	SlotUnhealthy = "unhealthy"
	// SlotReset records a provider reset of a releasing slot: OK ->
	// ready, else gone. A slot removed while held is deleted.
	SlotReset = "reset"
)

// SlotEvent is one change to a provider slot (#579: replaces AddSlot,
// RemoveSlot, SlotUnhealthy and SlotReset).
type SlotEvent struct {
	Op     string
	Pool   string
	ID     string
	Env    map[string]string // add
	Reason string            // unhealthy
	OK     bool              // reset
}

// ResourceReport is the resource state Resources reads (#579: replaces
// Slots, TicketSlots, ResourceWaits and SlotDemand).
type ResourceReport struct {
	// Use is each resource key in use or waited on.
	Use []ResourceUse
	// Waits lists tickets held back only by a full key, in claim order
	// per key (#408).
	Waits []ResourceWait
	// Slots lists every slot, or the pool's slots when one is asked for.
	Slots []Slot
	// Demand is the pool's waiting and slot counts; zero without a pool.
	Demand SlotDemand
}

// TicketSlots returns the slots t's live claim holds: t is in progress
// and each slot is held under t's fence.
func TicketSlots(r ResourceReport, t Ticket) []Slot {
	if t.Status != StatusInProgress {
		return nil
	}
	var out []Slot
	for _, sl := range r.Slots {
		if sl.State == SlotHeld && sl.Ticket == t.ID && sl.Fence == t.Fence {
			out = append(out, sl)
		}
	}
	return out
}
