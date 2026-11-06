# Remove the built-in Artifacts view (JOB-80)

2026-11-06. Desk's Artifacts view was rough and little used, so it is
gone: `ui/plugins/artifacts`, its built copy in `pkg/desk/dist/plugins`,
the nav entry (`g r`) and unread badge in `ui/shell/App.svelte`, its e2e
test, and `markdown()`/`bytes()` in `ui/kit/format.js` (nothing else
used them; chat has its own markdown).

Kept for agents and plugins: the artifact store, the `artifact.*` RPCs
and `jobshop artifact deliver/list/show/ack`, Desk's `/api/artifacts*`
endpoints, the `artifact` plugin slot and the Artifact types in
`ui/desk-plugin.d.ts`. A view plugin can rebuild the view from these:
list `api/artifacts`, render the body, and host renderers with
`api.mount(host, 'artifact', {type, ext}, {artifact, file})`.

Agent prompts now point humans at `jobshop artifact list` instead of the
view.
