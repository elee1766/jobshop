# pkg/control split: resource slot provider calls (JOB-158, 2026-11-08)

Part of the JOB-139 completion gate (found by the JOB-155 audit). No
behavior change: provider calls, demand spacing and idle release are as in
`407-provider-slots.md`.

## Boundary

- `pkg/control/slotprovider` owns the call transport (`Client.Call`: a
  command run with the call JSON on stdin, or an http POST, both bounded
  by the provider timeout), the call JSON (`Call`), demand throttling and
  removal tracking (`Client.Pass`), and the slot env and prompt an agent
  gets (`Env`, `Prompt`).
- It reaches the controller only through `slotprovider.Host`: `Now`,
  `Resources` (store slot and demand reads), `Go` (the controller's
  WaitGroup), `Context` (the controller's, so a demand or remove call
  outlives the pass), `Dir` and `Env` (command providers run in the root
  dir with the host-inherited env; the code never used `runIsolated`, so
  none is needed).
- Stays in `pkg/control/provider.go`: the admin RPC (`resource add`,
  `remove`, `unhealthy`, `list`), `resetSlots` (it records into the store
  and uses the off-pass runner), `slotEnv` (store read) and the pass step
  `providerPass`, which resets slots and then calls `Client.Pass`.
- `Controller.slots()` makes the client on first use, like `compressor()`.
  The package does not import `pkg/control`; its unit tests run on a fake
  host. The controller-level test stays in `provider_test.go`.
