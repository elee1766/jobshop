# JOB-150: Desk on React, Jotai, Wouter and Tailwind (2026-11-09)

Desk is now a client-rendered React app. The Go side is unchanged: the same
embedded `pkg/desk/dist`, the same `api/rpc`, `api/live` SSE and generated
client. The inventory, Svelte baseline and budgets are in
`JOB-150-desk-react-baseline.md`. This note records what was built, how it
measured against those budgets, and how to roll back.

## Layout and ownership

`ui/src` replaces `ui/shell`. `ui/kit` (Markdown, link and title parsing, with
their tests) and `packages/jobshop-client` are unchanged and shared.

| Path | Owns |
| --- | --- |
| `index.css`, `tokens.css`, `app.css` | Tailwind layers, the semantic tokens (still the one source for color, type, spacing, themes), the legacy element and layout CSS kept in the components layer. `index.css` maps the tokens into Tailwind's `@theme`. |
| `lib/` | Framework-neutral helpers and hooks: `api`, `model`, `trace`, `learn`, `session`, `push`, `reconcile`, `search` (Fuse), `route`, `shortcuts`, `attachments`, `useSize`, `useSessionFrames`. |
| `state/` | The one Jotai store and everything that writes to it: `entities` (collections), `live` (the stream and its atoms), `ui`, `actions` (every controller write), `chatBridge`. |
| `components/` | `LiveOwner` (connection and tab lifecycle), `Markdown`, `TitleText`, `TitleAction`, `StatusLetter`. |
| `screens/` | One file per screen: Home, Chat, Transcript, SessionViewer, TraceScreen, Graph, TraceSearch, TicketDetails, TicketHistory, Learn, System, HookAudit. |

Dependencies are exact pins in `ui/package.json`. Base UI is used for the chat
popovers and the dialog. TanStack Virtual is used only in `Transcript`.
TanStack Hotkeys is used only through `lib/shortcuts.ts`. Fuse is used only
through `lib/search.ts`. There is no query cache, form library, date library
or graph package: the graph keeps its own renderer with one camera state.

## State and route conventions

- Go to `api/live` to the one store to screens. `LiveOwner` sits above the
  routed screens and owns the single stream. It reopens only when the stream
  scope changes (the trace ref and inspected ticket, or an audit id), and it
  closes on `offline` and reopens on `online`. The e2e spec `shell.spec.js`
  asserts one `api/live` request across view changes and Back and Forward.
- Collections (attention, runs, chats) keep an ids atom for membership and
  order and one atom per entity, so a changed row re-renders that row and a
  list change re-renders the list. Unchanged payloads keep their reference
  (`lib/reconcile`). `resetLive` clears every scoped atom.
- The URL is the navigation state and is not mirrored in atoms. The contract
  is unchanged: `view`, `ref`, `inspect`, `audit`, `filter`, `since`, `chat`,
  `item`, and the `#waiting` and `#inflight` hashes. There are no path routes,
  so any proxy prefix works. `lib/route.ts` uses Wouter's browser-location
  primitives on the raw search string (Wouter's `useSearch` decodes first,
  which would turn an encoded `%41` into `A`). Typed filter text reaches the
  URL after 250 ms with `replaceState`.
- Shortcuts are registered by the component that owns the screen and are
  enabled only while it is active. Text fields, selects, contenteditable and IME
  composition are skipped by the library, Ctrl and Meta chords are left to the
  browser, and nothing registers Escape globally (Base UI owns it in dialogs).
- Native `confirm` and `prompt` stay. The e2e class, id and role locators stay.

## Compiler and Oxc status

- `oxc-transform-react` 0.156.0 is the compiler behind `@vitejs/plugin-react`
  6.2.1. It is experimental, so it is off by default and a build never depends
  on it: `DESK_COMPILER=1 npm run build` turns it on.
- With it on, the build succeeds and the whole e2e suite passes (63 passed,
  7 skipped, desktop and mobile), but the output is larger: 205 KB gzip for
  the JS files against 189 KB without it (`perf/build.mjs` measure). That is
  over the 190 KB budget, so it stays opt-in. No correctness claim rests on it.
- `react/incompatible-library` is turned off for `Transcript.tsx` only.
  `useVirtualizer` returns functions the compiler cannot memoize safely, so the
  compiler skips that one component. It is the only Virtual use.
- Other scoped Oxlint exceptions, each commented in `.oxlintrc.json`:
  `prefer-tag-over-role` for the one-letter status, the learn chip group and
  the graph axis (the specs address them by role), and the graph camera's
  interaction rules (one focusable region with its own key handling).
- The Svelte app declared an icon package but never imported it, so there is
  no icon set to port. No icon dependency was added.

## Results against the budgets

Production build, worst of three runs, 200 runs and 30 waiting items seeded
(`ui/perf/seed.sh`), the same scripts and host class as the baseline.

| Metric | Budget | Desktop | Mobile |
| --- | --- | --- | --- |
| First usable render | 350 ms | 342 | 347 |
| First contentful paint | 150 ms | 124 | 116 |
| Switch to learn, sys, home | 150 ms | 56, 51, 107 | 55, 56, 116 |
| In flight scroll p95 frame | 16.8 ms | 16.8 | 16.8 |
| 50-ticket burst, last row | 400 ms | 218 | 213 |
| Burst long tasks | 0 | 0 | 0 |
| `api/live` requests | 1 | 1 | 1 |
| Console or page errors | 0 | 0 | 0 |
| Script time at load | 300 ms | 179 | 176 |
| Heap idle, after burst | 9, 11 MB | 5.2, 6.2 | 5.2, 6.3 |
| DOM nodes, idle home | 5400 | 2720 | 2720 |

Transcript (`ui/perf/chat.mjs`, 200 turns, clean controller, five runs each,
worst of three): open to last turn 279 ms (budget 900), rows rendered at most
25 (budget 40), scroll to the start 172 ms (250), 557 DOM nodes (1500), heap
6.5 MB (12), no long tasks. On the controller that also carries the seeded
dataset the numbers are 448 ms, 25 rows, 206 ms, 3093 nodes and 7.6 MB:
the node count follows accumulated chat sessions, and the Svelte build on that
same controller measured 849 ms, 262 ms, 4924 nodes and 9.6 MB.

Bundle: JS 589.9 KB raw (budget 640), 189.3 KB gzip (190), 163.9 KB brotli
(165). CSS 6.7 KB gzip (14). The entry chunk is 182 KB gzip. The trace graph,
learn and sys screens are separate chunks (14 KB gzip together) and the app
fetches them when the browser is idle after the first render, so one entry is
the only initial request and a first switch finds them loaded. Without that
preload a first switch waited on the chunk fetch and a Suspense reveal and took
about 800 ms.

JS gzip is 1 KB under its budget, so the next dependency or screen needs a
deliberate look at the bundle first. The raw bytes still go out uncompressed
until #1088 lands.

Tooling: production build 0.53 s cold and warm (budget 4), `tsc` plus Oxlint
plus Oxfmt 4.0 s together, Oxlint plus Oxfmt about 30 ms (3 s), dev server
ready 218 ms (1000), first load through dev 679 ms (2500), HMR update 5 ms
median (150). `node_modules` is 233 MB, 55 packages (Svelte: 162 MB, 46). These
are budgets, not a claim that Oxc is faster than the Svelte compiler.

Tests: 18 unit tests (search, links, titles, client), and the Playwright suite
on desktop and mobile against an isolated controller started with `--tick=1s`:
58 passed and 6 skipped (the skips are the real-agent specs and the one spec
that needs a waiting ticket at start), the same as the Svelte baseline, plus
the new specs below. The suite also passes behind a path prefix
(`DESK_PREFIX=1`). New in this change: `e2e/shell.spec.js` (one live
connection, URL contract with an encoded filter, Back and Forward, shortcut
scoping) and `e2e/graph-waits.spec.js` (the wait edges and external cards that
#1095 added to the Svelte graph, ported here and checked in a browser, which
#1095 was not). Search has unit tests for exact-id
priority, weighted and typo matches, empty and no-result, rebuilt indexes with
additions and removals, and bounded latency on a large collection.

The seeded dataset also fails `home-sections` and `run-links` on the Svelte
build (6 specs, desktop and mobile), because 200 seeded runs push the new
ticket out of view. Run the e2e suite on a clean controller.

## Rollback

Svelte lives in git history only. To roll back, revert the cutover commits
(the React port, its lint refactor, lazy loading, the nix and CI changes) and
rebuild with `npm --prefix ui ci && npm --prefix ui run build`. The Go side,
the URL contract and the API did not change, so no data or config migration is
involved. If only the compiler misbehaves, build without `DESK_COMPILER`, which
is the default.
