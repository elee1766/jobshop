# Mapped home and slot allocation in sessionrt (JOB-153, 2026-11-08)

Slice 3 of JOB-146. JOB-146 stays open (JOB-152 and JOB-154 remain). No
behavior change.

## Moved to `pkg/control/sessionrt`

| File | Owns |
|---|---|
| `home.go` | `MappedUser` (uid and gids for a pool user and slot), `Homes` (`Dir`, `Prepare`, `RemoveSlot`, `CheckAccess`) and `ExpandHome` |
| `slot.go` | `SlotSource`, `AllocSlot`, `FreeSlots`: the lowest-free-slot rule over a narrow interface |

`Homes` holds the controller root dir and a logger. `sessionrt` still
imports `pkg/config` for `IsolationUser` (as `acl.go` does) but not
`pkg/store` or `pkg/control`.

## Stays in the controller

- `sessionuid.go`: `slotStore` implements `SlotSource` over the store
  (open sessions of a pool and their `SessionFields.Slot`).
  `sessionSlot` and `slotCapacity` read config and call `sessionrt`.
  `newSession` still holds `slotMu` from the pick until the row carries the
  slot, so two creates cannot choose the same slot.
- `mappedhome.go`: thin wrappers that read the live config, decide
  whether a ticket is a closed per_session session (`removeSlotHome`) and
  pass directories and ids down.

## Tests

`TestPrepareMappedHome` moved to `sessionrt` as `TestPrepareHome`, with
new tests for `MappedUser`, `Homes.Dir`, `RemoveSlot` and `AllocSlot`.
The slot tests that need a store (`TestSlotAllocation`, `TestSlotRestart`,
`TestRemoveSlotHome`) stay in `pkg/control`.
