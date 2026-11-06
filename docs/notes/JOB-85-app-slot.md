# Desk app slot (JOB-85)

2026-11-06. Desk plugin api 5.

- `ui/shell/App.svelte` is the runtime only: plugin loader, event
  stream, whoami, the attention badge, `--vvh`, toasts and confirm. It
  mounts the last `app` contribution (plugin name order) through
  `PluginSlot`, so a failing app gets the usual error card.
- Today's frame moved to the built-in plugin `ui/plugins/desk`
  (`Frame.svelte`, with `Drawer.svelte`, `SidePanel.svelte` and the
  panel width state). It uses only the api, like the other built-ins.
  Its markup and CSS did not change, so the e2e selectors still hold.
- New api: `contributions`, `badge`/`badges`, `live`, `theme`. `has`,
  `mount`, `slot` take `plugin`/`element`/`id` in `target` as names, so
  an app can mount a view by id and the frame mounts a drawer section
  or header widget by plugin and element.
- Theme moved to `ui/shell/lib/signals.js` and is applied before any
  app mounts, so a light theme does not flash dark.
- `defineElement` hides a built-in's element until its `style.css`
  loads. The frame is a plugin now, and without this a cold load showed
  it unstyled for a moment.
- kit `params()` fills from `api.params()` at once, so the frame's
  first render mounts the URL's view instead of Attention.
- Mocked e2e specs (`attention-chat`, `holder`, `pools`) list `desk`
  in their plugin list: with no `app` the shell shows "Desk has no app".
- Verified 2026-11-06 against a throwaway root on this branch: 89
  passed, 13 skipped (live agent sessions), at the root and under the
  `/desk/` prefix.
