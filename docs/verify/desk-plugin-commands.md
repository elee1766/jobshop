# Verify: Desk plugin commands and typed RPC (#542)

On the host running Desk (default `127.0.0.1:7480`):

1. Add to `jobshop.yaml`:
   `desk_commands: {echo: {command: cat}, fail: {command: "exit 3"}}`,
   run `jobshop config reload`.
2. `curl -s -XPOST -H 'X-Jobshop-Desk: 1' -H 'Content-Type: application/json' -d '{"name":"echo","input":{"a":1},"ticket":"<id>"}' 127.0.0.1:7480/api/rpc/desk.exec`
   returns `{"output":{"a":1}}`; `jobshop ticket show <id>` has a
   `desk_exec` history row with actor `desk:<user>`.
3. The same with `"name":"fail"` returns an error containing `exited 3`.
4. `curl -s 127.0.0.1:7480/api/rpc | jq '.schemas["ticket.create"].params, (."$defs" | keys | length)'`
   prints a `$ref` and a non-zero count.
5. Install the example: `cp -r examples/desk-plugins/csv-table <root>/desk/plugins/`,
   `printf 'a,b\n1,2\n' > /tmp/t.csv; jobshop artifact deliver "CSV" --attach /tmp/t.csv`.
   Desk, Artifacts, the new artifact: `t.csv` renders as a two-column
   table with "1 rows".
6. Playwright with the fixtures (`docs/verify/desk-plugins.md` step 6)
   and `desk_commands: {e2e-echo: {command: cat}}` in `jobshop.yaml`:
   `plugins.spec.js`'s `e2e-flow` test covers ticket create with meta,
   finding it by meta in `ticket.list`, its comment event, `api.exec`
   and a session stream.
