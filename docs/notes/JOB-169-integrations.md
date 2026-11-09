# JOB-169: provider clients under pkg/integrations

2026-11-09. Child of JOB-164, after JOB-166. `apicall`, `github`, `linear` and
`identity` moved to `pkg/integrations/`. Behavior, wire formats and config are
unchanged.

## What moved

| Was | Is |
|---|---|
| `pkg/apicall` | `pkg/integrations/apicall` |
| `pkg/github` | `pkg/integrations/github` |
| `pkg/linear` | `pkg/integrations/linear` |
| `pkg/identity` | `pkg/integrations/identity` |
| `hostToken` in `pkg/control/apiclients.go` | `identity.HostToken` |
| renew-on-401 and body-limit code in both clients | `pkg/integrations/internal/transport` |

Old paths are deleted, no wrappers. There is still one GitHub client and one
Linear client. `pkg/automation/github` and `pkg/automation/linear` are the
Starlark capability adapters and `pkg/ingest` is inbound delivery. They call the
clients and are not providers.

## Decisions from the caller trace

- **No tracker interface.** The clients differ (REST, ETags, Link pages and a
  same-host redirect policy against one GraphQL endpoint with no redirects, and
  different error types). The `provider-clients-are-siblings` rules stop either
  importing the other.
- **`internal/transport`** holds the two things both clients did the same way:
  `Authorized` (send, on a 401 ask the source to renew, resend once only if the
  token changed, then tell the context's observer the response was read) and
  `ReadLimited`. Only `github` and `linear` import it. Requests, headers, URLs
  and errors stay in each client. The `provider-clients-stay-small` rule limits
  both to `apicall`, `transport` and `telemetry`.
- **`apicall` stays public.** `automation`, `cli`, `control/linearsvc` and
  `control/activity` take `Source`, `Static`, `Observer` and
  `WithRequestTimeout` from outside the family. An `internal` placement would
  force a wider export. Observation is the narrow `Observer` struct on the
  context. No client internals are exported for it.
- **`identity` is credentials, not session ids.** Its only importers are
  `control` (and tests), so `identity-only-in-control` fails any other package
  that imports it: a client, hook or capability takes an `apicall.Source`.
  `HostToken` (GH_TOKEN, else `gh auth token`, cached ten minutes) moved there
  from the controller since it is a credential source. `identitySource` and
  `appJWT` stay in control because they read the live config and the controller
  clock.
- **Write keys are not provider bookkeeping.** `writekey` is imported by control
  only, for ingest dispatch, Linear writes and the `writekey.claim/finish` RPC,
  and it is a durable store collection with no provider import. It stays out of
  `integrations`, so integrations has no store dependency and the store has no
  provider dependency. JOB-173 moves it to `storage/store/writekey`.
- No provider SDK is in the module, and none of `domain`, `store` or `config`
  imports a provider package. `identity` imports `config` for `Identity`, which
  the `integrations` family may do.

## Checks

`archcheck` rows and the `lower-layers` and `independent-leaves` rules use the
new paths, and three rules are new (above). `TestRepositoryCatchesDeliberateEdges`
covers `linear -> github`, `github -> identity` and `hooks -> identity`. The
clients' existing httptest suites (renewal, Link pages, ETag, observer, limits,
rotation) pass unchanged, plus tests for `transport` and `HostToken`.
