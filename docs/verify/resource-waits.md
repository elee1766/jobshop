# Verify: resource waits (#408)

On gasserver, with `resources: {emu: {max: 1}}` and a test workflow whose
step has `needs_resources: [emu]`:

1. Start two runs; let the first step be claimed.
2. `jobshop run show <second run>`: the step's resources line ends
   `waiting on emu, 0 ahead (1/1 held)`.
3. With `attention: {resource_wait: 1m}` and a reload, after a minute
   `jobshop attention` lists the waiting step with kind `resource`.
4. Close the first step: the waiter leaves `run show` waits and
   attention.
