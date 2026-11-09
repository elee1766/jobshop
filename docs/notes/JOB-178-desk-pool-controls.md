# Desk pool scale and session cordon controls (JOB-178)

2026-11-09. The system view's pools table (`ui/src/components/PoolControls.tsx`)
shows cordoned and desired (auto or manual) per ticket pool. With
`pool.scale` allowed it adds a number field with `set`, and `auto`. Each
session row of a ticket pool gets `cordon` or `uncordon` when `session.cordon`
is allowed, and a `cordoned: <reason>` note.

- Calls go through `POST /api/rpc/{method}` (`pool.scale`, `session.cordon`),
  the same path as any other client, as the proxy user.
- "Grant" here is `desk.rpc` as `GET /api/rpc` lists it. Desk talks to the
  controller over the admin socket, so the pool admin grants of
  `docs/notes/281-pool-scaling.md` do not apply to it. A method removed by
  `desk.rpc.deny` or missing from `desk.rpc.allow` hides the controls, and
  the server refuses the call too. The list is read each time the view opens.
- A write marks its control pending (`setting…`, `working…`) and refuses a
  second submit until it returns. Success shows a notice; a refusal shows in
  the error bar. The new state arrives on the live stream.
- The target must be a whole number from 0 to `max_sessions`. Drain and
  chat-pool cordons are not offered.
- Browser check: `docs/verify/desk-pool-controls.md`.
