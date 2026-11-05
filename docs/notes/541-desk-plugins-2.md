# Desk plugins, part 2 (#541)

2026-11-05. Every Desk view is a plugin, the plugin API is version 3,
and `ticket.list` gained the filters a plugin needs to build a view
without fetching everything. Follows part 1 (`517-desk-plugins.md`)
and the #535 chat and attention plugins.

- **Layout:** `ui/shell` (was `ui/src`) is the frame: nav, drawer
  frame, session side panel, plugin loader, push, toasts and confirm.
  `ui/plugins/<name>` are the built-ins: attention, chat (both from
  #535), artifacts, work, activity, sessions, pools and ticket (the
  drawer details). `ui/kit` holds helpers each built-in bundles its own
  copy of: `defineElement`, `resource`, `query`, `params`, `slot`,
  `format.js` and `Trace`. `scripts/check-plugins.mjs` fails if a
  plugin imports outside its directory and `ui/kit`, anything in
  `ui/shell`, or if `ui/kit` imports outside itself.
- **Shell:** the nav lists `view` contributions, known ids (attention,
  artifacts, chat, work, activity, sessions, pools, deploys) in their
  old order with their icon and `g` key, then the rest. The attention
  count and unread artifacts badge stay in the shell. The drawer frame
  fetches `api/ticket/{id}` and gives every `drawer` and `action`
  plugin `{ticket, run, trace, comments}`; built-in sections go first.
- **API 3:** `params()` keeps every URL key; `go(patch)`, `onParams(fn)`;
  `query(method, params, {watch})` (an RPC result refetched on matching
  change events); `has(slot, target)` and `mount(host, slot, target,
  context)` so one plugin hosts another (artifact renderers inside
  Artifacts). Session I/O is #535's `api.session` (API 2).
- **`ticket.list` filters:** `meta: {"pack.key": value}` (up to 3,
  string values, as indexed in `ticket_meta`), `workflow`, `state`,
  `run`, `order: priority|newest`, `since` (on `updated_at`) and an
  `after` cursor. Both store engines cover them in the conformance
  suite. Meta filters join `ticket_meta`, not a JSON path scan.
- **Theme:** the `--desk-*` custom properties are listed in
  `docs/desk-plugins.md` (Theming).
