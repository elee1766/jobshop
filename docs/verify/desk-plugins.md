# Verify: Desk plugins (#517)

On the host running Desk (default `127.0.0.1:7480`):

1. `curl -s 127.0.0.1:7480/api/plugins | jq '.[].name'` lists `deploys`
   with `"source": "builtin"`.
2. `?view=deploys` renders the Deploys cards; the browser console shows
   no CSP errors.
3. Add `<root>/desk/plugins/hello/plugin.json`
   (`{"name":"hello","entry":"index.js","contributes":[{"slot":"view","element":"hello-view","id":"hello","title":"Hello"}]}`)
   and an `index.js` that defines `hello-view`. Reload Desk: a Hello nav
   entry appears. Make `index.js` throw: the Hello view shows an error
   card and the other views still work.
4. `curl -s -XPOST -H 'X-Jobshop-Desk: 1' -H 'Content-Type: application/json' -d '{}' 127.0.0.1:7480/api/rpc/claim`
   returns 403 with `"code":"forbidden"`; `.../api/rpc/version` returns
   the controller version.
5. Set `desk: {plugins: {deploys: {enabled: false}}}`, run
   `jobshop config reload`, reload Desk: Deploys is gone from the nav.
6. Playwright: `cp -r ui/e2e/plugins/* <root>/desk/plugins/`, restart
   Desk, then run `docs/verify/desk.md` step 3 with `DESK_PLUGINS=1`.
   `e2e/plugins.spec.js` covers a user view calling `api.get` and
   `api.call`, `claim` refused, a throwing plugin's error card, a user
   plugin replacing the built-in deploys, and a drawer section. Passed
   2026-11-05 against gasserver's controller: 10 of 10, and
   `desk.spec.js` 16 of 16 without the fixtures.
