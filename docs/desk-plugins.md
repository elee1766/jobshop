# Desk plugins

Desk is a plugin host (#517). Every view, drawer section and artifact
renderer is a plugin, including the built-in ones, and they all use the
same `api`. This guide covers writing one. The worked example is
[`examples/desk-plugins/csv-table`](../examples/desk-plugins/csv-table),
a CSV artifact renderer in plain JS with no build step. A pack ships a
view plugin with a `desk_commands` entry it runs the same way.

## The contract

A plugin is a directory holding `plugin.json` and ES modules:

```json
{
  "name": "csv-table",
  "version": "1.0.0",
  "entry": "index.js",
  "api": ">=1",
  "contributes": [
    { "slot": "artifact", "element": "csv-table-artifact", "match": { "type": ["text/csv"], "ext": ["csv"] } }
  ]
}
```

- `name` must equal the directory name.
- `entry` is the module Desk imports, once, before it creates any of the
  plugin's elements. It must call `customElements.define` for every
  `element` it contributes. Other files in the directory are served too,
  so the entry can import siblings with relative paths.
- `api` is the oldest Desk plugin API the plugin works with. A plugin
  needing a newer one gets an error card instead of half-working.
  Version 1 is part 1's api; version 2 adds `exec`.

Desk creates each element, then sets three properties on it:

| property | value |
| --- | --- |
| `api` | the Desk api object (below) |
| `config` | `desk.plugins.<name>.config` from `jobshop.yaml`, `{}` if unset |
| `context` | per slot, below; set again when it changes (another ticket in the drawer) |

Use setters or read them in `connectedCallback`. Any framework that
produces custom elements works: Lit, Svelte with `customElement: true`,
or plain classes as in the example.

## Slots

| slot | where | `context` | `match` keys |
| --- | --- | --- | --- |
| `view` | a nav entry, `?view=<id>` | `{}` | none; `id` and `title` name the entry |
| `drawer` | a section in the ticket or run drawer | `{ticket, run?}` | `kind`, `label`, `workflow`, `pool` |
| `action` | a button row in the drawer | `{ticket, run?}` | as drawer |
| `artifact` | an attachment renderer | `{artifact, file}` with `file.url` | `type` (MIME, `text/*` globs), `ext` |
| `header` | a widget in the nav bar | `{}` | none |
| `session` | the session panel body (#535); the app owns the frame, the last one wins | `{session, item, readonly}`, `item` the attention `kind:id` it resolves, or empty | none |
| `app` | the whole page (JOB-85); the shell mounts one, the last by plugin name, below | `{}` | none |
| `<plugin>:<name>` | a place a plugin's own page offers, below | per slot | any; checked against the `target` the host passes |

`match` lists accepted values per key; every key listed must accept one
of the target's values. No `match` accepts everything. A view whose `id`
is a built-in's id replaces that view.

### Plugin-defined slots

To add to a page without replacing it, contribute to a slot that page's
plugin defines (JOB-87). The name is `<plugin>:<name>`, lowercase
letters, digits and hyphens on both sides; Desk checks only the form, so
a typo is an empty slot, not a load error. Every matching contribution
renders, in plugin name order, each in its own box: one that fails to
load shows an error card in its box and the rest of the page keeps
working. Declare `"api": ">=4"`.

```json
{ "slot": "work:toolbar", "element": "my-export-button" }
```

Slots the built-in views offer:

| view | slot | where | `context` |
| --- | --- | --- | --- |
| Work | `work:toolbar` | the header bar, right of the state filter | `{state}`, the filter (`''` for all) |
| Work | `work:aside` | a column right of the lists (below them on a phone) | `{state, work}`, `work` the `api/work` result or null while loading |
| Chat | `chat:panel` | under the session list, above Escalations | `{session}`, the selected session or `''` |
| Chat (session panel) | `chat:composer-actions` | the composer's button row, left of Send | `{session, item, working}` |

A slot shows nothing when no plugin contributes, so a toolbar stays as
it was. To offer slots in your own view, call `api.slot` (below) on an
element and document the name and context.

### Replacing the whole page

The frame around the views (brand, nav, the main view, the session
panel, the ticket drawer and the phone tab bar) is itself a plugin: the
built-in `desk` contributes the `app` slot (JOB-85). The shell keeps
only the runtime: the plugin loader, the `api`, the event stream, URL
state, the RPC client with its CSRF header, toasts and confirm. It
mounts exactly one `app` contribution, the last by plugin name, in
`<body>`. Replace it in either way:

- name your plugin `desk` (a root or pack plugin with a built-in's name
  replaces it), or
- set `desk: {plugins: {desk: {enabled: false}}}` and contribute `app`
  from any plugin. Without that, an app plugin whose name sorts after
  `desk` still wins, since the last one does.

The app owns all layout, sizing and responsiveness. It builds its nav
from `api.contributions('view')` and puts a view on the page with
`api.mount(host, 'view', {id})`, the same elements the built-in frame
mounts; `api.badges`, `api.live` and `api.theme` give it what the
built-in frame shows. `--vvh` and `--vvtop` on `<html>` track the
visible viewport for a full-height page above a phone keyboard. Declare
`"api": ">=5"`. An app that fails to import, define or render shows an
error card in place of the page; with no `app` at all the shell says
so. [`examples/desk-plugins/mini-app`](../examples/desk-plugins/mini-app)
is a top-bar app in plain JS that hosts the built-in views.

## The api

The full typed surface is [`ui/desk-plugin.d.ts`](../ui/desk-plugin.d.ts),
generated from the Go RPC types. Reference it from JS with
`/** @type {import('./desk-plugin.d.ts').DeskApi} */` or copy it into a
TypeScript plugin.

- `get(path)`, `post(path, body)`: Desk's own HTTP API (`api/work`,
  `api/ticket/<id>`, ...). `post` adds the CSRF header.
- `call(method, params)`: an admin RPC method through Desk, acting as the
  proxy user. `GET api/rpc` lists the methods this Desk allows with a
  JSON schema for each one's params and result. Session-only methods
  (`claim`, `release`, `handoff`, ...) are never allowed;
  `desk.rpc.deny` and `desk.rpc.allow` narrow the rest.
- `exec(name, input, {ticket})`: runs a `desk_commands` entry (below)
  and resolves to its stdout, parsed if it is JSON.
- `subscribe(fn)`, `events(filter, fn)`: the change events Desk already
  streams (ticket, run, session, pool, activity). `filter` matches
  fields such as `{kind: "ticket", op: "close"}`. Both return an
  unsubscribe function; call it in `disconnectedCallback`.
- `navigate(view, params)`, `toast(msg)`, `confirm(msg)`, `url(path)`,
  `user`, `base`, `lastSeq`, `version`.

Added in api version 2 (#535; declare `"api": ">=2"`):

- `params()`: Desk's URL state (`view`, `ticket`, `session`, `item`).
- `openSession(id, {item?})`, `closeSession()`, `openTicket(id)`.
- `session.stream(id, {tail}, fn)`: the live stream; `fn` gets
  `{type: 'open'|'event'|'end'|'error', event}`, events are
  `rpc.SessionEvent` (`turn_start`, `text`, `reasoning`, `tool` with
  `input` and `output`, `turn_end`, `usage`, `state`). Returns a close
  function.
- `session.older(id, before, limit)`: `{events, first}`, the journal
  before seq `before`.
- `session.access(id)`, `session.send({session|key, text, interrupt})`,
  `session.cancel(id)` (Stop: cancels the turn, sends nothing).
- `push.status()`, `push.subscribe()`, `push.unsubscribe()`,
  `push.test()`: each resolves to `{state, error}`, `state` one of
  `unsupported`, `unavailable` (Desk without `--push`), `denied`, `off`,
  `on`.

Added in api version 3 (#541; declare `"api": ">=3"`). Every
built-in view (Work, Activity, Sessions, Pools, the ticket
details in the drawer) is now a plugin using only these:

- `params()` keeps every URL key, so a view stores its filters there;
  `go(patch)` changes them with a history entry and `onParams(fn)`
  follows them (back and forward included).
- `query(method, params, {watch})`: a reactive `call`. It refetches,
  coalesced to once a second, on an event for an id in the result or
  one matching `watch`. `subscribe(fn)` gets `{data, error, loading}`.
- `has(slot, target)`, `mount(host, slot, target, context)`: host
  another plugin's contribution, e.g. an `artifact` renderer.

Drawer plugins get `{ticket, run, trace, comments}`; built-ins come
first, so the ticket details lead.

Added in api version 4 (JOB-87; declare `"api": ">=4"`):

- `slot(host, 'work:aside', context, target?)`: mounts every
  contribution to a plugin-defined slot inside `host`, each in a
  `div.plugin-host` with `data-plugin` set, and returns `{context(c),
  destroy()}`, or null when none matches `target` (`{}` if omitted).
  `mount` still mounts only the first match, for a single renderer such
  as `artifact`. Call `destroy` in `disconnectedCallback`. Built-ins use
  the `slots` Svelte action in `ui/kit/plugin.svelte.js`.

Added in api version 5 (JOB-85; declare `"api": ">=5"`), for apps
and any plugin:

- `contributions(slot, target?)`: every enabled contribution to `slot`
  that `target` selects, each with `plugin` and `source`, in plugin name
  order.
- `has`, `mount`, `slot` and `contributions` take `plugin`, `element`
  and `id` in `target` as names: `mount(host, 'view', {id: 'work'})`
  mounts the Work view whatever its `match`.
- `badge(view, n)` sets the count on a view's nav entry (0 clears it);
  `badges(fn)` follows them. Desk sets `attention` from `api/attention`.
- `live(fn)` follows whether the event stream is connected.
- `theme`: `'dark'` or `'light'`; setting it applies and remembers it.

`badges` and `live` call `fn` at once and on each change, and return
an unsubscribe function.

Example: create a ticket with meta, then follow it.

```js
const t = await this.api.call('ticket.create', { title: 'Import rows', meta: { csv: { file: 'a.csv' } } });
this.off = this.api.events({ kind: 'ticket', id: t.id }, (ev) => this.refresh(ev));
```

## desk_commands

A plugin that needs to run something on the host (a report script, a
deploy check) does it through a named command the operator configured,
never a command string from the browser:

```yaml
desk_commands:
  row-count:
    command: jq length   # sh -c, in the root dir
    timeout: 10s         # default 30s
```

`api.exec('row-count', [1, 2, 3])` sends the input as JSON on stdin,
like hooks (#188). Stdout is the result. A non-zero exit or a timeout
rejects with the exit code and stderr. Each run is logged with the
actor, and with `{ticket}` it also adds a `desk_exec` row to that
ticket's history. Packs can contribute `desk_commands`, merged by name
like pools. `desk.exec` is a normal RPC method, so `desk.rpc.deny:
[desk.exec]` turns it off.

## Theming

Plugins render in Desk's document, so Desk's CSS custom properties
apply. Prefer the `--desk-*` tokens (#535), which Desk swaps for
`[data-theme='light']`: surfaces `--desk-bg`, `--desk-side`,
`--desk-panel`, `--desk-raised`, `--desk-hover`; `--desk-border`,
`--desk-border-strong`; text `--desk-text`, `--desk-muted`,
`--desk-faint`; `--desk-accent`, `--desk-accent-fg`,
`--desk-accent-soft`; states `--desk-ok`, `--desk-warn`, `--desk-bad`;
`--desk-shadow`, `--desk-radius`, `--desk-font`, `--desk-mono`,
`--desk-tabbar`. The short names (`var(--bad)`, `var(--muted)`) in
`ui/shell/app.css` still work. Classes
Desk defines (`card`, `faint`, `mono`) work too. A built-in's own CSS
loads through a `<link>` to its `style.css`, since the CSP blocks inline
styles; do the same, or set styles from JS with `element.style`.

## Loading order and config

Sources, later wins on the same name:

1. built-ins embedded in the binary,
2. each pack's `desk/` directory, in packs order,
3. `<root>/desk/plugins/<name>/`.

`GET /api/plugins` shows the winners with `source` and `overrides`.

```yaml
desk:
  plugins:
    deploys: { enabled: false }          # hide a plugin
    csv-table: { config: { maxRows: 500 } }
  rpc:
    deny: [desk.exec, learn.*]
```

Desk reads this through `config.show` per request, so `jobshop config
reload` applies it; reload the browser tab to pick up plugin changes.

## Trust

A plugin runs in Desk's page with Desk's privileges: it can call every
method Desk allows, as the signed-in user. Treat installing one like
adding a pack or a hook: only code you would run as the operator. Render
untrusted data (ticket text, artifact files) with `textContent`, not
`innerHTML`, as the CSV example does. Sandboxed plugins are out of
scope; the `api` shape stays compatible with them.

## Debugging

- A plugin that fails to import, never defines its element (10s), or
  throws from its module shows an error card in its slot with the
  message; the rest of Desk keeps working.
- The browser devtools Network tab shows `GET /plugins/<name>/<file>`;
  a 404 means a wrong `entry` or name. `curl /api/plugins` shows whether
  Desk found the plugin at all; a broken `plugin.json` is logged by
  Desk and skipped.
- `curl -s localhost:7480/api/rpc | jq '.methods'` shows what `call`
  may use; a refused call rejects with `forbidden`.
- `ui/e2e/plugins/e2e-test` is a fuller plain-JS plugin used by the
  Playwright tests; `ui/e2e/plugins/e2e-slots` adds to Work and Chat
  through their slots; `e2e/app.spec.js` runs `mini-app` as the app.

## Try the example

```sh
mkdir -p <root>/desk/plugins
cp -r examples/desk-plugins/csv-table <root>/desk/plugins/
printf 'a,b\n1,2\n' > /tmp/t.csv
jobshop artifact deliver "CSV test" --attach /tmp/t.csv
```

No built-in view hosts the `artifact` slot since JOB-80. A view plugin
that lists `api/artifacts` and calls `api.mount(host, 'artifact',
{type, ext}, {artifact, file})` for each attachment renders `t.csv` as a
table.
