# Resource pools: keyed and counted slots (#400)

2026-11-04. First slice of #400.

Config:

```yaml
resources:
  emulator: {max: 8}                         # 8 steps at once
  pr: {keyed: true, capacity_per_key: 1}     # one step per key
workflows:
  pr-work:
    steps:
      - {id: review, pool: dev, needs_resources: ["pr:${repo}#${issue}"], ...}
```

- `needs_resources` entries are templates filled from run vars at run
  start, like `pool` (#384). A keyed resource needs `name:<key>`, a
  counted one the bare name. An unknown resource, or the wrong form,
  fails the start and creates nothing.
- The filled keys are stored in the step's `fields.resources`.
- A step is ready and claimable only when every key has a free slot.
  The check runs in the ready and claim queries, so the supervisor never
  starts a session for a step that would wait on a slot, and
  `claim <id>` refuses with "waits for a resource slot".
- Occupancy is the in-progress tickets naming a key. A slot frees when
  its claim ends (close, release, lease expiry, dead session), from the
  same write. There is no separate slot table to leak.
- Capacities are in `resource_caps`, written from config at controller
  start and on reload. A key whose resource is not configured is
  unlimited. Lowering a cap never preempts a running step.
- `run show` lists each step's resource keys.

Not in this slice (follow-ups): provider-owned slots with env, reset and
health; waiting counts in stall checks and attention; required and
preferred placement.
