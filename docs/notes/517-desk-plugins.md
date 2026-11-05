# Desk plugins, part 1 (#517)

2026-11-05. Loader, slots, `api` and generic RPC, with Deploys moved to
a built-in plugin. Remaining built-ins, `api.query`, `api.session`,
`api.exec`, the generated `.d.ts` and the CSV example are follow-ups
linked on #517.

- **Plugin:** a directory with `plugin.json` (`name`, `entry`, `api`,
  `contributes`) and ES modules defining custom elements. The name must
  match the directory. A broken `plugin.json` is logged and skipped.
- **Sources**, later wins on the same name: embedded built-ins
  (`pkg/desk/dist/plugins`, built from `ui/plugins/`), each pack's
  `desk/` directory, `<root>/desk/plugins/`. `GET /api/plugins` lists the
  winners with `source` and `overrides`. Files are served from
  `GET /plugins/<name>/<file>` with `fs.ValidPath` and `os.Root`, so no
  `..` or symlink escape, nosniff and the same CSP.
- **Config:** `desk.plugins.<name>: {enabled, config}` and
  `desk.rpc: {allow, deny}` in `jobshop.yaml`. Desk reads them through
  `config.show` on each request, so `config reload` applies them.
- **Slots:** `view` (nav entry, `?view=<id>`; a built-in's id replaces
  it), `drawer` and `action` (ticket or run drawer, matched by `kind`,
  `label`, `workflow`, `pool`), `artifact` (attachment renderer by MIME
  `type`, `text/*` globs, or `ext`), `header` (nav widget). A plugin that
  fails to import, define its element in 10s, or throws from its module
  shows an error card in its slot; the rest of Desk keeps working.
- **`api` (version 1):** `get`, `post` (CSRF header added), `call(method,
  params)`, `subscribe(fn)`, `events(filter, fn)`, `navigate`, `toast`,
  `confirm`, `url`, `user`, `base`, `lastSeq`. Events come from the one
  stream Desk holds open; EventSource resumes from `Last-Event-ID`.
  `plugin.json` `"api": ">=2"` on this Desk gives an error card.
- **Generic RPC:** `POST /api/rpc/{method}` forwards to the admin socket
  with the proxy user as actor; a `_meta` in the body is dropped. Never
  allowed: `claim`, `heartbeat`, `release`, `handoff`, `writekey.*`,
  `ingest.claim`/`admit`, `guard`, and the two streams. Off unless
  allowed: `identity.token`, `controller.shutdown`. Then `deny` globs,
  then `allow` globs narrow. Refusal is 403 `{error, code: "forbidden"}`.
  `GET /api/rpc` lists what is allowed. `rpc.Methods()` is the method
  list; a control test keeps it equal to the handler table.
- **Built-ins:** Svelte components mounted by a small custom element
  class, built per plugin by `ui/scripts/build-plugins.mjs` with CSS in
  `style.css` loaded through a `<link>` (the CSP has no inline styles).
  `ui/scripts/check-plugins.mjs` fails `npm run check` and the build if a
  built-in imports outside its directory or from `ui/src`.
- **Trust:** plugins run in Desk's page with Desk's privileges, the same
  trust as packs and hooks. Install only code you would run as the
  operator. Sandboxed plugins are out of scope.
