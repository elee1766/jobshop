# Activity feed out of pkg/control (JOB-149, 2026-11-08)

Part of JOB-139. No behavior change: same `activity.json`, ETags, poll
intervals, window and per-source bounds, App-token endpoints, RPC
results and events.

## What moved to pkg/control/activity

- `Service`: the cache (`State`, written atomically by temp file and
  rename, read back on restart with its ETags), the poll loop (`Run`,
  `Poll`), `Admit` (items the admission pass let through), `Record`
  (controller-made items: recorded `jobshop gh` calls, redactions),
  `List` (filters, limit) and `LinkKey`.
- `Live`: the GitHub readers (events, completed runs, the per-resource
  endpoints for App tokens, JOB-108) and the Linear reader, with the
  event and response parsers. They use `pkg/github` and `pkg/linear`
  clients.
- Dependencies come in through two small interfaces: `Host` (config,
  clock, inbox, ingest pass, admitted-items hook, change event) and
  `Clients` (the client for the identity covering a repo or team, the
  host's login). `activity` does not import `pkg/control`.

## What stays in pkg/control

- `activityhost.go`: the adapters, the `activity.list` RPC (grant check,
  `bots` validation, sources), and `activityLinks`, which maps `repo#n`
  and Linear keys to runs and tickets from `meta` on every list, so links
  follow meta changes. `activityRecord` (RPC) stays in `ghcall.go`.
- Inbox write (`inboxActivity`) and `ingestPass`: new items still go
  through admission hooks before they reach the feed, so `Poll` writes
  nothing for unchanged events.
- Revert detection (`revert.go`, `revertSource`): it reads PRs and learns,
  which is controller work. `liveActivity` embeds `activity.Live` and adds
  the two PR reads it needs. `rebasecheck.go` uses the same source.
- Identity selection and token minting (`apiclients.go`, `githubSelf`).

## Coordination

- JOB-143 moved gh-CLI observation to `pkg/github`; the feed uses the REST
  `github.Client` on purpose (ETags, X-Poll-Interval), so the two readers
  stay separate.
