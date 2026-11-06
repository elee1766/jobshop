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
| `session` | the session panel body (#535); the shell owns the frame, the last one wins | `{session, item}`, `item` the attention `kind:id` it resolves, or empty | none |

`match` lists accepted values per key; every key listed must accept one
of the target's values. No `match` accepts everything. A view whose `id`
is a built-in's id replaces that view.

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
built-in view (Work, Activity, Sessions, Pools, Artifacts, the ticket
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
  Playwright tests.

## Try the example

```sh
mkdir -p <root>/desk/plugins
cp -r examples/desk-plugins/csv-table <root>/desk/plugins/
printf 'a,b\n1,2\n' > /tmp/t.csv
jobshop artifact deliver "CSV test" --attach /tmp/t.csv
```

Open Desk, Artifacts, the new artifact: `t.csv` renders as a table.
