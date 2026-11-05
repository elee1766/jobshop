# A2A escalation gateway (#547)

On gasserver, with a scratch A2A shim (not committed, under
`$JCODE_SCRATCH_DIR`) that serves `/.well-known/agent-card.json` with one
JSONRPC 1.0 interface, logs each request, and answers `SendMessage` with
a `working` task, then `GetTask` with `completed` and the text
"close it, drill acknowledged".

1. Add a gateway `network: a2a`, `url: http://127.0.0.1:<port>`,
   `token_file`, `pool: operator`, `escalations: [run]`, `poll: 5s`;
   `jobshop reload`, then `jobshop gateway <name>`.
2. Trigger the `drill` escalation (`docs/verify/escalation-drill.md`).
3. The shim logs `SendMessage` with `contextId` = the escalation ticket id
   and a `Bearer` header, then one `GetTask` per poll.
4. The reply reaches the escalation session, which resolves the
   escalation; the ticket closes and the shim gets the "was resolved" message.
5. `jobshop ticket show <id>`: `meta.gateway.a2a.<name>.done` is true and
   no comment or meta field holds the token.
