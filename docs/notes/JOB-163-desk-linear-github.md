# JOB-163: native Linear and GitHub views in Desk (2026-11-09)

Two top-level Desk tabs, `?view=linear` and `?view=github`, list and open
issues, pull requests, comments, reviews and checks from the configured
providers. They are native React screens, read only, and not iframes. Plan9
designs approved on JOB-163 are the reference.

## Scope and credentials

- Scope is only what the controller already has: each Linear team with an
  identity and each GitHub repo with an identity. `browse.scope` returns it.
  A request for any other team or repo is refused before a provider call.
- Tokens stay in the controller. The browser sees `browse.*` results only.
  Provider errors are rewritten token free before they leave `pkg/browse`.
- Writes are not offered. Reply, comment, merge and state controls are shown
  disabled with the reason "read only".

## Wire and code

| Piece | Where |
| --- | --- |
| Types and specs `browse.scope`, `browse.list`, `browse.show`, `browse.catalog` | `pkg/rpc/browse.go`, `spec.go`; generated `packages/jobshop-client/rpc.d.ts` |
| Service: cache, single flight, rate-limit and auth blocking, stale fallback | `pkg/browse` |
| Controller host, `browse` grant, run links, cache invalidation | `pkg/control/browsehost.go` |
| Typed `StatusError` with retry time | `pkg/linear`, `pkg/github` |
| `browse` live section (scope and revision) and chat refs | `pkg/desk/live.go`, `pkg/desk/chat.go` |
| Screens | `ui/src/screens/Browse.tsx`, `BrowseDetail.tsx`, `components/BrowseParts.tsx` |
| Data and helpers | `ui/src/state/browse.ts`, `ui/src/lib/browse.ts` |

`browse` is also a pool admin method (`PoolAdminMethods`), so a session can
read the same views. Desk and the admin socket are trusted and need no grant.

## Cache and limits

- TTLs: list 30 s, detail 20 s, catalog 10 min. A forced refresh is allowed
  every 5 s per key. A failed fetch serves the old copy for up to 30 min and
  marks it stale. At most 400 entries, least recently used out.
- One in-flight request per key. Each provider call is cut off at 25 s.
- A rate-limit answer blocks that scope until the provider's retry time
  (default 60 s, at most 15 min). An auth failure blocks it for 30 s, an
  outage for 10 s. The view shows the reason and the retry time.
- Pages are 50 rows (at most 100), threads 30 comments, query text 200
  characters, cursors 1024.

## Associations

Links to Jobshop work are real only. A run is linked when its meta names the
issue or pull request (`issue=JOB-163`, `pr=1114`). An item with no run has no
link. Chat context refs `linear:KEY` and `github:owner/repo#n` resolve through
`browse.show` (brief); an out-of-scope ref makes the send fail with 400.

## Live updates

The Desk stream carries a revision. It moves when the activity poller admits a
new item for a `(source, repo-or-team)`, which drops that cache entry, and the
open view fetches again. The poller runs every minute, so a change made at the
provider shows up within about a minute. The probe comment on JOB-163 appeared
in an open detail view 29 s after it was posted. A manual refresh asks the
provider at once.

## Bundle

Entry chunk 186.67 KB gzip (master 195.55 KB). Browse is its own lazy chunk of
about 10 KB gzip. Total JS is 214 KB gzip against 205 KB on master, so the old
JOB-150 budget of 190 KB, already exceeded on master, is not met.

## Tests

- Go: `pkg/browse` (fixtures recorded from the real providers in
  `pkg/browse/testdata`, re-recorded with `BROWSE_RECORD=1`), typed errors in
  `pkg/linear` and `pkg/github`, grant, scope and run links in
  `pkg/control`, chat refs in `pkg/desk`.
- Unit: `ui/src/lib/browse.test.ts`.
- Browser: `ui/e2e/browse.spec.js` runs on a mocked `/api/rpc` on desktop and
  mobile and is part of `e2e/run-isolated.mjs`. `browse-live.spec.js` needs real
  scope and skips without it (`docs/verify/desk-linear-github.md`).

## Live check, 2026-11-09

On an isolated controller with the real Linear and GitHub identities: scope,
lists and details match the providers, out-of-scope repos are refused, held
runs show their real run links, a Linear comment reached an open view, chat
send with real refs worked, and no credential appeared in any response.
