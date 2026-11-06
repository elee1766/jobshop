# Plugin-defined Desk slots (JOB-87)

2026-11-06. Desk plugin api 4.

- A contribution's `slot` is a built-in name or `<plugin>:<name>`
  (`pkg/desk/plugin.go`, `pluginSlotRe`). Go checks only the form: it
  does not check that a plugin with that name exists or hosts that slot,
  so a slot for a plugin that is disabled is simply empty.
- `api.slot(host, slot, context, target?)` mounts every matching
  contribution, in plugin name order, each in its own `div.plugin-host`.
  `api.mount` still mounts the first match only (artifact renderers).
- An import, define or module error, or an error event from the plugin's
  files while mounted, replaces that box with `.plugin-error`. Before
  this, `api.mount` showed the card only for load failures.
- Built-in slots: `work:toolbar`, `work:aside`, `chat:panel` (Chat view,
  under the session list) and `chat:composer-actions` (session panel).
  The issue suggested `work:aside` and `chat:composer-actions`;
  `chat:panel` was added so a plugin can add a panel to the Chat page
  itself, not only to the composer.
- Work and chat now declare `"api": ">=4"`.
