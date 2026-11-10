# Linear identity `api_url`

2026-11-10. A `linear` or `linear-app` identity may set `api_url`, the
GraphQL endpoint its controller client calls. Empty keeps
`https://api.linear.app/graphql`.

- Why: the `linear` built-in and every other controller-side Linear call
  had no endpoint setting outside Go tests, so a deployment's integration
  tests could point GitHub identities at a loopback fake (`api_url`) but
  not Linear ones. Callback code that writes to Linear could not be
  exercised against a real controller.
- The controller keeps one client per identity and rebuilds it when the
  endpoint changes on reload.
- The `jobshop linear` CLI is unchanged: it builds its own client and
  still reads the hidden `JOBSHOP_LINEAR_URL`.
- Production config does not set it.
