# Provider slots: env, reset and health (JOB-204)

2026-11-12. Last of the provider-slot follow-ups in
`400-resource-slots.md`. Env per slot and the reset between holders
shipped with #407 (`407-provider-slots.md`); this adds the health check.

```yaml
resources:
  emulator:
    provider:
      command: /srv/emu/provider
      health_every: 5m   # optional; empty: no checks
```

- `health_every` must be a positive duration (config load fails
  otherwise).
- Each supervisor pass, a ready slot idle `health_every` since its last
  check or state change (add, reset) gets `health <id>`: a command
  provider runs `<command> health <id>` with the call JSON on stdin, a
  url provider gets `{"op":"health",...}` as a POST. Exit 0 or 2xx: ok.
- The call runs off the pass, one per slot, bounded by `timeout`. The
  pass that finds it finished records it.
- A failed check marks the slot `gone`: no claim takes it until the
  provider adds it again (`jobshop resource add`).
- Only ready slots are checked. A held slot is checked by the reset after
  its claim. A result for a slot that changed while the check ran (taken
  by a claim, re-added) is dropped, so a check never takes a slot away
  from a claim; the store op (`SlotHealth`) also only touches a ready
  slot.
- Check times are in memory: after a controller restart, a slot is due
  `health_every` after its last state change.

Tests: `TestResourceChecks` (config), `TestHealthDue` (slotprovider),
storetest `SlotHealth` (both engines), `TestProviderSlotHealth`
(controller, supervisor passes on a fake clock).
