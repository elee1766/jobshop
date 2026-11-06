# Desk nav slot (JOB-86)

2026-11-06. Desk plugin api 6.

- The nav is the built-in plugin `ui/plugins/nav` (`desk-nav`). The desk
  app (`ui/plugins/desk/Frame.svelte`) mounts the last `nav` contribution
  by plugin name, so a plugin named after `nav` (`top-nav`) wins with no
  config. Disabled, or none: the page column fills the screen.
- `placement` is `{wide, narrow}`, each left, right, top, bottom or none,
  default left and bottom. The frame switches at the 768px breakpoint
  (`panel.mobile`) and sets `data-dock` on `[data-testid=desk-app]`. The
  nav's `context` is `{dock, narrow}`. A top bar, and a bottom bar on a
  wide page, is the CSS container `nav`; a phone's bottom dock is not,
  because the built-in tab bar draws its live dot and theme toggle into
  the top bar with `position: fixed`, and a container would trap it.
- Brand: drawn beside a left or right nav and as the phone top bar over a
  bottom nav; not drawn for top, none or no nav. Before this it was
  always drawn.
- `api.views()` lists views from the manifests, first one per id,
  sorted by `order` (1 and up, unset last) then plugin order. `view`
  contributions take `icon`, `key` and `order`. The `known` table in the
  frame is gone: the six built-in views declare `order` 1 to 6, so the
  list is in the same order as before. `deploys` (a user plugin) has no
  icon, key or order and so now lists last with a puzzle piece and no
  `g d` key. It had a rocket and `d`; a plugin that wants them sets
  `icon: rocket`, `key: d`.
- Go (`pkg/desk/plugin.go`): `nav` is a built-in slot; `icon`
  (`[a-z0-9-]`), `key` (one `[a-z0-9]`), `order` (0 to 10000) and
  `placement` are validated, and are errors on the wrong slot. A bad
  manifest is skipped with a log line, as before.
- Nav `g`+letter handling moved into the nav, so with another nav the
  keys are that nav's concern. Escape closing the session console stays
  in the frame.
- Verified 2026-11-06 against a throwaway root: 105 passed, 13 skipped
  (live agent sessions), at the root and under the `/desk/` prefix.
