# Verify: Desk app slot (JOB-85)

On the host running Desk (here `127.0.0.1:7480`):

1. `curl -s 127.0.0.1:7480/api/plugins | jq -r '.[] | select(.contributes[].slot == "app") | .name'`
   prints `desk`. Desk looks as before: nav, live dot, theme toggle,
   attention badge, drawer and session panel.
2. `cp -r examples/desk-plugins/mini-app <root>/desk/plugins/`, reload
   the tab: the page is a top bar ("Mini app", one button per view,
   live, theme) and Work renders under it. Buttons switch views, back
   works, the theme button flips light and dark.
3. Rename the copy's `plugin.json` name and directory to `desk`, or set
   `desk: {plugins: {desk: {enabled: false}}}` and run `jobshop config
   reload`: the same app shows. Remove the copy and re-enable desk: the
   built-in frame is back.
4. Make the copy's `index.js` throw on import: the page is one error card
   naming the plugin, not blank.
5. Playwright: `e2e/app.spec.js` (mocked, any Desk) covers 2 and 4 plus
   a plugin needing a newer api and no app at all. Run
   `docs/verify/desk.md` step 3 for the whole suite.
