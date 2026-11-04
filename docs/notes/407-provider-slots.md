# Resource pools: provider-owned slots (#407)

2026-11-04. Second slice of #400, after keyed and counted slots.

Config:

```yaml
resources:
  emulator:
    provider:
      command: /srv/emu/provider   # or url: https://host/provider
      timeout: 2m                  # one call; default 2m
      idle_release: 30m            # optional
      demand_every: 1m             # default 1m
```

A provider resource has no `max` or `keyed`: its slots are its capacity.
Steps name it in `needs_resources` like a counted resource.

- The provider owns slots: `jobshop resource add <pool> <id> [--env k=v]`,
  `resource remove`, `resource unhealthy [-r why]`, `resource list`
  (admin RPC `resource.*`, grantable to a pool with `admin:`).
- Slot states: ready -> held -> releasing -> ready or gone. A step is
  ready and claimable only while its resource has a ready slot. The
  claim takes the longest-idle one in the claim transaction, so two
  claimers cannot take one slot. Rows are in `resource_slots`, and the
  holder is recorded as ticket and fence.
- Env: a command step gets the slot's env plus `JOBSHOP_SLOT_<NAME>=<id>`
  in its process. An agent host outlives claims, so an agent step gets
  the same env as `export` lines at the top of its ticket prompt.
- Claim end, on any path (close, release, lease expiry, dead session,
  claimed again with a new fence): the next supervisor pass marks the
  slot releasing and calls `reset <id>` once. Exit 0 or 2xx: ready.
  Anything else: gone, until the provider adds it again.
- `resource unhealthy` while held reopens the ticket in the same
  transaction without counting a step failure, with a comment. The
  slot is gone.
- `resource remove` of an idle slot deletes it. A held one is marked
  and deleted after its claim, without a reset.
- Engine -> provider: `demand` with `{waiting, ready, held, oldest_wait}`
  when it changes or every `demand_every`, and `remove <id>` for a ready
  slot idle past `idle_release` while nothing waits. The provider
  answers a remove with `resource remove`.
- A command provider runs `sh -c '<command> "$@"' provider <op> [id]`
  in the root dir, with `JOBSHOP_RESOURCE`, `JOBSHOP_SLOT` and the call
  as JSON on stdin. A url provider gets the JSON as a POST.
- Reset calls run in the supervisor pass, so a slot is never reset
  twice. Demand and remove calls run in the background.

Not in this slice: waiting counts in stall checks and attention;
required and preferred placement.
