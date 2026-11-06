# Verify: Desk nav slot (JOB-86)

On the host running Desk (here `127.0.0.1:7480`):

1. `curl -s 127.0.0.1:7480/api/plugins | jq -r '.[] | select(.contributes[].slot == "nav") | .name'`
   prints `nav`; `jq '.[].contributes[] | select(.slot=="view") | [.id,.order,.icon,.key]'`
   lists the six built-in views with orders 1 to 6. Desk looks as before:
   sidebar, badge, live dot, theme toggle, `g w` opens Work, and on a
   phone the bottom tab bar with the brand as a top bar.
2. `cp -r examples/desk-plugins/top-nav <root>/desk/plugins/`, reload the
   tab: the links are a bar above the page, no sidebar, no brand, on a
   wide window and a phone; narrow the window and the labels shorten.
3. Set `desk: {plugins: {nav: {enabled: false}}}` (and remove top-nav),
   run `jobshop config reload`, reload: no nav, the page fills the
   screen. Re-enable it: the sidebar is back.
4. A `plugin.json` with `"placement": {"wide": "middle"}` or `"key":
   "ab"` is skipped; Desk logs `desk: plugin.json` with the reason.
5. Playwright: `e2e/nav.spec.js` (mocked, any Desk) covers the built-in
   nav's order, icons, `g` keys and docking on wide and phone widths, the
   example top bar on both, a self-placing (`none`) nav, no nav, and a
   broken nav. Run `docs/verify/desk.md` step 3 for the whole suite.
