# Desk React migration: inventory, baseline and budgets (JOB-150, 2026-11-08)

First slice of [JOB-150](https://linear.app/gfx-labs/issue/JOB-150): the screen
and behavior inventory, the measured Svelte baseline, and the numeric
acceptance budgets. It adds measurement scripts under `ui/perf/` and changes
no application code. The port starts after this lands, so the budgets below
are fixed before any React code is measured.

Agreed stack and constraints are in the issue. This note does not restate
them.

## 1. Inventory

### Routes and URL contract

Desk has no path routes. Every view is a query string on the Desk base path,
and the Go server serves `index.html` for any unknown path (`pkg/desk/desk.go`
`static`), with a `<base href>` for the proxy prefix (#516). Hash or path
routing is therefore not a server choice: **keep query-string state**. Wouter
reads it with `useSearch`, and `useLocation` is only used for the pathname.

| Param | Values | Set by |
| --- | --- | --- |
| `view` | `home` (default), `trace`, `improvements`, `system` | nav buttons, links, `desk navigate` |
| `ref` | ticket, session or external reference | trace links, trace search |
| `inspect` | node id inside the trace | graph node select |
| `audit` | automation invocation id | sys audit list, trace "automation invocation" |
| `filter` | in-flight filter text | in flight filter box, `desk filter` |
| `since` | `1` | catch-up toggle |
| `chat` | session id | active chat (replaceState, no history entry) |
| `item` | `<kind>:<id>` | push notification click (`./?item=...`), selects the waiting item |
| `#waiting`, `#inflight` | hash | phone tabs and section shortcuts, home only |

Compatibility rules for the port: every URL above keeps working (push
notifications and `jobshop desk navigate` generate them), `pushState` per
navigation and `replaceState` for `chat`, same-origin `a[href]` clicks to
`?view=trace&ref=` are intercepted and open the trace in place (`traceClick`),
and Back and Forward restore the whole state (`popstate` -> `restoreRoute`).

### Screens

| Screen | Source | Behavior to preserve |
| --- | --- | --- |
| Shell | `App.svelte`, `app.css`, `tokens.css` | top tag bar (`desk`, nav, sys status with tick time), error banner with dismiss, two columns (work, chats), phone bottom nav (waiting, chats, in flight, more), light/dark via `data-theme`, visual viewport vars `--vvh` `--vvtop`, soft-keyboard class, `online`/`offline`, focus POST on tab hide |
| Home: waiting | `App.svelte` | attention items (question, decision, ticket, run, wait, guard, pool), collapse toggle, option buttons, answer note, discuss, trace link, keys `1`-`5` answer, `d` discuss, `j`/`k` move, triage, "since" banner |
| Home: not moving | `App.svelte` | stalled runs, hide external holds |
| Home: in flight | `App.svelte`, `model.ts` | table with status letter, title action, issue and PR columns (stacked links on compact), step, waiting, age, sort oldest/newest, filter box, step chips, truncated notice, collapse toggle |
| Trace | `Trace.svelte`, `trace.ts`, `TraceSearch.svelte`, `TicketDetails.svelte`, `TicketHistory.svelte` | ticket search (recent, matches, exact ref open), temporal graph with creation time axis, hand pan (pointer), pinch and wheel zoom, keyboard pan and `+` `-` `0`, tooltip pin and `Escape`, visibility toggles, truncated-graph notice, details and history disclosure, relations |
| Session viewer | `SessionViewer.svelte`, `SessionTranscript.svelte` | read-only live session under a trace, virtualized transcript, older-history paging with scroll anchoring, raw prompt disclosures, close button |
| Learn | `Learn.svelte`, `learn.ts` | stage strip, filter, per-stage paging (10), problem and attempted fix, empty state |
| Sys | `App.svelte`, `HookAudit.svelte` | push toggle, theme toggle, automation audit list and detail (ordered event, invocation, attempt, call evidence, UNKNOWN warning, redacted data, resulting-ticket links), pools, resources and slots, ingest, scheduler passes, sessions with close (`window.confirm`) |
| Chat rail | `Chat.svelte` | picker popover (recent and archived, search, Enter picks first), new-chat popover (named draft, first send creates), rename and convert (`window.prompt`), archive, resolve (`window.confirm`), stop, attach (files, paste-as-document popover, clipboard paste), attached refs (`window.prompt`), journal link, streaming transcript, pending and queued message state, drafts, `Enter` sends and `Shift+Enter` newlines and no send during IME composition, hide and show (`localStorage desk.chat.hidden`), reload restore via `?chat=` |
| Titles and links | `kit/titles.js`, `kit/links.js`, `kit/markdown.js`, `TitleAction.svelte`, `TitleText.svelte` | Markdown links in titles, qualified GitHub references, `marked` + `DOMPurify` sanitizer |
| Service worker | `public/sw.js`, `lib/push.svelte.js` | Web Push opt-in, base-relative scope, no offline cache |

### Native dialogs

`window.confirm` (close session, resolve escalation), `window.prompt` (chat
name, attach reference). The e2e specs accept them with `page.once('dialog')`.
Base UI dialogs would change that contract: keep the native calls for these
four, or port the specs in the same change. Do not mix.

### Transport boundaries

All of it already sits behind `packages/jobshop-client` (framework free, no
global state): `get`, `post` (adds `X-Jobshop-Desk`), `call` (`api/rpc/<method>`),
`live` (one `api/live` SSE, reopened with new `trace`, `inspect` or `audit`
query when the view changes), `session` (per-session SSE with `tail`/`since`).

Direct `fetch` is not used outside the client. Other browser APIs: `EventSource`
(through the client), `navigator.onLine`, `visualViewport`,
`localStorage` (chat hidden flag, guarded), `Notification` + `PushManager` +
service worker, uploads read with `arrayBuffer` and `btoa` into a JSON body to
`api/chat/send` and `api/chat/manage` (no multipart), `clipboardData`,
`window.open` for the journal link.

Live event kinds the shell consumes: `attention`, `overview`, `system`, `chats`,
`improvements`, `desk` (agent commands: navigate, filter, select, expand,
collapse, attach), `scheduler`, `trace`, `automation`.

**Connection rule to keep**: one `api/live` per tab. It reconnects only when
the stream key changes (`trace:<ref>:<inspect>`, `audit:<id>`, or empty) or
connectivity returns. A React store must own the connection above the router so
navigation between views with the same key does not reopen it.

### Acceptance scenarios to carry over

Existing Playwright specs (`ui/e2e`, 13 files, 64 tests across the `desktop`
and `mobile` projects) already cover most of the parity list. Port them, do
not weaken them. They select mostly by role and accessible name (243
`getByRole` uses) but also by about 130 CSS class and id locators, listed
below. Those classes are an interface the port must keep or the specs must be
edited in the same change, keeping each assertion.

| Spec | Scenario |
| --- | --- |
| `desk.spec` | attachments local until send, layout and chat rail at 1000/790/700/412/360 px, direct view loads without console errors, blocked storage, new chat popover position, waiting reference to trace with Back and direct link, named draft creates on first send, transcript scrollbar gutter, archived chat table |
| `live.spec` | create/update/close ticket appears with no refresh GETs, reconnect snapshot, ordered event replay with SSE ids, graph keyboard and pointer, offline route |
| `chat-responsiveness.spec` | accepted message shows at once, working state, 85-turn history to the beginning with scroll anchoring and bounded rows (< 60), raw prompt disclosure survives virtualization, tool call spacing |
| `chat-dismiss.spec` | picker types (escalation, attention), close versus resolve, dialogs |
| `trace-session.spec` | converted escalation keeps transcript, claimed ticket shows live read-only session, history rows |
| `trace-axis.spec` | axis aligns with cards, endpoint labels, compact bars |
| `home-sections.spec` | in flight is a home section at every width, shortcuts, history |
| `run-links.spec` | Linear and PR links in desktop and compact, PR appears without reload |
| `title-links.spec` | Markdown links in titles open their destination, row action unaffected |
| `search.spec` | trace search finds completed tickets beyond the first page |
| `learn.spec` | needs-a-look first, filters, paging, long titles stay in viewport, empty state |
| `attachments.spec`, `chat.spec` | real agent round trips (opt in) |

Class and id locators used by specs (keep or edit in the same change):
`.transcript`, `.ticket-overview`, `.operator`, `.learn-row`, `.status`, `.tag`,
`.markdown`, `.ticket-status`, `.graph-card`, `.graph-stamp-now`,
`.graph-stamp-start`, `.graph-axis`, `.inspector`, `.chat-state`,
`#chat-message`, `.transcript-row` (with `data-row-key`, `data-index`),
`.ticket-parent`, `.ticket-label`, `.ticket-history`, `.history-raw`,
`.session-output`, `.chat-archive`, `.top`, `.ticket-graph`, `.ticket-comment`,
`.ref-stack`, `.raw-prompt`, `.message-delivery`, `.controller-context`,
`.table-scroll.refs`, `td.ref-col`.

### Coverage gaps the port must close

No spec exercises these today. Add focused regressions in the slice that ports
each one, because the port is where they can silently break:

- **Keyboard shortcuts on home**: `1`-`5`, `d`, `j`, `k`, and the guard that
  ignores them in inputs, with modifiers, and outside the home view.
- **Catch-up** toggle and its URL state, and the sys-page controls (push
  toggle, theme toggle, session close with confirm).
- **Audit detail**: ordered evidence, UNKNOWN state, redacted data, bounded
  artifacts and resulting-ticket links have no browser test.
- **Collapse toggles** on home sections and phone tabs `waiting`/`more`.
- **IME composition** on `Enter` in the composer, and `Shift+Enter`.
- **Cross-endpoint isolation** and a duplicate `api/live` check (the bench
  counts `api/live` requests; no spec asserts it).
- **Sanitizer**: `kit/markdown.js` has no test. Add a test that script, event
  handler and `javascript:` payloads are removed, run against both the current
  and the ported renderer.

### Environment facts that shape the port

- `jsconfig.json` + `svelte-check` is the type check today; `checkJs` covers
  `.js` files. The port needs `tsc --noEmit` over `.ts`/`.tsx` and the
  JSDoc-typed `.js` it keeps (`kit/*.js`, `lib/push`).
- `kit/*.test.js` run with `node --test` in CI. They stay framework free.
- CI `ui` job: `npm ci`, `npm run check`, `node --test ...`, `npm run build`,
  uploads `pkg/desk/dist`. Nix builds the same through `nix/desk-ui.nix` with an
  `npmDepsHash` and a fileset that lists `svelte.config.js` and `jsconfig.json`:
  both must change with the toolchain.
- The Go server sends assets uncompressed (see "Findings"), so transfer size
  today equals raw size.
- `packages/jobshop-client` and `rpc.d.ts` are shared with plugins and
  untouched by this migration.

## 2. Baseline

All numbers are from this host, recorded 2026-11-08, Svelte 5.57.3 on Vite
8.3.5, Node 22.23, Chromium 1217 headless via Playwright 1.59.1. Host: AMD
EPYC 7413, 6 visible cores, load average 2 to 15 from other sessions, so
timings carry noise. Each runtime figure is a median of 7 fresh browser
contexts. The values below are the **worst of two or three full repeats**, so
a budget cannot be met on a lucky run.

Dataset (isolated controller, `ui/perf/seed.sh`): 200 two-step runs (the
overview returns 100), 30 waiting `needs-human`
tickets, 0 chats for the page-load runs. Transcript run: one chat with 200
queued turns. The SSE burst creates 50
`needs-human` tickets in parallel and waits for the last to render. Desktop is
1280x800, mobile is the Pixel 7 profile.

### Build and tooling (`ui/perf/build.mjs`, `ui/perf/dev.mjs`)

| Metric | Svelte |
| --- | --- |
| Production build, no cache | 1.04 s |
| Production build, warm | 1.07 s |
| Type check (`svelte-check`) | 3.05 s |
| JS raw / gzip / brotli | 257.9 / 85.8 / 75.6 KB |
| CSS raw / gzip / brotli | 25.1 / 5.9 / 5.3 KB |
| Dev server ready (reported / wall) | 363 / 397 ms |
| First load through dev server | 1047 ms |
| Reload through dev server | 259 ms |
| HMR, edit to update message (median of 9) | 49 ms |
| `node_modules` | 46 packages, 162 MB |

### Runtime, production build (`ui/perf/bench.mjs`)

| Metric | Desktop | Mobile |
| --- | --- | --- |
| First usable render (waiting rows and `sys ok`) | 285 ms | 288 ms |
| First contentful paint | 88 ms | 80 ms |
| Script time at load | 170 ms | 178 ms |
| Transfer (uncompressed) | 1284 KB | 1284 KB |
| DOM nodes, idle | 4477 | 4477 |
| JS heap after GC, idle | 5.4 MB | 5.4 MB |
| Switch to learn / sys / home | 40 / 50 / 126 ms | 60 / 51 / 127 ms |
| In flight scroll, p95 frame | 16.8 ms | 16.8 ms |
| 50-ticket SSE burst, last row visible | 310 ms | 213 ms |
| Burst script / task time | 21 / 107 ms | 18 / 85 ms |
| Burst long tasks (over 50 ms) | 0 | 0 |
| Heap after burst | 6.5 MB | 6.5 MB |
| `api/live` requests across the session | 1 | 1 |
| Console or page errors | 0 | 0 |

(`transfer` includes about 1 MB of fonts and icons. JS and CSS are 283 KB of it.)

### Transcript (`ui/perf/chat.mjs`, 200 turns)

| Metric | Svelte |
| --- | --- |
| Open chat to last turn visible | 735 ms |
| Rows rendered at open / maximum while paging to the start | 21 / 26 |
| Scroll to start of conversation (history paging) | 179 ms |
| DOM nodes / heap | 1153 / 7.0 MB |
| Long tasks / longest | 1 / 524 ms |

The 524 ms long task is the open itself (first journal replay and Markdown
render), not scrolling. It is the number to beat or match.

### E2E parity baseline

`playwright test --workers=2`, desktop and mobile, against an isolated
controller with `DESK_MUTATE=1`, `DESK_ROOT`, `DESK_LINK_WORKFLOW`,
`DESK_ISSUE_WORKFLOW` and `serve --tick=1s`: **58 passed, 6 skipped, 0
failed** of 64, twice in a row. The 6 skips are the `DESK_CHAT` and
`DESK_ATTACHMENTS` specs, which start a real agent, and `desk.spec.js:110`,
which skips when no waiting ticket exists when the test starts.

Use `--tick=1s` for the isolated controller. With the default 30 s tick the
escalation chat specs (`chat-dismiss`, `trace-session`) failed in one
project per run (2 of 2 full runs, 2 or 3 failures): the controller opens an
escalation chat for a `needs-human` ticket on a tick, and the desktop and
mobile projects create their tickets at the same time. The likely cause is a
tick landing between the two projects' tickets (the log shows "attention
item left, closing its chat" right after an escalation chat opens), which I
did not prove. Raising `attention.prestart_max` did not help, `--tick=1s`
did. It is a property of the test setup, not of the Svelte app, so it is
recorded here instead of filed.

## 3. Findings that change the plan

- **Desk sends assets uncompressed.** `Content-Encoding` is absent, so a
  browser downloads 283 KB of JS and CSS, not the 92 KB gzip. A larger bundle
  hurts more than the compressed figures suggest. Filed as
  [#1088](https://github.com/elee1766/jobshop/issues/1088) (already open,
  `desk: gzip/brotli compress static UI assets`). That is a backend fix and
  independent of the framework, so it is not part of this migration. The
  budgets below are on raw and gzip both for that reason.
- **React alone is the size problem.** Standalone builds of each agreed
  library, gzip (`ui/perf` probe, not additive because they share code):

  | Library | gzip |
  | --- | --- |
  | react + react-dom 19.3.1 | 67.4 KB |
  | Base UI dialog, popover, menu, tooltip 1.9.0 | 68.0 KB (shares React) |
  | TanStack Virtual (react 3.14.26) | 13.4 KB |
  | TanStack Hotkeys (react 0.22.3) | 9.7 KB |
  | Fuse.js 7.5.0 | 9.2 KB |
  | wouter 3.13.0 | 5.2 KB |
  | jotai 3.0.1 | 3.8 KB |
  | lucide-react, 10 icons | 3.0 KB |
  | marked + dompurify (kept) | 25.4 KB |

  A combined probe app with all of them and four Base UI parts is **123.9 KB
  gzip** (393.5 KB raw) before any Desk code, against 85.8 KB for all of Svelte
  Desk today. Desk's own code will add to that. The current app's JS is also
  mostly Markdown and the virtualizer, which stay. So a bundle at or below
  today's size is **not achievable** with this stack. The budget is set on the
  stack's measured floor, and a numeric ceiling prevents drift.
- **Base UI is the heaviest part after React** (68 KB gzip with four parts).
  Import each part from its subpath (`@base-ui/react/dialog`), never the root,
  and use only the parts a screen needs. Native `popover` and `<dialog>` already
  carry the chat picker, new-chat and paste popovers today. Port those with
  Base UI only where it removes custom focus or Escape code, not by default.
- **TanStack Virtual**: the transcript uses `@tanstack/virtual-core` 3.19.11
  with `anchorTo: 'end'`, `followOnAppend` and `scrollEndThreshold`. The React
  adapter `@tanstack/react-virtual` 3.14.26 depends on `virtual-core` 3.19.11
  (checked), so the same core and options apply. The adapter's `useVirtualizer`
  re-renders on scroll; verify it against the transcript budget before
  committing to it, or keep driving `Virtualizer` from a hook.
- **`@tanstack/react-hotkeys` 0.22.3 is pre-1.0**: first published
  2026-02-10, 51 releases, latest published 2026-11-08 (today). Pin the exact
  version, wrap it in one `useShortcut` module so a replacement touches one
  file, and keep the home guard (inputs, modifiers, IME) in Desk code, because
  the library's scoping must be verified against the guard's current behavior
  (the guard has no test today).
- **Oxc React Compiler is experimental.** `@vitejs/plugin-react` 6.2.1 exposes
  `react({ compiler: true })` backed by `oxc-transform-react`, whose README
  says "Native React Compiler support is experimental". Its peer range is
  `>=0.152.0 <0.157.0`, so **0.157.0 (latest) is outside it**; 0.156.0 is the
  newest in range. Checked in a probe: with 0.156.0 the compiler transforms a
  component (memo slots appear in the output), with it off nothing does. The
  Babel path (`@rolldown/plugin-babel` 0.3.0 + `babel-plugin-react-compiler`
  1.0.0 + `@babel/core`) is the documented fallback. Compiler on or off must
  not change behavior: the port is written to be correct without it, the
  compiler is a measured optimization, and any gap goes in the PR notes.
- **Tailwind 4.3.5** with `@tailwindcss/vite` 4.3.5. Existing tokens are plain
  CSS custom properties in `tokens.css`; map them with `@theme inline` so the
  light and dark themes keep switching on `data-theme`. No second token source.
- **Oxlint 1.92.0 and Oxfmt 0.77.0** are current. Oxfmt is not 1.0.
- Versions checked on 2026-11-08 (npm): react 19.3.1, react-dom 19.3.1,
  jotai 3.0.1, wouter 3.13.0, tailwindcss 4.3.5, @tailwindcss/vite 4.3.5,
  @base-ui/react 1.9.0, @tanstack/react-virtual 3.14.26,
  @tanstack/react-hotkeys 0.22.3, fuse.js 7.5.0, lucide-react 1.70.0,
  @vitejs/plugin-react 6.2.1, oxc-transform-react 0.156.0 (in range),
  oxlint 1.92.0, oxfmt 0.77.0, vite 8.3.6 (repo pins 8.3.5).

## 4. Budgets

Set before the port, from the baseline above. A budget is a ceiling the
React build must stay at or under on the same dataset, same host class and
the same scripts, measured as the worst of three repeats. Anything over needs
a written reason in the PR and a decision, not a quiet bump.

Where the stack has a measured floor above the Svelte number (bundle size,
heap, script time), the budget is the floor plus headroom for Desk's own code,
and it says so. Everything a user waits on has a budget **at or near the
baseline**, since equal interactive performance is the target.

### User-facing runtime (production build)

| Metric | Svelte baseline | React budget | Basis |
| --- | --- | --- | --- |
| First usable render, desktop / mobile | 285 / 288 ms | at most 350 ms | +20% |
| First contentful paint | 88 / 80 ms | at most 150 ms | shell paints before data |
| Route switch (learn, sys, home) | 50 to 127 ms | at most 150 ms each | no switch slower than 150 ms |
| In flight scroll p95 frame | 16.8 ms | at most 16.8 ms (one frame at 60 Hz) | no dropped frames |
| 50-ticket SSE burst, last row | 310 / 213 ms | at most 400 ms | +30% |
| Burst long tasks over 50 ms | 0 | 0 | |
| Transcript open (200 turns) | 735 ms | at most 900 ms | +20% |
| Transcript rows rendered while paging | 26 max | at most 40, never 60 or more | virtualization holds (spec asserts < 60) |
| Transcript to start of conversation | 179 ms | at most 250 ms | |
| `api/live` requests per session | 1 | 1 | exact: no duplicate stream |
| Console or page errors | 0 | 0 | exact |

### Resource use (production build)

| Metric | Svelte baseline | React budget | Basis |
| --- | --- | --- | --- |
| Script time at load | 170 / 178 ms | at most 300 ms | React runtime start-up floor |
| JS heap after GC, idle | 5.4 MB | at most 9 MB | React + Fiber floor, measured in the probe app before Desk code |
| JS heap after burst | 6.5 MB | at most 11 MB | |
| Transcript heap | 7.0 MB | at most 12 MB | |
| DOM nodes, idle home | 4477 | at most 5400 | +20%: same markup |
| Transcript DOM nodes | 1153 | at most 1500 | |

### Bundle size

| Metric | Svelte baseline | React budget | Basis |
| --- | --- | --- | --- |
| JS gzip | 85.8 KB | at most 190 KB | probe app floor 123.9 KB + Desk code and Markdown |
| JS brotli | 75.6 KB | at most 165 KB | |
| JS raw | 257.9 KB | at most 640 KB | probe floor 393.5 KB |
| CSS gzip | 5.9 KB | at most 14 KB | Tailwind output kept to used utilities |
| Initial JS requests | 1 | at most 2 | one entry, one lazy chunk allowed |

JS is the one metric where the stack costs more than today and cannot match
it. Use code splitting only where it pays: lazy-load the trace graph, learn and
sys screens and the Base UI parts they alone use, so the home path stays near
the floor. A bundle over budget is a defect, not a tradeoff, unless a
dependency in the agreed list is shown to be the cause.

### Tooling

| Metric | Svelte baseline | React budget | Basis |
| --- | --- | --- | --- |
| Production build, no cache / warm | 1.04 / 1.07 s | at most 4 s each | the Babel path is slower; the Oxc path should be near 2 s |
| Type check | 3.05 s | at most 8 s | `tsc` on a larger typed tree |
| Dev server ready | 397 ms | at most 1000 ms | |
| Dev first load | 1047 ms | at most 2500 ms | |
| HMR update | 49 ms | at most 150 ms | |
| Lint + format check (Oxlint, Oxfmt) | n/a | at most 3 s together | new gate |
| `node_modules` size | 162 MB, 46 packages | reported, no limit | |

Tooling numbers are on a 6-core shared host, so these are generous. Tighten
them in the PR that adds the toolchain if the first measurement shows a floor
well under them. Do not claim Oxc is faster than the Svelte compiler from
them (the issue forbids that claim): they are budgets, not a comparison.

## 5. Port plan from here

This PR is slice 1. The remaining slices are separate PRs, each with the full
`ui` CI job, the shared e2e specs, and the bench run attached.

2. **Toolchain and shell** (draft for the representative screen): Vite React
   plugin (Oxc compiler on/off measured), Tailwind with `@theme inline`
   mapped from `tokens.css`, Oxlint and Oxfmt, `tsc`, Jotai store and live
   connection owner above Wouter, shell (top bar, status, phone nav, error
   banner), `useShortcut` module, shared primitives. One live screen: **home
   waiting + in flight**, since it exercises the SSE store, a long list,
   keyboard shortcuts and the URL contract. Svelte stays the shipped build
   (`ui/` builds both only behind a flag until parity), the candidate is
   served for comparison, no duplicate submissions or connections.
3. **Compare to the baseline** with `ui/perf/*` and the specs on this one
   screen before expanding. Stop and report if any budget is missed.
4. **Port** chat rail, trace and graph, session viewer, learn, sys and audit,
   one PR each, each adding the missing regression from the gap list.
5. **Cut over and remove Svelte**: `svelte`, the plugin, `svelte-check`,
   `svelte.config.js`, `jsconfig.json`, obsolete CSS; update `package.json`,
   lockfile, `nix/desk-ui.nix` (`npmDepsHash`, fileset), CI, `ui/AGENTS.md`,
   `docs/desk-design.md` paths, `docs/verify/*` and a rollback note (revert the
   cutover commit; the Svelte tree is deleted only in that commit).

## 6. Reproduce

Everything runs from `ui/` with Node 22 and Playwright's Chromium. Use an
isolated controller (`docs/verify/desk-redesign.md`), never the live one.

```sh
# Isolated controller with a pack that has a `work` workflow (vars: issue,
# one step in a pool that spawns nothing) and a chat pool on the ACP test agent
# (go test -c -o control.test ./pkg/control; agent args: [fake-agent]).
sh ui/perf/seed.sh <jobshop binary> <root> 200 30   # dataset
DESK_URL=http://127.0.0.1:7562 node ui/perf/bench.mjs --runs 7 --burst 50 --json bench.json
DESK_URL=http://127.0.0.1:7563 node ui/perf/chat.mjs --turns 200 --runs 5 --json chat.json
DESK_URL=http://127.0.0.1:7562 node ui/perf/dev.mjs shell/Learn.svelte --runs 9 --json dev.json
node ui/perf/build.mjs --json build.json            # no controller needed
```

`dev.mjs` edits the file it is given and restores it, so run it on a clean
tree. For the React app pass a `.tsx` file and `--snippet` with a visible
change, because a comment produces no update. Repeat each three times and keep
the worst run. Add `PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS=1` on hosts
without the browser dependencies. Run the e2e specs as in
`docs/verify/desk-redesign.md`, with the controller started `--tick=1s`.

The scripts are framework neutral: they find things by role and visible text
(`sys ok`, the Waiting and In flight regions, the Main and Mobile navigation,
`Choose chat`, `Chat transcript`), which the port must keep.
