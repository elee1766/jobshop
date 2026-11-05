# #579 step 3: resources and slots

2026-11-05. Sixth ticket-side group of #579.

- `SetResources(ResourceConfig)` replaces `SetResourceCaps` and
  `SetResourceProviders`. Caps and providers are written in one
  transaction, so a reload cannot leave one updated without the other.
- `Slot(actor, SlotEvent)` replaces `AddSlot`, `RemoveSlot`,
  `SlotUnhealthy` and `SlotReset` (ops `add`, `remove`, `unhealthy`,
  `reset`). Reset now also returns the slot.
- `Resources(pool)` returns a `ResourceReport` (use, waits, slots, and
  with a pool its demand) and replaces `Resources`, `ResourceWaits`,
  `Slots`, `TicketSlots` and `SlotDemand`. `store.TicketSlots(report,
  ticket)` filters a ticket's live slots in Go; the `TicketSlots` query
  is gone.
- `ReleaseSlots` is unchanged.
- `store.Store` is at 21 methods (29 before).
